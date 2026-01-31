package admin

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RedeemRuleHandler handles admin redeem rule management
type RedeemRuleHandler struct {
	ruleService *service.RedeemRuleService
}

// NewRedeemRuleHandler creates a new admin redeem rule handler
func NewRedeemRuleHandler(ruleService *service.RedeemRuleService) *RedeemRuleHandler {
	return &RedeemRuleHandler{
		ruleService: ruleService,
	}
}

// RedeemRuleResponse represents a redeem rule response
type RedeemRuleResponse struct {
	ID              int64   `json:"id"`
	Type            string  `json:"type"`
	TriggerValue    float64 `json:"trigger_value"`
	MaxTimesPerUser int     `json:"max_times_per_user"`
	FallbackValue   float64 `json:"fallback_value"`
	IsActive        bool    `json:"is_active"`
	Description     string  `json:"description"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// CreateRedeemRuleRequest represents create redeem rule request
type CreateRedeemRuleRequest struct {
	Type            string  `json:"type" binding:"required,oneof=balance concurrency subscription"`
	TriggerValue    float64 `json:"trigger_value" binding:"required,gt=0"`
	MaxTimesPerUser int     `json:"max_times_per_user" binding:"required,min=1"`
	FallbackValue   float64 `json:"fallback_value" binding:"required,gte=0"`
	IsActive        *bool   `json:"is_active"`
	Description     string  `json:"description"`
}

// UpdateRedeemRuleRequest represents update redeem rule request
// All fields are optional - only provided fields will be updated
type UpdateRedeemRuleRequest struct {
	TriggerValue    *float64 `json:"trigger_value" binding:"omitempty,gt=0"`
	MaxTimesPerUser *int     `json:"max_times_per_user" binding:"omitempty,min=1"`
	FallbackValue   *float64 `json:"fallback_value" binding:"omitempty,gte=0"`
	IsActive        *bool    `json:"is_active"`
	Description     *string  `json:"description"`
}

func ruleToResponse(rule *service.RedeemRule) *RedeemRuleResponse {
	if rule == nil {
		return nil
	}
	return &RedeemRuleResponse{
		ID:              rule.ID,
		Type:            rule.Type,
		TriggerValue:    rule.TriggerValue,
		MaxTimesPerUser: rule.MaxTimesPerUser,
		FallbackValue:   rule.FallbackValue,
		IsActive:        rule.IsActive,
		Description:     rule.Description,
		CreatedAt:       rule.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       rule.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// List handles listing all redeem rules
// GET /api/v1/admin/redeem-rules
func (h *RedeemRuleHandler) List(c *gin.Context) {
	rules, err := h.ruleService.List(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]RedeemRuleResponse, 0, len(rules))
	for i := range rules {
		out = append(out, *ruleToResponse(&rules[i]))
	}
	response.Success(c, out)
}

// GetByID handles getting a redeem rule by ID
// GET /api/v1/admin/redeem-rules/:id
func (h *RedeemRuleHandler) GetByID(c *gin.Context) {
	ruleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid rule ID")
		return
	}

	rule, err := h.ruleService.GetByID(c.Request.Context(), ruleID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, ruleToResponse(rule))
}

// Create handles creating a new redeem rule
// POST /api/v1/admin/redeem-rules
func (h *RedeemRuleHandler) Create(c *gin.Context) {
	var req CreateRedeemRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	rule, err := h.ruleService.Create(c.Request.Context(), &service.CreateRedeemRuleInput{
		Type:            req.Type,
		TriggerValue:    req.TriggerValue,
		MaxTimesPerUser: req.MaxTimesPerUser,
		FallbackValue:   req.FallbackValue,
		IsActive:        isActive,
		Description:     req.Description,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, ruleToResponse(rule))
}

// Update handles updating a redeem rule
// PUT /api/v1/admin/redeem-rules/:id
func (h *RedeemRuleHandler) Update(c *gin.Context) {
	ruleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid rule ID")
		return
	}

	var req UpdateRedeemRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	rule, err := h.ruleService.Update(c.Request.Context(), &service.UpdateRedeemRuleInput{
		ID:              ruleID,
		TriggerValue:    req.TriggerValue,
		MaxTimesPerUser: req.MaxTimesPerUser,
		FallbackValue:   req.FallbackValue,
		IsActive:        req.IsActive,
		Description:     req.Description,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, ruleToResponse(rule))
}

// Delete handles deleting a redeem rule
// DELETE /api/v1/admin/redeem-rules/:id
func (h *RedeemRuleHandler) Delete(c *gin.Context) {
	ruleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid rule ID")
		return
	}

	err = h.ruleService.Delete(c.Request.Context(), ruleID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Redeem rule deleted successfully"})
}
