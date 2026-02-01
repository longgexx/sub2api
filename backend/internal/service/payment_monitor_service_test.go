//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// TestPaymentMonitorService_NewService 测试服务创建
func TestPaymentMonitorService_NewService(t *testing.T) {
	t.Run("默认间隔为5秒", func(t *testing.T) {
		cfg := &config.Config{
			Payment: config.PaymentConfig{
				Monitor: config.PaymentMonitorConfig{
					ActiveIntervalSeconds: 0, // 未设置
				},
			},
		}

		svc := NewPaymentMonitorService(nil, nil, cfg)
		require.Equal(t, 5*time.Second, svc.activeInterval)
	})

	t.Run("自定义间隔", func(t *testing.T) {
		cfg := &config.Config{
			Payment: config.PaymentConfig{
				Monitor: config.PaymentMonitorConfig{
					ActiveIntervalSeconds: 10,
				},
			},
		}

		svc := NewPaymentMonitorService(nil, nil, cfg)
		require.Equal(t, 10*time.Second, svc.activeInterval)
	})

	t.Run("triggerChan 带缓冲", func(t *testing.T) {
		cfg := &config.Config{}
		svc := NewPaymentMonitorService(nil, nil, cfg)
		require.Equal(t, 1, cap(svc.triggerChan))
	})
}

// TestPaymentMonitorService_Trigger 测试触发机制
func TestPaymentMonitorService_Trigger(t *testing.T) {
	t.Run("非阻塞触发", func(t *testing.T) {
		cfg := &config.Config{}
		svc := NewPaymentMonitorService(nil, nil, cfg)

		// 第一次触发应该成功
		svc.Trigger()
		require.Len(t, svc.triggerChan, 1)

		// 第二次触发应该被忽略（channel 已满）
		svc.Trigger()
		require.Len(t, svc.triggerChan, 1) // 仍然是 1

		// 消费后可以再次触发
		<-svc.triggerChan
		svc.Trigger()
		require.Len(t, svc.triggerChan, 1)
	})
}

// TestPaymentMonitorService_StartStop 测试启动和停止
func TestPaymentMonitorService_StartStop(t *testing.T) {
	t.Run("启动和停止", func(t *testing.T) {
		cfg := &config.Config{
			Payment: config.PaymentConfig{
				Monitor: config.PaymentMonitorConfig{
					OrderTimeoutMins: 30,
				},
			},
		}
		// 创建一个带有 stub repo 的 PaymentService
		repo := &paymentOrderRepoStub{
			orders:     make(map[int64]*PaymentOrder),
			ordersByNo: make(map[string]*PaymentOrder),
			usedLogIDs: make(map[string]bool),
		}
		paymentSvc := NewPaymentService(repo, nil, nil, nil, nil, nil, cfg)

		svc := NewPaymentMonitorService(nil, paymentSvc, cfg)

		require.False(t, svc.IsRunning())

		svc.Start()
		require.True(t, svc.IsRunning())

		// 重复启动应该被忽略
		svc.Start()
		require.True(t, svc.IsRunning())

		svc.Stop()
		// 等待 goroutine 退出
		time.Sleep(100 * time.Millisecond)
		require.False(t, svc.IsRunning())
	})
}

// TestPaymentMonitorService_GetStatus 测试获取状态
func TestPaymentMonitorService_GetStatus(t *testing.T) {
	t.Run("初始状态", func(t *testing.T) {
		cfg := &config.Config{}
		svc := NewPaymentMonitorService(nil, nil, cfg)

		status := svc.GetStatus()
		require.False(t, status["running"].(bool))
		require.True(t, status["last_check_time"].(time.Time).IsZero())
	})
}

// TestPaymentMonitorService_RunCycle 测试运行周期
func TestPaymentMonitorService_RunCycle(t *testing.T) {
	t.Run("无支付宝客户端时跳过", func(t *testing.T) {
		cfg := &config.Config{
			Payment: config.PaymentConfig{
				Monitor: config.PaymentMonitorConfig{
					OrderTimeoutMins: 30,
				},
			},
		}
		repo := &paymentOrderRepoStub{
			orders:     make(map[int64]*PaymentOrder),
			ordersByNo: make(map[string]*PaymentOrder),
			usedLogIDs: make(map[string]bool),
		}
		paymentSvc := NewPaymentService(repo, nil, nil, nil, nil, nil, cfg)
		svc := NewPaymentMonitorService(nil, paymentSvc, cfg)

		// 不应该 panic
		svc.RunCycle()
	})
}

// TestPaymentMonitorService_GetLastCheckTime 测试获取最后检查时间
func TestPaymentMonitorService_GetLastCheckTime(t *testing.T) {
	t.Run("初始为零值", func(t *testing.T) {
		cfg := &config.Config{}
		svc := NewPaymentMonitorService(nil, nil, cfg)

		lastCheck := svc.GetLastCheckTime()
		require.True(t, lastCheck.IsZero())
	})
}

// TestPaymentMonitorService_CheckPendingOrders 测试检查待支付订单
func TestPaymentMonitorService_CheckPendingOrders(t *testing.T) {
	t.Run("有待支付订单时返回true", func(t *testing.T) {
		cfg := &config.Config{
			Payment: config.PaymentConfig{
				Monitor: config.PaymentMonitorConfig{
					OrderTimeoutMins: 30,
				},
			},
		}
		repo := &paymentOrderRepoStub{
			orders:       make(map[int64]*PaymentOrder),
			ordersByNo:   make(map[string]*PaymentOrder),
			usedLogIDs:   make(map[string]bool),
			pendingCount: 5,
		}
		paymentSvc := NewPaymentService(repo, nil, nil, nil, nil, nil, cfg)
		svc := NewPaymentMonitorService(nil, paymentSvc, cfg)

		result := svc.checkPendingOrders(t.Context())
		require.True(t, result)
	})

	t.Run("无待支付订单时返回false", func(t *testing.T) {
		cfg := &config.Config{
			Payment: config.PaymentConfig{
				Monitor: config.PaymentMonitorConfig{
					OrderTimeoutMins: 30,
				},
			},
		}
		repo := &paymentOrderRepoStub{
			orders:       make(map[int64]*PaymentOrder),
			ordersByNo:   make(map[string]*PaymentOrder),
			usedLogIDs:   make(map[string]bool),
			pendingCount: 0,
		}
		paymentSvc := NewPaymentService(repo, nil, nil, nil, nil, nil, cfg)
		svc := NewPaymentMonitorService(nil, paymentSvc, cfg)

		result := svc.checkPendingOrders(t.Context())
		require.False(t, result)
	})

	t.Run("数据库错误时返回true（保守策略）", func(t *testing.T) {
		cfg := &config.Config{
			Payment: config.PaymentConfig{
				Monitor: config.PaymentMonitorConfig{
					OrderTimeoutMins: 30,
				},
			},
		}
		repo := &paymentOrderRepoStub{
			orders:          make(map[int64]*PaymentOrder),
			ordersByNo:      make(map[string]*PaymentOrder),
			usedLogIDs:      make(map[string]bool),
			countPendingErr: errTestDB,
		}
		paymentSvc := NewPaymentService(repo, nil, nil, nil, nil, nil, cfg)
		svc := NewPaymentMonitorService(nil, paymentSvc, cfg)

		result := svc.checkPendingOrders(t.Context())
		require.True(t, result) // 保守策略：错误时假设有订单
	})
}

var errTestDB = &testError{msg: "database error"}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
