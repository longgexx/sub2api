// Package model provides shared data models that can be used across packages
// without creating circular dependencies.
package model

import "fmt"

// ScheduleRule 定义时间段调度规则
// weekdays 表示规则的"开始日"，跨天规则自动延续到次日
// 区间语义为 [start_minute, end_minute)，左闭右开
type ScheduleRule struct {
	Weekdays    []int `json:"weekdays"`     // 0=Sunday, 1=Monday, ..., 6=Saturday (规则开始日)
	StartMinute int   `json:"start_minute"` // 0-1439 (e.g., 9:00 = 540)
	EndMinute   int   `json:"end_minute"`   // 0-1439 (e.g., 18:00 = 1080)
}

// IsCrossDay 判断是否为跨天规则
func (r ScheduleRule) IsCrossDay() bool {
	return r.EndMinute <= r.StartMinute
}

// MinuteToTimeStr 将分钟数转换为 "HH:MM" 格式（用于 API 输出显示）
func MinuteToTimeStr(minute int) string {
	return fmt.Sprintf("%02d:%02d", minute/60, minute%60)
}
