package schema

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserRedeemStat holds the schema definition for the UserRedeemStat entity.
// 用户兑换统计表，记录用户各金额兑换码的使用次数
type UserRedeemStat struct {
	ent.Schema
}

func (UserRedeemStat) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_redeem_stats"},
	}
}

func (UserRedeemStat) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id").
			Comment("用户ID"),
		field.String("redeem_type").
			MaxLen(20).
			Default(service.RedeemTypeBalance).
			Comment("兑换码类型"),
		field.Float("value").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Comment("兑换码面值"),
		field.Int("used_count").
			Default(0).
			Comment("使用次数"),
		field.Time("last_used_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("最后使用时间"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (UserRedeemStat) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("redeem_stats").
			Field("user_id").
			Required().
			Unique(),
	}
}

func (UserRedeemStat) Indexes() []ent.Index {
	return []ent.Index{
		// 同一用户、类型、金额的统计唯一
		index.Fields("user_id", "redeem_type", "value").Unique(),
	}
}
