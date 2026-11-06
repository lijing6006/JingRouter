package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResetLogsMigration(t *testing.T) {
	content, err := FS.ReadFile("249_reset_logs.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS reset_logs")
	require.Contains(t, sql, "actor_user_id BIGINT")
	require.Contains(t, sql, "kind VARCHAR(32) NOT NULL")
	require.Contains(t, sql, "windows JSONB NOT NULL")
	require.Contains(t, sql, "amount_before JSONB NOT NULL")
	require.Contains(t, sql, "restored BOOLEAN NOT NULL DEFAULT false")
	require.Contains(t, sql, "CREATE INDEX IF NOT EXISTS idx_reset_logs_created_at")
	// 回填 luwei 2026-11-06 重置订阅 6 的记录（审计事件 id 1257），金额写死为核实值。
	require.Contains(t, sql, "admin.subscriptions.reset_quota.create")
	require.Contains(t, sql, "115.2679590720")
	require.Contains(t, sql, "317.3148984520")
	require.Contains(t, sql, "NOT EXISTS ( SELECT 1 FROM reset_logs r")
}
