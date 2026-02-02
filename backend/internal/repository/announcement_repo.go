package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/announcement"
	"github.com/Wei-Shaw/sub2api/ent/announcementread"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type announcementRepository struct {
	client *dbent.Client
}

func NewAnnouncementRepository(client *dbent.Client) service.AnnouncementRepository {
	return &announcementRepository{client: client}
}

func (r *announcementRepository) Create(ctx context.Context, a *service.Announcement) error {
	client := clientFromContext(ctx, r.client)
	builder := client.Announcement.Create().
		SetTitle(a.Title).
		SetContent(a.Content).
		SetType(a.Type).
		SetPriority(a.Priority).
		SetStatus(a.Status)

	if a.PublishAt != nil {
		builder.SetPublishAt(*a.PublishAt)
	}
	if a.ExpiresAt != nil {
		builder.SetExpiresAt(*a.ExpiresAt)
	}
	if a.CreatedBy != nil {
		builder.SetCreatedBy(*a.CreatedBy)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	a.ID = created.ID
	a.CreatedAt = created.CreatedAt
	a.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *announcementRepository) GetByID(ctx context.Context, id int64) (*service.Announcement, error) {
	m, err := r.client.Announcement.Query().
		Where(announcement.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrAnnouncementNotFound
		}
		return nil, err
	}
	return announcementEntityToService(m), nil
}

func (r *announcementRepository) Update(ctx context.Context, a *service.Announcement) error {
	client := clientFromContext(ctx, r.client)
	builder := client.Announcement.UpdateOneID(a.ID).
		SetTitle(a.Title).
		SetContent(a.Content).
		SetType(a.Type).
		SetPriority(a.Priority).
		SetStatus(a.Status)

	if a.PublishAt != nil {
		builder.SetPublishAt(*a.PublishAt)
	} else {
		builder.ClearPublishAt()
	}
	if a.ExpiresAt != nil {
		builder.SetExpiresAt(*a.ExpiresAt)
	} else {
		builder.ClearExpiresAt()
	}

	updated, err := builder.Save(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrAnnouncementNotFound
		}
		return err
	}

	a.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *announcementRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.Announcement.Delete().Where(announcement.IDEQ(id)).Exec(ctx)
	return err
}

func (r *announcementRepository) List(ctx context.Context, params pagination.PaginationParams, status string, search string) ([]service.Announcement, *pagination.PaginationResult, error) {
	q := r.client.Announcement.Query()

	if status != "" {
		q = q.Where(announcement.StatusEQ(status))
	}

	// 搜索标题或内容
	if search != "" {
		q = q.Where(
			announcement.Or(
				announcement.TitleContainsFold(search),
				announcement.ContentContainsFold(search),
			),
		)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	announcements, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(announcement.FieldPriority), dbent.Desc(announcement.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	out := announcementEntitiesToService(announcements)
	return out, paginationResultFromTotal(int64(total), params), nil
}

func (r *announcementRepository) GetActiveAnnouncements(ctx context.Context, now time.Time) ([]service.Announcement, error) {
	announcements, err := r.client.Announcement.Query().
		Where(
			announcement.StatusEQ(service.AnnouncementStatusPublished),
			announcement.Or(
				announcement.PublishAtIsNil(),
				announcement.PublishAtLTE(now),
			),
			announcement.Or(
				announcement.ExpiresAtIsNil(),
				announcement.ExpiresAtGT(now),
			),
		).
		Order(dbent.Desc(announcement.FieldPriority), dbent.Desc(announcement.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return announcementEntitiesToService(announcements), nil
}

func (r *announcementRepository) GetUnreadAnnouncements(ctx context.Context, userID int64, now time.Time) ([]service.Announcement, error) {
	q := r.client.Announcement.Query().
		Where(
			announcement.StatusEQ(service.AnnouncementStatusPublished),
			announcement.Or(
				announcement.PublishAtIsNil(),
				announcement.PublishAtLTE(now),
			),
			announcement.Or(
				announcement.ExpiresAtIsNil(),
				announcement.ExpiresAtGT(now),
			),
		)

	// 排除已读的公告：NOT EXISTS (announcement_reads where user_id = ?)
	q = q.Where(
		announcement.Not(
			announcement.HasReadsWith(
				announcementread.UserIDEQ(userID),
			),
		),
	)

	announcements, err := q.
		Order(dbent.Desc(announcement.FieldPriority), dbent.Desc(announcement.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return announcementEntitiesToService(announcements), nil
}

func (r *announcementRepository) CountUnread(ctx context.Context, userID int64, now time.Time) (int, error) {
	q := r.client.Announcement.Query().
		Where(
			announcement.StatusEQ(service.AnnouncementStatusPublished),
			announcement.Or(
				announcement.PublishAtIsNil(),
				announcement.PublishAtLTE(now),
			),
			announcement.Or(
				announcement.ExpiresAtIsNil(),
				announcement.ExpiresAtGT(now),
			),
		)

	// 排除已读的公告：NOT EXISTS (announcement_reads where user_id = ?)
	q = q.Where(
		announcement.Not(
			announcement.HasReadsWith(
				announcementread.UserIDEQ(userID),
			),
		),
	)

	return q.Count(ctx)
}

func (r *announcementRepository) GetUserReadIDs(ctx context.Context, userID int64) ([]int64, error) {
	reads, err := r.client.AnnouncementRead.Query().
		Where(announcementread.UserIDEQ(userID)).
		Select(announcementread.FieldAnnouncementID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(reads))
	for i, read := range reads {
		ids[i] = read.AnnouncementID
	}
	return ids, nil
}

func (r *announcementRepository) MarkAsRead(ctx context.Context, userID, announcementID int64) error {
	client := clientFromContext(ctx, r.client)

	// 使用 OnConflictColumns 实现 upsert，避免重复插入
	err := client.AnnouncementRead.Create().
		SetUserID(userID).
		SetAnnouncementID(announcementID).
		SetReadAt(time.Now()).
		OnConflictColumns(announcementread.FieldUserID, announcementread.FieldAnnouncementID).
		UpdateNewValues().
		Exec(ctx)

	return err
}

func (r *announcementRepository) MarkAllAsRead(ctx context.Context, userID int64, now time.Time) error {
	client := clientFromContext(ctx, r.client)

	// 获取所有未读的有效公告
	unreadAnnouncements, err := r.GetUnreadAnnouncements(ctx, userID, now)
	if err != nil {
		return err
	}

	if len(unreadAnnouncements) == 0 {
		return nil
	}

	// 使用批量创建
	readAt := time.Now()
	builders := make([]*dbent.AnnouncementReadCreate, 0, len(unreadAnnouncements))
	for _, a := range unreadAnnouncements {
		builders = append(builders, client.AnnouncementRead.Create().
			SetUserID(userID).
			SetAnnouncementID(a.ID).
			SetReadAt(readAt))
	}

	// 批量插入，忽略冲突
	return client.AnnouncementRead.CreateBulk(builders...).
		OnConflictColumns(announcementread.FieldUserID, announcementread.FieldAnnouncementID).
		DoNothing().
		Exec(ctx)
}

// BatchPublishScheduled 批量发布到期的定时公告
func (r *announcementRepository) BatchPublishScheduled(ctx context.Context, now time.Time) (int, error) {
	client := clientFromContext(ctx, r.client)

	// 更新所有 status=draft 且 publish_at <= now 的公告为 published
	affected, err := client.Announcement.Update().
		Where(
			announcement.StatusEQ(service.AnnouncementStatusDraft),
			announcement.PublishAtNotNil(),
			announcement.PublishAtLTE(now),
		).
		SetStatus(service.AnnouncementStatusPublished).
		Save(ctx)

	return affected, err
}

// Entity to Service conversions

func announcementEntityToService(m *dbent.Announcement) *service.Announcement {
	if m == nil {
		return nil
	}
	return &service.Announcement{
		ID:        m.ID,
		Title:     m.Title,
		Content:   m.Content,
		Type:      m.Type,
		Priority:  m.Priority,
		Status:    m.Status,
		PublishAt: m.PublishAt,
		ExpiresAt: m.ExpiresAt,
		CreatedBy: m.CreatedBy,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func announcementEntitiesToService(models []*dbent.Announcement) []service.Announcement {
	out := make([]service.Announcement, 0, len(models))
	for i := range models {
		if s := announcementEntityToService(models[i]); s != nil {
			out = append(out, *s)
		}
	}
	return out
}
