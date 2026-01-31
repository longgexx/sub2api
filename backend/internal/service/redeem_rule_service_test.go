//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// redeemRuleRepoStub 是兑换规则仓储的测试替身
type redeemRuleRepoStub struct {
	rules     map[int64]*RedeemRule
	nextID    int64
	createErr error
	updateErr error
	deleteErr error
}

func newRedeemRuleRepoStub() *redeemRuleRepoStub {
	return &redeemRuleRepoStub{
		rules:  make(map[int64]*RedeemRule),
		nextID: 1,
	}
}

func (s *redeemRuleRepoStub) GetByTypeAndValue(ctx context.Context, redeemType string, triggerValue float64) (*RedeemRule, error) {
	for _, rule := range s.rules {
		if rule.Type == redeemType && rule.TriggerValue == triggerValue {
			return rule, nil
		}
	}
	return nil, nil
}

func (s *redeemRuleRepoStub) Create(ctx context.Context, rule *RedeemRule) error {
	if s.createErr != nil {
		return s.createErr
	}
	rule.ID = s.nextID
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	s.nextID++
	s.rules[rule.ID] = rule
	return nil
}

func (s *redeemRuleRepoStub) GetByID(ctx context.Context, id int64) (*RedeemRule, error) {
	if rule, ok := s.rules[id]; ok {
		return rule, nil
	}
	return nil, nil
}

func (s *redeemRuleRepoStub) Update(ctx context.Context, rule *RedeemRule) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	rule.UpdatedAt = time.Now()
	s.rules[rule.ID] = rule
	return nil
}

func (s *redeemRuleRepoStub) Delete(ctx context.Context, id int64) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	delete(s.rules, id)
	return nil
}

func (s *redeemRuleRepoStub) List(ctx context.Context) ([]RedeemRule, error) {
	var result []RedeemRule
	for _, rule := range s.rules {
		result = append(result, *rule)
	}
	return result, nil
}

func TestRedeemRuleService_Create(t *testing.T) {
	t.Run("成功创建规则", func(t *testing.T) {
		repo := newRedeemRuleRepoStub()
		svc := NewRedeemRuleService(repo)

		input := &CreateRedeemRuleInput{
			Type:            RedeemTypeBalance,
			TriggerValue:    5.0,
			MaxTimesPerUser: 2,
			FallbackValue:   1.0,
			IsActive:        true,
			Description:     "试用用户优惠",
		}

		rule, err := svc.Create(context.Background(), input)
		require.NoError(t, err)
		require.NotNil(t, rule)
		require.Equal(t, RedeemTypeBalance, rule.Type)
		require.Equal(t, 5.0, rule.TriggerValue)
		require.Equal(t, 2, rule.MaxTimesPerUser)
		require.Equal(t, 1.0, rule.FallbackValue)
		require.True(t, rule.IsActive)
	})

	t.Run("默认值设置", func(t *testing.T) {
		repo := newRedeemRuleRepoStub()
		svc := NewRedeemRuleService(repo)

		input := &CreateRedeemRuleInput{
			TriggerValue:  5.0,
			FallbackValue: 1.0,
		}

		rule, err := svc.Create(context.Background(), input)
		require.NoError(t, err)
		require.Equal(t, RedeemTypeBalance, rule.Type)
		require.Equal(t, 1, rule.MaxTimesPerUser)
	})

	t.Run("仓储错误", func(t *testing.T) {
		repo := newRedeemRuleRepoStub()
		repo.createErr = errors.New("db error")
		svc := NewRedeemRuleService(repo)

		_, err := svc.Create(context.Background(), &CreateRedeemRuleInput{
			TriggerValue: 5.0,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "create redeem rule")
	})
}

func TestRedeemRuleService_GetByID(t *testing.T) {
	repo := newRedeemRuleRepoStub()
	svc := NewRedeemRuleService(repo)

	// 创建一个规则
	rule, _ := svc.Create(context.Background(), &CreateRedeemRuleInput{
		TriggerValue: 5.0,
		IsActive:     true,
	})

	t.Run("成功获取规则", func(t *testing.T) {
		got, err := svc.GetByID(context.Background(), rule.ID)
		require.NoError(t, err)
		require.Equal(t, rule.ID, got.ID)
	})

	t.Run("规则不存在", func(t *testing.T) {
		_, err := svc.GetByID(context.Background(), 999)
		require.ErrorIs(t, err, ErrRedeemRuleNotFound)
	})
}

func TestRedeemRuleService_Update(t *testing.T) {
	repo := newRedeemRuleRepoStub()
	svc := NewRedeemRuleService(repo)

	// 创建一个规则
	rule, _ := svc.Create(context.Background(), &CreateRedeemRuleInput{
		TriggerValue:    5.0,
		MaxTimesPerUser: 1,
		FallbackValue:   1.0,
		IsActive:        true,
		Description:     "原描述",
	})

	t.Run("成功更新规则", func(t *testing.T) {
		newMaxTimes := 3
		newFallback := 2.0
		newDesc := "新描述"
		isActive := false

		updated, err := svc.Update(context.Background(), &UpdateRedeemRuleInput{
			ID:              rule.ID,
			MaxTimesPerUser: &newMaxTimes,
			FallbackValue:   &newFallback,
			Description:     &newDesc,
			IsActive:        &isActive,
		})
		require.NoError(t, err)
		require.Equal(t, 3, updated.MaxTimesPerUser)
		require.Equal(t, 2.0, updated.FallbackValue)
		require.Equal(t, "新描述", updated.Description)
		require.False(t, updated.IsActive)
		// 未更新的字段保持不变
		require.Equal(t, 5.0, updated.TriggerValue)
	})

	t.Run("部分更新", func(t *testing.T) {
		newMaxTimes := 5
		updated, err := svc.Update(context.Background(), &UpdateRedeemRuleInput{
			ID:              rule.ID,
			MaxTimesPerUser: &newMaxTimes,
		})
		require.NoError(t, err)
		require.Equal(t, 5, updated.MaxTimesPerUser)
		// 其他字段保持不变
		require.Equal(t, 5.0, updated.TriggerValue)
	})

	t.Run("规则不存在", func(t *testing.T) {
		newMaxTimes := 1
		_, err := svc.Update(context.Background(), &UpdateRedeemRuleInput{
			ID:              999,
			MaxTimesPerUser: &newMaxTimes,
		})
		require.ErrorIs(t, err, ErrRedeemRuleNotFound)
	})
}

func TestRedeemRuleService_Delete(t *testing.T) {
	repo := newRedeemRuleRepoStub()
	svc := NewRedeemRuleService(repo)

	// 创建一个规则
	rule, _ := svc.Create(context.Background(), &CreateRedeemRuleInput{
		TriggerValue: 5.0,
	})

	t.Run("成功删除规则", func(t *testing.T) {
		err := svc.Delete(context.Background(), rule.ID)
		require.NoError(t, err)

		// 确认已删除
		_, err = svc.GetByID(context.Background(), rule.ID)
		require.ErrorIs(t, err, ErrRedeemRuleNotFound)
	})

	t.Run("规则不存在", func(t *testing.T) {
		err := svc.Delete(context.Background(), 999)
		require.ErrorIs(t, err, ErrRedeemRuleNotFound)
	})
}

func TestRedeemRuleService_List(t *testing.T) {
	repo := newRedeemRuleRepoStub()
	svc := NewRedeemRuleService(repo)

	t.Run("空列表", func(t *testing.T) {
		rules, err := svc.List(context.Background())
		require.NoError(t, err)
		require.Empty(t, rules)
	})

	t.Run("返回所有规则", func(t *testing.T) {
		// 创建几个规则
		_, _ = svc.Create(context.Background(), &CreateRedeemRuleInput{TriggerValue: 5.0})
		_, _ = svc.Create(context.Background(), &CreateRedeemRuleInput{TriggerValue: 10.0})

		rules, err := svc.List(context.Background())
		require.NoError(t, err)
		require.Len(t, rules, 2)
	})
}

func TestRedeemRule_Model(t *testing.T) {
	rule := &RedeemRule{
		ID:              1,
		Type:            RedeemTypeBalance,
		TriggerValue:    5.0,
		MaxTimesPerUser: 2,
		FallbackValue:   1.0,
		IsActive:        true,
		Description:     "测试规则",
	}

	require.Equal(t, int64(1), rule.ID)
	require.Equal(t, RedeemTypeBalance, rule.Type)
	require.Equal(t, 5.0, rule.TriggerValue)
	require.Equal(t, 2, rule.MaxTimesPerUser)
	require.Equal(t, 1.0, rule.FallbackValue)
	require.True(t, rule.IsActive)
}

func TestUserRedeemStat_Model(t *testing.T) {
	now := time.Now()
	stat := &UserRedeemStat{
		ID:         1,
		UserID:     123,
		RedeemType: RedeemTypeBalance,
		Value:      5.0,
		UsedCount:  3,
		LastUsedAt: &now,
	}

	require.Equal(t, int64(1), stat.ID)
	require.Equal(t, int64(123), stat.UserID)
	require.Equal(t, RedeemTypeBalance, stat.RedeemType)
	require.Equal(t, 5.0, stat.Value)
	require.Equal(t, 3, stat.UsedCount)
}

func TestRedeemResult_Model(t *testing.T) {
	result := &RedeemResult{
		RedeemCode: &RedeemCode{
			ID:    1,
			Code:  "TEST-CODE",
			Value: 5.0,
		},
		OriginalValue: 5.0,
		ActualValue:   3.0,
		IsDegraded:    true,
		Message:       "已降级",
	}

	require.Equal(t, 5.0, result.OriginalValue)
	require.Equal(t, 3.0, result.ActualValue)
	require.True(t, result.IsDegraded)
}

func TestRedeemPreview_Model(t *testing.T) {
	preview := &RedeemPreview{
		RedeemCode: &RedeemCode{
			ID:    1,
			Code:  "TEST-CODE",
			Value: 5.0,
		},
		OriginalValue:  5.0,
		ActualValue:    3.0,
		WillDegrade:    true,
		NeedsConfirm:   true,
		ConfirmMessage: "是否继续？",
	}

	require.Equal(t, 5.0, preview.OriginalValue)
	require.Equal(t, 3.0, preview.ActualValue)
	require.True(t, preview.WillDegrade)
	require.True(t, preview.NeedsConfirm)
}
