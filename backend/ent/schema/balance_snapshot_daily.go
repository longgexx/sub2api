package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// BalanceSnapshotDaily holds the schema definition for daily balance snapshots.
//
// 用于存储每日用户总余额的历史快照，支持余额趋势图功能。
type BalanceSnapshotDaily struct {
	ent.Schema
}

func (BalanceSnapshotDaily) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "balance_snapshots_daily"},
	}
}

func (BalanceSnapshotDaily) Fields() []ent.Field {
	return []ent.Field{
		field.Time("snapshot_date").
			SchemaType(map[string]string{
				dialect.Postgres: "date",
			}).
			Unique(),
		field.Float("total_balance").
			SchemaType(map[string]string{
				dialect.Postgres: "decimal(20,8)",
			}).
			Default(0),
		field.Int64("user_count").
			Default(0),
		field.Time("computed_at").
			Default(time.Now).
			SchemaType(map[string]string{
				dialect.Postgres: "timestamptz",
			}),
	}
}

func (BalanceSnapshotDaily) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("snapshot_date"),
	}
}
