-- 支付充值订单表
-- 用于记录用户通过支付宝经营码充值的订单

CREATE TABLE IF NOT EXISTS payment_orders (
    id BIGSERIAL PRIMARY KEY,
    trade_no VARCHAR(32) NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(20,8) NOT NULL,
    payment_amount DECIMAL(20,8) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ DEFAULT NULL,
    expired_at TIMESTAMPTZ NOT NULL,
    alipay_trade_no VARCHAR(64) DEFAULT NULL,
    notes TEXT DEFAULT NULL
);

-- 唯一索引：交易号
CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_orders_trade_no ON payment_orders(trade_no);

-- 普通索引
CREATE INDEX IF NOT EXISTS idx_payment_orders_user_id ON payment_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_orders_status ON payment_orders(status);
CREATE INDEX IF NOT EXISTS idx_payment_orders_expired_at ON payment_orders(expired_at);
CREATE INDEX IF NOT EXISTS idx_payment_orders_created_at ON payment_orders(created_at);

-- 复合索引：用于按金额匹配待支付订单
CREATE INDEX IF NOT EXISTS idx_payment_orders_payment_amount_status ON payment_orders(payment_amount, status);

-- 表注释
COMMENT ON TABLE payment_orders IS '支付充值订单表';
COMMENT ON COLUMN payment_orders.trade_no IS '系统交易号';
COMMENT ON COLUMN payment_orders.user_id IS '用户ID';
COMMENT ON COLUMN payment_orders.amount IS '用户请求充值金额';
COMMENT ON COLUMN payment_orders.payment_amount IS '实际支付金额（含偏移量，用于金额匹配）';
COMMENT ON COLUMN payment_orders.status IS '状态: pending待支付, paid已支付, expired已过期, cancelled已取消';
COMMENT ON COLUMN payment_orders.paid_at IS '支付时间';
COMMENT ON COLUMN payment_orders.expired_at IS '过期时间';
COMMENT ON COLUMN payment_orders.alipay_trade_no IS '支付宝交易号';
