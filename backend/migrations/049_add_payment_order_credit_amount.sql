-- Add credit_amount column to payment_orders table
-- This column stores the actual amount credited to user's balance

ALTER TABLE payment_orders
ADD COLUMN IF NOT EXISTS credit_amount DECIMAL(10,2) NULL;

-- Update historical data: set credit_amount to amount for paid orders (historical data with coefficient=1)
UPDATE payment_orders
SET credit_amount = amount
WHERE status = 'paid' AND credit_amount IS NULL;
