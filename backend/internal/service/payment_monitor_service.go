package service

import (
	"context"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/alipay"
)

// PaymentMonitorService 支付监控服务
// 定时查询支付宝账单，自动匹配并确认待支付订单
type PaymentMonitorService struct {
	alipayClient   *alipay.Client
	paymentService *PaymentService
	cfg            *config.Config
	running        bool
	mu             sync.Mutex
	ctx            context.Context
	cancel         context.CancelFunc
	lastCheckTime  time.Time
}

// NewPaymentMonitorService 创建支付监控服务
func NewPaymentMonitorService(
	alipayClient *alipay.Client,
	paymentService *PaymentService,
	cfg *config.Config,
) *PaymentMonitorService {
	return &PaymentMonitorService{
		alipayClient:   alipayClient,
		paymentService: paymentService,
		cfg:            cfg,
	}
}

// Start 启动监控服务
func (s *PaymentMonitorService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.mu.Unlock()

	log.Println("[PaymentMonitor] service started")

	interval := time.Duration(s.cfg.Payment.Monitor.IntervalSeconds) * time.Second

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// 立即执行一次
		s.RunCycle()

		for {
			select {
			case <-s.ctx.Done():
				log.Println("[PaymentMonitor] service stopped")
				return
			case <-ticker.C:
				s.RunCycle()
			}
		}
	}()
}

// Stop 停止监控服务
func (s *PaymentMonitorService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running && s.cancel != nil {
		s.cancel()
		s.running = false
	}
}

// IsRunning 检查是否运行中
func (s *PaymentMonitorService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// RunCycle 执行一次监控周期
func (s *PaymentMonitorService) RunCycle() {
	ctx := context.Background()

	// 清理过期订单
	s.cleanupExpiredOrders(ctx)

	// 检查支付宝客户端是否存在
	if s.alipayClient == nil {
		log.Println("[PaymentMonitor] alipay client not initialized, skipping")
		return
	}

	// 检查支付宝是否配置
	if !s.alipayClient.IsConfigured() {
		log.Println("[PaymentMonitor] alipay not configured, skipping")
		return
	}

	// 查询最近 30 分钟的账单
	bills, err := s.alipayClient.QueryRecentBills(30)
	if err != nil {
		log.Printf("[PaymentMonitor] query bills failed: %v", err)
		return
	}

	log.Printf("[PaymentMonitor] found %d bills", len(bills))

	// 处理账单（经营码模式）
	s.processBills(ctx, bills)

	// 修复竞态：加锁更新 lastCheckTime
	s.mu.Lock()
	s.lastCheckTime = time.Now()
	s.mu.Unlock()
}

// processBills 处理账单记录
// 经营码模式：根据金额匹配待支付订单
func (s *PaymentMonitorService) processBills(ctx context.Context, bills []alipay.BillRecord) {
	for _, bill := range bills {
		// 只处理收入类账单（支付宝可能返回 direction=in 或中文"收入"）
		direction := strings.TrimSpace(bill.TransDirection)
		if direction != "in" && direction != "收入" {
			continue
		}

		// 检查该账单是否已处理过（防止重复匹配）
		used, err := s.paymentService.IsTransLogIDUsed(ctx, bill.TransLogID)
		if err != nil {
			log.Printf("[PaymentMonitor] check trans_log_id failed: %v", err)
			continue
		}
		if used {
			// 已处理过，跳过
			continue
		}

		// 规范化金额（保留2位小数，避免浮点精度问题）
		amount := math.Round(math.Abs(bill.Amount)*100) / 100

		// 解析账单时间
		billTime, err := time.ParseInLocation("2006-01-02 15:04:05", bill.TransTime, time.Local)
		if err != nil {
			log.Printf("[PaymentMonitor] invalid bill time format: %s, error: %v", bill.TransTime, err)
			continue
		}

		// 根据金额匹配订单（订单创建时间必须早于账单时间，确保时序正确）
		order, err := s.paymentService.GetPendingOrderByAmountAfterTime(ctx, amount, billTime)
		if err != nil {
			// 没有匹配的订单，跳过
			continue
		}

		// 完成支付（alipayTradeNo 留空，经营码模式下没有交易号；保存账单流水号和付款人账户）
		if err := s.paymentService.CompletePayment(ctx, order.ID, "", bill.TransLogID, bill.OtherAccount); err != nil {
			log.Printf("[PaymentMonitor] complete payment failed: order=%s, error=%v", order.TradeNo, err)
			continue
		}

		log.Printf("[PaymentMonitor] payment completed: trade_no=%s, amount=%.2f, trans_log_id=%s, payer=%s",
			order.TradeNo, amount, bill.TransLogID, bill.OtherAccount)
	}
}

// cleanupExpiredOrders 清理过期订单
func (s *PaymentMonitorService) cleanupExpiredOrders(ctx context.Context) {
	deleted, err := s.paymentService.CleanupExpiredOrders(ctx)
	if err != nil {
		log.Printf("[PaymentMonitor] cleanup expired orders failed: %v", err)
		return
	}

	if deleted > 0 {
		log.Printf("[PaymentMonitor] cleaned up %d expired orders", deleted)
	}
}

// GetLastCheckTime 获取最后检查时间
func (s *PaymentMonitorService) GetLastCheckTime() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastCheckTime
}

// GetStatus 获取监控状态
func (s *PaymentMonitorService) GetStatus() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	status := map[string]interface{}{
		"running":         s.running,
		"last_check_time": s.lastCheckTime,
	}

	if !s.lastCheckTime.IsZero() {
		status["seconds_since_check"] = time.Since(s.lastCheckTime).Seconds()
	}

	return status
}
