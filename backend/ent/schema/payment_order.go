package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PaymentOrder holds the schema definition for the PaymentOrder entity.
// 支付充值订单表，用于记录用户通过支付宝充值的订单。
type PaymentOrder struct {
	ent.Schema
}

func (PaymentOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "payment_orders"},
	}
}

func (PaymentOrder) Fields() []ent.Field {
	return []ent.Field{
		// 系统交易号，唯一标识
		field.String("trade_no").
			MaxLen(32).
			NotEmpty().
			Unique(),
		// 用户ID
		field.Int64("user_id").
			Positive(),
		// 用户请求充值金额
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Positive(),
		// 实际支付金额（含偏移量，用于金额匹配）
		field.Float("payment_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Positive(),
		// 订单状态：pending/paid/expired/cancelled
		field.String("status").
			MaxLen(20).
			Default("pending"),
		// 创建时间
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		// 支付时间
		field.Time("paid_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		// 过期时间
		field.Time("expired_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		// 支付宝交易号（支付成功后填入）
		field.String("alipay_trade_no").
			MaxLen(64).
			Optional().
			Nillable(),
		// 支付宝账单流水号（用于防止重复匹配同一笔账单）
		field.String("alipay_trans_log_id").
			MaxLen(64).
			Optional().
			Nillable().
			Unique(),
		// 付款人账户（支付成功后从账单中获取）
		field.String("payer_account").
			MaxLen(128).
			Optional().
			Nillable(),
		// 实际到账金额（支付成功后计算存储）
		field.Float("credit_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Optional().
			Nillable(),
		// 创建订单时的充值系数（用于计算到账金额）
		field.Float("rate_coefficient").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}).
			Default(1.0),
		// 备注
		field.String("notes").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
	}
}

func (PaymentOrder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("payment_orders").
			Field("user_id").
			Required().
			Unique(),
	}
}

func (PaymentOrder) Indexes() []ent.Index {
	return []ent.Index{
		// trade_no 已在 Fields() 中声明 Unique()
		index.Fields("user_id"),
		index.Fields("status"),
		// 复合索引：用于按金额匹配待支付订单
		index.Fields("payment_amount", "status"),
		index.Fields("expired_at"),
		index.Fields("created_at"),
	}
}
