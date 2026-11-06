package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ResetLogHandler 重置记录查询接口。
// 重置记录对所有登录用户可见（不限管理员）：读接口挂在用户认证路由组，
// 而非 admin-only 路由组；重置操作本身仍保持 admin-only。
type ResetLogHandler struct {
	resetLogService *service.ResetLogService
}

// NewResetLogHandler 创建重置记录处理器。
func NewResetLogHandler(resetLogService *service.ResetLogService) *ResetLogHandler {
	return &ResetLogHandler{
		resetLogService: resetLogService,
	}
}

// List 分页查询重置记录（时间倒序）。
// GET /api/v1/admin/reset-logs
func (h *ResetLogHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	if pageSize > 200 {
		pageSize = 200
	}

	filter := &service.ResetLogFilter{
		Page:       page,
		PageSize:   pageSize,
		Kind:       strings.TrimSpace(c.Query("kind")),
		ActorEmail: strings.TrimSpace(c.Query("actor_email")),
	}

	if v := strings.TrimSpace(c.Query("actor_user_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid actor_user_id")
			return
		}
		filter.ActorUserID = &id
	}

	result, err := h.resetLogService.List(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Logs, int64(result.Total), result.Page, result.PageSize)
}
