package schema

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RedeemRule holds the schema definition for the RedeemRule entity.
// 兑换规则表，用于配置试用码的降级策略
type RedeemRule struct {
	ent.Schema
}

func (RedeemRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "redeem_rules"},
	}
}

func (RedeemRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("type").
			MaxLen(20).
			Default(service.RedeemTypeBalance).
			Comment("兑换码类型：balance/concurrency/subscription"),
		field.Float("trigger_value").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Comment("触发金额，如 5.00"),
		field.Int("max_times_per_user").
			Default(1).
			Comment("每用户可享受原价的次数"),
		field.Float("fallback_value").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Comment("降级金额，如 2.00"),
		field.Bool("is_active").
			Default(true).
			Comment("规则是否启用"),
		field.String("description").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Comment("规则说明"),
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

func (RedeemRule) Edges() []ent.Edge {
	return nil
}

func (RedeemRule) Indexes() []ent.Index {
	return []ent.Index{
		// 同一类型和触发金额的规则唯一
		index.Fields("type", "trigger_value").Unique(),
		index.Fields("is_active"),
	}
}
