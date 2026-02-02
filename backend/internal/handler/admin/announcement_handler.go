package admin

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// AnnouncementHandler 管理端公告处理器
type AnnouncementHandler struct {
	announcementService *service.AnnouncementService
}

// NewAnnouncementHandler 创建管理端公告处理器
func NewAnnouncementHandler(announcementService *service.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{
		announcementService: announcementService,
	}
}

// CreateAnnouncementRequest 创建公告请求
type CreateAnnouncementRequest struct {
	Title     string `json:"title" binding:"required,max=255"`
	Content   string `json:"content" binding:"required"`
	Type      string `json:"type" binding:"omitempty,oneof=info warning important"`
	Priority  int    `json:"priority"`
	PublishAt *int64 `json:"publish_at"` // Unix 时间戳（秒）
	ExpiresAt *int64 `json:"expires_at"` // Unix 时间戳（秒）
}

// UpdateAnnouncementRequest 更新公告请求
type UpdateAnnouncementRequest struct {
	Title     *string `json:"title" binding:"omitempty,max=255"`
	Content   *string `json:"content"`
	Type      *string `json:"type" binding:"omitempty,oneof=info warning important"`
	Priority  *int    `json:"priority"`
	PublishAt *int64  `json:"publish_at"` // Unix 时间戳（秒），0 表示清除
	ExpiresAt *int64  `json:"expires_at"` // Unix 时间戳（秒），0 表示清除
}

// List 获取公告列表
// GET /api/v1/admin/announcements
func (h *AnnouncementHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	status := c.Query("status")
	search := c.Query("search")

	params := pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	announcements, paginationResult, err := h.announcementService.List(c.Request.Context(), params, status, search)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, announcements, paginationResult.Total, page, pageSize)
}

// GetByID 获取公告详情
// GET /api/v1/admin/announcements/:id
func (h *AnnouncementHandler) GetByID(c *gin.Context) {
	announcementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid announcement ID")
		return
	}

	announcement, err := h.announcementService.GetByID(c.Request.Context(), announcementID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, announcement)
}

// Create 创建公告
// POST /api/v1/admin/announcements
func (h *AnnouncementHandler) Create(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	input := &service.CreateAnnouncementInput{
		Title:    req.Title,
		Content:  req.Content,
		Type:     req.Type,
		Priority: req.Priority,
	}

	var publishAt, expiresAt *time.Time
	if req.PublishAt != nil && *req.PublishAt > 0 {
		t := time.Unix(*req.PublishAt, 0)
		publishAt = &t
		input.PublishAt = &t
	}
	if req.ExpiresAt != nil && *req.ExpiresAt > 0 {
		t := time.Unix(*req.ExpiresAt, 0)
		expiresAt = &t
		input.ExpiresAt = &t
	}

	// Validate time constraints
	if publishAt != nil && expiresAt != nil && !expiresAt.After(*publishAt) {
		response.BadRequest(c, "Expiration time must be after publish time")
		return
	}

	announcement, err := h.announcementService.Create(c.Request.Context(), input, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, announcement)
}

// Update 更新公告
// PUT /api/v1/admin/announcements/:id
func (h *AnnouncementHandler) Update(c *gin.Context) {
	announcementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid announcement ID")
		return
	}

	var req UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	input := &service.UpdateAnnouncementInput{
		Title:    req.Title,
		Content:  req.Content,
		Type:     req.Type,
		Priority: req.Priority,
	}

	var publishAt, expiresAt *time.Time
	var clearPublishAt, clearExpiresAt bool

	if req.PublishAt != nil {
		if *req.PublishAt == 0 {
			// 0 表示清除
			clearPublishAt = true
			input.ClearPublishAt = true
		} else {
			t := time.Unix(*req.PublishAt, 0)
			publishAt = &t
			input.PublishAt = &t
		}
	}
	if req.ExpiresAt != nil {
		if *req.ExpiresAt == 0 {
			// 0 表示清除
			clearExpiresAt = true
			input.ClearExpiresAt = true
		} else {
			t := time.Unix(*req.ExpiresAt, 0)
			expiresAt = &t
			input.ExpiresAt = &t
		}
	}

	// Validate time constraints (only if both are being set, not cleared)
	if !clearPublishAt && !clearExpiresAt && publishAt != nil && expiresAt != nil && !expiresAt.After(*publishAt) {
		response.BadRequest(c, "Expiration time must be after publish time")
		return
	}

	announcement, err := h.announcementService.Update(c.Request.Context(), announcementID, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, announcement)
}

// Delete 删除公告
// DELETE /api/v1/admin/announcements/:id
func (h *AnnouncementHandler) Delete(c *gin.Context) {
	announcementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid announcement ID")
		return
	}

	if err := h.announcementService.Delete(c.Request.Context(), announcementID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Announcement deleted successfully"})
}

// Publish 发布公告
// POST /api/v1/admin/announcements/:id/publish
func (h *AnnouncementHandler) Publish(c *gin.Context) {
	announcementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid announcement ID")
		return
	}

	announcement, err := h.announcementService.Publish(c.Request.Context(), announcementID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, announcement)
}

// Archive 归档公告
// POST /api/v1/admin/announcements/:id/archive
func (h *AnnouncementHandler) Archive(c *gin.Context) {
	announcementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid announcement ID")
		return
	}

	announcement, err := h.announcementService.Archive(c.Request.Context(), announcementID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, announcement)
}
