package repository

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ProxyProbeServiceSuite struct {
	suite.Suite
	ctx      context.Context
	proxySrv *httptest.Server
	prober   *proxyProbeService
}

func (s *ProxyProbeServiceSuite) SetupTest() {
	s.ctx = context.Background()
	s.prober = &proxyProbeService{
		ipInfoURL:         "http://ipapi-co.test/json/",
		fallbackIPInfoURL: "http://ip-api.test/json/?lang=zh-CN",
		allowPrivateHosts: true,
	}
}

func (s *ProxyProbeServiceSuite) TearDownTest() {
	if s.proxySrv != nil {
		s.proxySrv.Close()
		s.proxySrv = nil
	}
}

func (s *ProxyProbeServiceSuite) setupProxyServer(handler http.HandlerFunc) {
	s.proxySrv = newLocalTestServer(s.T(), handler)
}

func (s *ProxyProbeServiceSuite) TestProbeProxy_InvalidProxyURL() {
	_, _, err := s.prober.ProbeProxy(s.ctx, "://bad")
	require.Error(s.T(), err)
	require.ErrorContains(s.T(), err, "failed to create proxy client")
}

func (s *ProxyProbeServiceSuite) TestProbeProxy_UnsupportedProxyScheme() {
	_, _, err := s.prober.ProbeProxy(s.ctx, "ftp://127.0.0.1:1")
	require.Error(s.T(), err)
	require.ErrorContains(s.T(), err, "failed to create proxy client")
}

func (s *ProxyProbeServiceSuite) TestProbeProxy_Success() {
	seen := make(chan string, 1)
	s.setupProxyServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen <- r.RequestURI
		w.Header().Set("Content-Type", "application/json")
		// 使用 ipapi.co 响应格式
		_, _ = io.WriteString(w, `{"ip":"1.2.3.4","city":"c","region":"r","country_name":"cc","country_code":"CC"}`)
	}))

	info, latencyMs, err := s.prober.ProbeProxy(s.ctx, s.proxySrv.URL)
	require.NoError(s.T(), err, "ProbeProxy")
	require.GreaterOrEqual(s.T(), latencyMs, int64(0), "unexpected latency")
	require.Equal(s.T(), "1.2.3.4", info.IP)
	require.Equal(s.T(), "c", info.City)
	require.Equal(s.T(), "r", info.Region)
	require.Equal(s.T(), "cc", info.Country)
	require.Equal(s.T(), "CC", info.CountryCode)

	// Verify proxy received the request
	select {
	case uri := <-seen:
		require.Contains(s.T(), uri, "ipapi-co.test", "expected request to go through proxy")
	default:
		require.Fail(s.T(), "expected proxy to receive request")
	}
}

func (s *ProxyProbeServiceSuite) TestProbeProxy_NonOKStatus() {
	requestCount := 0
	s.setupProxyServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))

	_, _, err := s.prober.ProbeProxy(s.ctx, s.proxySrv.URL)
	require.Error(s.T(), err)
	require.ErrorContains(s.T(), err, "status: 503")
	// HTTP 错误不应该触发 fallback
	require.Equal(s.T(), 1, requestCount, "HTTP error should not trigger fallback")
}

func (s *ProxyProbeServiceSuite) TestProbeProxy_InvalidJSON() {
	s.setupProxyServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, "not-json")
	}))

	_, _, err := s.prober.ProbeProxy(s.ctx, s.proxySrv.URL)
	require.Error(s.T(), err)
	require.ErrorContains(s.T(), err, "failed to parse response")
}

func (s *ProxyProbeServiceSuite) TestProbeProxy_InvalidIPInfoURL() {
	s.prober.ipInfoURL = "://invalid-url"
	s.setupProxyServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	_, _, err := s.prober.ProbeProxy(s.ctx, s.proxySrv.URL)
	require.Error(s.T(), err, "expected error for invalid ipInfoURL")
}

func (s *ProxyProbeServiceSuite) TestProbeProxy_ProxyServerClosed() {
	s.setupProxyServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	s.proxySrv.Close()

	_, _, err := s.prober.ProbeProxy(s.ctx, s.proxySrv.URL)
	require.Error(s.T(), err, "expected error when proxy server is closed")
}

func (s *ProxyProbeServiceSuite) TestProbeProxy_IPv6Success() {
	s.setupProxyServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// 使用 ipapi.co 响应格式
		_, _ = io.WriteString(w, `{"ip":"2001:db8::1","city":"Beijing","region":"Beijing","country_name":"China","country_code":"CN"}`)
	}))

	info, latencyMs, err := s.prober.ProbeProxy(s.ctx, s.proxySrv.URL)
	require.NoError(s.T(), err, "ProbeProxy with IPv6")
	require.GreaterOrEqual(s.T(), latencyMs, int64(0))
	require.Equal(s.T(), "2001:db8::1", info.IP)
	require.Equal(s.T(), "Beijing", info.City)
	require.Equal(s.T(), "China", info.Country)
	require.Equal(s.T(), "CN", info.CountryCode)
}

func (s *ProxyProbeServiceSuite) TestProbeProxy_FallbackOnPrimaryFailure() {
	requestCount := 0
	s.setupProxyServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		if requestCount == 1 {
			// 主服务 (ipapi.co) 返回失败
			_, _ = io.WriteString(w, `{"error":true,"reason":"rate limited"}`)
		} else {
			// 备用服务 (ip-api.com) 返回成功
			_, _ = io.WriteString(w, `{"status":"success","query":"2001:db8::1","city":"Beijing","regionName":"Beijing","country":"China","countryCode":"CN"}`)
		}
	}))

	info, latencyMs, err := s.prober.ProbeProxy(s.ctx, s.proxySrv.URL)
	require.NoError(s.T(), err, "ProbeProxy should succeed with fallback")
	require.GreaterOrEqual(s.T(), latencyMs, int64(0))
	require.Equal(s.T(), "2001:db8::1", info.IP)
	require.Equal(s.T(), "Beijing", info.City)
	require.Equal(s.T(), "China", info.Country)
	require.Equal(s.T(), "CN", info.CountryCode)
	require.Equal(s.T(), 2, requestCount, "should have made 2 requests (primary + fallback)")
}

func (s *ProxyProbeServiceSuite) TestProbeProxy_BothServicesFail() {
	requestCount := 0
	s.setupProxyServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		if requestCount == 1 {
			// 主服务 (ipapi.co) 返回失败
			_, _ = io.WriteString(w, `{"error":true,"reason":"rate limited"}`)
		} else {
			// 备用服务 (ip-api.com) 也返回失败
			_, _ = io.WriteString(w, `{"status":"fail","message":"reserved range"}`)
		}
	}))

	_, _, err := s.prober.ProbeProxy(s.ctx, s.proxySrv.URL)
	require.Error(s.T(), err)
	require.ErrorContains(s.T(), err, "all IP info services failed")
	require.Equal(s.T(), 2, requestCount, "should have tried both services")
}

func (s *ProxyProbeServiceSuite) TestProbeProxy_NoFallbackConfigured() {
	s.prober.fallbackIPInfoURL = ""
	s.setupProxyServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// 主服务 (ipapi.co) 返回失败
		_, _ = io.WriteString(w, `{"error":true,"reason":"rate limited"}`)
	}))

	_, _, err := s.prober.ProbeProxy(s.ctx, s.proxySrv.URL)
	require.Error(s.T(), err)
	require.ErrorContains(s.T(), err, "ipapi.co request failed")
}

func TestProxyProbeServiceSuite(t *testing.T) {
	suite.Run(t, new(ProxyProbeServiceSuite))
}

func TestParseIPAPIResponse(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantIP      string
		wantCountry string
		wantErr     bool
		errContains string
	}{
		{
			name:        "success with IPv4",
			body:        `{"status":"success","query":"1.2.3.4","city":"Beijing","regionName":"Beijing","country":"China","countryCode":"CN"}`,
			wantIP:      "1.2.3.4",
			wantCountry: "China",
		},
		{
			name:        "success with IPv6",
			body:        `{"status":"success","query":"2001:db8::1","city":"Tokyo","regionName":"Tokyo","country":"Japan","countryCode":"JP"}`,
			wantIP:      "2001:db8::1",
			wantCountry: "Japan",
		},
		{
			name:        "fail status",
			body:        `{"status":"fail","message":"reserved range"}`,
			wantErr:     true,
			errContains: "reserved range",
		},
		{
			name:        "invalid json",
			body:        `not-json`,
			wantErr:     true,
			errContains: "failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := parseIPAPIResponse([]byte(tt.body))
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantIP, info.IP)
			require.Equal(t, tt.wantCountry, info.Country)
		})
	}
}

func TestParseIPAPICoResponse(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantIP      string
		wantCountry string
		wantErr     bool
		errContains string
	}{
		{
			name:        "success with IPv4",
			body:        `{"ip":"1.2.3.4","city":"Beijing","region":"Beijing","country_name":"China","country_code":"CN"}`,
			wantIP:      "1.2.3.4",
			wantCountry: "China",
		},
		{
			name:        "success with IPv6",
			body:        `{"ip":"2001:db8::1","city":"Tokyo","region":"Tokyo","country_name":"Japan","country_code":"JP"}`,
			wantIP:      "2001:db8::1",
			wantCountry: "Japan",
		},
		{
			name:        "error response",
			body:        `{"error":true,"reason":"rate limited"}`,
			wantErr:     true,
			errContains: "rate limited",
		},
		{
			name:        "empty IP",
			body:        `{"ip":"","city":"Beijing"}`,
			wantErr:     true,
			errContains: "empty IP",
		},
		{
			name:        "invalid json",
			body:        `not-json`,
			wantErr:     true,
			errContains: "failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := parseIPAPICoResponse([]byte(tt.body))
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantIP, info.IP)
			require.Equal(t, tt.wantCountry, info.Country)
		})
	}
}
