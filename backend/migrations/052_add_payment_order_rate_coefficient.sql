-- Add rate_coefficient column to payment_orders table
-- This column stores the recharge coefficient at order creation time
-- Used to calculate the actual credit amount when payment is confirmed

ALTER TABLE payment_orders
ADD COLUMN IF NOT EXISTS rate_coefficient DECIMAL(10,4) NOT NULL DEFAULT 1.0;

COMMENT ON COLUMN payment_orders.rate_coefficient IS '创建订单时的充值系数，用于计算到账金额';
