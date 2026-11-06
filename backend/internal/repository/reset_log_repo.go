package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// resetLogRepository 重置记录仓储（raw SQL，append-only）。
// 刻意不提供删除：重置记录只允许追加，恢复状态通过 MarkRestored 回填。
type resetLogRepository struct {
	db *sql.DB
}

// NewResetLogRepository 创建重置记录仓储。
func NewResetLogRepository(db *sql.DB) service.ResetLogRepository {
	return &resetLogRepository{db: db}
}

const resetLogSelectColumns = `
	l.id,
	l.created_at,
	l.actor_user_id,
	l.actor_email,
	l.kind,
	l.target_id,
	l.target_label,
	l.windows,
	l.amount_before,
	l.restored,
	l.restored_at,
	l.restored_by
`

func resetLogJSONValue(v any) string {
	if v == nil {
		return "{}"
	}
	if encoded, err := json.Marshal(v); err == nil {
		return string(encoded)
	}
	return "{}"
}

// Insert 同步写入一条重置记录，写回自增 id。
func (r *resetLogRepository) Insert(ctx context.Context, log *service.ResetLog) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("nil reset log repository")
	}
	createdAt := log.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	var restoredAt any
	if log.RestoredAt != nil {
		restoredAt = log.RestoredAt.UTC()
	}
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO reset_logs
			(created_at, actor_user_id, actor_email, kind, target_id, target_label,
			 windows, amount_before, restored, restored_at, restored_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8::jsonb, $9, $10, $11)
			RETURNING id`,
		createdAt.UTC(),
		nullInt64Ptr(log.ActorUserID),
		truncateString(log.ActorEmail, 255),
		truncateString(log.Kind, 32),
		log.TargetID,
		truncateString(log.TargetLabel, 255),
		resetLogJSONValue(log.Windows),
		resetLogJSONValue(log.AmountBefore),
		log.Restored,
		restoredAt,
		truncateString(log.RestoredBy, 255),
	).Scan(&log.ID)
	if err != nil {
		return fmt.Errorf("insert reset log: %w", err)
	}
	return nil
}

// MarkRestored 将记录标记为已恢复（看守自动恢复或人工恢复后调用）。
func (r *resetLogRepository) MarkRestored(ctx context.Context, id int64, restoredBy string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("nil reset log repository")
	}
	result, err := r.db.ExecContext(ctx,
		`UPDATE reset_logs
			SET restored = true, restored_at = NOW(), restored_by = $2
			WHERE id = $1`,
		id, truncateString(restoredBy, 255))
	if err != nil {
		return fmt.Errorf("mark reset log restored: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("reset log %d not found", id)
	}
	return nil
}

// List 分页查询重置记录（时间倒序）。
func (r *resetLogRepository) List(ctx context.Context, filter *service.ResetLogFilter) (*service.ResetLogList, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil reset log repository")
	}
	if filter == nil {
		filter = &service.ResetLogFilter{}
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	where, args := buildResetLogsWhere(filter)
	countSQL := "SELECT COUNT(*) FROM reset_logs l " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	argsWithLimit := append(args, pageSize, offset)
	query := "SELECT" + resetLogSelectColumns + "\nFROM reset_logs l\n" + where + `
ORDER BY l.created_at DESC, l.id DESC
LIMIT $` + itoa(len(args)+1) + ` OFFSET $` + itoa(len(args)+2)

	rows, err := r.db.QueryContext(ctx, query, argsWithLimit...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	logs := make([]*service.ResetLog, 0, pageSize)
	for rows.Next() {
		item, err := scanResetLogRow(rows.Scan)
		if err != nil {
			return nil, err
		}
		logs = append(logs, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &service.ResetLogList{
		Logs:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func buildResetLogsWhere(filter *service.ResetLogFilter) (string, []any) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	clauses = append(clauses, "1=1")

	if v := strings.TrimSpace(filter.Kind); v != "" {
		args = append(args, v)
		clauses = append(clauses, "l.kind = $"+itoa(len(args)))
	}
	if filter.ActorUserID != nil {
		args = append(args, *filter.ActorUserID)
		clauses = append(clauses, "l.actor_user_id = $"+itoa(len(args)))
	}
	if v := strings.TrimSpace(filter.ActorEmail); v != "" {
		args = append(args, "%"+escapeLikePattern(v)+"%")
		clauses = append(clauses, "l.actor_email ILIKE $"+itoa(len(args)))
	}

	return "WHERE " + strings.Join(clauses, " AND "), args
}

func scanResetLogRow(scan func(dest ...any) error) (*service.ResetLog, error) {
	item := &service.ResetLog{}
	var actorUserID sql.NullInt64
	var restoredAt sql.NullTime
	var windowsRaw, amountBeforeRaw string
	if err := scan(
		&item.ID,
		&item.CreatedAt,
		&actorUserID,
		&item.ActorEmail,
		&item.Kind,
		&item.TargetID,
		&item.TargetLabel,
		&windowsRaw,
		&amountBeforeRaw,
		&item.Restored,
		&restoredAt,
		&item.RestoredBy,
	); err != nil {
		return nil, err
	}
	if actorUserID.Valid {
		v := actorUserID.Int64
		item.ActorUserID = &v
	}
	if restoredAt.Valid {
		v := restoredAt.Time
		item.RestoredAt = &v
	}
	windowsRaw = strings.TrimSpace(windowsRaw)
	if windowsRaw != "" && windowsRaw != "null" {
		windows := make(map[string]bool)
		if err := json.Unmarshal([]byte(windowsRaw), &windows); err == nil {
			item.Windows = windows
		}
	}
	if item.Windows == nil {
		item.Windows = map[string]bool{}
	}
	amountBeforeRaw = strings.TrimSpace(amountBeforeRaw)
	if amountBeforeRaw != "" && amountBeforeRaw != "null" {
		amounts := make(map[string]float64)
		if err := json.Unmarshal([]byte(amountBeforeRaw), &amounts); err == nil {
			item.AmountBefore = amounts
		}
	}
	if item.AmountBefore == nil {
		item.AmountBefore = map[string]float64{}
	}
	return item, nil
}
