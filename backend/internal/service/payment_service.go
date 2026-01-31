package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrPaymentOrderNotFound  = infraerrors.NotFound("PAYMENT_ORDER_NOT_FOUND", "payment order not found")
	ErrPaymentOrderExpired   = infraerrors.BadRequest("PAYMENT_ORDER_EXPIRED", "payment order has expired")
	ErrPaymentOrderPaid      = infraerrors.Conflict("PAYMENT_ORDER_PAID", "payment order already paid")
	ErrPaymentOrderCancelled = infraerrors.Conflict("PAYMENT_ORDER_CANCELLED", "payment order has been cancelled")
	ErrPaymentAmountInvalid  = infraerrors.BadRequest("PAYMENT_AMOUNT_INVALID", "payment amount is invalid")
	ErrPaymentDisabled       = infraerrors.BadRequest("PAYMENT_DISABLED", "payment is not enabled")
)

// PaymentOrderRepository 支付订单数据访问接口
type PaymentOrderRepository interface {
	Create(ctx context.Context, order *PaymentOrder) error
	GetByID(ctx context.Context, id int64) (*PaymentOrder, error)
	GetByTradeNo(ctx context.Context, tradeNo string) (*PaymentOrder, error)
	GetPendingByAmount(ctx context.Context, amount float64, withinMinutes int) ([]*PaymentOrder, error)
	GetPendingByAmountAfterTime(ctx context.Context, amount float64, billTime time.Time, withinMinutes int) ([]*PaymentOrder, error)
	GetUsedAmounts(ctx context.Context, baseAmount float64, withinMinutes int) ([]float64, error)
	IsTransLogIDUsed(ctx context.Context, transLogID string) (bool, error)
	UpdateToPaid(ctx context.Context, id int64, alipayTradeNo string, transLogID string, payerAccount string, creditAmount float64) error
	UpdateToExpired(ctx context.Context, id int64) error
	Cancel(ctx context.Context, id int64) error
	List(ctx context.Context, params pagination.PaginationParams, userID int64, status string) ([]PaymentOrder, *pagination.PaginationResult, error)
	ListAll(ctx context.Context, params pagination.PaginationParams, status, search string) ([]PaymentOrder, *pagination.PaginationResult, error)
	CleanupExpiredOrders(ctx context.Context) (int64, error)
	GetStats(ctx context.Context) (*PaymentStats, error)
}

// PaymentService 支付服务
type PaymentService struct {
	orderRepo            PaymentOrderRepository
	userRepo             UserRepository
	billingCacheService  *BillingCacheService
	authCacheInvalidator APIKeyAuthCacheInvalidator
	settingService       *SettingService
	entClient            *dbent.Client
	cfg                  *config.Config
	mu                   sync.Mutex // 用于金额分配
}

// NewPaymentService 创建支付服务
func NewPaymentService(
	orderRepo PaymentOrderRepository,
	userRepo UserRepository,
	billingCacheService *BillingCacheService,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
	settingService *SettingService,
	entClient *dbent.Client,
	cfg *config.Config,
) *PaymentService {
	return &PaymentService{
		orderRepo:            orderRepo,
		userRepo:             userRepo,
		billingCacheService:  billingCacheService,
		authCacheInvalidator: authCacheInvalidator,
		settingService:       settingService,
		entClient:            entClient,
		cfg:                  cfg,
	}
}

// GetConfig 获取支付配置
func (s *PaymentService) GetConfig(ctx context.Context) (*PaymentConfigResponse, error) {
	if !s.cfg.Payment.Enabled {
		return &PaymentConfigResponse{Enabled: false}, nil
	}

	// 获取系数和二维码配置
	rateCoefficient := 1.0
	qrCodeURL := s.cfg.Payment.Monitor.BusinessQRCode
	if s.settingService != nil {
		rateCoefficient = s.settingService.GetPaymentRateCoefficient(ctx)
		// 如果设置了自定义二维码，优先使用
		if customQR := s.settingService.GetPaymentQRCode(ctx); customQR != "" {
			qrCodeURL = customQR
		}
	}

	return &PaymentConfigResponse{
		Enabled:         true,
		MinAmount:       s.cfg.Payment.Monitor.MinAmount,
		MaxAmount:       s.cfg.Payment.Monitor.MaxAmount,
		QRCodeURL:       qrCodeURL,
		TimeoutMins:     s.cfg.Payment.Monitor.OrderTimeoutMins,
		RateCoefficient: rateCoefficient,
	}, nil
}

// PaymentConfigResponse 支付配置响应
type PaymentConfigResponse struct {
	Enabled         bool    `json:"enabled"`
	MinAmount       float64 `json:"min_amount"`
	MaxAmount       float64 `json:"max_amount"`
	QRCodeURL       string  `json:"qr_code_url"`
	TimeoutMins     int     `json:"timeout_mins"`
	RateCoefficient float64 `json:"rate_coefficient"` // 充值系数
}

// CreateOrder 创建充值订单
func (s *PaymentService) CreateOrder(ctx context.Context, userID int64, amount float64) (*PaymentOrder, error) {
	if !s.cfg.Payment.Enabled {
		return nil, ErrPaymentDisabled
	}

	// 验证金额范围
	if amount < s.cfg.Payment.Monitor.MinAmount || amount > s.cfg.Payment.Monitor.MaxAmount {
		return nil, ErrPaymentAmountInvalid
	}

	// 获取充值系数
	rateCoefficient := 1.0
	if s.settingService != nil {
		rateCoefficient = s.settingService.GetPaymentRateCoefficient(ctx)
	}

	// 计算基础支付金额（应用系数）并保留2位小数避免浮点精度问题
	basePaymentAmount := math.Round(amount*rateCoefficient*100) / 100

	// 分配唯一支付金额
	paymentAmount, err := s.allocateUniqueAmount(ctx, basePaymentAmount)
	if err != nil {
		return nil, fmt.Errorf("allocate payment amount: %w", err)
	}

	// 生成交易号
	tradeNo, err := s.generateTradeNo()
	if err != nil {
		return nil, fmt.Errorf("generate trade no: %w", err)
	}

	// 计算过期时间
	expiredAt := time.Now().Add(time.Duration(s.cfg.Payment.Monitor.OrderTimeoutMins) * time.Minute)

	order := &PaymentOrder{
		TradeNo:       tradeNo,
		UserID:        userID,
		Amount:        amount,
		PaymentAmount: paymentAmount,
		Status:        PaymentStatusPending,
		ExpiredAt:     expiredAt,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return order, nil
}

// allocateUniqueAmount 分配唯一支付金额
// 通过在基础金额上添加小数偏移量，确保同一时间段内每个订单的支付金额唯一
func (s *PaymentService) allocateUniqueAmount(ctx context.Context, baseAmount float64) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取最近 5 分钟内已使用的金额
	usedAmounts, err := s.orderRepo.GetUsedAmounts(ctx, baseAmount, 5)
	if err != nil {
		// 如果查询失败，返回基础金额
		return baseAmount, nil
	}

	// 创建已使用金额的 map
	usedMap := make(map[string]bool)
	for _, amt := range usedAmounts {
		usedMap[fmt.Sprintf("%.2f", amt)] = true
	}

	// 尝试分配唯一金额（基础金额 + 0.01 递增）
	amount := baseAmount
	for i := 0; i < 100; i++ {
		amountStr := fmt.Sprintf("%.2f", amount)
		if !usedMap[amountStr] {
			return amount, nil
		}
		amount += 0.01
	}

	// 如果超过 100 次尝试仍未找到，返回基础金额
	return baseAmount, nil
}

// generateTradeNo 生成交易号
// 格式：时间戳(14位) + 随机数(6位)
func (s *PaymentService) generateTradeNo() (string, error) {
	timestamp := time.Now().Format("20060102150405")

	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// 确保随机数为6位（取模1000000）
	randomNum := (int(b[0])<<16 | int(b[1])<<8 | int(b[2])) % 1000000
	return fmt.Sprintf("%s%06d", timestamp, randomNum), nil
}

// GetOrder 获取订单详情
func (s *PaymentService) GetOrder(ctx context.Context, userID int64, tradeNo string) (*PaymentOrder, error) {
	order, err := s.orderRepo.GetByTradeNo(ctx, tradeNo)
	if err != nil {
		return nil, ErrPaymentOrderNotFound
	}

	// 验证订单归属
	if order.UserID != userID {
		return nil, ErrPaymentOrderNotFound
	}

	return order, nil
}

// GetOrderByID 根据ID获取订单（管理员使用）
func (s *PaymentService) GetOrderByID(ctx context.Context, id int64) (*PaymentOrder, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrPaymentOrderNotFound
	}
	return order, nil
}

// CancelOrder 取消订单
func (s *PaymentService) CancelOrder(ctx context.Context, userID int64, tradeNo string) error {
	order, err := s.orderRepo.GetByTradeNo(ctx, tradeNo)
	if err != nil {
		return ErrPaymentOrderNotFound
	}

	// 验证订单归属
	if order.UserID != userID {
		return ErrPaymentOrderNotFound
	}

	// 只能取消待支付订单
	if !order.IsPending() {
		return infraerrors.Conflict("PAYMENT_ORDER_CANNOT_CANCEL", "only pending orders can be cancelled")
	}

	return s.orderRepo.Cancel(ctx, order.ID)
}

// ListUserOrders 获取用户订单列表
func (s *PaymentService) ListUserOrders(ctx context.Context, userID int64, params pagination.PaginationParams, status string) ([]PaymentOrder, *pagination.PaginationResult, error) {
	return s.orderRepo.List(ctx, params, userID, status)
}

// ListAllOrders 获取所有订单列表（管理员）
func (s *PaymentService) ListAllOrders(ctx context.Context, params pagination.PaginationParams, status, search string) ([]PaymentOrder, *pagination.PaginationResult, error) {
	return s.orderRepo.ListAll(ctx, params, status, search)
}

// ManualConfirm 手动确认支付（管理员）
// 允许确认任何未支付订单，包括已过期的订单（用于补录）
func (s *PaymentService) ManualConfirm(ctx context.Context, orderID int64) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return ErrPaymentOrderNotFound
	}

	// 只检查是否已支付，管理员可以确认过期订单（包括状态为 expired 的订单）
	if order.IsPaid() {
		return ErrPaymentOrderPaid
	}
	if order.Status == PaymentStatusCancelled {
		return ErrPaymentOrderCancelled
	}

	return s.completePayment(ctx, orderID, "MANUAL_CONFIRM", "", "", true)
}

// CompletePayment 完成支付
// 在事务中更新订单状态并增加用户余额
func (s *PaymentService) CompletePayment(ctx context.Context, orderID int64, alipayTradeNo string, transLogID string, payerAccount string) error {
	return s.completePayment(ctx, orderID, alipayTradeNo, transLogID, payerAccount, false)
}

func (s *PaymentService) completePayment(ctx context.Context, orderID int64, alipayTradeNo string, transLogID string, payerAccount string, allowExpired bool) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return ErrPaymentOrderNotFound
	}

	// 检查是否已过期（包括时间过期）
	if order.IsExpired() && !allowExpired {
		return ErrPaymentOrderExpired
	}
	if order.IsPaid() {
		return ErrPaymentOrderPaid
	}
	if order.Status == PaymentStatusCancelled {
		return ErrPaymentOrderCancelled
	}

	// 计算到账金额
	// 到账金额 = 请求金额 + 偏移量（按系数反推）
	creditAmount := s.calculateCreditAmount(ctx, order)

	// 使用数据库事务
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)

	// 更新订单状态
	if err := s.orderRepo.UpdateToPaid(txCtx, orderID, alipayTradeNo, transLogID, payerAccount, creditAmount); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	// 增加用户余额
	if err := s.userRepo.UpdateBalance(txCtx, order.UserID, creditAmount); err != nil {
		return fmt.Errorf("update user balance: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// 异步失效缓存
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if s.authCacheInvalidator != nil {
			s.authCacheInvalidator.InvalidateAuthCacheByUserID(cacheCtx, order.UserID)
		}
		if s.billingCacheService != nil {
			_ = s.billingCacheService.InvalidateUserBalance(cacheCtx, order.UserID)
		}
	}()

	return nil
}

// calculateCreditAmount 计算到账金额
// 到账金额 = 请求金额 + 偏移量（按系数反推）
// 例如：请求 $10，系数 0.5，支付 ¥5.01 → 到账 $10.02
func (s *PaymentService) calculateCreditAmount(ctx context.Context, order *PaymentOrder) float64 {
	// 获取充值系数
	rateCoefficient := 1.0
	if s.settingService != nil {
		rateCoefficient = s.settingService.GetPaymentRateCoefficient(ctx)
	}

	// 如果系数为1，直接返回实际支付金额（无需转换）
	if rateCoefficient == 1.0 {
		return order.PaymentAmount
	}

	// 计算基础支付金额（不含偏移）
	basePaymentAmount := math.Round(order.Amount*rateCoefficient*100) / 100

	// 计算支付偏移量
	paymentOffset := order.PaymentAmount - basePaymentAmount

	// 计算到账偏移量（按系数反推）
	var creditOffset float64
	if rateCoefficient > 0 {
		creditOffset = paymentOffset / rateCoefficient
	}

	// 到账金额 = 请求金额 + 到账偏移量，保留2位小数
	creditAmount := math.Round((order.Amount+creditOffset)*100) / 100

	return creditAmount
}

// GetStats 获取支付统计
func (s *PaymentService) GetStats(ctx context.Context) (*PaymentStats, error) {
	return s.orderRepo.GetStats(ctx)
}

// GetPendingOrderByAmount 根据金额查找待支付订单
func (s *PaymentService) GetPendingOrderByAmount(ctx context.Context, amount float64) (*PaymentOrder, error) {
	orders, err := s.orderRepo.GetPendingByAmount(ctx, amount, s.cfg.Payment.Monitor.OrderTimeoutMins)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, errors.New("no matching order found")
	}
	return orders[0], nil
}

// GetPendingOrderByAmountAfterTime 根据金额和时间约束查找待支付订单
// billTime: 账单时间，订单创建时间必须早于账单时间
func (s *PaymentService) GetPendingOrderByAmountAfterTime(ctx context.Context, amount float64, billTime time.Time) (*PaymentOrder, error) {
	orders, err := s.orderRepo.GetPendingByAmountAfterTime(ctx, amount, billTime, s.cfg.Payment.Monitor.OrderTimeoutMins)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, errors.New("no matching order found")
	}
	return orders[0], nil
}

// IsTransLogIDUsed 检查账单流水号是否已使用
func (s *PaymentService) IsTransLogIDUsed(ctx context.Context, transLogID string) (bool, error) {
	if transLogID == "" {
		return false, nil
	}
	return s.orderRepo.IsTransLogIDUsed(ctx, transLogID)
}

// CleanupExpiredOrders 清理过期订单
func (s *PaymentService) CleanupExpiredOrders(ctx context.Context) (int64, error) {
	return s.orderRepo.CleanupExpiredOrders(ctx)
}
