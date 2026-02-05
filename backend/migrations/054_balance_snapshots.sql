-- Balance snapshots: 记录每日用户总余额的历史快照
-- 用于支持余额趋势图功能

CREATE TABLE IF NOT EXISTS balance_snapshots_daily (
    -- 快照日期（UTC）
    snapshot_date DATE NOT NULL PRIMARY KEY,

    -- 快照数据
    total_balance DECIMAL(20,8) NOT NULL DEFAULT 0,  -- 所有用户余额总和
    user_count BIGINT NOT NULL DEFAULT 0,            -- 有余额的用户数

    -- 元数据
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()   -- 计算时间
);

CREATE INDEX IF NOT EXISTS idx_balance_snapshots_daily_date
    ON balance_snapshots_daily (snapshot_date DESC);

COMMENT ON TABLE balance_snapshots_daily IS '日级用户余额汇总快照，用于支持余额趋势图';
COMMENT ON COLUMN balance_snapshots_daily.snapshot_date IS '快照日期（UTC）';
COMMENT ON COLUMN balance_snapshots_daily.total_balance IS '所有用户余额总和';
COMMENT ON COLUMN balance_snapshots_daily.user_count IS '有余额的用户数';
COMMENT ON COLUMN balance_snapshots_daily.computed_at IS '快照计算时间';
