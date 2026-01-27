package service

import (
	"testing"
	"time"
)

func mustLoadLocation(tz string) *time.Location {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		panic(err)
	}
	return loc
}

func TestCheckScheduleWindow(t *testing.T) {
	tests := []struct {
		name           string
		account        *Account
		mockNow        time.Time
		expectedResult bool
		expectedReason string
	}{
		{
			name: "不启用调度应返回可用",
			account: &Account{
				ScheduleEnabled: false,
			},
			mockNow:        time.Now(),
			expectedResult: true,
		},
		{
			name: "启用调度但无规则应返回不可用",
			account: &Account{
				ScheduleEnabled:  true,
				ScheduleTimezone: "UTC",
				ScheduleRules:    []ScheduleRule{},
			},
			mockNow:        time.Now(),
			expectedResult: false,
			expectedReason: "invalid_config_no_rules",
		},
		{
			name: "无效时区应返回不可用",
			account: &Account{
				ScheduleEnabled:  true,
				ScheduleTimezone: "Invalid/Timezone",
				ScheduleRules: []ScheduleRule{
					{Weekdays: []int{1}, StartMinute: 540, EndMinute: 1080},
				},
			},
			mockNow:        time.Now(),
			expectedResult: false,
			expectedReason: "invalid_timezone",
		},
		{
			name: "工作日 09:00-18:00，周一 10:00 应可用",
			account: &Account{
				ScheduleEnabled:  true,
				ScheduleTimezone: "Asia/Shanghai",
				ScheduleRules: []ScheduleRule{
					{Weekdays: []int{1, 2, 3, 4, 5}, StartMinute: 540, EndMinute: 1080},
				},
			},
			// 2024-01-08 是周一
			mockNow:        time.Date(2024, 1, 8, 10, 0, 0, 0, mustLoadLocation("Asia/Shanghai")),
			expectedResult: true,
		},
		{
			name: "工作日 09:00-18:00，周一 18:00 整应不可用（左闭右开）",
			account: &Account{
				ScheduleEnabled:  true,
				ScheduleTimezone: "Asia/Shanghai",
				ScheduleRules: []ScheduleRule{
					{Weekdays: []int{1, 2, 3, 4, 5}, StartMinute: 540, EndMinute: 1080},
				},
			},
			mockNow:        time.Date(2024, 1, 8, 18, 0, 0, 0, mustLoadLocation("Asia/Shanghai")),
			expectedResult: false,
			expectedReason: "outside_schedule",
		},
		{
			name: "工作日 09:00-18:00，周六应不可用",
			account: &Account{
				ScheduleEnabled:  true,
				ScheduleTimezone: "Asia/Shanghai",
				ScheduleRules: []ScheduleRule{
					{Weekdays: []int{1, 2, 3, 4, 5}, StartMinute: 540, EndMinute: 1080},
				},
			},
			// 2024-01-13 是周六
			mockNow:        time.Date(2024, 1, 13, 10, 0, 0, 0, mustLoadLocation("Asia/Shanghai")),
			expectedResult: false,
			expectedReason: "outside_schedule",
		},
		{
			name: "跨天规则 22:00-02:00 周五，周五 23:00 应可用",
			account: &Account{
				ScheduleEnabled:  true,
				ScheduleTimezone: "UTC",
				ScheduleRules: []ScheduleRule{
					{Weekdays: []int{5}, StartMinute: 1320, EndMinute: 120},
				},
			},
			// 2024-01-12 是周五
			mockNow:        time.Date(2024, 1, 12, 23, 0, 0, 0, time.UTC),
			expectedResult: true,
		},
		{
			name: "跨天规则 22:00-02:00 周五，周六 01:00 应可用（延续到次日）",
			account: &Account{
				ScheduleEnabled:  true,
				ScheduleTimezone: "UTC",
				ScheduleRules: []ScheduleRule{
					{Weekdays: []int{5}, StartMinute: 1320, EndMinute: 120},
				},
			},
			// 2024-01-13 是周六凌晨，属于周五的跨天规则延续
			mockNow:        time.Date(2024, 1, 13, 1, 0, 0, 0, time.UTC),
			expectedResult: true,
		},
		{
			name: "跨天规则 22:00-02:00 周五，周六 02:00 整应不可用（左闭右开）",
			account: &Account{
				ScheduleEnabled:  true,
				ScheduleTimezone: "UTC",
				ScheduleRules: []ScheduleRule{
					{Weekdays: []int{5}, StartMinute: 1320, EndMinute: 120},
				},
			},
			mockNow:        time.Date(2024, 1, 13, 2, 0, 0, 0, time.UTC),
			expectedResult: false,
			expectedReason: "outside_schedule",
		},
		{
			name: "多规则：工作日 + 周六不同时间",
			account: &Account{
				ScheduleEnabled:  true,
				ScheduleTimezone: "UTC",
				ScheduleRules: []ScheduleRule{
					{Weekdays: []int{1, 2, 3, 4, 5}, StartMinute: 540, EndMinute: 1080}, // 周一至周五 09:00-18:00
					{Weekdays: []int{6}, StartMinute: 600, EndMinute: 840},              // 周六 10:00-14:00
				},
			},
			// 周六 11:00
			mockNow:        time.Date(2024, 1, 13, 11, 0, 0, 0, time.UTC),
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 注入模拟时间
			result := tt.account.checkScheduleWindowWithTime(tt.mockNow)

			if result.IsWithinWindow != tt.expectedResult {
				t.Errorf("IsWithinWindow = %v, want %v", result.IsWithinWindow, tt.expectedResult)
			}
			if tt.expectedReason != "" && result.Reason != tt.expectedReason {
				t.Errorf("Reason = %v, want %v", result.Reason, tt.expectedReason)
			}
		})
	}
}

func TestIsCrossDay(t *testing.T) {
	tests := []struct {
		rule     ScheduleRule
		expected bool
	}{
		{ScheduleRule{StartMinute: 540, EndMinute: 1080}, false},  // 09:00-18:00
		{ScheduleRule{StartMinute: 1320, EndMinute: 120}, true},   // 22:00-02:00
		{ScheduleRule{StartMinute: 0, EndMinute: 0}, true},        // 00:00-00:00 (跨天)
		{ScheduleRule{StartMinute: 1439, EndMinute: 60}, true},    // 23:59-01:00
		{ScheduleRule{StartMinute: 60, EndMinute: 1439}, false},   // 01:00-23:59
	}

	for _, tt := range tests {
		result := tt.rule.IsCrossDay()
		if result != tt.expected {
			t.Errorf("IsCrossDay(%d-%d) = %v, want %v",
				tt.rule.StartMinute, tt.rule.EndMinute, result, tt.expected)
		}
	}
}

func TestParseTimeToMinute(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"00:00", 0, false},
		{"09:00", 540, false},
		{"18:00", 1080, false},
		{"23:59", 1439, false},
		{"12:30", 750, false},
		{"invalid", 0, true},
		{"25:00", 0, true},
		{"9:00", 540, false}, // 单数字小时也应该解析成功
	}

	for _, tt := range tests {
		result, err := ParseTimeToMinute(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("ParseTimeToMinute(%s) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseTimeToMinute(%s) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("ParseTimeToMinute(%s) = %d, want %d", tt.input, result, tt.expected)
			}
		}
	}
}

func TestMinuteToTimeStr(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "00:00"},
		{540, "09:00"},
		{1080, "18:00"},
		{1439, "23:59"},
		{750, "12:30"},
	}

	for _, tt := range tests {
		result := MinuteToTimeStr(tt.input)
		if result != tt.expected {
			t.Errorf("MinuteToTimeStr(%d) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}
