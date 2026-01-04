// Package web provides common HTTP middleware for OpenHost.
// These middleware handle language detection, theme selection,
// authentication context, and other cross-cutting concerns.
package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LanguageMiddleware detects and sets the user's preferred language
func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check query parameter first (allows switching)
		if lang := c.Query("lang"); lang != "" {
			c.Set(ContextLangKey, lang)
			// Set cookie for persistence
			c.SetCookie("lang", lang, 86400*365, "/", "", false, false)
			c.Next()
			return
		}

		// Check cookie
		if lang, err := c.Cookie("lang"); err == nil && lang != "" {
			c.Set(ContextLangKey, lang)
			c.Next()
			return
		}

		// Check Accept-Language header
		if acceptLang := c.GetHeader("Accept-Language"); acceptLang != "" {
			// Parse first language code
			if len(acceptLang) >= 2 {
				lang := strings.ToLower(acceptLang[:2])
				c.Set(ContextLangKey, lang)
				c.Next()
				return
			}
		}

		// Default to English
		c.Set(ContextLangKey, "en")
		c.Next()
	}
}

// ThemeMiddleware sets the active theme based on query parameter or config
func ThemeMiddleware(defaultTheme string) gin.HandlerFunc {
	return func(c *gin.Context) {
		theme := defaultTheme

		// Allow theme preview via query parameter
		if previewTheme := c.Query("theme"); previewTheme != "" {
			if DefaultThemeManager.ThemeExists(previewTheme) {
				theme = previewTheme
			}
		}

		c.Set(ContextThemeKey, theme)
		c.Next()
	}
}

// CurrencyMiddleware sets the user's currency preference
func CurrencyMiddleware(defaultCurrency string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currency := defaultCurrency

		// Check query parameter
		if curr := c.Query("currency"); curr != "" {
			currency = strings.ToUpper(curr)
			c.SetCookie("currency", currency, 86400*365, "/", "", false, false)
		} else if curr, err := c.Cookie("currency"); err == nil && curr != "" {
			currency = curr
		}

		c.Set(ContextCurrencyKey, currency)
		c.Next()
	}
}

// SecurityHeaders adds common security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "SAMEORIGIN")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS filter
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions policy (formerly Feature-Policy)
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// CSPMiddleware adds Content-Security-Policy header
func CSPMiddleware(policy string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if policy == "" {
			policy = "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https://fonts.gstatic.com; connect-src 'self'"
		}
		c.Header("Content-Security-Policy", policy)
		c.Next()
	}
}

// NoCacheMiddleware prevents caching for specific routes
func NoCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// CacheMiddleware sets caching headers
func CacheMiddleware(maxAge time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		seconds := int(maxAge.Seconds())
		c.Header("Cache-Control", "public, max-age="+string(rune(seconds)))
		c.Next()
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists (e.g., from load balancer)
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// RecoveryMiddleware handles panics gracefully
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error
				// logger.Error("Panic recovered", "error", err, "stack", string(debug.Stack()))

				// Check if this is an API request
				if strings.HasPrefix(c.Request.URL.Path, "/api/") {
					ServerError(c, "Internal server error")
					return
				}

				// Render error page for web requests
				c.HTML(http.StatusInternalServerError, "errors/500.html", gin.H{
					"Title": "Server Error",
				})
			}
		}()
		c.Next()
	}
}

// NotFoundHandler handles 404 errors
func NotFoundHandler(c *gin.Context) {
	// Check if this is an API request
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		NotFound(c, "Resource not found")
		return
	}

	// Render 404 page for web requests
	Render(c, "errors/404.html", gin.H{
		"Title": "Page Not Found",
	})
}

// MethodNotAllowedHandler handles 405 errors
func MethodNotAllowedHandler(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.JSON(http.StatusMethodNotAllowed, APIResponse{
			Code:    http.StatusMethodNotAllowed,
			Message: "Method not allowed",
		})
		return
	}

	c.HTML(http.StatusMethodNotAllowed, "errors/405.html", gin.H{
		"Title": "Method Not Allowed",
	})
}

// MaintenanceModeMiddleware shows maintenance page when enabled
func MaintenanceModeMiddleware(enabled bool, allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}

		// Check if IP is allowed
		clientIP := c.ClientIP()
		for _, ip := range allowedIPs {
			if ip == clientIP {
				c.Next()
				return
			}
		}

		// Check if this is an API request
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			ServiceUnavailable(c, "Service under maintenance")
			return
		}

		// Render maintenance page
		c.HTML(http.StatusServiceUnavailable, "maintenance.html", gin.H{
			"Title": "Under Maintenance",
		})
		c.Abort()
	}
}

// Helper function to generate request ID
func generateRequestID() string {
	// Simple implementation - in production, use UUID
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
