//go:build integration

package repository

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/suite"
)

type AnnouncementRepoSuite struct {
	IntegrationDBSuite
	repo   service.AnnouncementRepository
	userID int64
}

func (s *AnnouncementRepoSuite) SetupTest() {
	s.IntegrationDBSuite.SetupTest()
	s.repo = NewAnnouncementRepository(s.client)

	// 创建测试用户用于外键约束
	user := mustCreateUser(s.T(), s.client, &service.User{
		Email: "announcement-test@example.com",
	})
	s.userID = user.ID
}

func TestAnnouncementRepoSuite(t *testing.T) {
	suite.Run(t, new(AnnouncementRepoSuite))
}

func (s *AnnouncementRepoSuite) mustCreateAnnouncement(a *service.Announcement) *service.Announcement {
	s.T().Helper()

	if a.Title == "" {
		a.Title = "Test Announcement"
	}
	if a.Content == "" {
		a.Content = "Test content"
	}
	if a.Type == "" {
		a.Type = service.AnnouncementTypeInfo
	}
	if a.Status == "" {
		a.Status = service.AnnouncementStatusDraft
	}

	s.Require().NoError(s.repo.Create(s.ctx, a), "create announcement")
	return a
}

// ==================== Create 测试 ====================

func (s *AnnouncementRepoSuite) TestCreate() {
	createdBy := s.userID
	a := &service.Announcement{
		Title:     "Test Title",
		Content:   "Test Content",
		Type:      service.AnnouncementTypeWarning,
		Priority:  10,
		Status:    service.AnnouncementStatusDraft,
		CreatedBy: &createdBy,
	}

	err := s.repo.Create(s.ctx, a)
	s.Require().NoError(err)
	s.Require().NotZero(a.ID)
	s.Require().Equal("Test Title", a.Title)
	s.Require().Equal("Test Content", a.Content)
	s.Require().Equal(service.AnnouncementTypeWarning, a.Type)
	s.Require().Equal(10, a.Priority)
	s.Require().Equal(service.AnnouncementStatusDraft, a.Status)
	s.Require().NotNil(a.CreatedBy)
	s.Require().Equal(s.userID, *a.CreatedBy)
	s.Require().False(a.CreatedAt.IsZero())
	s.Require().False(a.UpdatedAt.IsZero())
}

func (s *AnnouncementRepoSuite) TestCreate_WithPublishAt() {
	publishAt := time.Now().Add(time.Hour)
	a := &service.Announcement{
		Title:     "Scheduled",
		Content:   "Content",
		Type:      service.AnnouncementTypeInfo,
		Status:    service.AnnouncementStatusDraft,
		PublishAt: &publishAt,
	}

	err := s.repo.Create(s.ctx, a)
	s.Require().NoError(err)
	s.Require().NotNil(a.PublishAt)
	s.Require().WithinDuration(publishAt, *a.PublishAt, time.Second)
}

func (s *AnnouncementRepoSuite) TestCreate_WithExpiresAt() {
	expiresAt := time.Now().Add(24 * time.Hour)
	a := &service.Announcement{
		Title:     "Expiring",
		Content:   "Content",
		Type:      service.AnnouncementTypeInfo,
		Status:    service.AnnouncementStatusDraft,
		ExpiresAt: &expiresAt,
	}

	err := s.repo.Create(s.ctx, a)
	s.Require().NoError(err)
	s.Require().NotNil(a.ExpiresAt)
	s.Require().WithinDuration(expiresAt, *a.ExpiresAt, time.Second)
}

// ==================== GetByID 测试 ====================

func (s *AnnouncementRepoSuite) TestGetByID() {
	created := s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Find Me",
		Content: "Content",
	})

	found, err := s.repo.GetByID(s.ctx, created.ID)
	s.Require().NoError(err)
	s.Require().Equal(created.ID, found.ID)
	s.Require().Equal("Find Me", found.Title)
}

func (s *AnnouncementRepoSuite) TestGetByID_NotFound() {
	_, err := s.repo.GetByID(s.ctx, 999999)
	s.Require().ErrorIs(err, service.ErrAnnouncementNotFound)
}

// ==================== Update 测试 ====================

func (s *AnnouncementRepoSuite) TestUpdate() {
	created := s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Original",
		Content: "Original Content",
	})

	created.Title = "Updated"
	created.Content = "Updated Content"
	created.Type = service.AnnouncementTypeImportant
	created.Priority = 100

	err := s.repo.Update(s.ctx, created)
	s.Require().NoError(err)

	found, err := s.repo.GetByID(s.ctx, created.ID)
	s.Require().NoError(err)
	s.Require().Equal("Updated", found.Title)
	s.Require().Equal("Updated Content", found.Content)
	s.Require().Equal(service.AnnouncementTypeImportant, found.Type)
	s.Require().Equal(100, found.Priority)
}

func (s *AnnouncementRepoSuite) TestUpdate_ClearPublishAt() {
	publishAt := time.Now().Add(time.Hour)
	created := s.mustCreateAnnouncement(&service.Announcement{
		Title:     "With PublishAt",
		Content:   "Content",
		PublishAt: &publishAt,
	})
	s.Require().NotNil(created.PublishAt)

	// 清除 PublishAt
	created.PublishAt = nil
	err := s.repo.Update(s.ctx, created)
	s.Require().NoError(err)

	found, err := s.repo.GetByID(s.ctx, created.ID)
	s.Require().NoError(err)
	s.Require().Nil(found.PublishAt)
}

func (s *AnnouncementRepoSuite) TestUpdate_ClearExpiresAt() {
	expiresAt := time.Now().Add(24 * time.Hour)
	created := s.mustCreateAnnouncement(&service.Announcement{
		Title:     "With ExpiresAt",
		Content:   "Content",
		ExpiresAt: &expiresAt,
	})
	s.Require().NotNil(created.ExpiresAt)

	// 清除 ExpiresAt
	created.ExpiresAt = nil
	err := s.repo.Update(s.ctx, created)
	s.Require().NoError(err)

	found, err := s.repo.GetByID(s.ctx, created.ID)
	s.Require().NoError(err)
	s.Require().Nil(found.ExpiresAt)
}

func (s *AnnouncementRepoSuite) TestUpdate_NotFound() {
	a := &service.Announcement{
		ID:      999999,
		Title:   "Not Exist",
		Content: "Content",
		Type:    service.AnnouncementTypeInfo,
		Status:  service.AnnouncementStatusDraft,
	}

	err := s.repo.Update(s.ctx, a)
	s.Require().ErrorIs(err, service.ErrAnnouncementNotFound)
}

// ==================== Delete 测试 ====================

func (s *AnnouncementRepoSuite) TestDelete() {
	created := s.mustCreateAnnouncement(&service.Announcement{
		Title:   "To Delete",
		Content: "Content",
	})

	err := s.repo.Delete(s.ctx, created.ID)
	s.Require().NoError(err)

	_, err = s.repo.GetByID(s.ctx, created.ID)
	s.Require().ErrorIs(err, service.ErrAnnouncementNotFound)
}

// ==================== List 测试 ====================

func (s *AnnouncementRepoSuite) TestList() {
	// 创建多个公告
	for i := 0; i < 5; i++ {
		s.mustCreateAnnouncement(&service.Announcement{
			Title:    "List Test",
			Content:  "Content",
			Priority: i,
		})
	}

	announcements, result, err := s.repo.List(s.ctx, pagination.PaginationParams{Page: 1, PageSize: 10}, "", "")
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(announcements), 5)
	s.Require().NotNil(result)
	s.Require().GreaterOrEqual(result.Total, int64(5))
}

func (s *AnnouncementRepoSuite) TestList_FilterByStatus() {
	// 创建草稿公告
	s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Draft",
		Content: "Content",
		Status:  service.AnnouncementStatusDraft,
	})

	// 创建已发布公告
	s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Published",
		Content: "Content",
		Status:  service.AnnouncementStatusPublished,
	})

	// 只查询草稿
	drafts, _, err := s.repo.List(s.ctx, pagination.PaginationParams{Page: 1, PageSize: 10}, service.AnnouncementStatusDraft, "")
	s.Require().NoError(err)
	for _, a := range drafts {
		s.Require().Equal(service.AnnouncementStatusDraft, a.Status)
	}

	// 只查询已发布
	published, _, err := s.repo.List(s.ctx, pagination.PaginationParams{Page: 1, PageSize: 10}, service.AnnouncementStatusPublished, "")
	s.Require().NoError(err)
	for _, a := range published {
		s.Require().Equal(service.AnnouncementStatusPublished, a.Status)
	}
}

func (s *AnnouncementRepoSuite) TestList_Search() {
	s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Unique Search Term ABC123",
		Content: "Content",
	})

	s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Other",
		Content: "Content with Unique Search Term ABC123",
	})

	s.mustCreateAnnouncement(&service.Announcement{
		Title:   "No Match",
		Content: "Nothing here",
	})

	// 搜索标题
	results, _, err := s.repo.List(s.ctx, pagination.PaginationParams{Page: 1, PageSize: 10}, "", "ABC123")
	s.Require().NoError(err)
	s.Require().Len(results, 2)
}

func (s *AnnouncementRepoSuite) TestList_Pagination() {
	// 创建 15 个公告
	for i := 0; i < 15; i++ {
		s.mustCreateAnnouncement(&service.Announcement{
			Title:   "Pagination Test",
			Content: "Content",
		})
	}

	// 第一页
	page1, result1, err := s.repo.List(s.ctx, pagination.PaginationParams{Page: 1, PageSize: 5}, "", "")
	s.Require().NoError(err)
	s.Require().Len(page1, 5)
	s.Require().GreaterOrEqual(result1.Total, int64(15))

	// 第二页
	page2, _, err := s.repo.List(s.ctx, pagination.PaginationParams{Page: 2, PageSize: 5}, "", "")
	s.Require().NoError(err)
	s.Require().Len(page2, 5)

	// 确保两页内容不同
	s.Require().NotEqual(page1[0].ID, page2[0].ID)
}

// ==================== GetActiveAnnouncements 测试 ====================

func (s *AnnouncementRepoSuite) TestGetActiveAnnouncements() {
	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	// 已发布，无时间限制
	s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Active No Time",
		Content: "Content",
		Status:  service.AnnouncementStatusPublished,
	})

	// 已发布，已过发布时间
	s.mustCreateAnnouncement(&service.Announcement{
		Title:     "Active Past Publish",
		Content:   "Content",
		Status:    service.AnnouncementStatusPublished,
		PublishAt: &past,
	})

	// 已发布，未到发布时间（不应出现）
	s.mustCreateAnnouncement(&service.Announcement{
		Title:     "Future Publish",
		Content:   "Content",
		Status:    service.AnnouncementStatusPublished,
		PublishAt: &future,
	})

	// 已发布，已过期（不应出现）
	s.mustCreateAnnouncement(&service.Announcement{
		Title:     "Expired",
		Content:   "Content",
		Status:    service.AnnouncementStatusPublished,
		ExpiresAt: &past,
	})

	// 草稿（不应出现）
	s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Draft",
		Content: "Content",
		Status:  service.AnnouncementStatusDraft,
	})

	// 已归档（不应出现）
	s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Archived",
		Content: "Content",
		Status:  service.AnnouncementStatusArchived,
	})

	active, err := s.repo.GetActiveAnnouncements(s.ctx, now)
	s.Require().NoError(err)
	s.Require().Len(active, 2)

	titles := make(map[string]bool)
	for _, a := range active {
		titles[a.Title] = true
	}
	s.Require().True(titles["Active No Time"])
	s.Require().True(titles["Active Past Publish"])
}

// ==================== GetUnreadAnnouncements 测试 ====================

func (s *AnnouncementRepoSuite) TestGetUnreadAnnouncements() {
	now := time.Now()
	userID := s.userID

	// 创建两个已发布公告
	a1 := s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Unread 1",
		Content: "Content",
		Status:  service.AnnouncementStatusPublished,
	})

	a2 := s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Unread 2",
		Content: "Content",
		Status:  service.AnnouncementStatusPublished,
	})

	// 标记第一个为已读
	err := s.repo.MarkAsRead(s.ctx, userID, a1.ID)
	s.Require().NoError(err)

	// 获取未读公告
	unread, err := s.repo.GetUnreadAnnouncements(s.ctx, userID, now)
	s.Require().NoError(err)
	s.Require().Len(unread, 1)
	s.Require().Equal(a2.ID, unread[0].ID)
}

// ==================== CountUnread 测试 ====================

func (s *AnnouncementRepoSuite) TestCountUnread() {
	now := time.Now()
	userID := s.userID

	// 创建 3 个已发布公告
	for i := 0; i < 3; i++ {
		s.mustCreateAnnouncement(&service.Announcement{
			Title:   "Count Test",
			Content: "Content",
			Status:  service.AnnouncementStatusPublished,
		})
	}

	count, err := s.repo.CountUnread(s.ctx, userID, now)
	s.Require().NoError(err)
	s.Require().Equal(3, count)
}

// ==================== GetUserReadIDs 测试 ====================

func (s *AnnouncementRepoSuite) TestGetUserReadIDs() {
	userID := s.userID

	a1 := s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Read 1",
		Content: "Content",
		Status:  service.AnnouncementStatusPublished,
	})

	a2 := s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Read 2",
		Content: "Content",
		Status:  service.AnnouncementStatusPublished,
	})

	// 标记两个为已读
	s.Require().NoError(s.repo.MarkAsRead(s.ctx, userID, a1.ID))
	s.Require().NoError(s.repo.MarkAsRead(s.ctx, userID, a2.ID))

	readIDs, err := s.repo.GetUserReadIDs(s.ctx, userID)
	s.Require().NoError(err)
	s.Require().Len(readIDs, 2)
	s.Require().Contains(readIDs, a1.ID)
	s.Require().Contains(readIDs, a2.ID)
}

// ==================== MarkAsRead 测试 ====================

func (s *AnnouncementRepoSuite) TestMarkAsRead() {
	userID := s.userID
	a := s.mustCreateAnnouncement(&service.Announcement{
		Title:   "To Mark",
		Content: "Content",
		Status:  service.AnnouncementStatusPublished,
	})

	err := s.repo.MarkAsRead(s.ctx, userID, a.ID)
	s.Require().NoError(err)

	readIDs, err := s.repo.GetUserReadIDs(s.ctx, userID)
	s.Require().NoError(err)
	s.Require().Contains(readIDs, a.ID)
}

func (s *AnnouncementRepoSuite) TestMarkAsRead_Idempotent() {
	userID := s.userID
	a := s.mustCreateAnnouncement(&service.Announcement{
		Title:   "Idempotent",
		Content: "Content",
		Status:  service.AnnouncementStatusPublished,
	})

	// 多次标记同一公告为已读
	for i := 0; i < 3; i++ {
		err := s.repo.MarkAsRead(s.ctx, userID, a.ID)
		s.Require().NoError(err)
	}

	readIDs, err := s.repo.GetUserReadIDs(s.ctx, userID)
	s.Require().NoError(err)
	s.Require().Len(readIDs, 1)
}

// ==================== MarkAllAsRead 测试 ====================

func (s *AnnouncementRepoSuite) TestMarkAllAsRead() {
	now := time.Now()
	userID := s.userID

	// 创建 5 个已发布公告
	var ids []int64
	for i := 0; i < 5; i++ {
		a := s.mustCreateAnnouncement(&service.Announcement{
			Title:   "Mark All",
			Content: "Content",
			Status:  service.AnnouncementStatusPublished,
		})
		ids = append(ids, a.ID)
	}

	// 标记全部已读
	err := s.repo.MarkAllAsRead(s.ctx, userID, now)
	s.Require().NoError(err)

	// 验证全部已读
	readIDs, err := s.repo.GetUserReadIDs(s.ctx, userID)
	s.Require().NoError(err)
	s.Require().Len(readIDs, 5)
	for _, id := range ids {
		s.Require().Contains(readIDs, id)
	}
}

func (s *AnnouncementRepoSuite) TestMarkAllAsRead_Empty() {
	now := time.Now()
	userID := s.userID

	// 没有公告时调用不应报错
	err := s.repo.MarkAllAsRead(s.ctx, userID, now)
	s.Require().NoError(err)
}

// ==================== BatchPublishScheduled 测试 ====================

func (s *AnnouncementRepoSuite) TestBatchPublishScheduled() {
	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	// 创建到期的定时发布公告
	s.mustCreateAnnouncement(&service.Announcement{
		Title:     "Should Publish 1",
		Content:   "Content",
		Status:    service.AnnouncementStatusDraft,
		PublishAt: &past,
	})

	s.mustCreateAnnouncement(&service.Announcement{
		Title:     "Should Publish 2",
		Content:   "Content",
		Status:    service.AnnouncementStatusDraft,
		PublishAt: &past,
	})

	// 创建未到期的定时发布公告
	s.mustCreateAnnouncement(&service.Announcement{
		Title:     "Should Not Publish",
		Content:   "Content",
		Status:    service.AnnouncementStatusDraft,
		PublishAt: &future,
	})

	// 创建没有定时发布的草稿
	s.mustCreateAnnouncement(&service.Announcement{
		Title:   "No Schedule",
		Content: "Content",
		Status:  service.AnnouncementStatusDraft,
	})

	// 执行批量发布
	count, err := s.repo.BatchPublishScheduled(s.ctx, now)
	s.Require().NoError(err)
	s.Require().Equal(2, count)

	// 验证状态变更
	announcements, _, err := s.repo.List(s.ctx, pagination.PaginationParams{Page: 1, PageSize: 100}, "", "")
	s.Require().NoError(err)

	for _, a := range announcements {
		switch a.Title {
		case "Should Publish 1", "Should Publish 2":
			s.Require().Equal(service.AnnouncementStatusPublished, a.Status)
		case "Should Not Publish", "No Schedule":
			s.Require().Equal(service.AnnouncementStatusDraft, a.Status)
		}
	}
}

func (s *AnnouncementRepoSuite) TestBatchPublishScheduled_NoMatches() {
	now := time.Now()

	// 没有符合条件的公告
	count, err := s.repo.BatchPublishScheduled(s.ctx, now)
	s.Require().NoError(err)
	s.Require().Equal(0, count)
}
