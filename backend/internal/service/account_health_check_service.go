package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/util/urlvalidator"
)

// AccountHealthCheckService 定期检测错误状态的 API Key 账号
// 如果账号恢复正常，自动清除错误状态
type AccountHealthCheckService struct {
	accountRepo  AccountRepository
	httpUpstream HTTPUpstream
	cfg          *config.Config

	stopCh     chan struct{}
	stopOnce   sync.Once
	wg         sync.WaitGroup
	cancelFunc context.CancelFunc // 用于取消当前检测周期
	cancelMu   sync.Mutex         // 保护 cancelFunc
}

// NewAccountHealthCheckService 创建健康检测服务
func NewAccountHealthCheckService(
	accountRepo AccountRepository,
	httpUpstream HTTPUpstream,
	cfg *config.Config,
) *AccountHealthCheckService {
	return &AccountHealthCheckService{
		accountRepo:  accountRepo,
		httpUpstream: httpUpstream,
		cfg:          cfg,
		stopCh:       make(chan struct{}),
	}
}

// Start 启动健康检测服务
func (s *AccountHealthCheckService) Start() {
	if !s.cfg.AccountHealthCheck.Enabled {
		log.Println("[AccountHealthCheck] disabled")
		return
	}

	interval := time.Duration(s.cfg.AccountHealthCheck.CheckIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 10 * time.Second
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		log.Printf("[AccountHealthCheck] started, interval=%v, max_concurrency=%d",
			interval, s.cfg.AccountHealthCheck.MaxConcurrency)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// 启动时立即执行一次
		s.checkErrorAccounts()

		for {
			select {
			case <-ticker.C:
				s.checkErrorAccounts()
			case <-s.stopCh:
				log.Println("[AccountHealthCheck] stopped")
				return
			}
		}
	}()
}

// Stop 停止健康检测服务
func (s *AccountHealthCheckService) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		// 取消当前正在执行的检测周期
		s.cancelMu.Lock()
		if s.cancelFunc != nil {
			s.cancelFunc()
		}
		s.cancelMu.Unlock()
	})
	s.wg.Wait()
}

// checkErrorAccounts 检测所有错误状态的 API Key 账号
func (s *AccountHealthCheckService) checkErrorAccounts() {
	// 创建可取消的 context，用于优雅退出
	// 单轮检测最多运行 5 分钟，避免无限期运行
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	s.cancelMu.Lock()
	s.cancelFunc = cancel
	s.cancelMu.Unlock()
	defer cancel()

	// 使用 semaphore 控制并发
	maxConcurrency := s.cfg.AccountHealthCheck.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}
	sem := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	pageSize := 100
	page := 1
	totalChecked := 0

	// 分页循环，直到没有更多账号
	for {
		// 检查是否需要停止
		select {
		case <-s.stopCh:
			wg.Wait()
			return
		case <-ctx.Done():
			wg.Wait()
			return
		default:
		}

		// 查询错误状态的 API Key 账号
		accounts, _, err := s.accountRepo.ListWithFilters(ctx, pagination.PaginationParams{
			Page:     page,
			PageSize: pageSize,
		}, "", AccountTypeAPIKey, StatusError, "")
		if err != nil {
			log.Printf("[AccountHealthCheck] failed to list error accounts (page %d): %v", page, err)
			break
		}

		if len(accounts) == 0 {
			break
		}

		for i := range accounts {
			account := &accounts[i]
			// 只检测 API Key 类型的账号
			if account.Type != AccountTypeAPIKey {
				continue
			}

			// 检查是否需要停止
			select {
			case <-s.stopCh:
				wg.Wait()
				return
			case <-ctx.Done():
				wg.Wait()
				return
			default:
			}

			wg.Add(1)

			// 获取信号量（可取消）
			select {
			case sem <- struct{}{}:
				// 成功获取信号量
			case <-s.stopCh:
				wg.Done()
				wg.Wait()
				return
			case <-ctx.Done():
				wg.Done()
				wg.Wait()
				return
			}

			// 再次检查是否已取消，避免启动不必要的 goroutine
			select {
			case <-s.stopCh:
				<-sem // 释放信号量
				wg.Done()
				wg.Wait()
				return
			case <-ctx.Done():
				<-sem // 释放信号量
				wg.Done()
				wg.Wait()
				return
			default:
			}

			go func(acc *Account) {
				defer wg.Done()
				defer func() { <-sem }() // 释放信号量

				s.checkAndRecoverAccount(ctx, acc)
			}(account)
			totalChecked++
		}

		// 如果返回的账号数少于 pageSize，说明没有更多了
		if len(accounts) < pageSize {
			break
		}
		page++
	}

	wg.Wait()

	if totalChecked > 0 {
		log.Printf("[AccountHealthCheck] checked %d error accounts", totalChecked)
	}
}

// checkAndRecoverAccount 检测单个账号并尝试恢复
func (s *AccountHealthCheckService) checkAndRecoverAccount(ctx context.Context, account *Account) {
	// 为每个账号创建独立的超时 context，避免共享 context 超时影响其他账号
	checkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 根据平台选择验证方法
	var err error
	switch account.Platform {
	case PlatformAnthropic:
		err = s.verifyAnthropicAccount(checkCtx, account)
	case PlatformOpenAI:
		err = s.verifyOpenAIAccount(checkCtx, account)
	case PlatformGemini:
		err = s.verifyGeminiAccount(checkCtx, account)
	default:
		// 不支持的平台，跳过
		return
	}

	if err != nil {
		// 记录失败原因（debug 级别，避免日志过多）
		slog.Debug("account_health_check_failed",
			"account_id", account.ID,
			"platform", account.Platform,
			"error", err)
		return
	}

	// 验证成功，清除错误状态
	if clearErr := s.accountRepo.ClearError(checkCtx, account.ID); clearErr != nil {
		log.Printf("[AccountHealthCheck] failed to clear error for account %d: %v",
			account.ID, clearErr)
		return
	}

	log.Printf("[AccountHealthCheck] account %d (%s/%s) recovered successfully",
		account.ID, account.Platform, account.Name)
}

// verifyAnthropicAccount 验证 Anthropic API Key 账号
func (s *AccountHealthCheckService) verifyAnthropicAccount(ctx context.Context, account *Account) error {
	apiKey := account.GetCredential("api_key")
	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("no API key available")
	}

	baseURL := account.GetBaseURL()
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	// 校验 base_url
	normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	apiURL := normalizedBaseURL + "/v1/messages"

	// 构建最小请求
	payload := map[string]any{
		"model": claude.DefaultTestModel,
		"messages": []map[string]any{
			{
				"role":    "user",
				"content": "hi",
			},
		},
		"max_tokens": 1,
		"stream":     false,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("x-api-key", apiKey)

	// 获取代理 URL
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// 读取响应体（用于错误信息，限制大小防止内存问题）
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// 解析错误信息
	return fmt.Errorf("API returned %d: %s", resp.StatusCode, truncateHealthCheckString(string(body), 200))
}

// verifyOpenAIAccount 验证 OpenAI API Key 账号
func (s *AccountHealthCheckService) verifyOpenAIAccount(ctx context.Context, account *Account) error {
	apiKey := account.GetOpenAIApiKey()
	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("no API key available")
	}

	var apiURL string
	baseURL := account.GetOpenAIBaseURL()
	if baseURL == "" {
		// 默认使用 OpenAI 官方 API（已包含 /v1）
		apiURL = "https://api.openai.com/v1/responses"
	} else {
		// 自定义 base_url：校验后拼接 /responses（与网关逻辑一致）
		normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return fmt.Errorf("invalid base URL: %w", err)
		}
		apiURL = normalizedBaseURL + "/responses"
	}

	// 构建 Responses API 格式的请求
	payload := map[string]any{
		"model": "gpt-4o-mini", // 使用便宜的模型进行验证
		"input": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{"type": "input_text", "text": "hi"},
				},
			},
		},
		"max_output_tokens": 1, // 限制输出 token，减少费用
		"stream":            false,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 获取代理 URL
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// 读取响应体（用于错误信息，限制大小防止内存问题）
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("API returned %d: %s", resp.StatusCode, truncateHealthCheckString(string(body), 200))
}

// verifyGeminiAccount 验证 Gemini API Key 账号
func (s *AccountHealthCheckService) verifyGeminiAccount(ctx context.Context, account *Account) error {
	apiKey := account.GetCredential("api_key")
	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("no API key available")
	}

	baseURL := account.GetCredential("base_url")
	if baseURL == "" {
		baseURL = geminicli.AIStudioBaseURL
	}

	// 校验 base_url
	normalizedBaseURL, err := s.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	// 使用 generateContent（非流式）进行验证
	apiURL := fmt.Sprintf("%s/v1beta/models/%s:generateContent",
		normalizedBaseURL, geminicli.DefaultTestModel)

	// 构建最小请求
	payload := map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]any{
					{"text": "hi"},
				},
			},
		},
		"generationConfig": map[string]any{
			"maxOutputTokens": 1,
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	// 获取代理 URL
	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	resp, err := s.httpUpstream.Do(req, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// 读取响应体（用于错误信息，限制大小防止内存问题）
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("API returned %d: %s", resp.StatusCode, truncateHealthCheckString(string(body), 200))
}

// truncateHealthCheckString 截断字符串
func truncateHealthCheckString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// validateUpstreamBaseURL 校验上游 base_url，防止 SSRF 和误配置
func (s *AccountHealthCheckService) validateUpstreamBaseURL(raw string) (string, error) {
	if s.cfg == nil {
		return "", errors.New("config is not available")
	}
	if !s.cfg.Security.URLAllowlist.Enabled {
		normalized, err := urlvalidator.ValidateURLFormat(raw, s.cfg.Security.URLAllowlist.AllowInsecureHTTP)
		if err != nil {
			return "", fmt.Errorf("invalid base_url: %w", err)
		}
		return normalized, nil
	}
	normalized, err := urlvalidator.ValidateHTTPSURL(raw, urlvalidator.ValidationOptions{
		AllowedHosts:     s.cfg.Security.URLAllowlist.UpstreamHosts,
		RequireAllowlist: true,
		AllowPrivate:     s.cfg.Security.URLAllowlist.AllowPrivateHosts,
	})
	if err != nil {
		return "", fmt.Errorf("invalid base_url: %w", err)
	}
	return normalized, nil
}
