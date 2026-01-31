package service

import "time"

// RedeemRule 兑换规则模型
type RedeemRule struct {
	ID              int64
	Type            string  // 兑换码类型：balance/concurrency/subscription
	TriggerValue    float64 // 触发金额
	MaxTimesPerUser int     // 每用户可享受原价的次数
	FallbackValue   float64 // 降级金额
	IsActive        bool    // 是否启用
	Description     string  // 规则说明
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UserRedeemStat 用户兑换统计模型
type UserRedeemStat struct {
	ID         int64
	UserID     int64
	RedeemType string  // 兑换码类型
	Value      float64 // 兑换码面值
	UsedCount  int     // 使用次数
	LastUsedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// RedeemResult 兑换结果（扩展返回信息）
type RedeemResult struct {
	RedeemCode    *RedeemCode
	OriginalValue float64 // 原始面值
	ActualValue   float64 // 实际到账金额
	IsDegraded    bool    // 是否降级
	Message       string  // 提示消息
}

// RedeemPreview 兑换预检结果
type RedeemPreview struct {
	RedeemCode     *RedeemCode
	OriginalValue  float64 // 原始面值
	ActualValue    float64 // 实际到账金额（如降级）
	WillDegrade    bool    // 是否会降级
	NeedsConfirm   bool    // 是否需要用户确认
	ConfirmMessage string  // 确认提示消息
}
