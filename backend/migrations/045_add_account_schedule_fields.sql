-- 为 Anthropic OAuth/SetupToken 账号添加时间段调度功能
-- 仅在时间窗口内允许调度账号

-- 是否启用时间段调度
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS schedule_enabled BOOLEAN NOT NULL DEFAULT FALSE;

-- 调度时区（如 Asia/Shanghai, UTC）
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS schedule_timezone VARCHAR(50) NOT NULL DEFAULT 'UTC';

-- 调度规则（JSONB 数组）
-- 每条规则包含：weekdays（星期几，0=周日）、start_minute（开始分钟）、end_minute（结束分钟）
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS schedule_rules JSONB;

-- 为 schedule_enabled 添加索引，用于筛选启用时间段调度的账户
CREATE INDEX IF NOT EXISTS idx_accounts_schedule_enabled ON accounts(schedule_enabled);
