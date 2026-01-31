-- Add payer_account column to payment_orders table
-- This column stores the payer account information from Alipay bill

ALTER TABLE payment_orders
ADD COLUMN IF NOT EXISTS payer_account VARCHAR(128) NULL;
