-- 050_add_redeem_rules.sql
-- 新增兑换规则表和用户兑换统计表，用于实现试用码降级功能

-- 兑换规则表
CREATE TABLE IF NOT EXISTS redeem_rules (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL DEFAULT 'balance',
    trigger_value DECIMAL(20,8) NOT NULL,
    max_times_per_user INT NOT NULL DEFAULT 1,
    fallback_value DECIMAL(20,8) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 同一类型和触发金额的规则唯一
CREATE UNIQUE INDEX IF NOT EXISTS idx_redeem_rules_type_trigger_value ON redeem_rules(type, trigger_value);
CREATE INDEX IF NOT EXISTS idx_redeem_rules_is_active ON redeem_rules(is_active);

-- 用户兑换统计表
CREATE TABLE IF NOT EXISTS user_redeem_stats (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redeem_type VARCHAR(20) NOT NULL DEFAULT 'balance',
    value DECIMAL(20,8) NOT NULL,
    used_count INT NOT NULL DEFAULT 0,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 同一用户、类型、金额的统计唯一
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_redeem_stats_user_type_value ON user_redeem_stats(user_id, redeem_type, value);

-- 插入默认的5元试用规则
INSERT INTO redeem_rules (type, trigger_value, max_times_per_user, fallback_value, is_active, description)
VALUES ('balance', 5.00, 1, 2.00, true, '5元试用码限制：每用户限1次原价，超出后按2元兑换')
ON CONFLICT (type, trigger_value) DO NOTHING;
