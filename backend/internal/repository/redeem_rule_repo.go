package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/redeemrule"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// RedeemRuleRepository 兑换规则仓储接口
type RedeemRuleRepository interface {
	Create(ctx context.Context, rule *service.RedeemRule) error
	GetByID(ctx context.Context, id int64) (*service.RedeemRule, error)
	GetByTypeAndValue(ctx context.Context, redeemType string, triggerValue float64) (*service.RedeemRule, error)
	Update(ctx context.Context, rule *service.RedeemRule) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]service.RedeemRule, error)
}

type redeemRuleRepository struct {
	client *dbent.Client
}

func NewRedeemRuleRepository(client *dbent.Client) RedeemRuleRepository {
	return &redeemRuleRepository{client: client}
}

func (r *redeemRuleRepository) Create(ctx context.Context, rule *service.RedeemRule) error {
	created, err := r.client.RedeemRule.Create().
		SetType(rule.Type).
		SetTriggerValue(rule.TriggerValue).
		SetMaxTimesPerUser(rule.MaxTimesPerUser).
		SetFallbackValue(rule.FallbackValue).
		SetIsActive(rule.IsActive).
		SetNillableDescription(nilIfEmpty(rule.Description)).
		Save(ctx)
	if err != nil {
		return err
	}
	rule.ID = created.ID
	rule.CreatedAt = created.CreatedAt
	rule.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *redeemRuleRepository) GetByID(ctx context.Context, id int64) (*service.RedeemRule, error) {
	m, err := r.client.RedeemRule.Query().
		Where(redeemrule.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return redeemRuleEntityToService(m), nil
}

func (r *redeemRuleRepository) GetByTypeAndValue(ctx context.Context, redeemType string, triggerValue float64) (*service.RedeemRule, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.RedeemRule.Query().
		Where(
			redeemrule.TypeEQ(redeemType),
			redeemrule.TriggerValueEQ(triggerValue),
			redeemrule.IsActiveEQ(true),
		).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return redeemRuleEntityToService(m), nil
}

func (r *redeemRuleRepository) Update(ctx context.Context, rule *service.RedeemRule) error {
	up := r.client.RedeemRule.UpdateOneID(rule.ID).
		SetType(rule.Type).
		SetTriggerValue(rule.TriggerValue).
		SetMaxTimesPerUser(rule.MaxTimesPerUser).
		SetFallbackValue(rule.FallbackValue).
		SetIsActive(rule.IsActive).
		SetUpdatedAt(time.Now())

	if rule.Description != "" {
		up.SetDescription(rule.Description)
	} else {
		up.ClearDescription()
	}

	updated, err := up.Save(ctx)
	if err != nil {
		return err
	}
	rule.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *redeemRuleRepository) Delete(ctx context.Context, id int64) error {
	return r.client.RedeemRule.DeleteOneID(id).Exec(ctx)
}

func (r *redeemRuleRepository) List(ctx context.Context) ([]service.RedeemRule, error) {
	rules, err := r.client.RedeemRule.Query().
		Order(dbent.Desc(redeemrule.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return redeemRuleEntitiesToService(rules), nil
}

func redeemRuleEntityToService(m *dbent.RedeemRule) *service.RedeemRule {
	if m == nil {
		return nil
	}
	return &service.RedeemRule{
		ID:              m.ID,
		Type:            m.Type,
		TriggerValue:    m.TriggerValue,
		MaxTimesPerUser: m.MaxTimesPerUser,
		FallbackValue:   m.FallbackValue,
		IsActive:        m.IsActive,
		Description:     derefString(m.Description),
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func redeemRuleEntitiesToService(models []*dbent.RedeemRule) []service.RedeemRule {
	out := make([]service.RedeemRule, 0, len(models))
	for i := range models {
		if s := redeemRuleEntityToService(models[i]); s != nil {
			out = append(out, *s)
		}
	}
	return out
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
