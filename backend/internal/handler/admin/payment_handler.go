package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentHandler 支付订单管理 Handler
type PaymentHandler struct {
	paymentService        *service.PaymentService
	paymentMonitorService *service.PaymentMonitorService
}

// NewPaymentHandler 创建 PaymentHandler
func NewPaymentHandler(
	paymentService *service.PaymentService,
	paymentMonitorService *service.PaymentMonitorService,
) *PaymentHandler {
	return &PaymentHandler{
		paymentService:        paymentService,
		paymentMonitorService: paymentMonitorService,
	}
}

// ListOrders 获取所有订单列表
// GET /api/v1/admin/payment/orders
func (h *PaymentHandler) ListOrders(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	status := c.Query("status")
	search := c.Query("search")

	params := pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	orders, paginationResult, err := h.paymentService.ListAllOrders(c.Request.Context(), params, status, search)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, dto.PaymentOrdersFromService(orders), paginationResult.Total, page, pageSize)
}

// GetOrder 获取订单详情
// GET /api/v1/admin/payment/orders/:id
func (h *PaymentHandler) GetOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	order, err := h.paymentService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PaymentOrderFromService(order))
}

// ManualConfirm 手动确认支付
// POST /api/v1/admin/payment/orders/:id/confirm
func (h *PaymentHandler) ManualConfirm(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	if err := h.paymentService.ManualConfirm(c.Request.Context(), orderID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Payment confirmed successfully"})
}

// GetStats 获取支付统计
// GET /api/v1/admin/payment/stats
func (h *PaymentHandler) GetStats(c *gin.Context) {
	stats, err := h.paymentService.GetStats(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PaymentStatsFromService(stats))
}

// GetMonitorStatus 获取监控服务状态
// GET /api/v1/admin/payment/monitor/status
func (h *PaymentHandler) GetMonitorStatus(c *gin.Context) {
	if h.paymentMonitorService == nil {
		response.Success(c, gin.H{
			"running":   false,
			"available": false,
			"message":   "Payment monitor service is not configured",
		})
		return
	}

	status := h.paymentMonitorService.GetStatus()
	status["available"] = true
	response.Success(c, status)
}

// StartMonitor 启动监控服务
// POST /api/v1/admin/payment/monitor/start
func (h *PaymentHandler) StartMonitor(c *gin.Context) {
	if h.paymentMonitorService == nil {
		response.BadRequest(c, "Payment monitor service is not configured")
		return
	}

	if h.paymentMonitorService.IsRunning() {
		response.Success(c, gin.H{"message": "Monitor service is already running"})
		return
	}

	h.paymentMonitorService.Start()
	response.Success(c, gin.H{"message": "Monitor service started"})
}

// StopMonitor 停止监控服务
// POST /api/v1/admin/payment/monitor/stop
func (h *PaymentHandler) StopMonitor(c *gin.Context) {
	if h.paymentMonitorService == nil {
		response.BadRequest(c, "Payment monitor service is not configured")
		return
	}

	if !h.paymentMonitorService.IsRunning() {
		response.Success(c, gin.H{"message": "Monitor service is not running"})
		return
	}

	h.paymentMonitorService.Stop()
	response.Success(c, gin.H{"message": "Monitor service stopped"})
}
