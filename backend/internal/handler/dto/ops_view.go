package dto

import "time"

// OpsViewOverview 运营概览数据
type OpsViewOverview struct {
	// 余额指标
	TotalBalance float64 `json:"total_balance"` // 所有用户余额总和

	// 消耗指标
	TodayConsumption float64 `json:"today_consumption"` // 今日消耗
	WeekConsumption  float64 `json:"week_consumption"`  // 本周消耗
	MonthConsumption float64 `json:"month_consumption"` // 本月消耗

	// 充值指标
	TodayRecharge float64 `json:"today_recharge"` // 今日充值总额
	WeekRecharge  float64 `json:"week_recharge"`  // 本周充值总额
	MonthRecharge float64 `json:"month_recharge"` // 本月充值总额

	// 充值明细
	TodayPaymentRecharge float64 `json:"today_payment_recharge"` // 今日平台充值
	TodayRedeemRecharge  float64 `json:"today_redeem_recharge"`  // 今日兑换充值
	WeekPaymentRecharge  float64 `json:"week_payment_recharge"`  // 本周平台充值
	WeekRedeemRecharge   float64 `json:"week_redeem_recharge"`   // 本周兑换充值
	MonthPaymentRecharge float64 `json:"month_payment_recharge"` // 本月平台充值
	MonthRedeemRecharge  float64 `json:"month_redeem_recharge"`  // 本月兑换充值

	// 用户指标
	TotalUsers       int64   `json:"total_users"`        // 总用户数
	TodayActiveUsers int64   `json:"today_active_users"` // 今日活跃用户
	TodayNewUsers    int64   `json:"today_new_users"`    // 今日新增用户
	WeekActiveUsers  int64   `json:"week_active_users"`  // 本周活跃用户
	WeekNewUsers     int64   `json:"week_new_users"`     // 本周新增用户
	WeekPayingUsers  int64   `json:"week_paying_users"`  // 本周付费用户
	WeekRetention    float64 `json:"week_retention"`     // 周留存率 (%)
	MonthActiveUsers int64   `json:"month_active_users"` // 本月活跃用户
	MonthNewUsers    int64   `json:"month_new_users"`    // 本月新增用户
}

// OpsViewComparison 今日 vs 昨日对比数据
type OpsViewComparison struct {
	// 消耗对比
	TodayConsumption     float64 `json:"today_consumption"`
	YesterdayConsumption float64 `json:"yesterday_consumption"`
	ConsumptionChange    float64 `json:"consumption_change"` // 变化率 (%)

	// 充值对比
	TodayRecharge     float64 `json:"today_recharge"`
	YesterdayRecharge float64 `json:"yesterday_recharge"`
	RechargeChange    float64 `json:"recharge_change"` // 变化率 (%)

	// 活跃用户对比
	TodayActiveUsers     int64   `json:"today_active_users"`
	YesterdayActiveUsers int64   `json:"yesterday_active_users"`
	ActiveUsersChange    float64 `json:"active_users_change"` // 变化率 (%)

	// 新增用户对比
	TodayNewUsers     int64   `json:"today_new_users"`
	YesterdayNewUsers int64   `json:"yesterday_new_users"`
	NewUsersChange    float64 `json:"new_users_change"` // 变化率 (%)
}

// OpsViewTrendPoint 趋势数据点
type OpsViewTrendPoint struct {
	Date            string  `json:"date"`
	Consumption     float64 `json:"consumption"`      // 消耗金额
	PaymentRecharge float64 `json:"payment_recharge"` // 平台充值
	RedeemRecharge  float64 `json:"redeem_recharge"`  // 兑换充值
	TotalRecharge   float64 `json:"total_recharge"`   // 充值总额
}

// OpsViewUserGrowthPoint 用户增长数据点
type OpsViewUserGrowthPoint struct {
	Date         string  `json:"date"`
	NewUsers     int64   `json:"new_users"`     // 新增用户
	ActiveUsers  int64   `json:"active_users"`  // 活跃用户
	PayingUsers  int64   `json:"paying_users"`  // 付费用户
	RetentionD1  float64 `json:"retention_d1"`  // 次日留存率
	RetentionD7  float64 `json:"retention_d7"`  // 7日留存率
	RetentionD30 float64 `json:"retention_d30"` // 30日留存率
}

// OpsViewGroupConsumption 分组消耗数据
type OpsViewGroupConsumption struct {
	GroupID     int64   `json:"group_id"`
	GroupName   string  `json:"group_name"`
	Consumption float64 `json:"consumption"`
	Percentage  float64 `json:"percentage"` // 占比 (%)
	Requests    int64   `json:"requests"`
}

// OpsViewTopUser 高价值用户
type OpsViewTopUser struct {
	UserID       int64     `json:"user_id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	Consumption  float64   `json:"consumption"` // 总消耗
	Requests     int64     `json:"requests"`    // 请求数
	AvgCost      float64   `json:"avg_cost"`    // 平均单价
	LastActiveAt time.Time `json:"last_active_at"`
	RegisteredAt time.Time `json:"registered_at"`
}

// OpsViewModelRanking 模型使用排行
type OpsViewModelRanking struct {
	Model       string  `json:"model"`
	Requests    int64   `json:"requests"`
	Tokens      int64   `json:"tokens"`
	Consumption float64 `json:"consumption"`
	Percentage  float64 `json:"percentage"` // 占比 (%)
}

// OpsViewActivityHeatmapCell 活跃热力图单元
type OpsViewActivityHeatmapCell struct {
	Weekday  int   `json:"weekday"` // 0-6 (周日-周六)
	Hour     int   `json:"hour"`    // 0-23
	Requests int64 `json:"requests"`
}

// OpsViewRetentionCohort 留存队列数据
type OpsViewRetentionCohort struct {
	CohortDate   string  `json:"cohort_date"`
	CohortSize   int64   `json:"cohort_size"`
	RetentionD1  float64 `json:"retention_d1"`
	RetentionD7  float64 `json:"retention_d7"`
	RetentionD14 float64 `json:"retention_d14"`
	RetentionD30 float64 `json:"retention_d30"`
}
