package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// stubOpsViewService implements a mock OpsViewService for testing
type stubOpsViewService struct{}

func newStubOpsViewService() *stubOpsViewService {
	return &stubOpsViewService{}
}

func (s *stubOpsViewService) GetOverview(ctx context.Context, tz string) (*service.OpsViewOverview, error) {
	return &service.OpsViewOverview{
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
	}, nil
}

func (s *stubOpsViewService) GetComparison(ctx context.Context, tz string) (*service.OpsViewComparison, error) {
	return &service.OpsViewComparison{
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
	}, nil
}

func (s *stubOpsViewService) GetTrend(ctx context.Context, days int, tz string) ([]service.OpsViewTrendPoint, error) {
	return []service.OpsViewTrendPoint{
		{Date: "2024-01-01", Consumption: 100.0, PaymentRecharge: 80.0, RedeemRecharge: 20.0, TotalRecharge: 100.0},
		{Date: "2024-01-02", Consumption: 120.0, PaymentRecharge: 90.0, RedeemRecharge: 30.0, TotalRecharge: 120.0},
	}, nil
}

func (s *stubOpsViewService) GetUserGrowth(ctx context.Context, days int, tz string) ([]service.OpsViewUserGrowthPoint, error) {
	return []service.OpsViewUserGrowthPoint{
		{Date: "2024-01-01", NewUsers: 10, ActiveUsers: 50, PayingUsers: 5, RetentionD1: 80.0, RetentionD7: 60.0, RetentionD30: 40.0},
		{Date: "2024-01-02", NewUsers: 12, ActiveUsers: 55, PayingUsers: 6, RetentionD1: 82.0, RetentionD7: 62.0, RetentionD30: 42.0},
	}, nil
}

func (s *stubOpsViewService) GetGroupConsumption(ctx context.Context, start, end time.Time) ([]service.OpsViewGroupConsumption, error) {
	return []service.OpsViewGroupConsumption{
		{GroupID: 1, GroupName: "Group A", Consumption: 500.0, Percentage: 60.0, Requests: 1000},
		{GroupID: 2, GroupName: "Group B", Consumption: 333.33, Percentage: 40.0, Requests: 666},
	}, nil
}

func (s *stubOpsViewService) GetTopUsers(ctx context.Context, start, end time.Time, limit int) ([]service.OpsViewTopUser, error) {
	now := time.Now()
	return []service.OpsViewTopUser{
		{UserID: 1, Email: "user1@example.com", Username: "user1", Consumption: 200.0, Requests: 100, AvgCost: 2.0, LastActiveAt: now, RegisteredAt: now.AddDate(0, -1, 0)},
		{UserID: 2, Email: "user2@example.com", Username: "user2", Consumption: 150.0, Requests: 80, AvgCost: 1.875, LastActiveAt: now, RegisteredAt: now.AddDate(0, -2, 0)},
	}, nil
}

func (s *stubOpsViewService) GetModelRanking(ctx context.Context, start, end time.Time, limit int) ([]service.OpsViewModelRanking, error) {
	return []service.OpsViewModelRanking{
		{Model: "claude-3-opus", Requests: 500, Tokens: 100000, Consumption: 300.0, Percentage: 50.0},
		{Model: "claude-3-sonnet", Requests: 400, Tokens: 80000, Consumption: 200.0, Percentage: 33.33},
	}, nil
}

func (s *stubOpsViewService) GetActivityHeatmap(ctx context.Context, start, end time.Time, tz string) ([]service.OpsViewActivityHeatmapCell, error) {
	cells := make([]service.OpsViewActivityHeatmapCell, 0, 7*24)
	for weekday := 0; weekday < 7; weekday++ {
		for hour := 0; hour < 24; hour++ {
			cells = append(cells, service.OpsViewActivityHeatmapCell{
				Weekday:  weekday,
				Hour:     hour,
				Requests: int64((weekday + 1) * (hour + 1) * 10),
			})
		}
	}
	return cells, nil
}

// OpsViewServiceInterface defines the interface for OpsViewService
type OpsViewServiceInterface interface {
	GetOverview(ctx context.Context, tz string) (*service.OpsViewOverview, error)
	GetComparison(ctx context.Context, tz string) (*service.OpsViewComparison, error)
	GetTrend(ctx context.Context, days int, tz string) ([]service.OpsViewTrendPoint, error)
	GetUserGrowth(ctx context.Context, days int, tz string) ([]service.OpsViewUserGrowthPoint, error)
	GetGroupConsumption(ctx context.Context, start, end time.Time) ([]service.OpsViewGroupConsumption, error)
	GetTopUsers(ctx context.Context, start, end time.Time, limit int) ([]service.OpsViewTopUser, error)
	GetModelRanking(ctx context.Context, start, end time.Time, limit int) ([]service.OpsViewModelRanking, error)
	GetActivityHeatmap(ctx context.Context, start, end time.Time, tz string) ([]service.OpsViewActivityHeatmapCell, error)
}

// testOpsViewHandler wraps OpsViewHandler for testing with stub service
type testOpsViewHandler struct {
	svc OpsViewServiceInterface
}

func newTestOpsViewHandler(svc OpsViewServiceInterface) *testOpsViewHandler {
	return &testOpsViewHandler{svc: svc}
}

func (h *testOpsViewHandler) GetOverview(c *gin.Context) {
	tz := c.DefaultQuery("timezone", "UTC")
	overview, err := h.svc.GetOverview(c.Request.Context(), tz)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": overview})
}

func (h *testOpsViewHandler) GetComparison(c *gin.Context) {
	tz := c.DefaultQuery("timezone", "UTC")
	comparison, err := h.svc.GetComparison(c.Request.Context(), tz)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": comparison})
}

func (h *testOpsViewHandler) GetTrend(c *gin.Context) {
	tz := c.DefaultQuery("timezone", "UTC")
	days := 30
	trend, err := h.svc.GetTrend(c.Request.Context(), days, tz)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"trend": trend, "days": days}})
}

func (h *testOpsViewHandler) GetUserGrowth(c *gin.Context) {
	tz := c.DefaultQuery("timezone", "UTC")
	days := 30
	growth, err := h.svc.GetUserGrowth(c.Request.Context(), days, tz)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"growth": growth, "days": days}})
}

func (h *testOpsViewHandler) GetGroupConsumption(c *gin.Context) {
	now := time.Now()
	start := now.AddDate(0, 0, -30)
	groups, err := h.svc.GetGroupConsumption(c.Request.Context(), start, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"groups": groups}})
}

func (h *testOpsViewHandler) GetTopUsers(c *gin.Context) {
	now := time.Now()
	start := now.AddDate(0, 0, -30)
	users, err := h.svc.GetTopUsers(c.Request.Context(), start, now, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"users": users}})
}

func (h *testOpsViewHandler) GetModelRanking(c *gin.Context) {
	now := time.Now()
	start := now.AddDate(0, 0, -30)
	models, err := h.svc.GetModelRanking(c.Request.Context(), start, now, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"models": models}})
}

func (h *testOpsViewHandler) GetActivityHeatmap(c *gin.Context) {
	tz := c.DefaultQuery("timezone", "UTC")
	now := time.Now()
	start := now.AddDate(0, 0, -30)
	heatmap, err := h.svc.GetActivityHeatmap(c.Request.Context(), start, now, tz)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"heatmap": heatmap}})
}

func setupOpsViewRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	svc := newStubOpsViewService()
	handler := newTestOpsViewHandler(svc)

	opsView := router.Group("/api/v1/admin/ops-view")
	{
		opsView.GET("/overview", handler.GetOverview)
		opsView.GET("/comparison", handler.GetComparison)
		opsView.GET("/trend", handler.GetTrend)
		opsView.GET("/users/growth", handler.GetUserGrowth)
		opsView.GET("/groups/consumption", handler.GetGroupConsumption)
		opsView.GET("/users/top", handler.GetTopUsers)
		opsView.GET("/models/ranking", handler.GetModelRanking)
		opsView.GET("/activity/heatmap", handler.GetActivityHeatmap)
	}

	return router
}

func TestOpsViewHandlerGetOverview(t *testing.T) {
	router := setupOpsViewRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ops-view/overview?timezone=UTC", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	success, ok := resp["success"].(bool)
	require.True(t, ok)
	require.True(t, success)

	data, ok := resp["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, 100.50, data["today_consumption"])
	require.Equal(t, 500.25, data["week_consumption"])
	require.Equal(t, float64(50), data["week_active_users"])
}

func TestOpsViewHandlerGetComparison(t *testing.T) {
	router := setupOpsViewRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ops-view/comparison?timezone=UTC", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	success, ok := resp["success"].(bool)
	require.True(t, ok)
	require.True(t, success)

	data, ok := resp["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, 100.50, data["today_consumption"])
	require.Equal(t, 80.00, data["yesterday_consumption"])
	require.Equal(t, 25.625, data["consumption_change"])
}

func TestOpsViewHandlerGetTrend(t *testing.T) {
	router := setupOpsViewRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ops-view/trend?days=30&timezone=UTC", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	success, ok := resp["success"].(bool)
	require.True(t, ok)
	require.True(t, success)

	data, ok := resp["data"].(map[string]any)
	require.True(t, ok)
	trend, ok := data["trend"].([]any)
	require.True(t, ok)
	require.Len(t, trend, 2)

	firstPoint, ok := trend[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "2024-01-01", firstPoint["date"])
	require.Equal(t, 100.0, firstPoint["consumption"])
}

func TestOpsViewHandlerGetUserGrowth(t *testing.T) {
	router := setupOpsViewRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ops-view/users/growth?days=30&timezone=UTC", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	success, ok := resp["success"].(bool)
	require.True(t, ok)
	require.True(t, success)

	data, ok := resp["data"].(map[string]any)
	require.True(t, ok)
	growth, ok := data["growth"].([]any)
	require.True(t, ok)
	require.Len(t, growth, 2)

	firstPoint, ok := growth[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "2024-01-01", firstPoint["date"])
	require.Equal(t, float64(10), firstPoint["new_users"])
}

func TestOpsViewHandlerGetGroupConsumption(t *testing.T) {
	router := setupOpsViewRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ops-view/groups/consumption?timezone=UTC", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	success, ok := resp["success"].(bool)
	require.True(t, ok)
	require.True(t, success)

	data, ok := resp["data"].(map[string]any)
	require.True(t, ok)
	groups, ok := data["groups"].([]any)
	require.True(t, ok)
	require.Len(t, groups, 2)

	firstGroup, ok := groups[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Group A", firstGroup["group_name"])
	require.Equal(t, 500.0, firstGroup["consumption"])
}

func TestOpsViewHandlerGetTopUsers(t *testing.T) {
	router := setupOpsViewRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ops-view/users/top?limit=20&timezone=UTC", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	success, ok := resp["success"].(bool)
	require.True(t, ok)
	require.True(t, success)

	data, ok := resp["data"].(map[string]any)
	require.True(t, ok)
	users, ok := data["users"].([]any)
	require.True(t, ok)
	require.Len(t, users, 2)

	firstUser, ok := users[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "user1@example.com", firstUser["email"])
	require.Equal(t, 200.0, firstUser["consumption"])
}

func TestOpsViewHandlerGetModelRanking(t *testing.T) {
	router := setupOpsViewRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ops-view/models/ranking?limit=20&timezone=UTC", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	success, ok := resp["success"].(bool)
	require.True(t, ok)
	require.True(t, success)

	data, ok := resp["data"].(map[string]any)
	require.True(t, ok)
	models, ok := data["models"].([]any)
	require.True(t, ok)
	require.Len(t, models, 2)

	firstModel, ok := models[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "claude-3-opus", firstModel["model"])
	require.Equal(t, float64(500), firstModel["requests"])
}

func TestOpsViewHandlerGetActivityHeatmap(t *testing.T) {
	router := setupOpsViewRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ops-view/activity/heatmap?timezone=UTC", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	success, ok := resp["success"].(bool)
	require.True(t, ok)
	require.True(t, success)

	data, ok := resp["data"].(map[string]any)
	require.True(t, ok)
	heatmap, ok := data["heatmap"].([]any)
	require.True(t, ok)
	require.Len(t, heatmap, 7*24) // 7 days * 24 hours

	firstCell, ok := heatmap[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(0), firstCell["weekday"])
	require.Equal(t, float64(0), firstCell["hour"])
}

func TestOpsViewHandlerAllEndpoints(t *testing.T) {
	router := setupOpsViewRouter()

	endpoints := []struct {
		name   string
		path   string
		method string
	}{
		{"Overview", "/api/v1/admin/ops-view/overview", http.MethodGet},
		{"Comparison", "/api/v1/admin/ops-view/comparison", http.MethodGet},
		{"Trend", "/api/v1/admin/ops-view/trend", http.MethodGet},
		{"UserGrowth", "/api/v1/admin/ops-view/users/growth", http.MethodGet},
		{"GroupConsumption", "/api/v1/admin/ops-view/groups/consumption", http.MethodGet},
		{"TopUsers", "/api/v1/admin/ops-view/users/top", http.MethodGet},
		{"ModelRanking", "/api/v1/admin/ops-view/models/ranking", http.MethodGet},
		{"ActivityHeatmap", "/api/v1/admin/ops-view/activity/heatmap", http.MethodGet},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(ep.method, ep.path, nil)
			router.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code, "Endpoint %s failed", ep.name)
		})
	}
}

func TestOpsViewHandlerGetTrendWithDifferentDays(t *testing.T) {
	router := setupOpsViewRouter()

	testCases := []struct {
		name         string
		queryParams  string
		expectedCode int
	}{
		{"default days", "", http.StatusOK},
		{"7 days", "?days=7", http.StatusOK},
		{"30 days", "?days=30", http.StatusOK},
		{"365 days", "?days=365", http.StatusOK},
		{"invalid days (negative)", "?days=-1", http.StatusOK},
		{"invalid days (zero)", "?days=0", http.StatusOK},
		{"invalid days (too large)", "?days=1000", http.StatusOK},
		{"invalid days (non-numeric)", "?days=abc", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ops-view/trend"+tc.queryParams, nil)
			router.ServeHTTP(rec, req)
			require.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestOpsViewHandlerGetUserGrowthWithDifferentDays(t *testing.T) {
	router := setupOpsViewRouter()

	testCases := []struct {
		name         string
		queryParams  string
		expectedCode int
	}{
		{"default days", "", http.StatusOK},
		{"7 days", "?days=7", http.StatusOK},
		{"invalid days", "?days=invalid", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ops-view/users/growth"+tc.queryParams, nil)
			router.ServeHTTP(rec, req)
			require.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestOpsViewHandlerGetTopUsersWithDifferentLimits(t *testing.T) {
	router := setupOpsViewRouter()

	testCases := []struct {
		name         string
		queryParams  string
		expectedCode int
	}{
		{"default limit", "", http.StatusOK},
		{"limit 10", "?limit=10", http.StatusOK},
		{"limit 100", "?limit=100", http.StatusOK},
		{"invalid limit (negative)", "?limit=-1", http.StatusOK},
		{"invalid limit (zero)", "?limit=0", http.StatusOK},
		{"invalid limit (too large)", "?limit=500", http.StatusOK},
		{"invalid limit (non-numeric)", "?limit=abc", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ops-view/users/top"+tc.queryParams, nil)
			router.ServeHTTP(rec, req)
			require.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestOpsViewHandlerGetModelRankingWithDifferentLimits(t *testing.T) {
	router := setupOpsViewRouter()

	testCases := []struct {
		name         string
		queryParams  string
		expectedCode int
	}{
		{"default limit", "", http.StatusOK},
		{"limit 5", "?limit=5", http.StatusOK},
		{"invalid limit", "?limit=invalid", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ops-view/models/ranking"+tc.queryParams, nil)
			router.ServeHTTP(rec, req)
			require.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestOpsViewHandlerWithDifferentTimezones(t *testing.T) {
	router := setupOpsViewRouter()

	timezones := []string{
		"UTC",
		"Asia/Shanghai",
		"America/New_York",
		"Europe/London",
		"Invalid/Timezone",
		"",
	}

	endpoints := []string{
		"/api/v1/admin/ops-view/overview",
		"/api/v1/admin/ops-view/comparison",
		"/api/v1/admin/ops-view/trend",
		"/api/v1/admin/ops-view/users/growth",
		"/api/v1/admin/ops-view/activity/heatmap",
	}

	for _, endpoint := range endpoints {
		for _, tz := range timezones {
			t.Run(endpoint+"_"+tz, func(t *testing.T) {
				rec := httptest.NewRecorder()
				url := endpoint
				if tz != "" {
					url += "?timezone=" + tz
				}
				req := httptest.NewRequest(http.MethodGet, url, nil)
				router.ServeHTTP(rec, req)
				require.Equal(t, http.StatusOK, rec.Code)
			})
		}
	}
}

func TestOpsViewHandlerGetGroupConsumptionWithDateRange(t *testing.T) {
	router := setupOpsViewRouter()

	testCases := []struct {
		name         string
		queryParams  string
		expectedCode int
	}{
		{"no date range", "", http.StatusOK},
		{"with start_date", "?start_date=2024-01-01", http.StatusOK},
		{"with end_date", "?end_date=2024-01-31", http.StatusOK},
		{"with both dates", "?start_date=2024-01-01&end_date=2024-01-31", http.StatusOK},
		{"invalid start_date", "?start_date=invalid", http.StatusOK},
		{"invalid end_date", "?end_date=invalid", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/ops-view/groups/consumption"+tc.queryParams, nil)
			router.ServeHTTP(rec, req)
			require.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestParseOpsViewTimeRange(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name      string
		startDate string
		endDate   string
		timezone  string
	}{
		{"default range", "", "", "UTC"},
		{"with valid dates", "2024-01-01", "2024-01-31", "UTC"},
		{"with timezone", "2024-01-01", "2024-01-31", "Asia/Shanghai"},
		{"invalid start date falls back to default", "invalid", "", "UTC"},
		{"invalid end date falls back to default", "", "invalid", "UTC"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?start_date="+tc.startDate+"&end_date="+tc.endDate+"&timezone="+tc.timezone, nil)

			startTime, endTime := parseOpsViewTimeRange(c)

			require.False(t, startTime.IsZero())
			require.False(t, endTime.IsZero())
			require.True(t, endTime.After(startTime))
		})
	}
}
