package repository

import (
	"context"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// PaymentOrderRepository 支付订单数据访问接口
type PaymentOrderRepository interface {
	Create(ctx context.Context, order *service.PaymentOrder) error
	GetByID(ctx context.Context, id int64) (*service.PaymentOrder, error)
	GetByTradeNo(ctx context.Context, tradeNo string) (*service.PaymentOrder, error)
	GetPendingByAmount(ctx context.Context, amount float64, withinMinutes int) ([]*service.PaymentOrder, error)
	GetPendingByAmountAfterTime(ctx context.Context, amount float64, billTime time.Time, withinMinutes int) ([]*service.PaymentOrder, error)
	GetUsedAmounts(ctx context.Context, baseAmount float64, withinMinutes int) ([]float64, error)
	IsTransLogIDUsed(ctx context.Context, transLogID string) (bool, error)
	UpdateToPaid(ctx context.Context, id int64, alipayTradeNo string, transLogID string, payerAccount string, creditAmount float64) error
	UpdateToExpired(ctx context.Context, id int64) error
	Cancel(ctx context.Context, id int64) error
	List(ctx context.Context, params pagination.PaginationParams, userID int64, status string) ([]service.PaymentOrder, *pagination.PaginationResult, error)
	ListAll(ctx context.Context, params pagination.PaginationParams, status, search string) ([]service.PaymentOrder, *pagination.PaginationResult, error)
	CleanupExpiredOrders(ctx context.Context) (int64, error)
	GetStats(ctx context.Context) (*service.PaymentStats, error)
}

type paymentOrderRepository struct {
	client *dbent.Client
}

// NewPaymentOrderRepository 创建支付订单数据访问层
func NewPaymentOrderRepository(client *dbent.Client) *paymentOrderRepository {
	return &paymentOrderRepository{client: client}
}

func (r *paymentOrderRepository) Create(ctx context.Context, order *service.PaymentOrder) error {
	client := clientFromContext(ctx, r.client)
	builder := client.PaymentOrder.Create().
		SetTradeNo(order.TradeNo).
		SetUserID(order.UserID).
		SetAmount(order.Amount).
		SetPaymentAmount(order.PaymentAmount).
		SetStatus(order.Status).
		SetExpiredAt(order.ExpiredAt)

	if order.Notes != nil {
		builder.SetNotes(*order.Notes)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	order.ID = created.ID
	order.CreatedAt = created.CreatedAt
	return nil
}

func (r *paymentOrderRepository) GetByID(ctx context.Context, id int64) (*service.PaymentOrder, error) {
	m, err := r.client.PaymentOrder.Query().
		Where(paymentorder.IDEQ(id)).
		WithUser().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, fmt.Errorf("payment order not found")
		}
		return nil, err
	}
	return paymentOrderEntityToService(m), nil
}

func (r *paymentOrderRepository) GetByTradeNo(ctx context.Context, tradeNo string) (*service.PaymentOrder, error) {
	m, err := r.client.PaymentOrder.Query().
		Where(paymentorder.TradeNoEQ(tradeNo)).
		WithUser().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, fmt.Errorf("payment order not found")
		}
		return nil, err
	}
	return paymentOrderEntityToService(m), nil
}

func (r *paymentOrderRepository) GetPendingByAmount(ctx context.Context, amount float64, withinMinutes int) ([]*service.PaymentOrder, error) {
	threshold := time.Now().Add(-time.Duration(withinMinutes) * time.Minute)

	// 使用范围比较替代精确相等，避免浮点精度问题
	const epsilon = 0.001
	orders, err := r.client.PaymentOrder.Query().
		Where(
			paymentorder.StatusEQ(service.PaymentStatusPending),
			paymentorder.PaymentAmountGTE(amount-epsilon),
			paymentorder.PaymentAmountLTE(amount+epsilon),
			paymentorder.CreatedAtGTE(threshold),
		).
		Order(dbent.Asc(paymentorder.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*service.PaymentOrder, 0, len(orders))
	for _, o := range orders {
		result = append(result, paymentOrderEntityToService(o))
	}
	return result, nil
}

func (r *paymentOrderRepository) GetPendingByAmountAfterTime(ctx context.Context, amount float64, billTime time.Time, withinMinutes int) ([]*service.PaymentOrder, error) {
	threshold := time.Now().Add(-time.Duration(withinMinutes) * time.Minute)

	// 使用范围比较替代精确相等，避免浮点精度问题
	const epsilon = 0.001
	orders, err := r.client.PaymentOrder.Query().
		Where(
			paymentorder.StatusEQ(service.PaymentStatusPending),
			paymentorder.PaymentAmountGTE(amount-epsilon),
			paymentorder.PaymentAmountLTE(amount+epsilon),
			paymentorder.CreatedAtGTE(threshold),
			paymentorder.CreatedAtLTE(billTime),   // 订单创建时间必须早于账单时间
			paymentorder.ExpiredAtGT(time.Now()), // 未过期
		).
		Order(dbent.Asc(paymentorder.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*service.PaymentOrder, 0, len(orders))
	for _, o := range orders {
		result = append(result, paymentOrderEntityToService(o))
	}
	return result, nil
}

func (r *paymentOrderRepository) IsTransLogIDUsed(ctx context.Context, transLogID string) (bool, error) {
	count, err := r.client.PaymentOrder.Query().
		Where(paymentorder.AlipayTransLogIDEQ(transLogID)).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *paymentOrderRepository) GetUsedAmounts(ctx context.Context, baseAmount float64, withinMinutes int) ([]float64, error) {
	threshold := time.Now().Add(-time.Duration(withinMinutes) * time.Minute)

	// 查询在指定时间范围内的所有待支付订单的实际支付金额
	// 金额范围：baseAmount 到 baseAmount + 1（因为偏移量最多 0.99）
	orders, err := r.client.PaymentOrder.Query().
		Where(
			paymentorder.StatusEQ(service.PaymentStatusPending),
			paymentorder.PaymentAmountGTE(baseAmount),
			paymentorder.PaymentAmountLT(baseAmount+1),
			paymentorder.CreatedAtGTE(threshold),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	amounts := make([]float64, 0, len(orders))
	for _, o := range orders {
		amounts = append(amounts, o.PaymentAmount)
	}
	return amounts, nil
}

func (r *paymentOrderRepository) UpdateToPaid(ctx context.Context, id int64, alipayTradeNo string, transLogID string, payerAccount string, creditAmount float64) error {
	now := time.Now()
	client := clientFromContext(ctx, r.client)
	builder := client.PaymentOrder.Update().
		Where(
			paymentorder.IDEQ(id),
			paymentorder.StatusIn(service.PaymentStatusPending, service.PaymentStatusExpired),
		).
		SetStatus(service.PaymentStatusPaid).
		SetPaidAt(now).
		SetAlipayTradeNo(alipayTradeNo).
		SetCreditAmount(creditAmount)

	// 保存账单流水号（如果提供）
	if transLogID != "" {
		builder.SetAlipayTransLogID(transLogID)
	}

	// 保存付款人账户（如果提供）
	if payerAccount != "" {
		builder.SetPayerAccount(payerAccount)
	}

	affected, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("order not found or already processed")
	}
	return nil
}

func (r *paymentOrderRepository) UpdateToExpired(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	affected, err := client.PaymentOrder.Update().
		Where(
			paymentorder.IDEQ(id),
			paymentorder.StatusEQ(service.PaymentStatusPending),
		).
		SetStatus(service.PaymentStatusExpired).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("order not found or already processed")
	}
	return nil
}

func (r *paymentOrderRepository) Cancel(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	affected, err := client.PaymentOrder.Update().
		Where(
			paymentorder.IDEQ(id),
			paymentorder.StatusEQ(service.PaymentStatusPending),
		).
		SetStatus(service.PaymentStatusCancelled).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("order not found or already processed")
	}
	return nil
}

func (r *paymentOrderRepository) List(ctx context.Context, params pagination.PaginationParams, userID int64, status string) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	q := r.client.PaymentOrder.Query().
		Where(paymentorder.UserIDEQ(userID))

	if status != "" {
		q = q.Where(paymentorder.StatusEQ(status))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(paymentorder.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return paymentOrderEntitiesToService(orders), paginationResultFromTotal(int64(total), params), nil
}

func (r *paymentOrderRepository) ListAll(ctx context.Context, params pagination.PaginationParams, status, search string) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	q := r.client.PaymentOrder.Query()

	if status != "" {
		q = q.Where(paymentorder.StatusEQ(status))
	}
	if search != "" {
		q = q.Where(paymentorder.TradeNoContainsFold(search))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.
		WithUser().
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(paymentorder.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return paymentOrderEntitiesToService(orders), paginationResultFromTotal(int64(total), params), nil
}

func (r *paymentOrderRepository) CleanupExpiredOrders(ctx context.Context) (int64, error) {
	// 将所有已过期的待支付订单标记为过期状态
	affected, err := r.client.PaymentOrder.Update().
		Where(
			paymentorder.StatusEQ(service.PaymentStatusPending),
			paymentorder.ExpiredAtLT(time.Now()),
		).
		SetStatus(service.PaymentStatusExpired).
		Save(ctx)

	return int64(affected), err
}

func (r *paymentOrderRepository) GetStats(ctx context.Context) (*service.PaymentStats, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	stats := &service.PaymentStats{}

	// 今日订单数
	todayCount, err := r.client.PaymentOrder.Query().
		Where(paymentorder.CreatedAtGTE(todayStart)).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.TodayOrderCount = int64(todayCount)

	// 今日已支付订单数和金额（按支付时间筛选）
	todayPaidOrders, err := r.client.PaymentOrder.Query().
		Where(
			paymentorder.PaidAtGTE(todayStart),
			paymentorder.StatusEQ(service.PaymentStatusPaid),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	stats.TodayPaidCount = int64(len(todayPaidOrders))
	for _, o := range todayPaidOrders {
		stats.TodayPaidAmount += o.PaymentAmount
	}

	// 总订单数
	totalCount, err := r.client.PaymentOrder.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalOrderCount = int64(totalCount)

	// 总已支付订单数和金额
	allPaidOrders, err := r.client.PaymentOrder.Query().
		Where(paymentorder.StatusEQ(service.PaymentStatusPaid)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalPaidCount = int64(len(allPaidOrders))
	for _, o := range allPaidOrders {
		stats.TotalPaidAmount += o.PaymentAmount
	}

	// 待支付订单数
	pendingCount, err := r.client.PaymentOrder.Query().
		Where(paymentorder.StatusEQ(service.PaymentStatusPending)).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.PendingOrderCount = int64(pendingCount)

	return stats, nil
}

func paymentOrderEntityToService(m *dbent.PaymentOrder) *service.PaymentOrder {
	if m == nil {
		return nil
	}
	out := &service.PaymentOrder{
		ID:               m.ID,
		TradeNo:          m.TradeNo,
		UserID:           m.UserID,
		Amount:           m.Amount,
		PaymentAmount:    m.PaymentAmount,
		CreditAmount:     m.CreditAmount,
		Status:           m.Status,
		CreatedAt:        m.CreatedAt,
		PaidAt:           m.PaidAt,
		ExpiredAt:        m.ExpiredAt,
		AlipayTradeNo:    m.AlipayTradeNo,
		AlipayTransLogID: m.AlipayTransLogID,
		PayerAccount:     m.PayerAccount,
		Notes:            m.Notes,
	}
	if m.Edges.User != nil {
		out.User = userEntityToService(m.Edges.User)
	}
	return out
}

func paymentOrderEntitiesToService(models []*dbent.PaymentOrder) []service.PaymentOrder {
	out := make([]service.PaymentOrder, 0, len(models))
	for _, m := range models {
		if s := paymentOrderEntityToService(m); s != nil {
			out = append(out, *s)
		}
	}
	return out
}
