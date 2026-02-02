package service

import (
	"context"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// 公告相关错误
var (
	ErrAnnouncementNotFound                = infraerrors.NotFound("ANNOUNCEMENT_NOT_FOUND", "announcement not found")
	ErrAnnouncementTitleRequired           = infraerrors.BadRequest("ANNOUNCEMENT_TITLE_REQUIRED", "title is required")
	ErrAnnouncementContentRequired         = infraerrors.BadRequest("ANNOUNCEMENT_CONTENT_REQUIRED", "content is required")
	ErrAnnouncementInvalidType             = infraerrors.BadRequest("ANNOUNCEMENT_INVALID_TYPE", "invalid announcement type")
	ErrAnnouncementInvalidTimeRange        = infraerrors.BadRequest("ANNOUNCEMENT_INVALID_TIME_RANGE", "expiration time must be after publish time")
	ErrAnnouncementInvalidStatusTransition = infraerrors.BadRequest("ANNOUNCEMENT_INVALID_STATUS_TRANSITION", "invalid announcement status transition")
)

// 公告类型常量
const (
	AnnouncementTypeInfo      = "info"
	AnnouncementTypeWarning   = "warning"
	AnnouncementTypeImportant = "important"
)

// 公告状态常量
const (
	AnnouncementStatusDraft     = "draft"
	AnnouncementStatusPublished = "published"
	AnnouncementStatusArchived  = "archived"
)

// Announcement 公告模型
type Announcement struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Type      string     `json:"type"`
	Priority  int        `json:"priority"`
	Status    string     `json:"status"`
	PublishAt *time.Time `json:"publish_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedBy *int64     `json:"created_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	IsRead    bool       `json:"is_read,omitempty"` // 用户端使用，标记是否已读
}

// AnnouncementRead 公告已读记录
type AnnouncementRead struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	AnnouncementID int64     `json:"announcement_id"`
	ReadAt         time.Time `json:"read_at"`
}

// CreateAnnouncementInput 创建公告输入
type CreateAnnouncementInput struct {
	Title     string     `json:"title" binding:"required,max=255"`
	Content   string     `json:"content" binding:"required"`
	Type      string     `json:"type" binding:"omitempty,oneof=info warning important"`
	Priority  int        `json:"priority"`
	PublishAt *time.Time `json:"publish_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// UpdateAnnouncementInput 更新公告输入
type UpdateAnnouncementInput struct {
	Title          *string    `json:"title,omitempty" binding:"omitempty,max=255"`
	Content        *string    `json:"content,omitempty"`
	Type           *string    `json:"type,omitempty" binding:"omitempty,oneof=info warning important"`
	Priority       *int       `json:"priority,omitempty"`
	PublishAt      *time.Time `json:"publish_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	ClearPublishAt bool       `json:"-"` // 是否清除发布时间
	ClearExpiresAt bool       `json:"-"` // 是否清除过期时间
}

// AnnouncementRepository 公告仓储接口
type AnnouncementRepository interface {
	// 基础 CRUD
	Create(ctx context.Context, announcement *Announcement) error
	GetByID(ctx context.Context, id int64) (*Announcement, error)
	Update(ctx context.Context, announcement *Announcement) error
	Delete(ctx context.Context, id int64) error

	// 列表查询
	List(ctx context.Context, params pagination.PaginationParams, status string, search string) ([]Announcement, *pagination.PaginationResult, error)

	// 用户端查询
	GetActiveAnnouncements(ctx context.Context, now time.Time) ([]Announcement, error)
	GetUnreadAnnouncements(ctx context.Context, userID int64, now time.Time) ([]Announcement, error)
	CountUnread(ctx context.Context, userID int64, now time.Time) (int, error)
	GetUserReadIDs(ctx context.Context, userID int64) ([]int64, error)

	// 已读记录
	MarkAsRead(ctx context.Context, userID, announcementID int64) error
	MarkAllAsRead(ctx context.Context, userID int64, now time.Time) error

	// 定时发布
	BatchPublishScheduled(ctx context.Context, now time.Time) (int, error)
}
