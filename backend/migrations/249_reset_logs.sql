-- 重置操作日志：订阅 / 上游账号 / API 密钥三类用量重置记录。
-- 面板内所有登录用户可见（读接口只要求登录，不限管理员），用于公开谁在什么时候
-- 重置了什么、清掉了多少额度。记录永久保留，不走审计日志的保留期清理。
CREATE TABLE IF NOT EXISTS reset_logs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actor_user_id BIGINT,
    actor_email VARCHAR(255) NOT NULL DEFAULT '',
    kind VARCHAR(32) NOT NULL,
    target_id BIGINT NOT NULL,
    target_label VARCHAR(255) NOT NULL DEFAULT '',
    windows JSONB NOT NULL DEFAULT '{}'::jsonb,
    amount_before JSONB NOT NULL DEFAULT '{}'::jsonb,
    restored BOOLEAN NOT NULL DEFAULT false,
    restored_at TIMESTAMPTZ,
    restored_by VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_reset_logs_created_at ON reset_logs (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_reset_logs_kind ON reset_logs (kind);
CREATE INDEX IF NOT EXISTS idx_reset_logs_actor_user_id ON reset_logs (actor_user_id);

COMMENT ON TABLE reset_logs IS '重置操作日志：订阅/上游账号/API密钥的用量重置记录，所有登录用户可见';
COMMENT ON COLUMN reset_logs.kind IS '重置类型：subscription=订阅，account=上游账号，api_key=API密钥';
COMMENT ON COLUMN reset_logs.target_label IS '目标描述：订阅属主邮箱 / 上游账号名 / 密钥名';
COMMENT ON COLUMN reset_logs.windows IS '订阅重置的窗口，如 {"daily":true,"weekly":true,"monthly":true}；账号/密钥记重置部位，如 {"quota":true}';
COMMENT ON COLUMN reset_logs.amount_before IS '重置前用量，如订阅 {"daily":115.27,"weekly":317.31,"monthly":317.31}';
COMMENT ON COLUMN reset_logs.restored IS '是否已恢复（看守按 usage_logs 自动恢复后置 true）';

-- 回填 2026-11-06 luwei 自行重置订阅 6 的记录（审计事件 id 1257）。
-- amount_before 直接写死为重置前核实的真实值：订阅 6 的计数器事后已被恢复，
-- 不能用当前值反推（重置后若有新用量，当前值会比重置前大）。
-- restored_at 为人工恢复完成的大致时间（当晚约 21:00）。
INSERT INTO reset_logs (created_at, actor_user_id, actor_email, kind, target_id, target_label, windows, amount_before, restored, restored_at, restored_by)
SELECT a.created_at, a.actor_user_id, a.actor_email, 'subscription',
       6, COALESCE(u.email, ''), '{"daily":true,"weekly":true,"monthly":true}'::jsonb,
       '{"daily":115.2679590720,"weekly":317.3148984520,"monthly":317.3148984520}'::jsonb,
       true, '2026-11-06 21:00:00+08'::timestamptz, 'manual'
FROM audit_logs a
JOIN user_subscriptions s ON s.id = 6
LEFT JOIN users u ON u.id = s.user_id
WHERE a.id = 1257
  AND a.action = 'admin.subscriptions.reset_quota.create'
  AND NOT EXISTS (
      SELECT 1 FROM reset_logs r
      WHERE r.kind = 'subscription' AND r.target_id = 6 AND r.created_at = a.created_at
  );
