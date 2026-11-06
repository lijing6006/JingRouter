package service

import (
	"context"
	"time"
)

// 重置记录类型。
const (
	// ResetLogKindSubscription 订阅用量重置。
	ResetLogKindSubscription = "subscription"
	// ResetLogKindAccount 上游账号配额重置。
	ResetLogKindAccount = "account"
	// ResetLogKindAPIKey API 密钥用量重置。
	ResetLogKindAPIKey = "api_key"
)

// ResetLog 一条用量重置记录。
// 由三处重置路径写入（订阅 / 上游账号 / API 密钥），面板内所有登录用户可见。
// 只追加、不删除；恢复状态由重置看守（或人工恢复）回填。
type ResetLog struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	ActorUserID *int64    `json:"actor_user_id,omitempty"`
	ActorEmail  string    `json:"actor_email"`
	Kind        string    `json:"kind"`
	TargetID    int64     `json:"target_id"`
	TargetLabel string    `json:"target_label"`
	// Windows 本次重置的部位：订阅为 daily/weekly/monthly，
	// 账号为 quota，密钥为 quota/rate_limit。
	Windows map[string]bool `json:"windows"`
	// AmountBefore 重置前用量：订阅为 daily/weekly/monthly（美元），
	// 账号为 quota_used/quota_daily_used/quota_weekly_used，
	// 密钥为 quota_used 与 usage_5h/usage_1d/usage_7d。
	AmountBefore map[string]float64 `json:"amount_before"`
	Restored     bool               `json:"restored"`
	RestoredAt   *time.Time         `json:"restored_at,omitempty"`
	RestoredBy   string             `json:"restored_by,omitempty"`
}

// ResetLogFilter 重置记录列表查询条件。
type ResetLogFilter struct {
	Page     int
	PageSize int

	Kind        string
	ActorUserID *int64
	ActorEmail  string
}

// ResetLogList 分页结果。
type ResetLogList struct {
	Logs     []*ResetLog
	Total    int
	Page     int
	PageSize int
}

// ResetLogRepository 重置记录持久化端口（只追加、不删除）。
type ResetLogRepository interface {
	Insert(ctx context.Context, log *ResetLog) error
	List(ctx context.Context, filter *ResetLogFilter) (*ResetLogList, error)
	// MarkRestored 将记录标记为已恢复（看守自动恢复或人工恢复后调用）。
	MarkRestored(ctx context.Context, id int64, restoredBy string) error
}

// resetActorKey 是请求上下文中重置操作人的 key。
type resetActorKey struct{}

// ResetActor 是执行重置的操作人身份，由 handler 层在调用 service 前
// 写入请求上下文（service 层本身拿不到 gin 的认证信息）。
type ResetActor struct {
	UserID int64
	Email  string
}

// WithResetActor 把操作人写入 ctx。
func WithResetActor(ctx context.Context, userID int64, email string) context.Context {
	return context.WithValue(ctx, resetActorKey{}, ResetActor{UserID: userID, Email: email})
}

// ResetActorFromContext 取出 ctx 中的操作人；没有时返回 false。
func ResetActorFromContext(ctx context.Context) (ResetActor, bool) {
	actor, ok := ctx.Value(resetActorKey{}).(ResetActor)
	return actor, ok
}
