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
