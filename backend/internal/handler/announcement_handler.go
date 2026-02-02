package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// AnnouncementHandler 用户端公告处理器
type AnnouncementHandler struct {
	announcementService *service.AnnouncementService
}

// NewAnnouncementHandler 创建用户端公告处理器
func NewAnnouncementHandler(announcementService *service.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{
		announcementService: announcementService,
	}
}

// GetActiveAnnouncements 获取当前有效公告
// GET /api/v1/announcements
func (h *AnnouncementHandler) GetActiveAnnouncements(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	announcements, err := h.announcementService.GetActiveAnnouncements(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, announcements)
}

// GetUnreadAnnouncements 获取未读公告
// GET /api/v1/announcements/unread
func (h *AnnouncementHandler) GetUnreadAnnouncements(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	announcements, err := h.announcementService.GetUnreadAnnouncements(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, announcements)
}

// GetUnreadCount 获取未读公告数量
// GET /api/v1/announcements/unread/count
func (h *AnnouncementHandler) GetUnreadCount(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	count, err := h.announcementService.GetUnreadCount(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"count": count})
}

// MarkAsRead 标记公告为已读
// POST /api/v1/announcements/:id/read
func (h *AnnouncementHandler) MarkAsRead(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	announcementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid announcement ID")
		return
	}

	if err := h.announcementService.MarkAsRead(c.Request.Context(), subject.UserID, announcementID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Marked as read"})
}

// MarkAllAsRead 标记所有公告为已读
// POST /api/v1/announcements/read-all
func (h *AnnouncementHandler) MarkAllAsRead(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	if err := h.announcementService.MarkAllAsRead(c.Request.Context(), subject.UserID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "All marked as read"})
}
