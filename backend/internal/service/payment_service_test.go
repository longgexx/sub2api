//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

// paymentOrderRepoStub 是支付订单仓储的测试替身
type paymentOrderRepoStub struct {
	orders       map[int64]*PaymentOrder
	ordersByNo   map[string]*PaymentOrder
	usedAmounts  []float64
	usedLogIDs   map[string]bool
	createErr    error
	nextID       int64
	updateCalled bool
}

func newPaymentOrderRepoStub() *paymentOrderRepoStub {
	return &paymentOrderRepoStub{
		orders:     make(map[int64]*PaymentOrder),
		ordersByNo: make(map[string]*PaymentOrder),
		usedLogIDs: make(map[string]bool),
		nextID:     1,
	}
}

func (s *paymentOrderRepoStub) Create(ctx context.Context, order *PaymentOrder) error {
	if s.createErr != nil {
		return s.createErr
	}
	order.ID = s.nextID
	s.nextID++
	s.orders[order.ID] = order
	s.ordersByNo[order.TradeNo] = order
	return nil
}

func (s *paymentOrderRepoStub) GetByID(ctx context.Context, id int64) (*PaymentOrder, error) {
	if order, ok := s.orders[id]; ok {
		return order, nil
	}
	return nil, ErrPaymentOrderNotFound
}

func (s *paymentOrderRepoStub) GetByTradeNo(ctx context.Context, tradeNo string) (*PaymentOrder, error) {
	if order, ok := s.ordersByNo[tradeNo]; ok {
		return order, nil
	}
	return nil, ErrPaymentOrderNotFound
}

func (s *paymentOrderRepoStub) GetPendingByAmount(ctx context.Context, amount float64, withinMinutes int) ([]*PaymentOrder, error) {
	return nil, nil
}

func (s *paymentOrderRepoStub) GetPendingByAmountAfterTime(ctx context.Context, amount float64, billTime time.Time, withinMinutes int) ([]*PaymentOrder, error) {
	return nil, nil
}

func (s *paymentOrderRepoStub) GetUsedAmounts(ctx context.Context, baseAmount float64, withinMinutes int) ([]float64, error) {
	return s.usedAmounts, nil
}

func (s *paymentOrderRepoStub) IsTransLogIDUsed(ctx context.Context, transLogID string) (bool, error) {
	return s.usedLogIDs[transLogID], nil
}

func (s *paymentOrderRepoStub) UpdateToPaid(ctx context.Context, id int64, alipayTradeNo string, transLogID string, payerAccount string, creditAmount float64) error {
	s.updateCalled = true
	if order, ok := s.orders[id]; ok {
		order.Status = PaymentStatusPaid
		now := time.Now()
		order.PaidAt = &now
		return nil
	}
	return ErrPaymentOrderNotFound
}

func (s *paymentOrderRepoStub) UpdateToExpired(ctx context.Context, id int64) error {
	if order, ok := s.orders[id]; ok {
		order.Status = PaymentStatusExpired
		return nil
	}
	return ErrPaymentOrderNotFound
}

func (s *paymentOrderRepoStub) Cancel(ctx context.Context, id int64) error {
	if order, ok := s.orders[id]; ok {
		order.Status = PaymentStatusCancelled
		return nil
	}
	return ErrPaymentOrderNotFound
}

func (s *paymentOrderRepoStub) List(ctx context.Context, params pagination.PaginationParams, userID int64, status string) ([]PaymentOrder, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{}, nil
}

func (s *paymentOrderRepoStub) ListAll(ctx context.Context, params pagination.PaginationParams, status, search string) ([]PaymentOrder, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{}, nil
}

func (s *paymentOrderRepoStub) CleanupExpiredOrders(ctx context.Context) (int64, error) {
	return 0, nil
}

func (s *paymentOrderRepoStub) GetStats(ctx context.Context) (*PaymentStats, error) {
	return &PaymentStats{}, nil
}

// paymentUserRepoStub 用于支付测试的用户仓储替身
type paymentUserRepoStub struct {
	balance        float64
	updateBalanceErr error
}

func (s *paymentUserRepoStub) GetByID(ctx context.Context, id int64) (*User, error) {
	return &User{ID: id, Balance: s.balance}, nil
}

func (s *paymentUserRepoStub) UpdateBalance(ctx context.Context, userID int64, amount float64) error {
	if s.updateBalanceErr != nil {
		return s.updateBalanceErr
	}
	s.balance += amount
	return nil
}

func (s *paymentUserRepoStub) GetByEmail(ctx context.Context, email string) (*User, error) {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) Create(ctx context.Context, user *User) error {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) Update(ctx context.Context, user *User) error {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) UpdateConcurrency(ctx context.Context, userID int64, concurrency int) error {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) List(ctx context.Context, params pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) ListWithFilters(ctx context.Context, params pagination.PaginationParams, status, search string) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) Delete(ctx context.Context, id int64) error {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) GetByIDWithAllowedGroups(ctx context.Context, id int64) (*User, error) {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) GetUserDetailsByID(ctx context.Context, id int64) (*User, error) {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) PasswordLogin(ctx context.Context, email, password string) (*User, error) {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) ChangePassword(ctx context.Context, userID int64, newPasswordHash string) error {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) GetByExternalProviderAndID(ctx context.Context, provider, externalID string) (*User, error) {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) CreateWithExternalProvider(ctx context.Context, user *User, provider, externalID string) error {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) LinkExternalProvider(ctx context.Context, userID int64, provider, externalID string) error {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) UnlinkExternalProvider(ctx context.Context, userID int64, provider string) error {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) GetLinkedProviders(ctx context.Context, userID int64) ([]string, error) {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) UpdateTotp(ctx context.Context, userID int64, secret, recoveryCode string) error {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) DisableTotp(ctx context.Context, userID int64) error {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) GetByAPIKey(ctx context.Context, apiKey string) (*User, error) {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) UpdateAllowedGroups(ctx context.Context, userID int64, groupIDs []int64) error {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) CreateAnonymous(ctx context.Context) (*User, error) {
	panic("unexpected call")
}

func (s *paymentUserRepoStub) LinkAnonymousUser(ctx context.Context, anonymousUserID int64, email, passwordHash string) error {
	panic("unexpected call")
}

func TestPaymentService_GetConfig_Disabled(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			Enabled: false,
		},
	}

	svc := NewPaymentService(nil, nil, nil, nil, nil, nil, cfg)

	result, err := svc.GetConfig(context.Background())
	require.NoError(t, err)
	require.False(t, result.Enabled)
}

func TestPaymentService_GetConfig_Enabled(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			Enabled: true,
			Monitor: config.PaymentMonitorConfig{
				MinAmount:        1.0,
				MaxAmount:        100.0,
				BusinessQRCode:   "https://example.com/qr.png",
				OrderTimeoutMins: 30,
			},
		},
	}

	svc := NewPaymentService(nil, nil, nil, nil, nil, nil, cfg)

	result, err := svc.GetConfig(context.Background())
	require.NoError(t, err)
	require.True(t, result.Enabled)
	require.Equal(t, 1.0, result.MinAmount)
	require.Equal(t, 100.0, result.MaxAmount)
	require.Equal(t, "https://example.com/qr.png", result.QRCodeURL)
	require.Equal(t, 30, result.TimeoutMins)
	require.Equal(t, 1.0, result.RateCoefficient)
}

func TestPaymentService_CreateOrder_Disabled(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			Enabled: false,
		},
	}

	svc := NewPaymentService(nil, nil, nil, nil, nil, nil, cfg)

	_, err := svc.CreateOrder(context.Background(), 1, 10.0)
	require.ErrorIs(t, err, ErrPaymentDisabled)
}

func TestPaymentService_CreateOrder_InvalidAmount(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			Enabled: true,
			Monitor: config.PaymentMonitorConfig{
				MinAmount: 5.0,
				MaxAmount: 100.0,
			},
		},
	}

	svc := NewPaymentService(nil, nil, nil, nil, nil, nil, cfg)

	// 金额过小
	_, err := svc.CreateOrder(context.Background(), 1, 1.0)
	require.ErrorIs(t, err, ErrPaymentAmountInvalid)

	// 金额过大
	_, err = svc.CreateOrder(context.Background(), 1, 200.0)
	require.ErrorIs(t, err, ErrPaymentAmountInvalid)
}

func TestPaymentService_CreateOrder_Success(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			Enabled: true,
			Monitor: config.PaymentMonitorConfig{
				MinAmount:        1.0,
				MaxAmount:        100.0,
				OrderTimeoutMins: 30,
			},
		},
	}

	repo := newPaymentOrderRepoStub()
	svc := NewPaymentService(repo, nil, nil, nil, nil, nil, cfg)

	order, err := svc.CreateOrder(context.Background(), 123, 10.0)
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, int64(123), order.UserID)
	require.Equal(t, 10.0, order.Amount)
	require.Equal(t, PaymentStatusPending, order.Status)
	require.NotEmpty(t, order.TradeNo)
	require.Len(t, order.TradeNo, 20) // 14位时间戳 + 6位随机数
}

func TestPaymentService_AllocateUniqueAmount(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			Enabled: true,
			Monitor: config.PaymentMonitorConfig{
				MinAmount:        1.0,
				MaxAmount:        100.0,
				OrderTimeoutMins: 30,
			},
		},
	}

	t.Run("无已用金额时返回基础金额", func(t *testing.T) {
		repo := newPaymentOrderRepoStub()
		repo.usedAmounts = []float64{}
		svc := NewPaymentService(repo, nil, nil, nil, nil, nil, cfg)

		order, err := svc.CreateOrder(context.Background(), 1, 10.0)
		require.NoError(t, err)
		require.Equal(t, 10.0, order.PaymentAmount)
	})

	t.Run("基础金额已被使用时分配递增金额", func(t *testing.T) {
		repo := newPaymentOrderRepoStub()
		repo.usedAmounts = []float64{10.0}
		svc := NewPaymentService(repo, nil, nil, nil, nil, nil, cfg)

		order, err := svc.CreateOrder(context.Background(), 1, 10.0)
		require.NoError(t, err)
		require.Equal(t, 10.01, order.PaymentAmount)
	})

	t.Run("多个金额已被使用时跳过", func(t *testing.T) {
		repo := newPaymentOrderRepoStub()
		repo.usedAmounts = []float64{10.0, 10.01, 10.02}
		svc := NewPaymentService(repo, nil, nil, nil, nil, nil, cfg)

		order, err := svc.CreateOrder(context.Background(), 1, 10.0)
		require.NoError(t, err)
		require.Equal(t, 10.03, order.PaymentAmount)
	})
}

func TestPaymentService_GetOrder(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			Enabled: true,
			Monitor: config.PaymentMonitorConfig{
				MinAmount:        1.0,
				MaxAmount:        100.0,
				OrderTimeoutMins: 30,
			},
		},
	}

	repo := newPaymentOrderRepoStub()
	svc := NewPaymentService(repo, nil, nil, nil, nil, nil, cfg)

	// 创建订单
	order, err := svc.CreateOrder(context.Background(), 123, 10.0)
	require.NoError(t, err)

	t.Run("正确用户可获取订单", func(t *testing.T) {
		got, err := svc.GetOrder(context.Background(), 123, order.TradeNo)
		require.NoError(t, err)
		require.Equal(t, order.ID, got.ID)
	})

	t.Run("错误用户无法获取订单", func(t *testing.T) {
		_, err := svc.GetOrder(context.Background(), 999, order.TradeNo)
		require.ErrorIs(t, err, ErrPaymentOrderNotFound)
	})

	t.Run("不存在的交易号返回错误", func(t *testing.T) {
		_, err := svc.GetOrder(context.Background(), 123, "invalid")
		require.ErrorIs(t, err, ErrPaymentOrderNotFound)
	})
}

func TestPaymentService_CancelOrder(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			Enabled: true,
			Monitor: config.PaymentMonitorConfig{
				MinAmount:        1.0,
				MaxAmount:        100.0,
				OrderTimeoutMins: 30,
			},
		},
	}

	repo := newPaymentOrderRepoStub()
	svc := NewPaymentService(repo, nil, nil, nil, nil, nil, cfg)

	// 创建订单
	order, err := svc.CreateOrder(context.Background(), 123, 10.0)
	require.NoError(t, err)

	t.Run("正确用户可取消订单", func(t *testing.T) {
		err := svc.CancelOrder(context.Background(), 123, order.TradeNo)
		require.NoError(t, err)
		require.Equal(t, PaymentStatusCancelled, repo.orders[order.ID].Status)
	})

	t.Run("错误用户无法取消订单", func(t *testing.T) {
		// 重新创建一个订单
		order2, _ := svc.CreateOrder(context.Background(), 123, 10.0)
		err := svc.CancelOrder(context.Background(), 999, order2.TradeNo)
		require.ErrorIs(t, err, ErrPaymentOrderNotFound)
	})
}

func TestPaymentOrder_Status(t *testing.T) {
	t.Run("IsPending", func(t *testing.T) {
		order := &PaymentOrder{Status: PaymentStatusPending}
		require.True(t, order.IsPending())
		order.Status = PaymentStatusPaid
		require.False(t, order.IsPending())
	})

	t.Run("IsPaid", func(t *testing.T) {
		order := &PaymentOrder{Status: PaymentStatusPaid}
		require.True(t, order.IsPaid())
		order.Status = PaymentStatusPending
		require.False(t, order.IsPaid())
	})

	t.Run("IsExpired 根据状态", func(t *testing.T) {
		order := &PaymentOrder{
			Status:    PaymentStatusExpired,
			ExpiredAt: time.Now().Add(time.Hour),
		}
		require.True(t, order.IsExpired())
	})

	t.Run("IsExpired 根据时间", func(t *testing.T) {
		order := &PaymentOrder{
			Status:    PaymentStatusPending,
			ExpiredAt: time.Now().Add(-time.Hour),
		}
		require.True(t, order.IsExpired())
	})
}

func TestPaymentService_CalculateCreditAmount(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			Enabled: true,
		},
	}

	t.Run("系数为1时直接返回支付金额", func(t *testing.T) {
		svc := NewPaymentService(nil, nil, nil, nil, nil, nil, cfg)

		order := &PaymentOrder{
			Amount:        10.0,
			PaymentAmount: 10.02,
		}

		result := svc.calculateCreditAmount(context.Background(), order)
		require.Equal(t, 10.02, result)
	})
}

func TestPaymentService_IsTransLogIDUsed(t *testing.T) {
	cfg := &config.Config{
		Payment: config.PaymentConfig{
			Enabled: true,
		},
	}

	repo := newPaymentOrderRepoStub()
	repo.usedLogIDs["existing-log-id"] = true

	svc := NewPaymentService(repo, nil, nil, nil, nil, nil, cfg)

	t.Run("空字符串返回 false", func(t *testing.T) {
		used, err := svc.IsTransLogIDUsed(context.Background(), "")
		require.NoError(t, err)
		require.False(t, used)
	})

	t.Run("已存在的 ID 返回 true", func(t *testing.T) {
		used, err := svc.IsTransLogIDUsed(context.Background(), "existing-log-id")
		require.NoError(t, err)
		require.True(t, used)
	})

	t.Run("不存在的 ID 返回 false", func(t *testing.T) {
		used, err := svc.IsTransLogIDUsed(context.Background(), "new-log-id")
		require.NoError(t, err)
		require.False(t, used)
	})
}
