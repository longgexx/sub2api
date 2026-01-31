package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/userredeemstat"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// UserRedeemStatRepository 用户兑换统计仓储接口
type UserRedeemStatRepository interface {
	GetUserStat(ctx context.Context, userID int64, redeemType string, value float64) (*service.UserRedeemStat, error)
	IncrementCount(ctx context.Context, userID int64, redeemType string, value float64) error
}

type userRedeemStatRepository struct {
	client *dbent.Client
}

func NewUserRedeemStatRepository(client *dbent.Client) UserRedeemStatRepository {
	return &userRedeemStatRepository{client: client}
}

func (r *userRedeemStatRepository) GetUserStat(ctx context.Context, userID int64, redeemType string, value float64) (*service.UserRedeemStat, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.UserRedeemStat.Query().
		Where(
			userredeemstat.UserIDEQ(userID),
			userredeemstat.RedeemTypeEQ(redeemType),
			userredeemstat.ValueEQ(value),
		).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return userRedeemStatEntityToService(m), nil
}

func (r *userRedeemStatRepository) IncrementCount(ctx context.Context, userID int64, redeemType string, value float64) error {
	client := clientFromContext(ctx, r.client)
	now := time.Now()

	// 使用 upsert 语义：存在则更新，不存在则创建
	err := client.UserRedeemStat.Create().
		SetUserID(userID).
		SetRedeemType(redeemType).
		SetValue(value).
		SetUsedCount(1).
		SetLastUsedAt(now).
		OnConflictColumns(userredeemstat.FieldUserID, userredeemstat.FieldRedeemType, userredeemstat.FieldValue).
		Update(func(u *dbent.UserRedeemStatUpsert) {
			u.AddUsedCount(1)
			u.SetLastUsedAt(now)
			u.SetUpdatedAt(now)
		}).
		Exec(ctx)

	return err
}

func userRedeemStatEntityToService(m *dbent.UserRedeemStat) *service.UserRedeemStat {
	if m == nil {
		return nil
	}
	return &service.UserRedeemStat{
		ID:         m.ID,
		UserID:     m.UserID,
		RedeemType: m.RedeemType,
		Value:      m.Value,
		UsedCount:  m.UsedCount,
		LastUsedAt: m.LastUsedAt,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}
