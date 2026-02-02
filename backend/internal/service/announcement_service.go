package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// AnnouncementService 公告服务
type AnnouncementService struct {
	announcementRepo AnnouncementRepository
}

// NewAnnouncementService 创建公告服务实例
func NewAnnouncementService(announcementRepo AnnouncementRepository) *AnnouncementService {
	return &AnnouncementService{
		announcementRepo: announcementRepo,
	}
}

func normalizeAnnouncementType(t string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(t))
	if normalized == "" {
		return AnnouncementTypeInfo, nil
	}
	switch normalized {
	case AnnouncementTypeInfo, AnnouncementTypeWarning, AnnouncementTypeImportant:
		return normalized, nil
	default:
		return "", ErrAnnouncementInvalidType
	}
}

func validateAnnouncementBasicFields(title, content string) error {
	if strings.TrimSpace(title) == "" {
		return ErrAnnouncementTitleRequired
	}
	if strings.TrimSpace(content) == "" {
		return ErrAnnouncementContentRequired
	}
	return nil
}

func validateAnnouncementTimeRange(publishAt, expiresAt *time.Time) error {
	if publishAt != nil && expiresAt != nil && !expiresAt.After(*publishAt) {
		return ErrAnnouncementInvalidTimeRange
	}
	return nil
}

func isAnnouncementActiveAt(a *Announcement, now time.Time) bool {
	if a == nil {
		return false
	}
	if a.Status != AnnouncementStatusPublished {
		return false
	}
	if a.PublishAt != nil && a.PublishAt.After(now) {
		return false
	}
	if a.ExpiresAt != nil && !a.ExpiresAt.After(now) {
		return false
	}
	return true
}

// ==================== 用户端方法 ====================

// GetActiveAnnouncements 获取当前有效公告（带已读状态）
func (s *AnnouncementService) GetActiveAnnouncements(ctx context.Context, userID int64) ([]Announcement, error) {
	now := time.Now()
	announcements, err := s.announcementRepo.GetActiveAnnouncements(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("get active announcements: %w", err)
	}

	// 获取用户已读记录
	readIDs, err := s.announcementRepo.GetUserReadIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user read IDs: %w", err)
	}
	readSet := make(map[int64]bool)
	for _, id := range readIDs {
		readSet[id] = true
	}

	// 标记已读状态
	for i := range announcements {
		announcements[i].IsRead = readSet[announcements[i].ID]
	}

	return announcements, nil
}

// GetUnreadAnnouncements 获取未读公告
func (s *AnnouncementService) GetUnreadAnnouncements(ctx context.Context, userID int64) ([]Announcement, error) {
	now := time.Now()
	announcements, err := s.announcementRepo.GetUnreadAnnouncements(ctx, userID, now)
	if err != nil {
		return nil, fmt.Errorf("get unread announcements: %w", err)
	}
	return announcements, nil
}

// GetUnreadCount 获取未读公告数量
func (s *AnnouncementService) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	now := time.Now()
	count, err := s.announcementRepo.CountUnread(ctx, userID, now)
	if err != nil {
		return 0, fmt.Errorf("count unread announcements: %w", err)
	}
	return count, nil
}

// MarkAsRead 标记公告为已读
func (s *AnnouncementService) MarkAsRead(ctx context.Context, userID, announcementID int64) error {
	a, err := s.announcementRepo.GetByID(ctx, announcementID)
	if err != nil {
		return err
	}
	if !isAnnouncementActiveAt(a, time.Now()) {
		// 避免向客户端泄露公告是否存在/已归档等信息
		return ErrAnnouncementNotFound
	}
	if err := s.announcementRepo.MarkAsRead(ctx, userID, announcementID); err != nil {
		return fmt.Errorf("mark as read: %w", err)
	}
	return nil
}

// MarkAllAsRead 标记所有公告为已读
func (s *AnnouncementService) MarkAllAsRead(ctx context.Context, userID int64) error {
	now := time.Now()
	if err := s.announcementRepo.MarkAllAsRead(ctx, userID, now); err != nil {
		return fmt.Errorf("mark all as read: %w", err)
	}
	return nil
}

// ==================== 管理端方法 ====================

// Create 创建公告
func (s *AnnouncementService) Create(ctx context.Context, input *CreateAnnouncementInput, createdBy int64) (*Announcement, error) {
	title := strings.TrimSpace(input.Title)
	if err := validateAnnouncementBasicFields(title, input.Content); err != nil {
		return nil, err
	}

	announcementType, err := normalizeAnnouncementType(input.Type)
	if err != nil {
		return nil, err
	}
	if input.Priority < 0 {
		return nil, infraerrors.BadRequest("ANNOUNCEMENT_INVALID_PRIORITY", "priority must be >= 0")
	}
	if err := validateAnnouncementTimeRange(input.PublishAt, input.ExpiresAt); err != nil {
		return nil, err
	}

	announcement := &Announcement{
		Title:     title,
		Content:   input.Content,
		Type:      announcementType,
		Priority:  input.Priority,
		Status:    AnnouncementStatusDraft,
		PublishAt: input.PublishAt,
		ExpiresAt: input.ExpiresAt,
		CreatedBy: &createdBy,
	}

	if err := s.announcementRepo.Create(ctx, announcement); err != nil {
		return nil, fmt.Errorf("create announcement: %w", err)
	}

	return announcement, nil
}

// GetByID 根据ID获取公告
func (s *AnnouncementService) GetByID(ctx context.Context, id int64) (*Announcement, error) {
	announcement, err := s.announcementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return announcement, nil
}

// Update 更新公告
func (s *AnnouncementService) Update(ctx context.Context, id int64, input *UpdateAnnouncementInput) (*Announcement, error) {
	announcement, err := s.announcementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		announcement.Title = strings.TrimSpace(*input.Title)
	}
	if input.Content != nil {
		announcement.Content = *input.Content
	}
	if input.Type != nil {
		normalized, err := normalizeAnnouncementType(*input.Type)
		if err != nil {
			return nil, err
		}
		announcement.Type = normalized
	}
	if input.Priority != nil {
		if *input.Priority < 0 {
			return nil, infraerrors.BadRequest("ANNOUNCEMENT_INVALID_PRIORITY", "priority must be >= 0")
		}
		announcement.Priority = *input.Priority
	}
	if input.ClearPublishAt {
		announcement.PublishAt = nil
	} else if input.PublishAt != nil {
		announcement.PublishAt = input.PublishAt
	}
	if input.ClearExpiresAt {
		announcement.ExpiresAt = nil
	} else if input.ExpiresAt != nil {
		announcement.ExpiresAt = input.ExpiresAt
	}

	if err := validateAnnouncementBasicFields(announcement.Title, announcement.Content); err != nil {
		return nil, err
	}
	if err := validateAnnouncementTimeRange(announcement.PublishAt, announcement.ExpiresAt); err != nil {
		return nil, err
	}

	if err := s.announcementRepo.Update(ctx, announcement); err != nil {
		return nil, fmt.Errorf("update announcement: %w", err)
	}

	return announcement, nil
}

// Delete 删除公告（软删除）
func (s *AnnouncementService) Delete(ctx context.Context, id int64) error {
	if err := s.announcementRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete announcement: %w", err)
	}
	return nil
}

// List 获取公告列表
func (s *AnnouncementService) List(ctx context.Context, params pagination.PaginationParams, status string, search string) ([]Announcement, *pagination.PaginationResult, error) {
	return s.announcementRepo.List(ctx, params, status, search)
}

// Publish 发布公告
func (s *AnnouncementService) Publish(ctx context.Context, id int64) (*Announcement, error) {
	announcement, err := s.announcementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 只有草稿状态的公告可以发布
	if announcement.Status != AnnouncementStatusDraft {
		return nil, ErrAnnouncementInvalidStatusTransition.WithMetadata(map[string]string{
			"current_status": announcement.Status,
			"action":         "publish",
		})
	}

	announcement.Status = AnnouncementStatusPublished

	if err := s.announcementRepo.Update(ctx, announcement); err != nil {
		return nil, fmt.Errorf("publish announcement: %w", err)
	}

	return announcement, nil
}

// Archive 归档公告
func (s *AnnouncementService) Archive(ctx context.Context, id int64) (*Announcement, error) {
	announcement, err := s.announcementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 只有已发布状态的公告可以归档
	if announcement.Status != AnnouncementStatusPublished {
		return nil, ErrAnnouncementInvalidStatusTransition.WithMetadata(map[string]string{
			"current_status": announcement.Status,
			"action":         "archive",
		})
	}

	announcement.Status = AnnouncementStatusArchived

	if err := s.announcementRepo.Update(ctx, announcement); err != nil {
		return nil, fmt.Errorf("archive announcement: %w", err)
	}

	return announcement, nil
}
