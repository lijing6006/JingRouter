package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"golang.org/x/sync/singleflight"
)

const (
	// workbuddyQuotaUpstreamTimeout 拉取 Magpie 额度接口的单次超时。
	workbuddyQuotaUpstreamTimeout = 15 * time.Second
	// workbuddyCheckinTimeout 触发 Magpie 签到接口的单次超时。签到要逐号调
	// 供应商，Magpie 侧 handler 上下文是 1 分钟，这里略放宽到 70 秒。
	workbuddyCheckinTimeout = 70 * time.Second
	// workbuddyWebCookieName Magpie Web UI 的签到 Cookie 名：固定为
	// "magpie_web_" + Web UI 监听端口（本部署为 3430）。见 Magpie 源码
	// internal/gui/web.go 的 webGuard（Cookie 值即 MAGPIE_WEB_KEY）。
	workbuddyWebCookieName = "magpie_web_3430"
	// workbuddyQuotaMaxBodyBytes Magpie 额度接口响应体上限（4 个账号的窗口数据很小）。
	workbuddyQuotaMaxBodyBytes = 256 * 1024
	// workbuddySettingsMaxBodyBytes Magpie 设置接口响应体上限：GET /api/settings
	// 返回整份设置（含各插件的 schema），比额度接口大得多，放宽到 2MB。
	workbuddySettingsMaxBodyBytes = 2 * 1024 * 1024
	// workbuddyQuotaSingleflightKey 并发去重 key。
	workbuddyQuotaSingleflightKey = "workbuddy-quotas"
)

// WorkbuddyQuotaWindow 是 Magpie 额度接口返回的单个额度窗口
// （Credits / Extra credits / Total credits / 各模型 tokens 估算）。
type WorkbuddyQuotaWindow struct {
	Name       string  `json:"name"`
	Used       float64 `json:"used"`
	Remaining  float64 `json:"remaining"`
	Limit      float64 `json:"limit"`
	Display    string  `json:"display,omitempty"`     // 如 "24.88 / 1100"（供应商口径的已用/上限）
	Unit       string  `json:"unit,omitempty"`        // 如 "credits"
	Amount     float64 `json:"amount,omitempty"`      // 供应商口径的已用量（与 used 的归一化值不同）
	ResetsAt   string  `json:"resets_at,omitempty"`   // RFC3339，窗口重置时间
	RunsOutAt  string  `json:"runs_out_at,omitempty"` // RFC3339，余额耗尽预计时间
	RunsOut    float64 `json:"runs_out,omitempty"`    // 预计耗尽时的剩余量
	HoldsWhy   string  `json:"holds_why,omitempty"`   // 额度偏低时的原因标记
}

// WorkbuddyQuotaAccount 是单个 WorkBuddy 账号的额度快照。
type WorkbuddyQuotaAccount struct {
	Provider string                 `json:"provider"` // 固定 "workbuddy"
	Name     string                 `json:"name"`     // 如 "WorkBuddy"
	Plan     string                 `json:"plan"`     // 如 "Free plan"
	User     string                 `json:"user"`     // 账号名，如 "小牛来咯"
	Kind     string                 `json:"kind,omitempty"`
	ReadAt   string                 `json:"read_at,omitempty"`
	Windows  []WorkbuddyQuotaWindow `json:"windows"`
	// CheckinKey 是该账号在 Magpie 签到开关里的 key（如 "workbuddy|workbuddy#wycm4m"），
	// 来自 Magpie GET /api/settings 的 checkinAccounts（按 user 匹配）。
	// 为空表示未知（后端未配 MAGPIE_WEB_KEY、取不到开关状态），前端此时把调度开关置为只读。
	CheckinKey string `json:"checkin_key,omitempty"`
	// CheckinEnabled 是该账号的每日签到开关是否开启。仅当 CheckinKey 非空时有效。
	CheckinEnabled bool `json:"checkin_enabled"`
}

// magpieCheckinAccount 是 Magpie GET /api/settings 返回的 checkinAccounts 元素：
// 单个 WorkBuddy 账号的签到开关（key 用于调 POST /api/settings/checkin-account）。
type magpieCheckinAccount struct {
	Key  string `json:"key"`
	User string `json:"user"`
	On   bool   `json:"on"`
}

// magpieSettingsCheckin 只取 Magpie GET /api/settings 里与签到开关有关的字段。
type magpieSettingsCheckin struct {
	CheckinAccounts []magpieCheckinAccount `json:"checkinAccounts"`
}

// WorkbuddyCheckinResult 是 Magpie 一次 WorkBuddy 签到的单个账号结果
//（POST /api/usage/plugin-checkin 返回的数组元素；只取展示需要的字段）。
// Outcome 取值与 Magpie 一致：claimed（本次领到）、done（今日已签）、
// ineligible（不符合条件）、inactive（当前无签到活动）、captcha（需验证码）、
// own-app（需去 WorkBuddy App 内签）、failed（失败，原因见 Msg）。
type WorkbuddyCheckinResult struct {
	User    string  `json:"user,omitempty"`
	Day     string  `json:"day"`
	At      string  `json:"at,omitempty"`
	Outcome string  `json:"outcome"`
	Credit  float64 `json:"credit,omitempty"`
	Streak  int     `json:"streak,omitempty"`
	Msg     string  `json:"msg,omitempty"`
	Asked   bool    `json:"asked,omitempty"`
}

// WorkbuddyQuotaResult 是 GET /admin/workbuddy/accounts 的返回结构。
type WorkbuddyQuotaResult struct {
	Accounts  []WorkbuddyQuotaAccount `json:"accounts"`
	FetchedAt int64                  `json:"fetched_at"` // Unix 秒，本次拉取时间
}

// workbuddyQuotasResponse 是 Magpie GET /v1/magpie/quotas 的原始返回（只取 data）。
type workbuddyQuotasResponse struct {
	Data []WorkbuddyQuotaAccount `json:"data"`
}

// WorkbuddyQuotaService 从本地 Magpie 网关拉取 WorkBuddy 各账号的额度。
//
// 只读，不写 Magpie、不改账号。Magpie 网关密钥只在后端使用：放进调 Magpie
// 接口的 Authorization 请求头，绝不写日志、不返回给前端。
type WorkbuddyQuotaService struct {
	cfg    *config.Config
	client *http.Client
	flight singleflight.Group
}

// ProvideWorkbuddyQuotaService 构造 WorkBuddy 额度查询服务。
func ProvideWorkbuddyQuotaService(cfg *config.Config) *WorkbuddyQuotaService {
	return NewWorkbuddyQuotaService(cfg)
}

// NewWorkbuddyQuotaService 构造 WorkBuddy 额度查询服务。
func NewWorkbuddyQuotaService(cfg *config.Config) *WorkbuddyQuotaService {
	return &WorkbuddyQuotaService{
		cfg:    cfg,
		client: &http.Client{Timeout: workbuddyQuotaUpstreamTimeout},
	}
}

// ListAccounts 拉取全部 WorkBuddy 账号的额度快照（只保留 provider 为 workbuddy 的卡片）。
func (s *WorkbuddyQuotaService) ListAccounts(ctx context.Context) (*WorkbuddyQuotaResult, error) {
	v, err, _ := s.flight.Do(workbuddyQuotaSingleflightKey, func() (any, error) {
		return s.fetchAccounts(ctx)
	})
	if err != nil {
		return nil, err
	}
	result, ok := v.(*WorkbuddyQuotaResult)
	if !ok || result == nil {
		return nil, infraerrors.New(http.StatusInternalServerError, "WORKBUDDY_QUOTA_INVALID", "invalid workbuddy quota result")
	}
	return result, nil
}

func (s *WorkbuddyQuotaService) fetchAccounts(ctx context.Context) (*WorkbuddyQuotaResult, error) {
	baseURL, gatewayKey := s.magpieEndpoint()
	if gatewayKey == "" {
		return nil, infraerrors.New(http.StatusServiceUnavailable, "WORKBUDDY_NOT_CONFIGURED", "magpie gateway key is not configured (set MAGPIE_GATEWAY_KEY)")
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/v1/magpie/quotas"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, infraerrors.Newf(http.StatusInternalServerError, "WORKBUDDY_REQUEST_BUILD_FAILED", "build request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+gatewayKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, infraerrors.Newf(http.StatusBadGateway, "WORKBUDDY_REQUEST_FAILED", "magpie quotas request failed: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, workbuddyQuotaMaxBodyBytes))
	if err != nil {
		return nil, infraerrors.Newf(http.StatusBadGateway, "WORKBUDDY_READ_FAILED", "read magpie quotas response: %v", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, infraerrors.New(http.StatusBadGateway, "WORKBUDDY_AUTH_FAILED", "magpie gateway key was rejected (401/403)")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, infraerrors.Newf(http.StatusBadGateway, "WORKBUDDY_UPSTREAM_ERROR", "magpie quotas returned status %d", resp.StatusCode)
	}
	var raw workbuddyQuotasResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, infraerrors.Newf(http.StatusBadGateway, "WORKBUDDY_RESPONSE_INVALID", "parse magpie quotas response: %v", err)
	}
	accounts := make([]WorkbuddyQuotaAccount, 0, len(raw.Data))
	for _, acc := range raw.Data {
		if !strings.EqualFold(strings.TrimSpace(acc.Provider), "workbuddy") {
			continue
		}
		if acc.Windows == nil {
			acc.Windows = []WorkbuddyQuotaWindow{}
		}
		accounts = append(accounts, acc)
	}
	// 顺带取各账号的签到开关状态（按 user 名匹配）。取不到也不影响额度本身：
	// 此时 CheckinKey 为空，前端把调度开关置为只读。
	if states := s.fetchCheckinStates(ctx); len(states) > 0 {
		for i := range accounts {
			if st, ok := states[accounts[i].User]; ok {
				accounts[i].CheckinKey = st.Key
				accounts[i].CheckinEnabled = st.On
			}
		}
	}
	return &WorkbuddyQuotaResult{
		Accounts:  accounts,
		FetchedAt: time.Now().Unix(),
	}, nil
}

// fetchCheckinStates 从 Magpie GET /api/settings 拉取各 WorkBuddy 账号的签到开关
// （checkinAccounts：key、user、on），返回按 user 名索引的映射。
//
// Best-effort：任何失败（未配 WEB_KEY、Magpie 报错、解析失败）都返回空映射，
// 调用方不因此失败，只把开关置为只读。WEB_KEY 只放 Cookie，绝不写日志。
func (s *WorkbuddyQuotaService) fetchCheckinStates(ctx context.Context) map[string]magpieCheckinAccount {
	baseURL, webKey := s.magpieCheckinEndpoint()
	if webKey == "" {
		return nil
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/settings"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cookie", workbuddyWebCookieName+"="+webKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, workbuddySettingsMaxBodyBytes))
	if err != nil {
		return nil
	}
	var settings magpieSettingsCheckin
	if err := json.Unmarshal(body, &settings); err != nil {
		return nil
	}
	states := make(map[string]magpieCheckinAccount, len(settings.CheckinAccounts))
	for _, acc := range settings.CheckinAccounts {
		if acc.Key == "" || acc.User == "" {
			continue
		}
		states[acc.User] = acc
	}
	return states
}

// SetCheckinEnabled 开关单个 WorkBuddy 账号的每日签到（调 Magpie Web UI 的
// POST /api/settings/checkin-account，body {key, on}，Cookie 认证）。
// 只改该账号是否参与每日签到，不会立即触发一次签到；开关状态实时生效。
func (s *WorkbuddyQuotaService) SetCheckinEnabled(ctx context.Context, key string, on bool) error {
	baseURL, webKey := s.magpieCheckinEndpoint()
	if webKey == "" {
		return infraerrors.New(http.StatusServiceUnavailable, "WORKBUDDY_CHECKIN_NOT_CONFIGURED", "magpie web key is not configured (set MAGPIE_WEB_KEY)")
	}
	if strings.TrimSpace(key) == "" {
		return infraerrors.New(http.StatusBadRequest, "WORKBUDDY_CHECKIN_KEY_REQUIRED", "checkin account key is required")
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/settings/checkin-account"
	payload, _ := json.Marshal(map[string]any{"key": key, "on": on})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return infraerrors.Newf(http.StatusInternalServerError, "WORKBUDDY_CHECKIN_TOGGLE_REQUEST_BUILD_FAILED", "build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cookie", workbuddyWebCookieName+"="+webKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return infraerrors.Newf(http.StatusBadGateway, "WORKBUDDY_CHECKIN_TOGGLE_REQUEST_FAILED", "magpie checkin-account request failed: %v", err)
	}
	defer resp.Body.Close()
	// 响应体是整份 Magpie 设置（很大），不解析、不留存，只确认状态码。
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, workbuddySettingsMaxBodyBytes))
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return infraerrors.New(http.StatusBadGateway, "WORKBUDDY_CHECKIN_AUTH_FAILED", "magpie web key was rejected (401/403)")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return infraerrors.Newf(http.StatusBadGateway, "WORKBUDDY_CHECKIN_TOGGLE_UPSTREAM_ERROR", "magpie checkin-account returned status %d", resp.StatusCode)
	}
	return nil
}

// Checkin 触发一次 Magpie WorkBuddy 签到，覆盖全部已开启的账号（幂等，
// 当天已签过的账号返回 outcome=done，不会重复领）。
//
// 走 Magpie Web UI 的接口 POST /api/usage/plugin-checkin（provider=workbuddy）。
// 注意：本站的 WorkBuddy 账号是经 opencode-workbuddy-auth 插件登录的，签到由
// 插件自己负责，所以 Magpie 自带的 POST /api/usage/workbuddy-checkin 对这批
// 账号返回空数组；插件账号的签到必须走 plugin-checkin（实测 5 个号均返回）。
// 认证用 Cookie（magpie_web_3430=<WEB_KEY>），WEB_KEY 由环境变量
// MAGPIE_WEB_KEY 提供；调用走与额度接口相同的 base_url（经 Caddy 反代到
// Magpie Web UI 3430）。WEB_KEY 只放请求头，绝不写入日志、不返回给前端。
func (s *WorkbuddyQuotaService) Checkin(ctx context.Context) ([]WorkbuddyCheckinResult, error) {
	baseURL, webKey := s.magpieCheckinEndpoint()
	if webKey == "" {
		return nil, infraerrors.New(http.StatusServiceUnavailable, "WORKBUDDY_CHECKIN_NOT_CONFIGURED", "magpie web key is not configured (set MAGPIE_WEB_KEY)")
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/usage/plugin-checkin"
	payload, _ := json.Marshal(map[string]string{"provider": "workbuddy"})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, infraerrors.Newf(http.StatusInternalServerError, "WORKBUDDY_CHECKIN_REQUEST_BUILD_FAILED", "build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cookie", workbuddyWebCookieName+"="+webKey)

	client := &http.Client{Timeout: workbuddyCheckinTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, infraerrors.Newf(http.StatusBadGateway, "WORKBUDDY_CHECKIN_REQUEST_FAILED", "magpie checkin request failed: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, workbuddyQuotaMaxBodyBytes))
	if err != nil {
		return nil, infraerrors.Newf(http.StatusBadGateway, "WORKBUDDY_CHECKIN_READ_FAILED", "read magpie checkin response: %v", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, infraerrors.New(http.StatusBadGateway, "WORKBUDDY_CHECKIN_AUTH_FAILED", "magpie web key was rejected (401/403)")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, infraerrors.Newf(http.StatusBadGateway, "WORKBUDDY_CHECKIN_UPSTREAM_ERROR", "magpie checkin returned status %d", resp.StatusCode)
	}
	var results []WorkbuddyCheckinResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, infraerrors.Newf(http.StatusBadGateway, "WORKBUDDY_CHECKIN_RESPONSE_INVALID", "parse magpie checkin response: %v", err)
	}
	if results == nil {
		results = []WorkbuddyCheckinResult{}
	}
	return results, nil
}

// magpieCheckinEndpoint 返回 Magpie Web UI 地址与 Web 密钥（签到接口用）。
func (s *WorkbuddyQuotaService) magpieCheckinEndpoint() (string, string) {
	baseURL := "https://wk.jingrouer.top"
	webKey := ""
	if s.cfg != nil {
		if u := strings.TrimSpace(s.cfg.Magpie.BaseURL); u != "" {
			if _, err := url.Parse(u); err == nil {
				baseURL = u
			}
		}
		webKey = strings.TrimSpace(s.cfg.Magpie.WebKey)
	}
	return baseURL, webKey
}

// magpieEndpoint 返回 Magpie 网关地址与密钥。地址默认 https://wk.jingrouer.top，
// 可由环境变量 MAGPIE_BASE_URL 覆盖；密钥只能由环境变量 MAGPIE_GATEWAY_KEY 提供。
func (s *WorkbuddyQuotaService) magpieEndpoint() (string, string) {
	baseURL := "https://wk.jingrouer.top"
	gatewayKey := ""
	if s.cfg != nil {
		if u := strings.TrimSpace(s.cfg.Magpie.BaseURL); u != "" {
			if _, err := url.Parse(u); err == nil {
				baseURL = u
			}
		}
		gatewayKey = strings.TrimSpace(s.cfg.Magpie.GatewayKey)
	}
	return baseURL, gatewayKey
}
