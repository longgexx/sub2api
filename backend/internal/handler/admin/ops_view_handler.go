package admin

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// OpsViewHandler 运营视图处理器
type OpsViewHandler struct {
	opsViewService *service.OpsViewService
}

// NewOpsViewHandler 创建运营视图处理器
func NewOpsViewHandler(opsViewService *service.OpsViewService) *OpsViewHandler {
	return &OpsViewHandler{
		opsViewService: opsViewService,
	}
}

// GetOverview 获取运营概览数据
// GET /api/v1/admin/ops-view/overview
func (h *OpsViewHandler) GetOverview(c *gin.Context) {
	tz := c.DefaultQuery("timezone", "UTC")

	overview, err := h.opsViewService.GetOverview(c.Request.Context(), tz)
	if err != nil {
		response.InternalError(c, "Failed to get overview: "+err.Error())
		return
	}

	response.Success(c, overview)
}

// GetComparison 获取今日 vs 昨日对比数据
// GET /api/v1/admin/ops-view/comparison
func (h *OpsViewHandler) GetComparison(c *gin.Context) {
	tz := c.DefaultQuery("timezone", "UTC")

	comparison, err := h.opsViewService.GetComparison(c.Request.Context(), tz)
	if err != nil {
		response.InternalError(c, "Failed to get comparison: "+err.Error())
		return
	}

	response.Success(c, comparison)
}

// GetTrend 获取趋势数据
// GET /api/v1/admin/ops-view/trend
// Query params: days (default 30), timezone
func (h *OpsViewHandler) GetTrend(c *gin.Context) {
	tz := c.DefaultQuery("timezone", "UTC")
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 365 {
		days = 30
	}

	trend, err := h.opsViewService.GetTrend(c.Request.Context(), days, tz)
	if err != nil {
		response.InternalError(c, "Failed to get trend: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"trend": trend,
		"days":  days,
	})
}

// GetUserGrowth 获取用户增长数据
// GET /api/v1/admin/ops-view/users/growth
// Query params: days (default 30), timezone
func (h *OpsViewHandler) GetUserGrowth(c *gin.Context) {
	tz := c.DefaultQuery("timezone", "UTC")
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 365 {
		days = 30
	}

	growth, err := h.opsViewService.GetUserGrowth(c.Request.Context(), days, tz)
	if err != nil {
		response.InternalError(c, "Failed to get user growth: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"growth": growth,
		"days":   days,
	})
}

// GetGroupConsumption 获取分组消耗分布
// GET /api/v1/admin/ops-view/groups/consumption
// Query params: start_date, end_date, timezone
func (h *OpsViewHandler) GetGroupConsumption(c *gin.Context) {
	startTime, endTime := parseOpsViewTimeRange(c)

	consumption, err := h.opsViewService.GetGroupConsumption(c.Request.Context(), startTime, endTime)
	if err != nil {
		response.InternalError(c, "Failed to get group consumption: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"groups":     consumption,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.Add(-24 * time.Hour).Format("2006-01-02"),
	})
}

// GetTopUsers 获取高价值用户排行
// GET /api/v1/admin/ops-view/users/top
// Query params: start_date, end_date, limit (default 20), timezone
func (h *OpsViewHandler) GetTopUsers(c *gin.Context) {
	startTime, endTime := parseOpsViewTimeRange(c)
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	users, err := h.opsViewService.GetTopUsers(c.Request.Context(), startTime, endTime, limit)
	if err != nil {
		response.InternalError(c, "Failed to get top users: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"users":      users,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.Add(-24 * time.Hour).Format("2006-01-02"),
		"limit":      limit,
	})
}

// GetModelRanking 获取模型使用排行
// GET /api/v1/admin/ops-view/models/ranking
// Query params: start_date, end_date, limit (default 20), timezone
func (h *OpsViewHandler) GetModelRanking(c *gin.Context) {
	startTime, endTime := parseOpsViewTimeRange(c)
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	ranking, err := h.opsViewService.GetModelRanking(c.Request.Context(), startTime, endTime, limit)
	if err != nil {
		response.InternalError(c, "Failed to get model ranking: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"models":     ranking,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.Add(-24 * time.Hour).Format("2006-01-02"),
		"limit":      limit,
	})
}

// GetActivityHeatmap 获取活跃时段热力图
// GET /api/v1/admin/ops-view/activity/heatmap
// Query params: start_date, end_date, timezone
func (h *OpsViewHandler) GetActivityHeatmap(c *gin.Context) {
	tz := c.DefaultQuery("timezone", "UTC")
	startTime, endTime := parseOpsViewTimeRange(c)

	heatmap, err := h.opsViewService.GetActivityHeatmap(c.Request.Context(), startTime, endTime, tz)
	if err != nil {
		response.InternalError(c, "Failed to get activity heatmap: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"heatmap":    heatmap,
		"start_date": startTime.Format("2006-01-02"),
		"end_date":   endTime.Add(-24 * time.Hour).Format("2006-01-02"),
	})
}

// parseOpsViewTimeRange 解析时间范围参数
func parseOpsViewTimeRange(c *gin.Context) (time.Time, time.Time) {
	userTZ := c.Query("timezone")
	now := timezone.NowInUserLocation(userTZ)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var startTime, endTime time.Time

	if startDate != "" {
		if t, err := timezone.ParseInUserLocation("2006-01-02", startDate, userTZ); err == nil {
			startTime = t
		} else {
			startTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -30), userTZ)
		}
	} else {
		startTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, -30), userTZ)
	}

	if endDate != "" {
		if t, err := timezone.ParseInUserLocation("2006-01-02", endDate, userTZ); err == nil {
			endTime = t.Add(24 * time.Hour)
		} else {
			endTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)
		}
	} else {
		endTime = timezone.StartOfDayInUserLocation(now.AddDate(0, 0, 1), userTZ)
	}

	return startTime, endTime
}
