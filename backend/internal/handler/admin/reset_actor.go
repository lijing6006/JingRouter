package admin

import (
	"context"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// withResetActor 把当前操作人（gin 认证信息）写入给定的请求上下文，
// 供 service 层的重置方法记录“谁点的重置”。
// 拿不到身份时原样返回 ctx，记录里操作人为空（不阻断重置）。
// 调用方须传入真正要用的 ctx（不要用 c.Request.Context() 覆盖带超时/幂等的 ctx）。
func withResetActor(c *gin.Context, ctx context.Context) context.Context {
	var userID int64
	if subject, ok := middleware2.GetAuthSubjectFromContext(c); ok {
		userID = subject.UserID
	}
	email := c.GetString(middleware2.ContextKeyAuthEmail)
	if userID <= 0 && email == "" {
		return ctx
	}
	return service.WithResetActor(ctx, userID, email)
}
