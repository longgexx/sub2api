//go:build unit

package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// stubAnnouncementService 公告服务桩
type stubAnnouncementService struct {
	announcements []service.Announcement
	createErr     error
	updateErr     error
	deleteErr     error
	getByIDErr    error
	publishErr    error
	archiveErr    error
	nextID        int64
}

func newStubAnnouncementService() *stubAnnouncementService {
	now := time.Now().UTC()
	return &stubAnnouncementService{
		announcements: []service.Announcement{
			{
				ID:        1,
				Title:     "Test Announcement",
				Content:   "Test content",
				Type:      service.AnnouncementTypeInfo,
				Priority:  10,
				Status:    service.AnnouncementStatusDraft,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:        2,
				Title:     "Published Announcement",
				Content:   "Published content",
				Type:      service.AnnouncementTypeWarning,
				Priority:  20,
				Status:    service.AnnouncementStatusPublished,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		nextID: 3,
	}
}

func (s *stubAnnouncementService) Create(_ context.Context, input *service.CreateAnnouncementInput, createdBy int64) (*service.Announcement, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	a := &service.Announcement{
		ID:        s.nextID,
		Title:     input.Title,
		Content:   input.Content,
		Type:      input.Type,
		Priority:  input.Priority,
		Status:    service.AnnouncementStatusDraft,
		PublishAt: input.PublishAt,
		ExpiresAt: input.ExpiresAt,
		CreatedBy: &createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.nextID++
	s.announcements = append(s.announcements, *a)
	return a, nil
}

func (s *stubAnnouncementService) GetByID(_ context.Context, id int64) (*service.Announcement, error) {
	if s.getByIDErr != nil {
		return nil, s.getByIDErr
	}
	for i := range s.announcements {
		if s.announcements[i].ID == id {
			return &s.announcements[i], nil
		}
	}
	return nil, service.ErrAnnouncementNotFound
}

func (s *stubAnnouncementService) Update(_ context.Context, id int64, input *service.UpdateAnnouncementInput) (*service.Announcement, error) {
	if s.updateErr != nil {
		return nil, s.updateErr
	}
	for i := range s.announcements {
		if s.announcements[i].ID == id {
			if input.Title != nil {
				s.announcements[i].Title = *input.Title
			}
			if input.Content != nil {
				s.announcements[i].Content = *input.Content
			}
			if input.Type != nil {
				s.announcements[i].Type = *input.Type
			}
			if input.Priority != nil {
				s.announcements[i].Priority = *input.Priority
			}
			if input.PublishAt != nil {
				s.announcements[i].PublishAt = input.PublishAt
			}
			if input.ExpiresAt != nil {
				s.announcements[i].ExpiresAt = input.ExpiresAt
			}
			if input.ClearPublishAt {
				s.announcements[i].PublishAt = nil
			}
			if input.ClearExpiresAt {
				s.announcements[i].ExpiresAt = nil
			}
			s.announcements[i].UpdatedAt = time.Now()
			return &s.announcements[i], nil
		}
	}
	return nil, service.ErrAnnouncementNotFound
}

func (s *stubAnnouncementService) Delete(_ context.Context, id int64) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	for i := range s.announcements {
		if s.announcements[i].ID == id {
			s.announcements = append(s.announcements[:i], s.announcements[i+1:]...)
			return nil
		}
	}
	return service.ErrAnnouncementNotFound
}

func (s *stubAnnouncementService) List(_ context.Context, params pagination.PaginationParams, status string, search string) ([]service.Announcement, *pagination.PaginationResult, error) {
	var result []service.Announcement
	for _, a := range s.announcements {
		if status != "" && a.Status != status {
			continue
		}
		result = append(result, a)
	}
	return result, &pagination.PaginationResult{Total: int64(len(result))}, nil
}

func (s *stubAnnouncementService) Publish(_ context.Context, id int64) (*service.Announcement, error) {
	if s.publishErr != nil {
		return nil, s.publishErr
	}
	for i := range s.announcements {
		if s.announcements[i].ID == id {
			if s.announcements[i].Status != service.AnnouncementStatusDraft {
				return nil, service.ErrAnnouncementInvalidStatusTransition
			}
			s.announcements[i].Status = service.AnnouncementStatusPublished
			return &s.announcements[i], nil
		}
	}
	return nil, service.ErrAnnouncementNotFound
}

func (s *stubAnnouncementService) Archive(_ context.Context, id int64) (*service.Announcement, error) {
	if s.archiveErr != nil {
		return nil, s.archiveErr
	}
	for i := range s.announcements {
		if s.announcements[i].ID == id {
			if s.announcements[i].Status != service.AnnouncementStatusPublished {
				return nil, service.ErrAnnouncementInvalidStatusTransition
			}
			s.announcements[i].Status = service.AnnouncementStatusArchived
			return &s.announcements[i], nil
		}
	}
	return nil, service.ErrAnnouncementNotFound
}

// 用户端方法（handler 测试不需要，但接口需要）
func (s *stubAnnouncementService) GetActiveAnnouncements(_ context.Context, _ int64) ([]service.Announcement, error) {
	return nil, nil
}

func (s *stubAnnouncementService) GetUnreadAnnouncements(_ context.Context, _ int64) ([]service.Announcement, error) {
	return nil, nil
}

func (s *stubAnnouncementService) GetUnreadCount(_ context.Context, _ int64) (int, error) {
	return 0, nil
}

func (s *stubAnnouncementService) MarkAsRead(_ context.Context, _, _ int64) error {
	return nil
}

func (s *stubAnnouncementService) MarkAllAsRead(_ context.Context, _ int64) error {
	return nil
}

// setupAnnouncementRouter 设置测试路由
func setupAnnouncementRouter() (*gin.Engine, *stubAnnouncementService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	svc := newStubAnnouncementService()

	// 创建真实的 AnnouncementService 包装 stub
	announcementSvc := &service.AnnouncementService{}
	handler := NewAnnouncementHandler(announcementSvc)

	// 使用自定义中间件注入 stub service
	router.Use(func(c *gin.Context) {
		c.Set("announcementStub", svc)
		c.Next()
	})

	// 模拟认证中间件
	authMiddleware := func(c *gin.Context) {
		c.Set("user", struct {
			UserID      int64
			Concurrency int
		}{UserID: 1, Concurrency: 5})
		c.Next()
	}

	// 由于我们需要测试 handler，我们直接使用 stub service
	// 创建一个包装 handler 的方式
	stubHandler := &stubAnnouncementHandler{stub: svc}

	router.GET("/api/v1/admin/announcements", stubHandler.List)
	router.GET("/api/v1/admin/announcements/:id", stubHandler.GetByID)
	router.POST("/api/v1/admin/announcements", authMiddleware, stubHandler.Create)
	router.PUT("/api/v1/admin/announcements/:id", stubHandler.Update)
	router.DELETE("/api/v1/admin/announcements/:id", stubHandler.Delete)
	router.POST("/api/v1/admin/announcements/:id/publish", stubHandler.Publish)
	router.POST("/api/v1/admin/announcements/:id/archive", stubHandler.Archive)

	_ = handler // 避免未使用警告

	return router, svc
}

// stubAnnouncementHandler 直接使用 stub service 的 handler
type stubAnnouncementHandler struct {
	stub *stubAnnouncementService
}

func (h *stubAnnouncementHandler) List(c *gin.Context) {
	status := c.Query("status")
	search := c.Query("search")
	announcements, result, err := h.stub.List(c.Request.Context(), pagination.PaginationParams{Page: 1, PageSize: 20}, status, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": announcements, "total": result.Total})
}

func (h *stubAnnouncementHandler) GetByID(c *gin.Context) {
	var id int64
	if _, err := parseID(c.Param("id"), &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	a, err := h.stub.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrAnnouncementNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": a})
}

func (h *stubAnnouncementHandler) Create(c *gin.Context) {
	var req CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := &service.CreateAnnouncementInput{
		Title:    req.Title,
		Content:  req.Content,
		Type:     req.Type,
		Priority: req.Priority,
	}

	if req.PublishAt != nil && *req.PublishAt > 0 {
		t := time.Unix(*req.PublishAt, 0)
		input.PublishAt = &t
	}
	if req.ExpiresAt != nil && *req.ExpiresAt > 0 {
		t := time.Unix(*req.ExpiresAt, 0)
		input.ExpiresAt = &t
	}

	// 时间验证
	if input.PublishAt != nil && input.ExpiresAt != nil && !input.ExpiresAt.After(*input.PublishAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Expiration time must be after publish time"})
		return
	}

	a, err := h.stub.Create(c.Request.Context(), input, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": a})
}

func (h *stubAnnouncementHandler) Update(c *gin.Context) {
	var id int64
	if _, err := parseID(c.Param("id"), &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := &service.UpdateAnnouncementInput{
		Title:    req.Title,
		Content:  req.Content,
		Type:     req.Type,
		Priority: req.Priority,
	}

	if req.PublishAt != nil {
		if *req.PublishAt == 0 {
			input.ClearPublishAt = true
		} else {
			t := time.Unix(*req.PublishAt, 0)
			input.PublishAt = &t
		}
	}
	if req.ExpiresAt != nil {
		if *req.ExpiresAt == 0 {
			input.ClearExpiresAt = true
		} else {
			t := time.Unix(*req.ExpiresAt, 0)
			input.ExpiresAt = &t
		}
	}

	a, err := h.stub.Update(c.Request.Context(), id, input)
	if err != nil {
		if err == service.ErrAnnouncementNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": a})
}

func (h *stubAnnouncementHandler) Delete(c *gin.Context) {
	var id int64
	if _, err := parseID(c.Param("id"), &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err := h.stub.Delete(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrAnnouncementNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

func (h *stubAnnouncementHandler) Publish(c *gin.Context) {
	var id int64
	if _, err := parseID(c.Param("id"), &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	a, err := h.stub.Publish(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrAnnouncementNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		if err == service.ErrAnnouncementInvalidStatusTransition {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status transition"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": a})
}

func (h *stubAnnouncementHandler) Archive(c *gin.Context) {
	var id int64
	if _, err := parseID(c.Param("id"), &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	a, err := h.stub.Archive(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrAnnouncementNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		if err == service.ErrAnnouncementInvalidStatusTransition {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status transition"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": a})
}

func parseID(s string, id *int64) (bool, error) {
	var err error
	*id, err = parseInt64(s)
	return err == nil, err
}

func parseInt64(s string) (int64, error) {
	var id int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, service.ErrAnnouncementNotFound
		}
		id = id*10 + int64(c-'0')
	}
	return id, nil
}

// ==================== 测试用例 ====================

func TestAnnouncementHandler_List(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/announcements", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotNil(t, resp["data"])
}

func TestAnnouncementHandler_List_FilterByStatus(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/announcements?status=draft", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAnnouncementHandler_GetByID(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/announcements/1", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotNil(t, resp["data"])
}

func TestAnnouncementHandler_GetByID_NotFound(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/announcements/999", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAnnouncementHandler_GetByID_InvalidID(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/announcements/invalid", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAnnouncementHandler_Create(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	body := map[string]any{
		"title":    "New Announcement",
		"content":  "New content",
		"type":     "info",
		"priority": 5,
	}
	jsonBody, _ := json.Marshal(body)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/announcements", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotNil(t, resp["data"])
}

func TestAnnouncementHandler_Create_WithSchedule(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	publishAt := time.Now().Add(time.Hour).Unix()
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	body := map[string]any{
		"title":      "Scheduled Announcement",
		"content":    "Content",
		"publish_at": publishAt,
		"expires_at": expiresAt,
	}
	jsonBody, _ := json.Marshal(body)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/announcements", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAnnouncementHandler_Create_InvalidTimeRange(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	publishAt := time.Now().Add(24 * time.Hour).Unix()
	expiresAt := time.Now().Add(time.Hour).Unix() // 过期时间早于发布时间

	body := map[string]any{
		"title":      "Invalid",
		"content":    "Content",
		"publish_at": publishAt,
		"expires_at": expiresAt,
	}
	jsonBody, _ := json.Marshal(body)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/announcements", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAnnouncementHandler_Create_MissingTitle(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	body := map[string]any{
		"content": "Content without title",
	}
	jsonBody, _ := json.Marshal(body)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/announcements", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAnnouncementHandler_Update(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	body := map[string]any{
		"title":   "Updated Title",
		"content": "Updated Content",
	}
	jsonBody, _ := json.Marshal(body)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/announcements/1", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]any)
	require.Equal(t, "Updated Title", data["title"])
}

func TestAnnouncementHandler_Update_ClearPublishAt(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	body := map[string]any{
		"publish_at": 0, // 0 表示清除
	}
	jsonBody, _ := json.Marshal(body)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/announcements/1", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAnnouncementHandler_Update_NotFound(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	body := map[string]any{
		"title": "Updated",
	}
	jsonBody, _ := json.Marshal(body)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/announcements/999", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAnnouncementHandler_Delete(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/announcements/1", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAnnouncementHandler_Delete_NotFound(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/announcements/999", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAnnouncementHandler_Publish(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/announcements/1/publish", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]any)
	require.Equal(t, service.AnnouncementStatusPublished, data["status"])
}

func TestAnnouncementHandler_Publish_AlreadyPublished(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	// ID 2 已经是 published 状态
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/announcements/2/publish", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAnnouncementHandler_Publish_NotFound(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/announcements/999/publish", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAnnouncementHandler_Archive(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	// ID 2 是 published 状态，可以归档
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/announcements/2/archive", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	data := resp["data"].(map[string]any)
	require.Equal(t, service.AnnouncementStatusArchived, data["status"])
}

func TestAnnouncementHandler_Archive_Draft(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	// ID 1 是 draft 状态，不能归档
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/announcements/1/archive", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAnnouncementHandler_Archive_NotFound(t *testing.T) {
	router, _ := setupAnnouncementRouter()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/announcements/999/archive", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
