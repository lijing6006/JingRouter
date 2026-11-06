package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// WorkbuddyHandler 暴露 WorkBuddy 账号额度查询与签到端点（数据来自本地 Magpie 网关）。
//
//   - GET /admin/workbuddy/accounts  全部 WorkBuddy 账号的额度快照
//   - POST /admin/workbuddy/checkin  触发一次 Magpie WorkBuddy 签到（覆盖全部已开启的账号，幂等）
//   - POST /admin/workbuddy/checkin-account  单个账号的签到开关（调 Magpie 的签到开关接口）
//
// Magpie 网关密钥与 Web 密钥只在后端 service 里使用，前端只调本接口、拿不到密钥。
type WorkbuddyHandler struct {
	quotaService *service.WorkbuddyQuotaService
}

func NewWorkbuddyHandler(quotaService *service.WorkbuddyQuotaService) *WorkbuddyHandler {
	return &WorkbuddyHandler{
		quotaService: quotaService,
	}
}

// ListAccounts 拉取全部 WorkBuddy 账号的额度快照。
func (h *WorkbuddyHandler) ListAccounts(c *gin.Context) {
	if h == nil || h.quotaService == nil {
		response.Error(c, 503, "workbuddy quota service is not enabled")
		return
	}
	result, err := h.quotaService.ListAccounts(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// Checkin 触发一次 Magpie WorkBuddy 签到：覆盖全部已开启的账号（不是只签当前卡片
// 那个号），当天已签过的返回 done，不会重复领。签完前端会刷新一次额度。
func (h *WorkbuddyHandler) Checkin(c *gin.Context) {
	if h == nil || h.quotaService == nil {
		response.Error(c, 503, "workbuddy quota service is not enabled")
		return
	}
	results, err := h.quotaService.Checkin(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, results)
}

// SetCheckinEnabledRequest 是 POST /admin/workbuddy/checkin-account 的请求体：
// key 为 Magpie 签到开关的账号 key（GET /admin/workbuddy/accounts 返回的
// checkin_key），on 为开关目标状态。
type SetCheckinEnabledRequest struct {
	Key string `json:"key"`
	On  bool   `json:"on"`
}

// SetCheckinEnabled 开关单个 WorkBuddy 账号的每日签到（调 Magpie
// POST /api/settings/checkin-account）。只改该账号是否参与每日签到，
// 不会立即触发一次签到；开关状态实时生效。
func (h *WorkbuddyHandler) SetCheckinEnabled(c *gin.Context) {
	if h == nil || h.quotaService == nil {
		response.Error(c, 503, "workbuddy quota service is not enabled")
		return
	}
	var req SetCheckinEnabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid request body")
		return
	}
	if err := h.quotaService.SetCheckinEnabled(c.Request.Context(), req.Key, req.On); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"key": req.Key, "on": req.On})
}
