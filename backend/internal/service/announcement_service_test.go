//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

// announcementRepoStub 公告仓储桩
type announcementRepoStub struct {
	announcements       map[int64]*Announcement
	readRecords         map[int64]map[int64]bool // userID -> announcementID -> read
	createErr           error
	updateErr           error
	deleteErr           error
	getByIDErr          error
	markAsReadErr       error
	markAllAsReadErr    error
	batchPublishErr     error
	batchPublishCount   int
	nextID              int64
}

func newAnnouncementRepoStub() *announcementRepoStub {
	return &announcementRepoStub{
		announcements: make(map[int64]*Announcement),
		readRecords:   make(map[int64]map[int64]bool),
		nextID:        1,
	}
}

func (r *announcementRepoStub) Create(_ context.Context, a *Announcement) error {
	if r.createErr != nil {
		return r.createErr
	}
	a.ID = r.nextID
	r.nextID++
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
	r.announcements[a.ID] = a
	return nil
}

func (r *announcementRepoStub) GetByID(_ context.Context, id int64) (*Announcement, error) {
	if r.getByIDErr != nil {
		return nil, r.getByIDErr
	}
	a, ok := r.announcements[id]
	if !ok {
		return nil, ErrAnnouncementNotFound
	}
	return a, nil
}

func (r *announcementRepoStub) Update(_ context.Context, a *Announcement) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	if _, ok := r.announcements[a.ID]; !ok {
		return ErrAnnouncementNotFound
	}
	a.UpdatedAt = time.Now()
	r.announcements[a.ID] = a
	return nil
}

func (r *announcementRepoStub) Delete(_ context.Context, id int64) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	if _, ok := r.announcements[id]; !ok {
		return ErrAnnouncementNotFound
	}
	delete(r.announcements, id)
	return nil
}

func (r *announcementRepoStub) List(_ context.Context, _ pagination.PaginationParams, status string, _ string) ([]Announcement, *pagination.PaginationResult, error) {
	var result []Announcement
	for _, a := range r.announcements {
		if status == "" || a.Status == status {
			result = append(result, *a)
		}
	}
	return result, &pagination.PaginationResult{Total: int64(len(result))}, nil
}

func (r *announcementRepoStub) GetActiveAnnouncements(_ context.Context, now time.Time) ([]Announcement, error) {
	var result []Announcement
	for _, a := range r.announcements {
		if a.Status == AnnouncementStatusPublished {
			if a.PublishAt != nil && a.PublishAt.After(now) {
				continue
			}
			if a.ExpiresAt != nil && !a.ExpiresAt.After(now) {
				continue
			}
			result = append(result, *a)
		}
	}
	return result, nil
}

func (r *announcementRepoStub) GetUnreadAnnouncements(_ context.Context, userID int64, now time.Time) ([]Announcement, error) {
	var result []Announcement
	userReads := r.readRecords[userID]
	for _, a := range r.announcements {
		if a.Status == AnnouncementStatusPublished {
			if a.PublishAt != nil && a.PublishAt.After(now) {
				continue
			}
			if a.ExpiresAt != nil && !a.ExpiresAt.After(now) {
				continue
			}
			if userReads != nil && userReads[a.ID] {
				continue
			}
			result = append(result, *a)
		}
	}
	return result, nil
}

func (r *announcementRepoStub) CountUnread(_ context.Context, userID int64, now time.Time) (int, error) {
	announcements, _ := r.GetUnreadAnnouncements(context.Background(), userID, now)
	return len(announcements), nil
}

func (r *announcementRepoStub) GetUserReadIDs(_ context.Context, userID int64) ([]int64, error) {
	var ids []int64
	if userReads, ok := r.readRecords[userID]; ok {
		for id := range userReads {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (r *announcementRepoStub) MarkAsRead(_ context.Context, userID, announcementID int64) error {
	if r.markAsReadErr != nil {
		return r.markAsReadErr
	}
	if r.readRecords[userID] == nil {
		r.readRecords[userID] = make(map[int64]bool)
	}
	r.readRecords[userID][announcementID] = true
	return nil
}

func (r *announcementRepoStub) MarkAllAsRead(_ context.Context, userID int64, _ time.Time) error {
	if r.markAllAsReadErr != nil {
		return r.markAllAsReadErr
	}
	if r.readRecords[userID] == nil {
		r.readRecords[userID] = make(map[int64]bool)
	}
	for id, a := range r.announcements {
		if a.Status == AnnouncementStatusPublished {
			r.readRecords[userID][id] = true
		}
	}
	return nil
}

func (r *announcementRepoStub) BatchPublishScheduled(_ context.Context, _ time.Time) (int, error) {
	if r.batchPublishErr != nil {
		return 0, r.batchPublishErr
	}
	return r.batchPublishCount, nil
}

// ==================== 辅助函数测试 ====================

func TestNormalizeAnnouncementType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"empty returns info", "", AnnouncementTypeInfo, false},
		{"info", "info", AnnouncementTypeInfo, false},
		{"warning", "warning", AnnouncementTypeWarning, false},
		{"important", "important", AnnouncementTypeImportant, false},
		{"uppercase INFO", "INFO", AnnouncementTypeInfo, false},
		{"mixed case Warning", "Warning", AnnouncementTypeWarning, false},
		{"with spaces", "  important  ", AnnouncementTypeImportant, false},
		{"invalid type", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeAnnouncementType(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrAnnouncementInvalidType)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidateAnnouncementBasicFields(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		content string
		wantErr error
	}{
		{"valid", "Title", "Content", nil},
		{"empty title", "", "Content", ErrAnnouncementTitleRequired},
		{"whitespace title", "   ", "Content", ErrAnnouncementTitleRequired},
		{"empty content", "Title", "", ErrAnnouncementContentRequired},
		{"whitespace content", "Title", "   ", ErrAnnouncementContentRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAnnouncementBasicFields(tt.title, tt.content)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateAnnouncementTimeRange(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name      string
		publishAt *time.Time
		expiresAt *time.Time
		wantErr   bool
	}{
		{"both nil", nil, nil, false},
		{"only publishAt", &now, nil, false},
		{"only expiresAt", nil, &future, false},
		{"valid range", &past, &future, false},
		{"same time", &now, &now, true},
		{"expires before publish", &future, &past, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAnnouncementTimeRange(tt.publishAt, tt.expiresAt)
			if tt.wantErr {
				require.ErrorIs(t, err, ErrAnnouncementInvalidTimeRange)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsAnnouncementActiveAt(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name     string
		a        *Announcement
		expected bool
	}{
		{"nil announcement", nil, false},
		{"draft status", &Announcement{Status: AnnouncementStatusDraft}, false},
		{"archived status", &Announcement{Status: AnnouncementStatusArchived}, false},
		{"published no time constraints", &Announcement{Status: AnnouncementStatusPublished}, true},
		{"published future publish_at", &Announcement{Status: AnnouncementStatusPublished, PublishAt: &future}, false},
		{"published past publish_at", &Announcement{Status: AnnouncementStatusPublished, PublishAt: &past}, true},
		{"published expired", &Announcement{Status: AnnouncementStatusPublished, ExpiresAt: &past}, false},
		{"published not expired", &Announcement{Status: AnnouncementStatusPublished, ExpiresAt: &future}, true},
		{"published valid range", &Announcement{Status: AnnouncementStatusPublished, PublishAt: &past, ExpiresAt: &future}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAnnouncementActiveAt(tt.a, now)
			require.Equal(t, tt.expected, result)
		})
	}
}

// ==================== Service 方法测试 ====================

func TestAnnouncementService_Create_Success(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	input := &CreateAnnouncementInput{
		Title:    "Test Announcement",
		Content:  "Test content",
		Type:     "info",
		Priority: 10,
	}

	a, err := svc.Create(context.Background(), input, 1)
	require.NoError(t, err)
	require.NotNil(t, a)
	require.Equal(t, int64(1), a.ID)
	require.Equal(t, "Test Announcement", a.Title)
	require.Equal(t, "Test content", a.Content)
	require.Equal(t, AnnouncementTypeInfo, a.Type)
	require.Equal(t, 10, a.Priority)
	require.Equal(t, AnnouncementStatusDraft, a.Status)
	require.Equal(t, int64(1), *a.CreatedBy)
}

func TestAnnouncementService_Create_EmptyTitle(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	input := &CreateAnnouncementInput{
		Title:   "",
		Content: "Test content",
	}

	_, err := svc.Create(context.Background(), input, 1)
	require.ErrorIs(t, err, ErrAnnouncementTitleRequired)
}

func TestAnnouncementService_Create_EmptyContent(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	input := &CreateAnnouncementInput{
		Title:   "Test",
		Content: "",
	}

	_, err := svc.Create(context.Background(), input, 1)
	require.ErrorIs(t, err, ErrAnnouncementContentRequired)
}

func TestAnnouncementService_Create_InvalidType(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	input := &CreateAnnouncementInput{
		Title:   "Test",
		Content: "Content",
		Type:    "invalid",
	}

	_, err := svc.Create(context.Background(), input, 1)
	require.ErrorIs(t, err, ErrAnnouncementInvalidType)
}

func TestAnnouncementService_Create_NegativePriority(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	input := &CreateAnnouncementInput{
		Title:    "Test",
		Content:  "Content",
		Priority: -1,
	}

	_, err := svc.Create(context.Background(), input, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "priority must be >= 0")
}

func TestAnnouncementService_Create_InvalidTimeRange(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	input := &CreateAnnouncementInput{
		Title:     "Test",
		Content:   "Content",
		PublishAt: &future,
		ExpiresAt: &past,
	}

	_, err := svc.Create(context.Background(), input, 1)
	require.ErrorIs(t, err, ErrAnnouncementInvalidTimeRange)
}

func TestAnnouncementService_Create_RepoError(t *testing.T) {
	repo := newAnnouncementRepoStub()
	repo.createErr = errors.New("db error")
	svc := NewAnnouncementService(repo)

	input := &CreateAnnouncementInput{
		Title:   "Test",
		Content: "Content",
	}

	_, err := svc.Create(context.Background(), input, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "create announcement")
}

func TestAnnouncementService_GetByID_Success(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 先创建一个公告
	input := &CreateAnnouncementInput{Title: "Test", Content: "Content"}
	created, _ := svc.Create(context.Background(), input, 1)

	// 获取公告
	a, err := svc.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, a.ID)
}

func TestAnnouncementService_GetByID_NotFound(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	_, err := svc.GetByID(context.Background(), 999)
	require.ErrorIs(t, err, ErrAnnouncementNotFound)
}

func TestAnnouncementService_Update_Success(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建公告
	input := &CreateAnnouncementInput{Title: "Original", Content: "Content"}
	created, _ := svc.Create(context.Background(), input, 1)

	// 更新公告
	newTitle := "Updated"
	newContent := "New Content"
	updateInput := &UpdateAnnouncementInput{
		Title:   &newTitle,
		Content: &newContent,
	}

	updated, err := svc.Update(context.Background(), created.ID, updateInput)
	require.NoError(t, err)
	require.Equal(t, "Updated", updated.Title)
	require.Equal(t, "New Content", updated.Content)
}

func TestAnnouncementService_Update_ClearPublishAt(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建带发布时间的公告
	publishAt := time.Now().Add(time.Hour)
	input := &CreateAnnouncementInput{Title: "Test", Content: "Content", PublishAt: &publishAt}
	created, _ := svc.Create(context.Background(), input, 1)
	require.NotNil(t, created.PublishAt)

	// 清除发布时间
	updateInput := &UpdateAnnouncementInput{ClearPublishAt: true}
	updated, err := svc.Update(context.Background(), created.ID, updateInput)
	require.NoError(t, err)
	require.Nil(t, updated.PublishAt)
}

func TestAnnouncementService_Update_ClearExpiresAt(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建带过期时间的公告
	expiresAt := time.Now().Add(time.Hour)
	input := &CreateAnnouncementInput{Title: "Test", Content: "Content", ExpiresAt: &expiresAt}
	created, _ := svc.Create(context.Background(), input, 1)
	require.NotNil(t, created.ExpiresAt)

	// 清除过期时间
	updateInput := &UpdateAnnouncementInput{ClearExpiresAt: true}
	updated, err := svc.Update(context.Background(), created.ID, updateInput)
	require.NoError(t, err)
	require.Nil(t, updated.ExpiresAt)
}

func TestAnnouncementService_Update_NotFound(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	newTitle := "Updated"
	_, err := svc.Update(context.Background(), 999, &UpdateAnnouncementInput{Title: &newTitle})
	require.ErrorIs(t, err, ErrAnnouncementNotFound)
}

func TestAnnouncementService_Delete_Success(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建公告
	input := &CreateAnnouncementInput{Title: "Test", Content: "Content"}
	created, _ := svc.Create(context.Background(), input, 1)

	// 删除公告
	err := svc.Delete(context.Background(), created.ID)
	require.NoError(t, err)

	// 验证已删除
	_, err = svc.GetByID(context.Background(), created.ID)
	require.ErrorIs(t, err, ErrAnnouncementNotFound)
}

func TestAnnouncementService_Delete_NotFound(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	err := svc.Delete(context.Background(), 999)
	require.ErrorIs(t, err, ErrAnnouncementNotFound)
}

func TestAnnouncementService_Publish_Success(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建草稿公告
	input := &CreateAnnouncementInput{Title: "Test", Content: "Content"}
	created, _ := svc.Create(context.Background(), input, 1)
	require.Equal(t, AnnouncementStatusDraft, created.Status)

	// 发布公告
	published, err := svc.Publish(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, AnnouncementStatusPublished, published.Status)
}

func TestAnnouncementService_Publish_AlreadyPublished(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建并发布公告
	input := &CreateAnnouncementInput{Title: "Test", Content: "Content"}
	created, _ := svc.Create(context.Background(), input, 1)
	_, _ = svc.Publish(context.Background(), created.ID)

	// 再次发布应该失败
	_, err := svc.Publish(context.Background(), created.ID)
	require.ErrorIs(t, err, ErrAnnouncementInvalidStatusTransition)
}

func TestAnnouncementService_Publish_Archived(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建、发布、归档公告
	input := &CreateAnnouncementInput{Title: "Test", Content: "Content"}
	created, _ := svc.Create(context.Background(), input, 1)
	_, _ = svc.Publish(context.Background(), created.ID)
	_, _ = svc.Archive(context.Background(), created.ID)

	// 尝试发布已归档的公告应该失败
	_, err := svc.Publish(context.Background(), created.ID)
	require.ErrorIs(t, err, ErrAnnouncementInvalidStatusTransition)
}

func TestAnnouncementService_Archive_Success(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建并发布公告
	input := &CreateAnnouncementInput{Title: "Test", Content: "Content"}
	created, _ := svc.Create(context.Background(), input, 1)
	_, _ = svc.Publish(context.Background(), created.ID)

	// 归档公告
	archived, err := svc.Archive(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, AnnouncementStatusArchived, archived.Status)
}

func TestAnnouncementService_Archive_Draft(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建草稿公告
	input := &CreateAnnouncementInput{Title: "Test", Content: "Content"}
	created, _ := svc.Create(context.Background(), input, 1)

	// 尝试归档草稿应该失败
	_, err := svc.Archive(context.Background(), created.ID)
	require.ErrorIs(t, err, ErrAnnouncementInvalidStatusTransition)
}

func TestAnnouncementService_Archive_AlreadyArchived(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建、发布、归档公告
	input := &CreateAnnouncementInput{Title: "Test", Content: "Content"}
	created, _ := svc.Create(context.Background(), input, 1)
	_, _ = svc.Publish(context.Background(), created.ID)
	_, _ = svc.Archive(context.Background(), created.ID)

	// 再次归档应该失败
	_, err := svc.Archive(context.Background(), created.ID)
	require.ErrorIs(t, err, ErrAnnouncementInvalidStatusTransition)
}

func TestAnnouncementService_MarkAsRead_Success(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建并发布公告
	input := &CreateAnnouncementInput{Title: "Test", Content: "Content"}
	created, _ := svc.Create(context.Background(), input, 1)
	_, _ = svc.Publish(context.Background(), created.ID)

	// 标记已读
	err := svc.MarkAsRead(context.Background(), 100, created.ID)
	require.NoError(t, err)

	// 验证已读
	require.True(t, repo.readRecords[100][created.ID])
}

func TestAnnouncementService_MarkAsRead_NotActive(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建草稿公告（未发布）
	input := &CreateAnnouncementInput{Title: "Test", Content: "Content"}
	created, _ := svc.Create(context.Background(), input, 1)

	// 尝试标记未发布的公告为已读应该失败
	err := svc.MarkAsRead(context.Background(), 100, created.ID)
	require.ErrorIs(t, err, ErrAnnouncementNotFound)
}

func TestAnnouncementService_MarkAllAsRead_Success(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建并发布多个公告
	for i := 0; i < 3; i++ {
		input := &CreateAnnouncementInput{Title: "Test", Content: "Content"}
		created, _ := svc.Create(context.Background(), input, 1)
		_, _ = svc.Publish(context.Background(), created.ID)
	}

	// 标记全部已读
	err := svc.MarkAllAsRead(context.Background(), 100)
	require.NoError(t, err)

	// 验证全部已读
	require.Len(t, repo.readRecords[100], 3)
}

func TestAnnouncementService_GetUnreadCount(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建并发布多个公告
	for i := 0; i < 5; i++ {
		input := &CreateAnnouncementInput{Title: "Test", Content: "Content"}
		created, _ := svc.Create(context.Background(), input, 1)
		_, _ = svc.Publish(context.Background(), created.ID)
	}

	// 获取未读数量
	count, err := svc.GetUnreadCount(context.Background(), 100)
	require.NoError(t, err)
	require.Equal(t, 5, count)

	// 标记一个已读
	_ = svc.MarkAsRead(context.Background(), 100, 1)

	// 再次获取未读数量
	count, err = svc.GetUnreadCount(context.Background(), 100)
	require.NoError(t, err)
	require.Equal(t, 4, count)
}

func TestAnnouncementService_GetActiveAnnouncements(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建并发布公告
	input := &CreateAnnouncementInput{Title: "Test", Content: "Content"}
	created, _ := svc.Create(context.Background(), input, 1)
	_, _ = svc.Publish(context.Background(), created.ID)

	// 创建草稿公告（不应该出现在活跃列表中）
	_, _ = svc.Create(context.Background(), &CreateAnnouncementInput{Title: "Draft", Content: "Content"}, 1)

	// 获取活跃公告
	announcements, err := svc.GetActiveAnnouncements(context.Background(), 100)
	require.NoError(t, err)
	require.Len(t, announcements, 1)
	require.Equal(t, "Test", announcements[0].Title)
}

func TestAnnouncementService_GetUnreadAnnouncements(t *testing.T) {
	repo := newAnnouncementRepoStub()
	svc := NewAnnouncementService(repo)

	// 创建并发布两个公告
	input1 := &CreateAnnouncementInput{Title: "Test1", Content: "Content"}
	created1, _ := svc.Create(context.Background(), input1, 1)
	_, _ = svc.Publish(context.Background(), created1.ID)

	input2 := &CreateAnnouncementInput{Title: "Test2", Content: "Content"}
	created2, _ := svc.Create(context.Background(), input2, 1)
	_, _ = svc.Publish(context.Background(), created2.ID)

	// 标记第一个为已读
	_ = svc.MarkAsRead(context.Background(), 100, created1.ID)

	// 获取未读公告
	announcements, err := svc.GetUnreadAnnouncements(context.Background(), 100)
	require.NoError(t, err)
	require.Len(t, announcements, 1)
	require.Equal(t, "Test2", announcements[0].Title)
}
