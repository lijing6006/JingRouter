package service

import (
	"context"
	"fmt"
	"time"
)

// ResetLogService 重置记录服务：写入（同步落库）与分页读取。
// 写入失败只返回错误，不阻断调用方的重置操作——调用方只记日志。
type ResetLogService struct {
	repo ResetLogRepository
}

// NewResetLogService 创建重置记录服务。
func NewResetLogService(repo ResetLogRepository) *ResetLogService {
	return &ResetLogService{repo: repo}
}

// Record 同步写入一条重置记录。
func (s *ResetLogService) Record(ctx context.Context, log *ResetLog) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("nil reset log service")
	}
	if log == nil {
		return fmt.Errorf("nil reset log")
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}
	if log.Windows == nil {
		log.Windows = map[string]bool{}
	}
	if log.AmountBefore == nil {
		log.AmountBefore = map[string]float64{}
	}
	return s.repo.Insert(ctx, log)
}

// List 分页查询重置记录（时间倒序）。
func (s *ResetLogService) List(ctx context.Context, filter *ResetLogFilter) (*ResetLogList, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("nil reset log service")
	}
	return s.repo.List(ctx, filter)
}

// MarkRestored 将记录标记为已恢复（看守自动恢复或人工恢复后调用）。
func (s *ResetLogService) MarkRestored(ctx context.Context, id int64, restoredBy string) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("nil reset log service")
	}
	return s.repo.MarkRestored(ctx, id, restoredBy)
}
