package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// WorkbuddyHandler 暴露 WorkBuddy 账号额度查询端点（数据来自本地 Magpie 网关）。
//
//   - GET /admin/workbuddy/accounts  全部 WorkBuddy 账号的额度快照
//
// Magpie 网关密钥只在后端 service 里使用，前端只调本接口、拿不到密钥。
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
