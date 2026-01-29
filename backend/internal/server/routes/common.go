package routes

import (
	"context"
	_ "embed"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed scripts/claude_code_setup.sh
var bashScriptTemplate string

//go:embed scripts/claude_code_setup.ps1
var powershellScriptTemplate string

// SimpleSettingsProvider is a simple implementation for script generation
type SimpleSettingsProvider interface {
	GetScriptSettings(ctx context.Context) (siteName, apiBaseURL string, err error)
}

// RegisterCommonRoutes 注册通用路由（健康检查、状态等）
func RegisterCommonRoutes(r *gin.Engine, settingsProvider SimpleSettingsProvider) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Claude Code 遥测日志（忽略，直接返回200）
	r.POST("/api/event_logging/batch", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup status endpoint (always returns needs_setup: false in normal mode)
	// This is used by the frontend to detect when the service has restarted after setup
	r.GET("/setup/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"needs_setup": false,
				"step":        "completed",
			},
		})
	})

	// Claude Code 配置脚本下载
	scripts := r.Group("/scripts")
	{
		// Bash 脚本 (macOS/Linux)
		scripts.GET("/claude-code-setup.sh", func(c *gin.Context) {
			script := generateScript(c, settingsProvider, bashScriptTemplate)
			c.Header("Content-Type", "text/x-shellscript; charset=utf-8")
			c.Header("Content-Disposition", "inline; filename=claude-code-setup.sh")
			c.String(http.StatusOK, script)
		})

		// PowerShell 脚本 (Windows)
		scripts.GET("/claude-code-setup.ps1", func(c *gin.Context) {
			script := generateScript(c, settingsProvider, powershellScriptTemplate)
			c.Header("Content-Type", "text/plain; charset=utf-8")
			c.Header("Content-Disposition", "inline; filename=claude-code-setup.ps1")
			c.String(http.StatusOK, script)
		})
	}
}

// generateScript generates script content with dynamic placeholders replaced
func generateScript(c *gin.Context, provider SimpleSettingsProvider, template string) string {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	// Get settings from provider
	siteName := "Sub2API"
	apiBaseURL := ""

	if provider != nil {
		name, baseURL, err := provider.GetScriptSettings(ctx)
		if err == nil {
			if name != "" {
				siteName = name
			}
			if baseURL != "" {
				apiBaseURL = baseURL
			}
		}
	}

	// If API Base URL is not configured, use request origin
	if apiBaseURL == "" {
		apiBaseURL = getServerURL(c)
	}

	// Get server URL from request
	serverURL := getServerURL(c)

	// Replace placeholders
	script := template
	script = strings.ReplaceAll(script, "{{SITE_NAME}}", siteName)
	script = strings.ReplaceAll(script, "{{API_BASE_URL}}", apiBaseURL)
	script = strings.ReplaceAll(script, "{{SERVER_URL}}", serverURL)

	return script
}

// getServerURL extracts the server URL from the request
func getServerURL(c *gin.Context) string {
	scheme := "https"
	if c.Request.TLS == nil {
		// Check X-Forwarded-Proto header (common with reverse proxies)
		if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		} else if c.Request.URL.Scheme != "" {
			scheme = c.Request.URL.Scheme
		} else {
			// Default to http for local development
			scheme = "http"
		}
	}

	host := c.Request.Host
	// Check X-Forwarded-Host header
	if forwardedHost := c.GetHeader("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}

	return scheme + "://" + host
}
