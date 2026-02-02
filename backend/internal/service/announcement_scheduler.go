package service

import (
	"context"
	"log"
	"sync"
	"time"
)

// AnnouncementSchedulerService 定时发布公告服务
type AnnouncementSchedulerService struct {
	announcementRepo AnnouncementRepository
	interval         time.Duration
	stopCh           chan struct{}
	stopOnce         sync.Once
	wg               sync.WaitGroup
}

// NewAnnouncementSchedulerService 创建定时发布服务实例
func NewAnnouncementSchedulerService(announcementRepo AnnouncementRepository, interval time.Duration) *AnnouncementSchedulerService {
	return &AnnouncementSchedulerService{
		announcementRepo: announcementRepo,
		interval:         interval,
		stopCh:           make(chan struct{}),
	}
}

// Start 启动定时发布服务
func (s *AnnouncementSchedulerService) Start() {
	if s == nil || s.announcementRepo == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// 启动时立即执行一次
		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop 停止定时发布服务
func (s *AnnouncementSchedulerService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

// runOnce 执行一次定时发布检查
func (s *AnnouncementSchedulerService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	updated, err := s.announcementRepo.BatchPublishScheduled(ctx, now)
	if err != nil {
		log.Printf("[AnnouncementScheduler] Batch publish scheduled announcements failed: %v", err)
		return
	}
	if updated > 0 {
		log.Printf("[AnnouncementScheduler] Published %d scheduled announcements", updated)
	}
}
