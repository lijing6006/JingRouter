package service

import (
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
	// workbuddyQuotaMaxBodyBytes Magpie 额度接口响应体上限（4 个账号的窗口数据很小）。
	workbuddyQuotaMaxBodyBytes = 256 * 1024
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
	Provider string                  `json:"provider"` // 固定 "workbuddy"
	Name     string                  `json:"name"`     // 如 "WorkBuddy"
	Plan     string                  `json:"plan"`     // 如 "Free plan"
	User     string                  `json:"user"`     // 账号名，如 "小牛来咯"
	Kind     string                  `json:"kind,omitempty"`
	ReadAt   string                  `json:"read_at,omitempty"`
	Windows  []WorkbuddyQuotaWindow  `json:"windows"`
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
	return &WorkbuddyQuotaResult{
		Accounts:  accounts,
		FetchedAt: time.Now().Unix(),
	}, nil
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
