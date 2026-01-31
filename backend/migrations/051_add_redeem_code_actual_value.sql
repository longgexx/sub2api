-- 051_add_redeem_code_actual_value.sql
-- 新增 actual_value 字段，用于记录兑换码的实际到账金额（降级兑换时与原始面值不同）

ALTER TABLE redeem_codes ADD COLUMN IF NOT EXISTS actual_value DECIMAL(20,8);

-- 为已使用的兑换码，将 actual_value 设置为与 value 相同（历史数据兼容）
UPDATE redeem_codes SET actual_value = value WHERE status = 'used' AND actual_value IS NULL;
