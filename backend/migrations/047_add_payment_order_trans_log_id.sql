-- Add alipay_trans_log_id column to payment_orders table
-- This column stores the Alipay bill transaction log ID to prevent duplicate matching
-- (one bill should only match one order)

ALTER TABLE payment_orders
ADD COLUMN IF NOT EXISTS alipay_trans_log_id VARCHAR(64) NULL;

-- Add unique constraint to prevent the same bill from matching multiple orders
-- Using partial index to allow multiple NULL values
CREATE UNIQUE INDEX IF NOT EXISTS payment_orders_alipay_trans_log_id_key
ON payment_orders (alipay_trans_log_id)
WHERE alipay_trans_log_id IS NOT NULL;
