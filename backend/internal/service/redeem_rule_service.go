package service

import (
	"context"
	"fmt"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrRedeemRuleNotFound = infraerrors.NotFound("REDEEM_RULE_NOT_FOUND", "redeem rule not found")
)

// RedeemRuleService 兑换规则管理服务
type RedeemRuleService struct {
	ruleRepo RedeemRuleRepository
}

// RedeemRuleFullRepository 完整的规则仓储接口（包含 CRUD）
type RedeemRuleFullRepository interface {
	RedeemRuleRepository
	Create(ctx context.Context, rule *RedeemRule) error
	GetByID(ctx context.Context, id int64) (*RedeemRule, error)
	Update(ctx context.Context, rule *RedeemRule) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]RedeemRule, error)
}

// NewRedeemRuleService 创建兑换规则管理服务
func NewRedeemRuleService(ruleRepo RedeemRuleFullRepository) *RedeemRuleService {
	return &RedeemRuleService{
		ruleRepo: ruleRepo,
	}
}

// CreateRedeemRuleInput 创建规则输入
type CreateRedeemRuleInput struct {
	Type            string
	TriggerValue    float64
	MaxTimesPerUser int
	FallbackValue   float64
	IsActive        bool
	Description     string
}

// Create 创建规则
func (s *RedeemRuleService) Create(ctx context.Context, input *CreateRedeemRuleInput) (*RedeemRule, error) {
	if input.Type == "" {
		input.Type = RedeemTypeBalance
	}
	if input.MaxTimesPerUser <= 0 {
		input.MaxTimesPerUser = 1
	}

	rule := &RedeemRule{
		Type:            input.Type,
		TriggerValue:    input.TriggerValue,
		MaxTimesPerUser: input.MaxTimesPerUser,
		FallbackValue:   input.FallbackValue,
		IsActive:        input.IsActive,
		Description:     input.Description,
	}

	fullRepo, ok := s.ruleRepo.(RedeemRuleFullRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not support Create")
	}

	if err := fullRepo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("create redeem rule: %w", err)
	}

	return rule, nil
}

// GetByID 获取规则详情
func (s *RedeemRuleService) GetByID(ctx context.Context, id int64) (*RedeemRule, error) {
	fullRepo, ok := s.ruleRepo.(RedeemRuleFullRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not support GetByID")
	}

	rule, err := fullRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get redeem rule: %w", err)
	}
	if rule == nil {
		return nil, ErrRedeemRuleNotFound
	}

	return rule, nil
}

// UpdateRedeemRuleInput 更新规则输入
// All fields are optional pointers - only non-nil fields will be updated
type UpdateRedeemRuleInput struct {
	ID              int64
	TriggerValue    *float64
	MaxTimesPerUser *int
	FallbackValue   *float64
	IsActive        *bool
	Description     *string
}

// Update 更新规则
func (s *RedeemRuleService) Update(ctx context.Context, input *UpdateRedeemRuleInput) (*RedeemRule, error) {
	fullRepo, ok := s.ruleRepo.(RedeemRuleFullRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not support Update")
	}

	existing, err := fullRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("get redeem rule: %w", err)
	}
	if existing == nil {
		return nil, ErrRedeemRuleNotFound
	}

	// Only update fields that are provided (non-nil)
	if input.TriggerValue != nil {
		existing.TriggerValue = *input.TriggerValue
	}
	if input.MaxTimesPerUser != nil {
		existing.MaxTimesPerUser = *input.MaxTimesPerUser
	}
	if input.FallbackValue != nil {
		existing.FallbackValue = *input.FallbackValue
	}
	if input.IsActive != nil {
		existing.IsActive = *input.IsActive
	}
	if input.Description != nil {
		existing.Description = *input.Description
	}

	if err := fullRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("update redeem rule: %w", err)
	}

	return existing, nil
}

// Delete 删除规则
func (s *RedeemRuleService) Delete(ctx context.Context, id int64) error {
	fullRepo, ok := s.ruleRepo.(RedeemRuleFullRepository)
	if !ok {
		return fmt.Errorf("repository does not support Delete")
	}

	existing, err := fullRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get redeem rule: %w", err)
	}
	if existing == nil {
		return ErrRedeemRuleNotFound
	}

	if err := fullRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete redeem rule: %w", err)
	}

	return nil
}

// List 获取所有规则
func (s *RedeemRuleService) List(ctx context.Context) ([]RedeemRule, error) {
	fullRepo, ok := s.ruleRepo.(RedeemRuleFullRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not support List")
	}

	rules, err := fullRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list redeem rules: %w", err)
	}

	return rules, nil
}
