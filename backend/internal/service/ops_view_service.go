package service

import (
	"context"
	"sort"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/ent/redeemcode"
	"github.com/Wei-Shaw/sub2api/ent/usagelog"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

// OpsViewService 运营视图服务
type OpsViewService struct {
	entClient *dbent.Client
}

// NewOpsViewService 创建运营视图服务
func NewOpsViewService(entClient *dbent.Client) *OpsViewService {
	return &OpsViewService{
		entClient: entClient,
	}
}

// GetOverview 获取运营概览数据
func (s *OpsViewService) GetOverview(ctx context.Context, tz string) (*OpsViewOverview, error) {
	now := timezone.NowInUserLocation(tz)
	todayStart := timezone.StartOfDayInUserLocation(now, tz)
	weekStart := timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -int(now.Weekday())), tz)
	monthStart := timezone.StartOfDayInUserLocation(time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), tz)
	tomorrow := todayStart.Add(24 * time.Hour)

	overview := &OpsViewOverview{}

	// 总余额
	overview.TotalBalance, _ = s.getTotalBalance(ctx)

	// 消耗金额统计
	overview.TodayConsumption, _ = s.getConsumption(ctx, todayStart, tomorrow)
	overview.WeekConsumption, _ = s.getConsumption(ctx, weekStart, tomorrow)
	overview.MonthConsumption, _ = s.getConsumption(ctx, monthStart, tomorrow)

	// 充值金额统计
	todayPayment, _ := s.getPaymentRecharge(ctx, todayStart, tomorrow)
	todayRedeem, _ := s.getRedeemRecharge(ctx, todayStart, tomorrow)
	weekPayment, _ := s.getPaymentRecharge(ctx, weekStart, tomorrow)
	weekRedeem, _ := s.getRedeemRecharge(ctx, weekStart, tomorrow)
	monthPayment, _ := s.getPaymentRecharge(ctx, monthStart, tomorrow)
	monthRedeem, _ := s.getRedeemRecharge(ctx, monthStart, tomorrow)

	overview.TodayPaymentRecharge = todayPayment
	overview.TodayRedeemRecharge = todayRedeem
	overview.TodayRecharge = todayPayment + todayRedeem
	overview.WeekPaymentRecharge = weekPayment
	overview.WeekRedeemRecharge = weekRedeem
	overview.WeekRecharge = weekPayment + weekRedeem
	overview.MonthPaymentRecharge = monthPayment
	overview.MonthRedeemRecharge = monthRedeem
	overview.MonthRecharge = monthPayment + monthRedeem

	// 用户指标
	overview.TotalUsers, _ = s.getTotalUsers(ctx)
	overview.TodayActiveUsers, _ = s.getActiveUsers(ctx, todayStart, tomorrow)
	overview.TodayNewUsers, _ = s.getNewUsers(ctx, todayStart, tomorrow)
	overview.WeekActiveUsers, _ = s.getActiveUsers(ctx, weekStart, tomorrow)
	overview.WeekNewUsers, _ = s.getNewUsers(ctx, weekStart, tomorrow)
	overview.WeekPayingUsers, _ = s.getPayingUsers(ctx, weekStart, tomorrow)
	overview.MonthActiveUsers, _ = s.getActiveUsers(ctx, monthStart, tomorrow)
	overview.MonthNewUsers, _ = s.getNewUsers(ctx, monthStart, tomorrow)

	// 周留存率
	overview.WeekRetention, _ = s.getRetentionRate(ctx, weekStart, 7, tz)

	return overview, nil
}

// GetComparison 获取今日 vs 昨日对比数据
func (s *OpsViewService) GetComparison(ctx context.Context, tz string) (*OpsViewComparison, error) {
	now := timezone.NowInUserLocation(tz)
	todayStart := timezone.StartOfDayInUserLocation(now, tz)
	yesterdayStart := todayStart.Add(-24 * time.Hour)
	tomorrow := todayStart.Add(24 * time.Hour)

	comparison := &OpsViewComparison{}

	// 消耗对比
	comparison.TodayConsumption, _ = s.getConsumption(ctx, todayStart, tomorrow)
	comparison.YesterdayConsumption, _ = s.getConsumption(ctx, yesterdayStart, todayStart)
	comparison.ConsumptionChange = s.calcChangeRate(comparison.YesterdayConsumption, comparison.TodayConsumption)

	// 充值对比
	todayPayment, _ := s.getPaymentRecharge(ctx, todayStart, tomorrow)
	todayRedeem, _ := s.getRedeemRecharge(ctx, todayStart, tomorrow)
	yesterdayPayment, _ := s.getPaymentRecharge(ctx, yesterdayStart, todayStart)
	yesterdayRedeem, _ := s.getRedeemRecharge(ctx, yesterdayStart, todayStart)

	comparison.TodayRecharge = todayPayment + todayRedeem
	comparison.YesterdayRecharge = yesterdayPayment + yesterdayRedeem
	comparison.RechargeChange = s.calcChangeRate(comparison.YesterdayRecharge, comparison.TodayRecharge)

	// 活跃用户对比
	comparison.TodayActiveUsers, _ = s.getActiveUsers(ctx, todayStart, tomorrow)
	comparison.YesterdayActiveUsers, _ = s.getActiveUsers(ctx, yesterdayStart, todayStart)
	comparison.ActiveUsersChange = s.calcChangeRate(float64(comparison.YesterdayActiveUsers), float64(comparison.TodayActiveUsers))

	// 新增用户对比
	comparison.TodayNewUsers, _ = s.getNewUsers(ctx, todayStart, tomorrow)
	comparison.YesterdayNewUsers, _ = s.getNewUsers(ctx, yesterdayStart, todayStart)
	comparison.NewUsersChange = s.calcChangeRate(float64(comparison.YesterdayNewUsers), float64(comparison.TodayNewUsers))

	return comparison, nil
}

// GetTrend 获取趋势数据
func (s *OpsViewService) GetTrend(ctx context.Context, days int, tz string) ([]OpsViewTrendPoint, error) {
	now := timezone.NowInUserLocation(tz)
	todayStart := timezone.StartOfDayInUserLocation(now, tz)

	points := make([]OpsViewTrendPoint, 0, days)

	for i := days - 1; i >= 0; i-- {
		dayStart := todayStart.Add(-time.Duration(i) * 24 * time.Hour)
		dayEnd := dayStart.Add(24 * time.Hour)

		point := OpsViewTrendPoint{
			Date: dayStart.Format("2006-01-02"),
		}

		point.Consumption, _ = s.getConsumption(ctx, dayStart, dayEnd)
		point.PaymentRecharge, _ = s.getPaymentRecharge(ctx, dayStart, dayEnd)
		point.RedeemRecharge, _ = s.getRedeemRecharge(ctx, dayStart, dayEnd)
		point.TotalRecharge = point.PaymentRecharge + point.RedeemRecharge

		points = append(points, point)
	}

	return points, nil
}

// GetUserGrowth 获取用户增长数据
func (s *OpsViewService) GetUserGrowth(ctx context.Context, days int, tz string) ([]OpsViewUserGrowthPoint, error) {
	now := timezone.NowInUserLocation(tz)
	todayStart := timezone.StartOfDayInUserLocation(now, tz)

	points := make([]OpsViewUserGrowthPoint, 0, days)

	for i := days - 1; i >= 0; i-- {
		dayStart := todayStart.Add(-time.Duration(i) * 24 * time.Hour)
		dayEnd := dayStart.Add(24 * time.Hour)

		point := OpsViewUserGrowthPoint{
			Date: dayStart.Format("2006-01-02"),
		}

		point.NewUsers, _ = s.getNewUsers(ctx, dayStart, dayEnd)
		point.ActiveUsers, _ = s.getActiveUsers(ctx, dayStart, dayEnd)
		point.PayingUsers, _ = s.getPayingUsers(ctx, dayStart, dayEnd)

		// 计算留存率
		point.RetentionD1, _ = s.getRetentionRate(ctx, dayStart, 1, tz)
		point.RetentionD7, _ = s.getRetentionRate(ctx, dayStart, 7, tz)
		point.RetentionD30, _ = s.getRetentionRate(ctx, dayStart, 30, tz)

		points = append(points, point)
	}

	return points, nil
}

// GetGroupConsumption 获取分组消耗分布
func (s *OpsViewService) GetGroupConsumption(ctx context.Context, start, end time.Time) ([]OpsViewGroupConsumption, error) {
	// 查询分组消耗数据
	// ent Aggregate 生成的列名是 sum_actual_cost 和 count
	type groupStat struct {
		GroupID       int64   `json:"group_id"`
		SumActualCost float64 `json:"sum_actual_cost"`
		Count         int64   `json:"count"`
	}

	var stats []groupStat
	err := s.entClient.UsageLog.Query().
		Where(
			usagelog.CreatedAtGTE(start),
			usagelog.CreatedAtLT(end),
			usagelog.GroupIDNotNil(),
		).
		GroupBy(usagelog.FieldGroupID).
		Aggregate(
			dbent.As(dbent.Sum(usagelog.FieldActualCost), "sum_actual_cost"),
			dbent.As(dbent.Count(), "count"),
		).
		Scan(ctx, &stats)
	if err != nil {
		return nil, err
	}

	// 计算总消耗
	var totalConsumption float64
	for _, stat := range stats {
		totalConsumption += stat.SumActualCost
	}

	// 批量获取分组名称
	groupIDs := make([]int64, 0, len(stats))
	for _, stat := range stats {
		if stat.GroupID > 0 {
			groupIDs = append(groupIDs, stat.GroupID)
		}
	}
	groupMap := make(map[int64]string)
	if len(groupIDs) > 0 {
		groups, _ := s.entClient.Group.Query().
			Where(group.IDIn(groupIDs...)).
			All(ctx)
		for _, g := range groups {
			groupMap[g.ID] = g.Name
		}
	}

	// 构建结果
	result := make([]OpsViewGroupConsumption, 0, len(stats))
	for _, stat := range stats {
		groupName := "未知分组"
		if name, ok := groupMap[stat.GroupID]; ok {
			groupName = name
		}

		percentage := float64(0)
		if totalConsumption > 0 {
			percentage = (stat.SumActualCost / totalConsumption) * 100
		}

		result = append(result, OpsViewGroupConsumption{
			GroupID:     stat.GroupID,
			GroupName:   groupName,
			Consumption: stat.SumActualCost,
			Percentage:  percentage,
			Requests:    stat.Count,
		})
	}

	return result, nil
}

// GetTopUsers 获取高价值用户排行
func (s *OpsViewService) GetTopUsers(ctx context.Context, start, end time.Time, limit int) ([]OpsViewTopUser, error) {
	// 查询用户消耗统计
	// ent Aggregate 生成的列名是 sum_actual_cost 和 count
	type userStat struct {
		UserID        int64   `json:"user_id"`
		SumActualCost float64 `json:"sum_actual_cost"`
		Count         int64   `json:"count"`
	}

	var stats []userStat
	err := s.entClient.UsageLog.Query().
		Where(
			usagelog.CreatedAtGTE(start),
			usagelog.CreatedAtLT(end),
		).
		GroupBy(usagelog.FieldUserID).
		Aggregate(
			dbent.As(dbent.Sum(usagelog.FieldActualCost), "sum_actual_cost"),
			dbent.As(dbent.Count(), "count"),
		).
		Scan(ctx, &stats)
	if err != nil {
		return nil, err
	}

	// 按消耗排序并取前N个 (使用 sort.Slice 替代冒泡排序)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].SumActualCost > stats[j].SumActualCost
	})

	if len(stats) > limit {
		stats = stats[:limit]
	}

	// 批量获取用户详情
	userIDs := make([]int64, len(stats))
	for i, stat := range stats {
		userIDs[i] = stat.UserID
	}

	users, _ := s.entClient.User.Query().
		Where(user.IDIn(userIDs...)).
		All(ctx)
	userMap := make(map[int64]*dbent.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// 批量获取最后活跃时间
	type lastActiveInfo struct {
		UserID    int64     `json:"user_id"`
		CreatedAt time.Time `json:"created_at"`
	}
	var lastActiveLogs []lastActiveInfo
	// 使用子查询获取每个用户的最后活跃时间
	for _, userID := range userIDs {
		lastLog, _ := s.entClient.UsageLog.Query().
			Where(usagelog.UserIDEQ(userID)).
			Order(dbent.Desc(usagelog.FieldCreatedAt)).
			First(ctx)
		if lastLog != nil {
			lastActiveLogs = append(lastActiveLogs, lastActiveInfo{
				UserID:    userID,
				CreatedAt: lastLog.CreatedAt,
			})
		}
	}
	lastActiveMap := make(map[int64]time.Time)
	for _, info := range lastActiveLogs {
		lastActiveMap[info.UserID] = info.CreatedAt
	}

	// 构建结果
	result := make([]OpsViewTopUser, 0, len(stats))
	for _, stat := range stats {
		u, ok := userMap[stat.UserID]
		if !ok {
			continue
		}

		avgCost := float64(0)
		if stat.Count > 0 {
			avgCost = stat.SumActualCost / float64(stat.Count)
		}

		lastActiveAt := u.CreatedAt
		if t, ok := lastActiveMap[stat.UserID]; ok {
			lastActiveAt = t
		}

		result = append(result, OpsViewTopUser{
			UserID:       stat.UserID,
			Email:        u.Email,
			Username:     u.Username,
			Consumption:  stat.SumActualCost,
			Requests:     stat.Count,
			AvgCost:      avgCost,
			LastActiveAt: lastActiveAt,
			RegisteredAt: u.CreatedAt,
		})
	}

	return result, nil
}

// GetModelRanking 获取模型使用排行
func (s *OpsViewService) GetModelRanking(ctx context.Context, start, end time.Time, limit int) ([]OpsViewModelRanking, error) {
	// 查询模型使用统计
	// ent Aggregate 生成的列名是 count, sum_input_tokens, sum_output_tokens, sum_actual_cost
	type modelStat struct {
		Model           string  `json:"model"`
		Count           int64   `json:"count"`
		SumInputTokens  int64   `json:"sum_input_tokens"`
		SumOutputTokens int64   `json:"sum_output_tokens"`
		SumActualCost   float64 `json:"sum_actual_cost"`
	}

	var stats []modelStat
	err := s.entClient.UsageLog.Query().
		Where(
			usagelog.CreatedAtGTE(start),
			usagelog.CreatedAtLT(end),
		).
		GroupBy(usagelog.FieldModel).
		Aggregate(
			dbent.As(dbent.Count(), "count"),
			dbent.As(dbent.Sum(usagelog.FieldInputTokens), "sum_input_tokens"),
			dbent.As(dbent.Sum(usagelog.FieldOutputTokens), "sum_output_tokens"),
			dbent.As(dbent.Sum(usagelog.FieldActualCost), "sum_actual_cost"),
		).
		Scan(ctx, &stats)
	if err != nil {
		return nil, err
	}

	// 计算总消耗
	var totalConsumption float64
	for _, stat := range stats {
		totalConsumption += stat.SumActualCost
	}

	// 按消耗排序 (使用 sort.Slice 替代冒泡排序)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].SumActualCost > stats[j].SumActualCost
	})

	if len(stats) > limit {
		stats = stats[:limit]
	}

	// 构建结果
	result := make([]OpsViewModelRanking, 0, len(stats))
	for _, stat := range stats {
		percentage := float64(0)
		if totalConsumption > 0 {
			percentage = (stat.SumActualCost / totalConsumption) * 100
		}

		result = append(result, OpsViewModelRanking{
			Model:       stat.Model,
			Requests:    stat.Count,
			Tokens:      stat.SumInputTokens + stat.SumOutputTokens,
			Consumption: stat.SumActualCost,
			Percentage:  percentage,
		})
	}

	return result, nil
}

// GetActivityHeatmap 获取活跃时段热力图
func (s *OpsViewService) GetActivityHeatmap(ctx context.Context, start, end time.Time, tz string) ([]OpsViewActivityHeatmapCell, error) {
	// 初始化热力图数据 (7天 x 24小时)
	heatmap := make([]OpsViewActivityHeatmapCell, 0, 7*24)
	for weekday := 0; weekday < 7; weekday++ {
		for hour := 0; hour < 24; hour++ {
			heatmap = append(heatmap, OpsViewActivityHeatmapCell{
				Weekday:  weekday,
				Hour:     hour,
				Requests: 0,
			})
		}
	}

	// 查询使用日志
	logs, err := s.entClient.UsageLog.Query().
		Where(
			usagelog.CreatedAtGTE(start),
			usagelog.CreatedAtLT(end),
		).
		Select(usagelog.FieldCreatedAt).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 统计每个时段的请求数
	loc, _ := time.LoadLocation(tz)
	if loc == nil {
		loc = time.UTC
	}

	for _, log := range logs {
		localTime := log.CreatedAt.In(loc)
		weekday := int(localTime.Weekday())
		hour := localTime.Hour()
		index := weekday*24 + hour
		if index < len(heatmap) {
			heatmap[index].Requests++
		}
	}

	return heatmap, nil
}

// getTotalBalance 获取所有用户余额总和
func (s *OpsViewService) getTotalBalance(ctx context.Context) (float64, error) {
	var result []struct {
		Sum float64 `json:"sum"`
	}

	err := s.entClient.User.Query().
		Where(user.DeletedAtIsNil()).
		Aggregate(dbent.Sum(user.FieldBalance)).
		Scan(ctx, &result)
	if err != nil {
		return 0, err
	}

	if len(result) > 0 {
		return result[0].Sum, nil
	}
	return 0, nil
}

// getTotalUsers 获取总用户数
func (s *OpsViewService) getTotalUsers(ctx context.Context) (int64, error) {
	count, err := s.entClient.User.Query().
		Where(user.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}

// getConsumption 获取消耗金额
func (s *OpsViewService) getConsumption(ctx context.Context, start, end time.Time) (float64, error) {
	var result []struct {
		Sum float64 `json:"sum"`
	}

	err := s.entClient.UsageLog.Query().
		Where(
			usagelog.CreatedAtGTE(start),
			usagelog.CreatedAtLT(end),
		).
		Aggregate(dbent.Sum(usagelog.FieldActualCost)).
		Scan(ctx, &result)
	if err != nil {
		return 0, err
	}

	if len(result) > 0 {
		return result[0].Sum, nil
	}
	return 0, nil
}

// getPaymentRecharge 获取平台充值金额
func (s *OpsViewService) getPaymentRecharge(ctx context.Context, start, end time.Time) (float64, error) {
	var result []struct {
		Sum float64 `json:"sum"`
	}

	err := s.entClient.PaymentOrder.Query().
		Where(
			paymentorder.StatusEQ("paid"),
			paymentorder.PaidAtGTE(start),
			paymentorder.PaidAtLT(end),
		).
		Aggregate(dbent.Sum(paymentorder.FieldCreditAmount)).
		Scan(ctx, &result)
	if err != nil {
		return 0, err
	}

	if len(result) > 0 {
		return result[0].Sum, nil
	}
	return 0, nil
}

// getRedeemRecharge 获取兑换充值金额
func (s *OpsViewService) getRedeemRecharge(ctx context.Context, start, end time.Time) (float64, error) {
	var result []struct {
		Sum float64 `json:"sum"`
	}

	err := s.entClient.RedeemCode.Query().
		Where(
			redeemcode.StatusEQ("used"),
			redeemcode.UsedAtGTE(start),
			redeemcode.UsedAtLT(end),
		).
		Aggregate(dbent.Sum(redeemcode.FieldActualValue)).
		Scan(ctx, &result)
	if err != nil {
		return 0, err
	}

	if len(result) > 0 {
		return result[0].Sum, nil
	}
	return 0, nil
}

// getActiveUsers 获取活跃用户数
func (s *OpsViewService) getActiveUsers(ctx context.Context, start, end time.Time) (int64, error) {
	count, err := s.entClient.UsageLog.Query().
		Where(
			usagelog.CreatedAtGTE(start),
			usagelog.CreatedAtLT(end),
		).
		Unique(true).
		Select(usagelog.FieldUserID).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}

// getNewUsers 获取新增用户数
func (s *OpsViewService) getNewUsers(ctx context.Context, start, end time.Time) (int64, error) {
	count, err := s.entClient.User.Query().
		Where(
			user.CreatedAtGTE(start),
			user.CreatedAtLT(end),
			user.DeletedAtIsNil(),
		).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}

// getPayingUsers 获取付费用户数
func (s *OpsViewService) getPayingUsers(ctx context.Context, start, end time.Time) (int64, error) {
	// 统计在时间范围内有付费行为的用户数
	count, err := s.entClient.PaymentOrder.Query().
		Where(
			paymentorder.StatusEQ("paid"),
			paymentorder.PaidAtGTE(start),
			paymentorder.PaidAtLT(end),
		).
		Unique(true).
		Select(paymentorder.FieldUserID).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}

// getRetentionRate 计算留存率
func (s *OpsViewService) getRetentionRate(ctx context.Context, cohortDate time.Time, days int, tz string) (float64, error) {
	cohortEnd := cohortDate.Add(24 * time.Hour)

	// 计算 N 天后的留存日期
	retentionStart := cohortDate.Add(time.Duration(days) * 24 * time.Hour)
	retentionEnd := retentionStart.Add(24 * time.Hour)

	// 检查留存日期是否超过当前时间，如果超过则返回 -1 表示数据不可用
	now := timezone.NowInUserLocation(tz)
	if retentionStart.After(now) {
		return -1, nil // -1 表示数据尚不可用（未来日期）
	}

	// 获取队列用户（在 cohortDate 当天注册的用户）
	cohortUsers, err := s.entClient.User.Query().
		Where(
			user.CreatedAtGTE(cohortDate),
			user.CreatedAtLT(cohortEnd),
			user.DeletedAtIsNil(),
		).
		Select(user.FieldID).
		All(ctx)
	if err != nil {
		return 0, err
	}

	if len(cohortUsers) == 0 {
		return 0, nil
	}

	userIDs := make([]int64, len(cohortUsers))
	for i, u := range cohortUsers {
		userIDs[i] = u.ID
	}

	// 统计在留存日期内活跃的队列用户数
	activeCount, err := s.entClient.UsageLog.Query().
		Where(
			usagelog.UserIDIn(userIDs...),
			usagelog.CreatedAtGTE(retentionStart),
			usagelog.CreatedAtLT(retentionEnd),
		).
		Unique(true).
		Select(usagelog.FieldUserID).
		Count(ctx)
	if err != nil {
		return 0, err
	}

	return float64(activeCount) / float64(len(cohortUsers)) * 100, nil
}

// calcChangeRate 计算变化率
func (s *OpsViewService) calcChangeRate(oldValue, newValue float64) float64 {
	if oldValue == 0 {
		if newValue == 0 {
			return 0
		}
		return 100 // 从0增长视为100%增长
	}
	return ((newValue - oldValue) / oldValue) * 100
}
