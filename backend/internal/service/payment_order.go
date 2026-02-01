package service

import "time"

// PaymentOrder 支付订单常量
const (
	PaymentStatusPending   = "pending"
	PaymentStatusPaid      = "paid"
	PaymentStatusExpired   = "expired"
	PaymentStatusCancelled = "cancelled"
)

// PaymentOrder 支付订单模型
type PaymentOrder struct {
	ID               int64
	TradeNo          string
	UserID           int64
	Amount           float64  // 用户请求充值金额
	PaymentAmount    float64  // 实际支付金额（含偏移量）
	CreditAmount     *float64 // 实际到账金额（支付成功后计算存储）
	RateCoefficient  float64  // 创建订单时的充值系数
	Status           string
	CreatedAt        time.Time
	PaidAt           *time.Time
	ExpiredAt        time.Time
	AlipayTradeNo    *string
	AlipayTransLogID *string // 支付宝账单流水号（用于防止重复匹配）
	PayerAccount     *string // 付款人账户
	Notes            *string

	// 关联
	User *User
}

// IsPending 检查是否为待支付状态
func (p *PaymentOrder) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// IsPaid 检查是否已支付
func (p *PaymentOrder) IsPaid() bool {
	return p.Status == PaymentStatusPaid
}

// IsExpired 检查是否已过期
func (p *PaymentOrder) IsExpired() bool {
	return p.Status == PaymentStatusExpired || time.Now().After(p.ExpiredAt)
}

// PaymentStats 支付统计
type PaymentStats struct {
	TodayOrderCount   int64
	TodayPaidCount    int64
	TodayPaidAmount   float64
	TotalOrderCount   int64
	TotalPaidCount    int64
	TotalPaidAmount   float64
	PendingOrderCount int64
}
