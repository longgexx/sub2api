package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// PaymentOrder 支付订单 DTO
type PaymentOrder struct {
	ID            int64      `json:"id"`
	TradeNo       string     `json:"trade_no"`
	UserID        int64      `json:"user_id"`
	Amount        float64    `json:"amount"`
	PaymentAmount float64    `json:"payment_amount"`
	CreditAmount  *float64   `json:"credit_amount,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	ExpiredAt     time.Time  `json:"expired_at"`
	AlipayTradeNo *string    `json:"alipay_trade_no,omitempty"`
	PayerAccount  *string    `json:"payer_account,omitempty"`

	User *User `json:"user,omitempty"`
}

// PaymentOrderFromService 从 service 模型转换为 DTO
func PaymentOrderFromService(m *service.PaymentOrder) *PaymentOrder {
	if m == nil {
		return nil
	}
	out := &PaymentOrder{
		ID:            m.ID,
		TradeNo:       m.TradeNo,
		UserID:        m.UserID,
		Amount:        m.Amount,
		PaymentAmount: m.PaymentAmount,
		CreditAmount:  m.CreditAmount,
		Status:        m.Status,
		CreatedAt:     m.CreatedAt,
		PaidAt:        m.PaidAt,
		ExpiredAt:     m.ExpiredAt,
		AlipayTradeNo: m.AlipayTradeNo,
		PayerAccount:  m.PayerAccount,
	}
	if m.User != nil {
		out.User = UserFromService(m.User)
	}
	return out
}

// PaymentOrdersFromService 批量转换
func PaymentOrdersFromService(models []service.PaymentOrder) []PaymentOrder {
	out := make([]PaymentOrder, 0, len(models))
	for i := range models {
		if p := PaymentOrderFromService(&models[i]); p != nil {
			out = append(out, *p)
		}
	}
	return out
}

// PaymentConfig 支付配置 DTO
type PaymentConfig struct {
	Enabled         bool    `json:"enabled"`
	MinAmount       float64 `json:"min_amount"`
	MaxAmount       float64 `json:"max_amount"`
	QRCodeURL       string  `json:"qr_code_url"`
	TimeoutMins     int     `json:"timeout_mins"`
	RateCoefficient float64 `json:"rate_coefficient"` // 充值系数
}

// PaymentConfigFromService 从 service 响应转换为 DTO
func PaymentConfigFromService(m *service.PaymentConfigResponse) *PaymentConfig {
	if m == nil {
		return nil
	}
	return &PaymentConfig{
		Enabled:         m.Enabled,
		MinAmount:       m.MinAmount,
		MaxAmount:       m.MaxAmount,
		QRCodeURL:       m.QRCodeURL,
		TimeoutMins:     m.TimeoutMins,
		RateCoefficient: m.RateCoefficient,
	}
}

// PaymentStats 支付统计 DTO
type PaymentStats struct {
	TodayOrderCount   int64   `json:"today_order_count"`
	TodayPaidCount    int64   `json:"today_paid_count"`
	TodayPaidAmount   float64 `json:"today_paid_amount"`
	TotalOrderCount   int64   `json:"total_order_count"`
	TotalPaidCount    int64   `json:"total_paid_count"`
	TotalPaidAmount   float64 `json:"total_paid_amount"`
	PendingOrderCount int64   `json:"pending_order_count"`
}

// PaymentStatsFromService 从 service 模型转换为 DTO
func PaymentStatsFromService(m *service.PaymentStats) *PaymentStats {
	if m == nil {
		return nil
	}
	return &PaymentStats{
		TodayOrderCount:   m.TodayOrderCount,
		TodayPaidCount:    m.TodayPaidCount,
		TodayPaidAmount:   m.TodayPaidAmount,
		TotalOrderCount:   m.TotalOrderCount,
		TotalPaidCount:    m.TotalPaidCount,
		TotalPaidAmount:   m.TotalPaidAmount,
		PendingOrderCount: m.PendingOrderCount,
	}
}
