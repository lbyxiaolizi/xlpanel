// Package web provides the HTML template rendering system for OpenHost.
// It supports multiple themes, internationalization, and a flexible hook system
// for extending template functionality.
package web

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"

	"github.com/openhost/openhost/internal/infrastructure/i18n"
)

// Context keys for storing values in gin.Context
const (
	ContextUserKey       = "user"
	ContextCSRFKey       = "csrf_token"
	ContextCurrencyKey   = "currency"
	ContextThemeKey      = "theme"
	ContextLangKey       = "lang"
	ContextFlashKey      = "flash"
	ContextBreadcrumbKey = "breadcrumbs"
	ContextSiteConfigKey = "site_config"
)

// Layout types for different areas of the application
const (
	LayoutPublic = "public"
	LayoutClient = "client"
	LayoutAdmin  = "admin"
	LayoutAuth   = "auth"
)

// Flash represents a flash message
type Flash struct {
	Type    string // success, error, warning, info
	Message string
}

// Breadcrumb represents a navigation breadcrumb
type Breadcrumb struct {
	Title string
	URL   string
	Icon  string
}

// SiteConfig contains site-wide configuration passed to templates
type SiteConfig struct {
	Name         string
	Tagline      string
	LogoURL      string
	FaviconURL   string
	SupportEmail string
	SupportPhone string
	Address      string
	SocialLinks  map[string]string
	FooterText   string
	CustomCSS    string
	CustomJS     string
}

// RenderOptions configures how a template should be rendered
type RenderOptions struct {
	Layout      string      // Layout to use: public, client, admin, auth
	Title       string      // Page title
	Description string      // Meta description
	StatusCode  int         // HTTP status code (default 200)
	ContentType string      // Response content type
	Headers     http.Header // Additional headers
}

// Renderer handles template rendering with theme and i18n support
type Renderer struct {
	mu            sync.RWMutex
	theme         string
	basePath      string
	providers     []HookProvider
	i18nManager   *i18n.Manager
	themeManager  *ThemeManager
	siteConfig    *SiteConfig
	templateCache map[string]*template.Template
	cacheEnabled  bool
	funcMap       template.FuncMap
}

// defaultRenderer is the global default renderer instance
var defaultRenderer = NewRenderer("default")

// NewRenderer creates a new Renderer with the specified theme
func NewRenderer(theme string, providers ...HookProvider) *Renderer {
	if theme == "" {
		theme = "default"
	}

	// Initialize i18n manager with built-in translations
	i18nMgr := i18n.NewManager("./locales")
	_ = i18nMgr.LoadBuiltinLanguages()

	r := &Renderer{
		theme:         theme,
		basePath:      "./themes",
		providers:     providers,
		i18nManager:   i18nMgr,
		themeManager:  DefaultThemeManager,
		siteConfig:    &SiteConfig{Name: "OpenHost", Tagline: "Modern Hosting Platform"},
		templateCache: make(map[string]*template.Template),
		cacheEnabled:  false, // Disable cache by default for development
	}

	// Initialize default template functions
	r.initFuncMap()

	return r
}

// initFuncMap initializes the default template function map
func (r *Renderer) initFuncMap() {
	r.funcMap = template.FuncMap{
		// Internationalization
		"t": func(key string, args ...any) string { return key },
		"T": func(key string, args ...any) string { return key },

		// Template helpers
		"dict":       templateDict,
		"list":       templateList,
		"safe":       templateSafe,
		"safeURL":    templateSafeURL,
		"safeJS":     templateSafeJS,
		"safeCSS":    templateSafeCSS,
		"html":       templateHTML,
		"urlEncode":  templateURLEncode,
		"jsonEncode": templateJSONEncode,

		// Comparison
		"eq": templateEq,
		"ne": templateNe,
		"lt": templateLt,
		"le": templateLe,
		"gt": templateGt,
		"ge": templateGe,

		// Math
		"add": templateAdd,
		"sub": templateSub,
		"mul": templateMul,
		"div": templateDiv,
		"mod": templateMod,

		// String manipulation
		"lower":     strings.ToLower,
		"upper":     strings.ToUpper,
		"title":     strings.Title,
		"trim":      strings.TrimSpace,
		"replace":   strings.ReplaceAll,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"split":     strings.Split,
		"join":      strings.Join,
		"truncate":  templateTruncate,
		"pluralize": templatePluralize,

		// Date/Time formatting
		"formatDate":     templateFormatDate,
		"formatDateTime": templateFormatDateTime,
		"formatTime":     templateFormatTime,
		"timeAgo":        templateTimeAgo,
		"now":            time.Now,

		// Number formatting
		"formatNumber":   templateFormatNumber,
		"formatCurrency": templateFormatCurrency,
		"formatPercent":  templateFormatPercent,
		"formatBytes":    templateFormatBytes,

		// Conditional helpers
		"default":  templateDefault,
		"coalesce": templateCoalesce,
		"ternary":  templateTernary,

		// Collection helpers
		"first":   templateFirst,
		"last":    templateLast,
		"slice":   templateSlice,
		"len":     templateLen,
		"reverse": templateReverse,
		"sortBy":  templateSortBy,
		"groupBy": templateGroupBy,
		"pluck":   templatePluck,
		"unique":  templateUnique,
		"in":      templateIn,
		"notIn":   templateNotIn,
		"range":   templateRange,

		// URL helpers
		"asset":       r.assetURL,
		"url":         templateURL,
		"currentURL":  func() string { return "" },
		"isActiveURL": templateIsActiveURL,

		// Form helpers
		"selected": templateSelected,
		"checked":  templateChecked,
		"disabled": templateDisabled,

		// Hook system
		"hook": r.hook,

		// Debug (only in development)
		"dump":  templateDump,
		"debug": templateDebug,
	}
}

// SetI18nManager sets the i18n manager for this renderer
func (r *Renderer) SetI18nManager(mgr *i18n.Manager) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.i18nManager = mgr
}

// SetThemeManager sets the theme manager for this renderer
func (r *Renderer) SetThemeManager(tm *ThemeManager) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.themeManager = tm
}

// SetSiteConfig sets the site-wide configuration
func (r *Renderer) SetSiteConfig(config *SiteConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.siteConfig = config
}

// SetCacheEnabled enables or disables template caching
func (r *Renderer) SetCacheEnabled(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cacheEnabled = enabled
	if !enabled {
		r.templateCache = make(map[string]*template.Template)
	}
}

// AddHookProvider adds a hook provider to the renderer
func (r *Renderer) AddHookProvider(provider HookProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers = append(r.providers, provider)
}

// RegisterFunc registers a custom template function
func (r *Renderer) RegisterFunc(name string, fn interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.funcMap[name] = fn
}

// SetRenderer sets the global default renderer
func SetRenderer(renderer *Renderer) {
	if renderer != nil {
		defaultRenderer = renderer
	}
}

// GetRenderer returns the global default renderer
func GetRenderer() *Renderer {
	return defaultRenderer
}

// Render renders a template using the default renderer
func Render(c *gin.Context, templateName string, data gin.H) {
	defaultRenderer.Render(c, templateName, data)
}

// RenderWithOptions renders a template with custom options
func RenderWithOptions(c *gin.Context, templateName string, data gin.H, opts RenderOptions) {
	defaultRenderer.RenderWithOptions(c, templateName, data, opts)
}

// Render renders a template with the given data
func (r *Renderer) Render(c *gin.Context, templateName string, data gin.H) {
	r.RenderWithOptions(c, templateName, data, RenderOptions{})
}

// RenderWithOptions renders a template with custom options
func (r *Renderer) RenderWithOptions(c *gin.Context, templateName string, data gin.H, opts RenderOptions) {
	if data == nil {
		data = gin.H{}
	}

	if opts.Layout == "" {
		opts.Layout = r.inferLayout(templateName)
	}

	// Set status code
	statusCode := opts.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	// Inject context values
	r.injectContextData(c, data)

	// Get language and translator
	lang := r.getLanguage(c)
	data["Lang"] = lang
	data["Languages"] = r.i18nManager.GetLanguages()

	translator := r.i18nManager.GetTranslator(lang)
	data["T"] = translator.Func()

	// Get active theme
	activeTheme := r.getActiveTheme(c)
	data["Theme"] = activeTheme

	// Load theme configuration
	themeConfig, err := r.themeManager.GetTheme(activeTheme)
	if err != nil {
		themeConfig, _ = r.themeManager.GetTheme(r.themeManager.GetFallbackTheme())
	}
	data["ThemeConfig"] = themeConfig

	// Set site configuration
	data["Site"] = r.siteConfig

	// Set page metadata
	if opts.Title != "" {
		data["Title"] = opts.Title
	}
	if opts.Description != "" {
		data["Description"] = opts.Description
	}
	data["Year"] = time.Now().Year()

	// Build template with translator functions
	tmpl, err := r.loadTemplates(activeTheme, templateName, translator, opts.Layout)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Determine which template to execute
	execName := r.determineExecName(activeTheme, templateName, opts.Layout)

	// Set custom headers
	for key, values := range opts.Headers {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Render the template
	c.Render(statusCode, render.HTML{
		Template: tmpl,
		Name:     execName,
		Data:     data,
	})
}

func (r *Renderer) inferLayout(templateName string) string {
	switch {
	case strings.HasPrefix(templateName, "admin/"):
		return LayoutAdmin
	case strings.HasPrefix(templateName, "client/"):
		return LayoutClient
	case templateName == "login.html" || templateName == "register.html" || templateName == "install.html":
		return LayoutAuth
	default:
		return LayoutPublic
	}
}

// injectContextData injects common context values into the data map
func (r *Renderer) injectContextData(c *gin.Context, data gin.H) {
	data["User"] = contextValue(c, ContextUserKey)
	data["CSRFToken"] = contextValue(c, ContextCSRFKey)
	data["Currency"] = contextValue(c, ContextCurrencyKey)

	// Flash messages
	if flash := contextValue(c, ContextFlashKey); flash != nil {
		data["Flash"] = flash
	}

	// Breadcrumbs
	if breadcrumbs := contextValue(c, ContextBreadcrumbKey); breadcrumbs != nil {
		data["Breadcrumbs"] = breadcrumbs
	}

	// Request info
	data["Request"] = gin.H{
		"URL":     c.Request.URL.String(),
		"Path":    c.Request.URL.Path,
		"Method":  c.Request.Method,
		"Host":    c.Request.Host,
		"IsHTTPS": c.Request.TLS != nil,
	}
}

// getActiveTheme determines the active theme from context or defaults
func (r *Renderer) getActiveTheme(c *gin.Context) string {
	// Check context override
	if override, ok := c.Get(ContextThemeKey); ok {
		if theme, ok := override.(string); ok && theme != "" {
			return theme
		}
	}

	// Check query parameter (for theme preview)
	if theme := c.Query("theme"); theme != "" && r.themeManager.ThemeExists(theme) {
		return theme
	}

	return r.theme
}

// getLanguage determines the language to use for rendering
func (r *Renderer) getLanguage(c *gin.Context) string {
	// 1. Check context (set by middleware)
	if lang, ok := c.Get(ContextLangKey); ok {
		if l, ok := lang.(string); ok && l != "" {
			return l
		}
	}

	// 2. Check query parameter
	if lang := c.Query("lang"); lang != "" {
		// Set cookie for persistence
		c.SetCookie("lang", lang, 86400*365, "/", "", false, false)
		return lang
	}

	// 3. Check cookie
	if lang, err := c.Cookie("lang"); err == nil && lang != "" {
		return lang
	}

	// 4. Check Accept-Language header
	if lang := c.GetHeader("Accept-Language"); lang != "" {
		if len(lang) >= 2 {
			code := strings.ToLower(lang[:2])
			for _, l := range r.i18nManager.GetLanguages() {
				if l.Code == code {
					return code
				}
			}
		}
	}

	// 5. Default to English
	return "en"
}

// loadTemplates loads and parses templates for a given theme and page
func (r *Renderer) loadTemplates(theme, templateName string, translator *i18n.Translator, layout string) (*template.Template, error) {
	cacheKey := fmt.Sprintf("%s:%s:%s:%s", theme, templateName, layout, translator.Lang())

	// Check cache
	if r.cacheEnabled {
		r.mu.RLock()
		if tmpl, ok := r.templateCache[cacheKey]; ok {
			r.mu.RUnlock()
			return tmpl, nil
		}
		r.mu.RUnlock()
	}

	themeDir := filepath.Join(r.basePath, theme)
	files, err := r.resolveTemplateFiles(themeDir, templateName, layout)
	if err != nil {
		return nil, err
	}

	// Create function map with translator
	funcMap := make(template.FuncMap)
	for k, v := range r.funcMap {
		funcMap[k] = v
	}
	funcMap["t"] = translator.T
	funcMap["T"] = translator.T

	tmpl, err := template.New("templates").Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	// Cache the template
	if r.cacheEnabled {
		r.mu.Lock()
		r.templateCache[cacheKey] = tmpl
		r.mu.Unlock()
	}

	return tmpl, nil
}

// resolveTemplateFiles determines which template files to load
func (r *Renderer) resolveTemplateFiles(themeDir, templateName, layout string) ([]string, error) {
	var files []string

	// Add layout file based on type
	layoutFile := "base.html"
	if layout != "" {
		layoutFile = layout + ".html"
	}

	// Check for layout in layouts directory
	layoutPath := filepath.Join(themeDir, "layouts", layoutFile)
	if fileExists(layoutPath) {
		files = append(files, layoutPath)
	} else {
		// Fallback to base.html
		basePath := filepath.Join(themeDir, "layouts", "base.html")
		if fileExists(basePath) {
			files = append(files, basePath)
		}
	}

	// Add partials directory if it exists
	partialsDir := filepath.Join(themeDir, "partials")
	if dirExists(partialsDir) {
		partialFiles, _ := filepath.Glob(filepath.Join(partialsDir, "*.html"))
		files = append(files, partialFiles...)
	}

	// Add the page template
	pagePath := filepath.Join(themeDir, "pages", templateName)
	if fileExists(pagePath) {
		files = append(files, pagePath)
		return files, nil
	}

	// Try without pages prefix
	altPath := filepath.Join(themeDir, templateName)
	if fileExists(altPath) {
		files = append(files, altPath)
		return files, nil
	}

	// If no specific template found, return error
	if len(files) == 0 {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}

	return files, nil
}

// determineExecName determines which template name to execute
func (r *Renderer) determineExecName(theme, templateName, layout string) string {
	themeDir := filepath.Join(r.basePath, theme)

	// Check for layout file
	layoutFile := "base.html"
	if layout != "" {
		layoutFile = layout + ".html"
	}

	layoutPath := filepath.Join(themeDir, "layouts", layoutFile)
	if fileExists(layoutPath) {
		return layoutFile
	}

	basePath := filepath.Join(themeDir, "layouts", "base.html")
	if fileExists(basePath) {
		return "base.html"
	}

	return templateName
}

// assetURL returns the full URL for a theme asset
func (r *Renderer) assetURL(path string) string {
	return fmt.Sprintf("/static/%s", strings.TrimPrefix(path, "/"))
}

// hook executes registered hook providers and returns combined HTML
func (r *Renderer) hook(name string, data any) template.HTML {
	var output strings.Builder
	for _, provider := range r.providers {
		html, ok, err := provider.HookHTML(name, data)
		if err != nil || !ok || html == "" {
			continue
		}
		output.WriteString(html)
	}
	return template.HTML(output.String())
}

// ClearTemplateCache clears the template cache
func (r *Renderer) ClearTemplateCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.templateCache = make(map[string]*template.Template)
}

// Helper functions

func contextValue(c *gin.Context, key string) any {
	if value, ok := c.Get(key); ok {
		return value
	}
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// SetFlash sets a flash message in the context
func SetFlash(c *gin.Context, flashType, message string) {
	c.Set(ContextFlashKey, &Flash{Type: flashType, Message: message})
}

// SetBreadcrumbs sets breadcrumbs in the context
func SetBreadcrumbs(c *gin.Context, breadcrumbs []Breadcrumb) {
	c.Set(ContextBreadcrumbKey, breadcrumbs)
}

// AddBreadcrumb adds a breadcrumb to the context
func AddBreadcrumb(c *gin.Context, title, url string) {
	existing, _ := c.Get(ContextBreadcrumbKey)
	breadcrumbs, _ := existing.([]Breadcrumb)
	breadcrumbs = append(breadcrumbs, Breadcrumb{Title: title, URL: url})
	c.Set(ContextBreadcrumbKey, breadcrumbs)
}
