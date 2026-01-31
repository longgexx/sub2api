package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentHandler 支付相关请求处理器
type PaymentHandler struct {
	paymentService *service.PaymentService
}

// NewPaymentHandler 创建 PaymentHandler
func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// GetConfig 获取支付配置
// GET /api/v1/payment/config
func (h *PaymentHandler) GetConfig(c *gin.Context) {
	cfg, err := h.paymentService.GetConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PaymentConfigFromService(cfg))
}

// CreateOrder 创建充值订单
// POST /api/v1/payment/orders
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	order, err := h.paymentService.CreateOrder(c.Request.Context(), subject.UserID, req.Amount)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PaymentOrderFromService(order))
}

// ListOrders 获取用户订单列表
// GET /api/v1/payment/orders
func (h *PaymentHandler) ListOrders(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)
	status := c.Query("status")

	params := pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	orders, paginationResult, err := h.paymentService.ListUserOrders(c.Request.Context(), subject.UserID, params, status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, dto.PaymentOrdersFromService(orders), paginationResult.Total, page, pageSize)
}

// GetOrder 获取订单详情
// GET /api/v1/payment/orders/:trade_no
func (h *PaymentHandler) GetOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	tradeNo := c.Param("trade_no")
	if tradeNo == "" {
		response.BadRequest(c, "trade_no is required")
		return
	}

	order, err := h.paymentService.GetOrder(c.Request.Context(), subject.UserID, tradeNo)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PaymentOrderFromService(order))
}

// CancelOrder 取消订单
// DELETE /api/v1/payment/orders/:trade_no
func (h *PaymentHandler) CancelOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	tradeNo := c.Param("trade_no")
	if tradeNo == "" {
		response.BadRequest(c, "trade_no is required")
		return
	}

	if err := h.paymentService.CancelOrder(c.Request.Context(), subject.UserID, tradeNo); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Order cancelled successfully"})
}
