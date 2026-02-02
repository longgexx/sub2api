package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/httpclient"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func NewProxyExitInfoProber(cfg *config.Config) service.ProxyExitInfoProber {
	insecure := false
	allowPrivate := false
	validateResolvedIP := true
	if cfg != nil {
		insecure = cfg.Security.ProxyProbe.InsecureSkipVerify
		allowPrivate = cfg.Security.URLAllowlist.AllowPrivateHosts
		validateResolvedIP = cfg.Security.URLAllowlist.Enabled
	}
	if insecure {
		log.Printf("[ProxyProbe] Warning: insecure_skip_verify is not allowed and will cause probe failure.")
	}
	return &proxyProbeService{
		ipInfoURL:          defaultIPInfoURL,
		fallbackIPInfoURL:  fallbackIPInfoURL,
		insecureSkipVerify: insecure,
		allowPrivateHosts:  allowPrivate,
		validateResolvedIP: validateResolvedIP,
	}
}

const (
	defaultIPInfoURL         = "https://ipapi.co/json/"
	fallbackIPInfoURL        = "http://ip-api.com/json/?lang=zh-CN"
	defaultProxyProbeTimeout = 30 * time.Second
)

type proxyProbeService struct {
	ipInfoURL          string
	fallbackIPInfoURL  string
	insecureSkipVerify bool
	allowPrivateHosts  bool
	validateResolvedIP bool
}

func (s *proxyProbeService) ProbeProxy(ctx context.Context, proxyURL string) (*service.ProxyExitInfo, int64, error) {
	client, err := httpclient.GetClient(httpclient.Options{
		ProxyURL:           proxyURL,
		Timeout:            defaultProxyProbeTimeout,
		InsecureSkipVerify: s.insecureSkipVerify,
		ProxyStrict:        true,
		ValidateResolvedIP: s.validateResolvedIP,
		AllowPrivateHosts:  s.allowPrivateHosts,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create proxy client: %w", err)
	}

	// 尝试主服务 (ipapi.co - 支持 IPv6)
	info, latencyMs, shouldFallback, err := s.probeWithService(ctx, client, s.ipInfoURL, parseIPAPICoResponse)
	if err == nil {
		log.Printf("[ProxyProbe] Success: IP=%s, City=%s, Region=%s, Country=%s (%s), Latency=%dms",
			info.IP, info.City, info.Region, info.Country, info.CountryCode, latencyMs)
		return info, latencyMs, nil
	}

	// 只有在 IP 查询失败时才尝试备用服务
	// HTTP 错误（如 500、429）不触发 fallback，直接返回错误
	if shouldFallback && s.fallbackIPInfoURL != "" {
		log.Printf("[ProxyProbe] Primary service failed (IP query error): %v, trying fallback", err)
		info, fallbackLatencyMs, _, fallbackErr := s.probeWithService(ctx, client, s.fallbackIPInfoURL, parseIPAPIResponse)
		if fallbackErr == nil {
			log.Printf("[ProxyProbe] Fallback success: IP=%s, City=%s, Region=%s, Country=%s (%s), Latency=%dms",
				info.IP, info.City, info.Region, info.Country, info.CountryCode, fallbackLatencyMs)
			return info, fallbackLatencyMs, nil
		}
		return nil, latencyMs, fmt.Errorf("all IP info services failed: primary=%v, fallback=%v", err, fallbackErr)
	}

	return nil, latencyMs, err
}

// responseParser 定义响应解析函数类型
type responseParser func(body []byte) (*service.ProxyExitInfo, error)

// probeWithService 使用指定的服务进行 IP 探测
// 返回值：info, latencyMs, shouldFallback, error
// shouldFallback 为 true 表示是 IP 查询失败（如 IPv6 不支持），应该尝试备用服务
// shouldFallback 为 false 表示是网络/HTTP 错误，不应该 fallback
func (s *proxyProbeService) probeWithService(ctx context.Context, client *http.Client, serviceURL string, parser responseParser) (*service.ProxyExitInfo, int64, bool, error) {
	startTime := time.Now()
	req, err := http.NewRequestWithContext(ctx, "GET", serviceURL, nil)
	if err != nil {
		return nil, 0, false, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置 User-Agent，ipapi.co 需要此头部否则返回 429
	req.Header.Set("User-Agent", "ipapi.co/#go-v1.5")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, false, fmt.Errorf("proxy connection failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	latencyMs := time.Since(startTime).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		// HTTP 错误不触发 fallback
		return nil, latencyMs, false, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, latencyMs, false, fmt.Errorf("failed to read response: %w", err)
	}

	info, err := parser(body)
	if err != nil {
		// 解析错误或 IP 查询失败，应该尝试 fallback
		return nil, latencyMs, true, err
	}

	return info, latencyMs, false, nil
}

// parseIPAPIResponse 解析 ip-api.com 响应
func parseIPAPIResponse(body []byte) (*service.ProxyExitInfo, error) {
	var ipInfo struct {
		Status      string `json:"status"`
		Message     string `json:"message"`
		Query       string `json:"query"`
		City        string `json:"city"`
		Region      string `json:"region"`
		RegionName  string `json:"regionName"`
		Country     string `json:"country"`
		CountryCode string `json:"countryCode"`
	}

	if err := json.Unmarshal(body, &ipInfo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if strings.ToLower(ipInfo.Status) != "success" {
		msg := ipInfo.Message
		if msg == "" {
			msg = "ip-api request failed"
		}
		return nil, fmt.Errorf("ip-api request failed: %s", msg)
	}

	region := ipInfo.RegionName
	if region == "" {
		region = ipInfo.Region
	}

	return &service.ProxyExitInfo{
		IP:          ipInfo.Query,
		City:        ipInfo.City,
		Region:      region,
		Country:     ipInfo.Country,
		CountryCode: ipInfo.CountryCode,
	}, nil
}

// parseIPAPICoResponse 解析 ipapi.co 响应（支持 IPv6）
func parseIPAPICoResponse(body []byte) (*service.ProxyExitInfo, error) {
	var ipInfo struct {
		IP          string `json:"ip"`
		City        string `json:"city"`
		Region      string `json:"region"`
		CountryName string `json:"country_name"`
		CountryCode string `json:"country_code"`
		Error       bool   `json:"error"`
		Reason      string `json:"reason"`
	}

	if err := json.Unmarshal(body, &ipInfo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if ipInfo.Error {
		reason := ipInfo.Reason
		if reason == "" {
			reason = "ipapi.co request failed"
		}
		return nil, fmt.Errorf("ipapi.co request failed: %s", reason)
	}

	if ipInfo.IP == "" {
		return nil, fmt.Errorf("ipapi.co returned empty IP")
	}

	return &service.ProxyExitInfo{
		IP:          ipInfo.IP,
		City:        ipInfo.City,
		Region:      ipInfo.Region,
		Country:     ipInfo.CountryName,
		CountryCode: ipInfo.CountryCode,
	}, nil
}
