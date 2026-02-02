//go:build unit

package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

// schedulerRepoStub 定时发布测试用的仓储桩
type schedulerRepoStub struct {
	batchPublishCalls int32
	batchPublishErr   error
	batchPublishCount int
}

func (r *schedulerRepoStub) Create(_ context.Context, _ *Announcement) error {
	return nil
}

func (r *schedulerRepoStub) GetByID(_ context.Context, _ int64) (*Announcement, error) {
	return nil, ErrAnnouncementNotFound
}

func (r *schedulerRepoStub) Update(_ context.Context, _ *Announcement) error {
	return nil
}

func (r *schedulerRepoStub) Delete(_ context.Context, _ int64) error {
	return nil
}

func (r *schedulerRepoStub) List(_ context.Context, _ pagination.PaginationParams, _ string, _ string) ([]Announcement, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *schedulerRepoStub) GetActiveAnnouncements(_ context.Context, _ time.Time) ([]Announcement, error) {
	return nil, nil
}

func (r *schedulerRepoStub) GetUnreadAnnouncements(_ context.Context, _ int64, _ time.Time) ([]Announcement, error) {
	return nil, nil
}

func (r *schedulerRepoStub) CountUnread(_ context.Context, _ int64, _ time.Time) (int, error) {
	return 0, nil
}

func (r *schedulerRepoStub) GetUserReadIDs(_ context.Context, _ int64) ([]int64, error) {
	return nil, nil
}

func (r *schedulerRepoStub) MarkAsRead(_ context.Context, _, _ int64) error {
	return nil
}

func (r *schedulerRepoStub) MarkAllAsRead(_ context.Context, _ int64, _ time.Time) error {
	return nil
}

func (r *schedulerRepoStub) BatchPublishScheduled(_ context.Context, _ time.Time) (int, error) {
	atomic.AddInt32(&r.batchPublishCalls, 1)
	if r.batchPublishErr != nil {
		return 0, r.batchPublishErr
	}
	return r.batchPublishCount, nil
}

func (r *schedulerRepoStub) getBatchPublishCalls() int {
	return int(atomic.LoadInt32(&r.batchPublishCalls))
}

// ==================== Scheduler 测试 ====================

func TestAnnouncementSchedulerService_NewService(t *testing.T) {
	repo := &schedulerRepoStub{}
	svc := NewAnnouncementSchedulerService(repo, time.Minute)

	require.NotNil(t, svc)
	require.Equal(t, time.Minute, svc.interval)
	require.NotNil(t, svc.stopCh)
}

func TestAnnouncementSchedulerService_StartStop(t *testing.T) {
	repo := &schedulerRepoStub{batchPublishCount: 0}
	svc := NewAnnouncementSchedulerService(repo, 50*time.Millisecond)

	// 启动服务
	svc.Start()

	// 等待至少执行一次
	time.Sleep(100 * time.Millisecond)

	// 停止服务
	svc.Stop()

	// 验证至少执行了一次（启动时立即执行）
	calls := repo.getBatchPublishCalls()
	require.GreaterOrEqual(t, calls, 1, "should have called BatchPublishScheduled at least once")
}

func TestAnnouncementSchedulerService_StartNilService(t *testing.T) {
	var svc *AnnouncementSchedulerService
	// 不应该 panic
	svc.Start()
}

func TestAnnouncementSchedulerService_StartNilRepo(t *testing.T) {
	svc := &AnnouncementSchedulerService{
		announcementRepo: nil,
		interval:         time.Minute,
		stopCh:           make(chan struct{}),
	}
	// 不应该 panic，应该直接返回
	svc.Start()
}

func TestAnnouncementSchedulerService_StartZeroInterval(t *testing.T) {
	repo := &schedulerRepoStub{}
	svc := NewAnnouncementSchedulerService(repo, 0)
	// 不应该 panic，应该直接返回
	svc.Start()
	require.Equal(t, 0, repo.getBatchPublishCalls())
}

func TestAnnouncementSchedulerService_StartNegativeInterval(t *testing.T) {
	repo := &schedulerRepoStub{}
	svc := NewAnnouncementSchedulerService(repo, -time.Minute)
	// 不应该 panic，应该直接返回
	svc.Start()
	require.Equal(t, 0, repo.getBatchPublishCalls())
}

func TestAnnouncementSchedulerService_StopNilService(t *testing.T) {
	var svc *AnnouncementSchedulerService
	// 不应该 panic
	svc.Stop()
}

func TestAnnouncementSchedulerService_StopMultipleTimes(t *testing.T) {
	repo := &schedulerRepoStub{}
	svc := NewAnnouncementSchedulerService(repo, 50*time.Millisecond)

	svc.Start()
	time.Sleep(100 * time.Millisecond)

	// 多次调用 Stop 不应该 panic
	svc.Stop()
	svc.Stop()
	svc.Stop()
}

func TestAnnouncementSchedulerService_RunOnceWithError(t *testing.T) {
	repo := &schedulerRepoStub{
		batchPublishErr: errors.New("db error"),
	}
	svc := NewAnnouncementSchedulerService(repo, 50*time.Millisecond)

	svc.Start()
	time.Sleep(100 * time.Millisecond)
	svc.Stop()

	// 即使有错误，也应该继续运行
	require.GreaterOrEqual(t, repo.getBatchPublishCalls(), 1)
}

func TestAnnouncementSchedulerService_RunOnceWithPublishedCount(t *testing.T) {
	repo := &schedulerRepoStub{
		batchPublishCount: 5,
	}
	svc := NewAnnouncementSchedulerService(repo, 50*time.Millisecond)

	svc.Start()
	time.Sleep(100 * time.Millisecond)
	svc.Stop()

	// 验证执行了
	require.GreaterOrEqual(t, repo.getBatchPublishCalls(), 1)
}

func TestAnnouncementSchedulerService_PeriodicExecution(t *testing.T) {
	repo := &schedulerRepoStub{batchPublishCount: 0}
	svc := NewAnnouncementSchedulerService(repo, 30*time.Millisecond)

	svc.Start()

	// 等待足够长的时间让定时器触发多次
	time.Sleep(150 * time.Millisecond)

	svc.Stop()

	// 应该执行了多次（启动时1次 + 定时器触发多次）
	calls := repo.getBatchPublishCalls()
	require.GreaterOrEqual(t, calls, 3, "should have called BatchPublishScheduled multiple times")
}
