package schema

import (
	"time"

	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Announcement holds the schema definition for the Announcement entity.
// 公告表，用于管理员发布公告通知。
type Announcement struct {
	ent.Schema
}

func (Announcement) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "announcements"},
	}
}

func (Announcement) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Announcement) Fields() []ent.Field {
	return []ent.Field{
		// 公告标题
		field.String("title").
			MaxLen(255).
			NotEmpty(),
		// 公告内容（支持 Markdown）
		field.String("content").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			NotEmpty(),
		// 公告类型：info-普通, warning-警告, important-重要
		field.String("type").
			MaxLen(20).
			Default("info"),
		// 优先级，数值越大越优先显示
		field.Int("priority").
			Default(0),
		// 状态：draft-草稿, published-已发布, archived-已归档
		field.String("status").
			MaxLen(20).
			Default("draft"),
		// 定时发布时间，NULL 表示立即发布
		field.Time("publish_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		// 过期时间，NULL 表示永不过期
		field.Time("expires_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		// 创建者用户ID
		field.Int64("created_by").
			Optional().
			Nillable(),
	}
}

func (Announcement) Edges() []ent.Edge {
	return []ent.Edge{
		// 已读记录
		edge.To("reads", AnnouncementRead.Type),
		// 创建者
		edge.From("creator", User.Type).
			Ref("announcements").
			Field("created_by").
			Unique(),
	}
}

func (Announcement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("publish_at"),
		index.Fields("expires_at"),
		index.Fields("priority"),
		index.Fields("deleted_at"),
		index.Fields("created_at"),
	}
}

// AnnouncementRead holds the schema definition for the AnnouncementRead entity.
// 公告已读记录表，记录用户阅读公告的状态。
type AnnouncementRead struct {
	ent.Schema
}

func (AnnouncementRead) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "announcement_reads"},
	}
}

func (AnnouncementRead) Fields() []ent.Field {
	return []ent.Field{
		// 用户ID
		field.Int64("user_id"),
		// 公告ID
		field.Int64("announcement_id"),
		// 阅读时间
		field.Time("read_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (AnnouncementRead) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联用户
		edge.From("user", User.Type).
			Ref("announcement_reads").
			Field("user_id").
			Required().
			Unique(),
		// 关联公告
		edge.From("announcement", Announcement.Type).
			Ref("reads").
			Field("announcement_id").
			Required().
			Unique(),
	}
}

func (AnnouncementRead) Indexes() []ent.Index {
	return []ent.Index{
		// 用户和公告的唯一组合
		index.Fields("user_id", "announcement_id").Unique(),
		index.Fields("user_id"),
		index.Fields("announcement_id"),
	}
}
