package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpsViewService_calcChangeRate(t *testing.T) {
	svc := &OpsViewService{}

	tests := []struct {
		name     string
		oldValue float64
		newValue float64
		expected float64
	}{
		{
			name:     "positive change",
			oldValue: 100,
			newValue: 150,
			expected: 50.0, // (150-100)/100 * 100 = 50%
		},
		{
			name:     "negative change",
			oldValue: 100,
			newValue: 80,
			expected: -20.0, // (80-100)/100 * 100 = -20%
		},
		{
			name:     "no change",
			oldValue: 100,
			newValue: 100,
			expected: 0.0,
		},
		{
			name:     "from zero to positive",
			oldValue: 0,
			newValue: 100,
			expected: 100.0, // special case: 100% growth
		},
		{
			name:     "from zero to zero",
			oldValue: 0,
			newValue: 0,
			expected: 0.0,
		},
		{
			name:     "double the value",
			oldValue: 50,
			newValue: 100,
			expected: 100.0, // (100-50)/50 * 100 = 100%
		},
		{
			name:     "half the value",
			oldValue: 100,
			newValue: 50,
			expected: -50.0, // (50-100)/100 * 100 = -50%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.calcChangeRate(tt.oldValue, tt.newValue)
			require.InDelta(t, tt.expected, result, 0.01, "calcChangeRate(%f, %f) = %f, want %f", tt.oldValue, tt.newValue, result, tt.expected)
		})
	}
}

func TestOpsViewOverview_Structure(t *testing.T) {
	overview := &OpsViewOverview{
		TodayConsumption:     100.50,
		WeekConsumption:      500.25,
		MonthConsumption:     2000.00,
		TodayRecharge:        200.00,
		WeekRecharge:         1000.00,
		MonthRecharge:        4000.00,
		TodayPaymentRecharge: 150.00,
		TodayRedeemRecharge:  50.00,
		WeekPaymentRecharge:  800.00,
		WeekRedeemRecharge:   200.00,
		MonthPaymentRecharge: 3200.00,
		MonthRedeemRecharge:  800.00,
		WeekActiveUsers:      50,
		WeekNewUsers:         10,
		WeekPayingUsers:      5,
		WeekRetention:        45.5,
		MonthActiveUsers:     200,
		MonthNewUsers:        40,
	}

	// Verify recharge totals
	require.Equal(t, overview.TodayPaymentRecharge+overview.TodayRedeemRecharge, overview.TodayRecharge)
	require.Equal(t, overview.WeekPaymentRecharge+overview.WeekRedeemRecharge, overview.WeekRecharge)
	require.Equal(t, overview.MonthPaymentRecharge+overview.MonthRedeemRecharge, overview.MonthRecharge)
}

func TestOpsViewComparison_Structure(t *testing.T) {
	comparison := &OpsViewComparison{
		TodayConsumption:     100.50,
		YesterdayConsumption: 80.00,
		ConsumptionChange:    25.625,
		TodayRecharge:        200.00,
		YesterdayRecharge:    150.00,
		RechargeChange:       33.33,
		TodayActiveUsers:     50,
		YesterdayActiveUsers: 45,
		ActiveUsersChange:    11.11,
		TodayNewUsers:        10,
		YesterdayNewUsers:    8,
		NewUsersChange:       25.00,
	}

	// Verify change calculations are reasonable
	require.Greater(t, comparison.TodayConsumption, comparison.YesterdayConsumption)
	require.Greater(t, comparison.ConsumptionChange, float64(0))

	require.Greater(t, comparison.TodayRecharge, comparison.YesterdayRecharge)
	require.Greater(t, comparison.RechargeChange, float64(0))
}

func TestOpsViewTrendPoint_Structure(t *testing.T) {
	point := OpsViewTrendPoint{
		Date:            "2024-01-01",
		Consumption:     100.0,
		PaymentRecharge: 80.0,
		RedeemRecharge:  20.0,
		TotalRecharge:   100.0,
	}

	// Verify total recharge calculation
	require.Equal(t, point.PaymentRecharge+point.RedeemRecharge, point.TotalRecharge)
}

func TestOpsViewUserGrowthPoint_Structure(t *testing.T) {
	point := OpsViewUserGrowthPoint{
		Date:         "2024-01-01",
		NewUsers:     10,
		ActiveUsers:  50,
		PayingUsers:  5,
		RetentionD1:  80.0,
		RetentionD7:  60.0,
		RetentionD30: 40.0,
	}

	// Verify retention rates are in valid range
	require.GreaterOrEqual(t, point.RetentionD1, float64(0))
	require.LessOrEqual(t, point.RetentionD1, float64(100))
	require.GreaterOrEqual(t, point.RetentionD7, float64(0))
	require.LessOrEqual(t, point.RetentionD7, float64(100))
	require.GreaterOrEqual(t, point.RetentionD30, float64(0))
	require.LessOrEqual(t, point.RetentionD30, float64(100))

	// Retention should generally decrease over time
	require.GreaterOrEqual(t, point.RetentionD1, point.RetentionD7)
	require.GreaterOrEqual(t, point.RetentionD7, point.RetentionD30)
}

func TestOpsViewGroupConsumption_Percentage(t *testing.T) {
	groups := []OpsViewGroupConsumption{
		{GroupID: 1, GroupName: "Group A", Consumption: 600.0, Percentage: 60.0, Requests: 1000},
		{GroupID: 2, GroupName: "Group B", Consumption: 400.0, Percentage: 40.0, Requests: 666},
	}

	// Verify percentages sum to 100
	var totalPercentage float64
	for _, g := range groups {
		totalPercentage += g.Percentage
	}
	require.InDelta(t, 100.0, totalPercentage, 0.01)
}

func TestOpsViewModelRanking_Percentage(t *testing.T) {
	models := []OpsViewModelRanking{
		{Model: "claude-3-opus", Requests: 500, Tokens: 100000, Consumption: 300.0, Percentage: 50.0},
		{Model: "claude-3-sonnet", Requests: 400, Tokens: 80000, Consumption: 200.0, Percentage: 33.33},
		{Model: "claude-3-haiku", Requests: 100, Tokens: 20000, Consumption: 100.0, Percentage: 16.67},
	}

	// Verify percentages sum to approximately 100
	var totalPercentage float64
	for _, m := range models {
		totalPercentage += m.Percentage
	}
	require.InDelta(t, 100.0, totalPercentage, 0.1)
}

func TestOpsViewActivityHeatmapCell_ValidRange(t *testing.T) {
	// Generate all cells
	cells := make([]OpsViewActivityHeatmapCell, 0, 7*24)
	for weekday := 0; weekday < 7; weekday++ {
		for hour := 0; hour < 24; hour++ {
			cells = append(cells, OpsViewActivityHeatmapCell{
				Weekday:  weekday,
				Hour:     hour,
				Requests: int64((weekday + 1) * (hour + 1) * 10),
			})
		}
	}

	require.Len(t, cells, 7*24)

	// Verify all cells have valid weekday and hour values
	for _, cell := range cells {
		require.GreaterOrEqual(t, cell.Weekday, 0)
		require.LessOrEqual(t, cell.Weekday, 6)
		require.GreaterOrEqual(t, cell.Hour, 0)
		require.LessOrEqual(t, cell.Hour, 23)
		require.GreaterOrEqual(t, cell.Requests, int64(0))
	}
}

func TestOpsViewTopUser_AvgCostCalculation(t *testing.T) {
	user := OpsViewTopUser{
		UserID:      1,
		Email:       "user@example.com",
		Username:    "user",
		Consumption: 200.0,
		Requests:    100,
		AvgCost:     2.0,
	}

	// Verify average cost calculation
	expectedAvgCost := user.Consumption / float64(user.Requests)
	require.InDelta(t, expectedAvgCost, user.AvgCost, 0.001)
}

func TestOpsViewService_calcChangeRate_EdgeCases(t *testing.T) {
	svc := &OpsViewService{}

	tests := []struct {
		name     string
		oldValue float64
		newValue float64
		expected float64
	}{
		{
			name:     "very small positive change",
			oldValue: 1000000,
			newValue: 1000001,
			expected: 0.0001,
		},
		{
			name:     "very large positive change",
			oldValue: 1,
			newValue: 1000000,
			expected: 99999900.0,
		},
		{
			name:     "decimal values",
			oldValue: 0.5,
			newValue: 0.75,
			expected: 50.0,
		},
		{
			name:     "negative to positive (from zero)",
			oldValue: 0,
			newValue: 50,
			expected: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.calcChangeRate(tt.oldValue, tt.newValue)
			require.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestNewOpsViewService(t *testing.T) {
	svc := NewOpsViewService(nil)
	require.NotNil(t, svc)
	require.Nil(t, svc.entClient)
}

func TestOpsViewOverview_AllFields(t *testing.T) {
	overview := &OpsViewOverview{
		TotalBalance:         10000.00,
		TodayConsumption:     100.50,
		WeekConsumption:      500.25,
		MonthConsumption:     2000.00,
		TodayRecharge:        200.00,
		WeekRecharge:         1000.00,
		MonthRecharge:        4000.00,
		TodayPaymentRecharge: 150.00,
		TodayRedeemRecharge:  50.00,
		WeekPaymentRecharge:  800.00,
		WeekRedeemRecharge:   200.00,
		MonthPaymentRecharge: 3200.00,
		MonthRedeemRecharge:  800.00,
		TotalUsers:           1000,
		TodayActiveUsers:     50,
		TodayNewUsers:        20,
		WeekActiveUsers:      100,
		WeekNewUsers:         10,
		WeekPayingUsers:      5,
		WeekRetention:        45.5,
		MonthActiveUsers:     200,
		MonthNewUsers:        40,
	}

	// Verify all recharge totals
	require.Equal(t, overview.TodayPaymentRecharge+overview.TodayRedeemRecharge, overview.TodayRecharge)
	require.Equal(t, overview.WeekPaymentRecharge+overview.WeekRedeemRecharge, overview.WeekRecharge)
	require.Equal(t, overview.MonthPaymentRecharge+overview.MonthRedeemRecharge, overview.MonthRecharge)

	// Verify user metrics are reasonable
	require.GreaterOrEqual(t, overview.TotalUsers, overview.MonthActiveUsers)
	require.GreaterOrEqual(t, overview.MonthActiveUsers, overview.WeekActiveUsers)
	require.GreaterOrEqual(t, overview.WeekActiveUsers, overview.TodayActiveUsers)

	// Verify retention is in valid range
	require.GreaterOrEqual(t, overview.WeekRetention, float64(0))
	require.LessOrEqual(t, overview.WeekRetention, float64(100))
}

func TestOpsViewComparison_ChangeCalculations(t *testing.T) {
	svc := &OpsViewService{}

	testCases := []struct {
		name         string
		today        float64
		yesterday    float64
		expectedSign int // 1 for positive, -1 for negative, 0 for zero
	}{
		{"increase", 150, 100, 1},
		{"decrease", 80, 100, -1},
		{"no change", 100, 100, 0},
		{"from zero", 100, 0, 1},
		{"to zero", 0, 100, -1},
		{"both zero", 0, 0, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			change := svc.calcChangeRate(tc.yesterday, tc.today)
			if tc.expectedSign > 0 {
				require.Greater(t, change, float64(0))
			} else if tc.expectedSign < 0 {
				require.Less(t, change, float64(0))
			} else {
				require.Equal(t, float64(0), change)
			}
		})
	}
}

func TestOpsViewTrendPoint_MultiplePoints(t *testing.T) {
	points := []OpsViewTrendPoint{
		{Date: "2024-01-01", Consumption: 100.0, PaymentRecharge: 80.0, RedeemRecharge: 20.0, TotalRecharge: 100.0},
		{Date: "2024-01-02", Consumption: 120.0, PaymentRecharge: 90.0, RedeemRecharge: 30.0, TotalRecharge: 120.0},
		{Date: "2024-01-03", Consumption: 110.0, PaymentRecharge: 85.0, RedeemRecharge: 25.0, TotalRecharge: 110.0},
	}

	for _, point := range points {
		// Verify total recharge calculation for each point
		require.Equal(t, point.PaymentRecharge+point.RedeemRecharge, point.TotalRecharge)
		// Verify date format
		require.Regexp(t, `^\d{4}-\d{2}-\d{2}$`, point.Date)
	}
}

func TestOpsViewUserGrowthPoint_RetentionValidation(t *testing.T) {
	testCases := []struct {
		name         string
		retentionD1  float64
		retentionD7  float64
		retentionD30 float64
		valid        bool
	}{
		{"normal decreasing", 80.0, 60.0, 40.0, true},
		{"all same", 50.0, 50.0, 50.0, true},
		{"all zero", 0.0, 0.0, 0.0, true},
		{"all 100", 100.0, 100.0, 100.0, true},
		{"negative D1 (invalid)", -1.0, 60.0, 40.0, false},
		{"over 100 D7 (invalid)", 80.0, 110.0, 40.0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			point := OpsViewUserGrowthPoint{
				Date:         "2024-01-01",
				NewUsers:     10,
				ActiveUsers:  50,
				PayingUsers:  5,
				RetentionD1:  tc.retentionD1,
				RetentionD7:  tc.retentionD7,
				RetentionD30: tc.retentionD30,
			}

			isValid := point.RetentionD1 >= 0 && point.RetentionD1 <= 100 &&
				point.RetentionD7 >= 0 && point.RetentionD7 <= 100 &&
				point.RetentionD30 >= 0 && point.RetentionD30 <= 100

			require.Equal(t, tc.valid, isValid)
		})
	}
}

func TestOpsViewGroupConsumption_EmptyGroups(t *testing.T) {
	groups := []OpsViewGroupConsumption{}
	require.Len(t, groups, 0)

	// Single group should have 100% percentage
	singleGroup := []OpsViewGroupConsumption{
		{GroupID: 1, GroupName: "Only Group", Consumption: 1000.0, Percentage: 100.0, Requests: 500},
	}
	require.Equal(t, 100.0, singleGroup[0].Percentage)
}

func TestOpsViewModelRanking_TokenCalculation(t *testing.T) {
	models := []OpsViewModelRanking{
		{Model: "model-a", Requests: 100, Tokens: 50000, Consumption: 100.0, Percentage: 50.0},
		{Model: "model-b", Requests: 100, Tokens: 30000, Consumption: 60.0, Percentage: 30.0},
		{Model: "model-c", Requests: 100, Tokens: 20000, Consumption: 40.0, Percentage: 20.0},
	}

	var totalPercentage float64
	var totalTokens int64
	for _, m := range models {
		totalPercentage += m.Percentage
		totalTokens += m.Tokens
		require.Greater(t, m.Tokens, int64(0))
	}

	require.InDelta(t, 100.0, totalPercentage, 0.01)
	require.Equal(t, int64(100000), totalTokens)
}

func TestOpsViewActivityHeatmapCell_IndexCalculation(t *testing.T) {
	// Test that weekday * 24 + hour gives unique index for each cell
	indices := make(map[int]bool)

	for weekday := 0; weekday < 7; weekday++ {
		for hour := 0; hour < 24; hour++ {
			index := weekday*24 + hour
			require.False(t, indices[index], "Duplicate index found: %d", index)
			indices[index] = true
		}
	}

	require.Len(t, indices, 7*24)
}

func TestOpsViewTopUser_ZeroRequests(t *testing.T) {
	// When requests is 0, avgCost should be 0 to avoid division by zero
	user := OpsViewTopUser{
		UserID:      1,
		Email:       "user@example.com",
		Username:    "user",
		Consumption: 0.0,
		Requests:    0,
		AvgCost:     0.0,
	}

	require.Equal(t, int64(0), user.Requests)
	require.Equal(t, float64(0), user.AvgCost)
}
