package web

import (
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"

	"github.com/openhost/openhost/internal/infrastructure/i18n"
)

const (
	ContextUserKey     = "user"
	ContextCSRFKey     = "csrf_token"
	ContextCurrencyKey = "currency"
	ContextThemeKey    = "theme"
	ContextLangKey     = "lang"
)

type Renderer struct {
	theme       string
	basePath    string
	providers   []HookProvider
	i18nManager *i18n.Manager
}

var defaultRenderer = NewRenderer("default")

func NewRenderer(theme string, providers ...HookProvider) *Renderer {
	if theme == "" {
		theme = "default"
	}
	// Initialize i18n manager with built-in translations
	i18nMgr := i18n.NewManager("./locales")
	_ = i18nMgr.LoadBuiltinLanguages()

	return &Renderer{
		theme:       theme,
		basePath:    "./themes",
		providers:   providers,
		i18nManager: i18nMgr,
	}
}

// SetI18nManager sets the i18n manager for this renderer
func (r *Renderer) SetI18nManager(mgr *i18n.Manager) {
	r.i18nManager = mgr
}

func SetRenderer(renderer *Renderer) {
	if renderer != nil {
		defaultRenderer = renderer
	}
}

func Render(c *gin.Context, templateName string, data gin.H) {
	defaultRenderer.Render(c, templateName, data)
}

func (r *Renderer) Render(c *gin.Context, templateName string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["User"] = contextValue(c, ContextUserKey)
	data["CSRFToken"] = contextValue(c, ContextCSRFKey)
	data["Currency"] = contextValue(c, ContextCurrencyKey)
	data["Year"] = time.Now().Year()

	// Get language from context, query param, cookie, or default to "en"
	lang := r.getLanguage(c)
	data["Lang"] = lang
	data["Languages"] = r.i18nManager.GetLanguages()

	// Get translator for the current language
	translator := r.i18nManager.GetTranslator(lang)
	data["T"] = translator.Func()

	activeTheme := r.theme
	if override, ok := c.Get(ContextThemeKey); ok {
		if theme, ok := override.(string); ok && theme != "" {
			activeTheme = theme
		}
	}
	data["Theme"] = activeTheme

	tmpl, err := r.loadTemplates(activeTheme, templateName, translator)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	name := templateName
	basePath := filepath.Join(r.basePath, activeTheme, "layouts", "base.html")
	if fileExists(basePath) {
		name = "base.html"
	}
	c.Render(200, render.HTML{
		Template: tmpl,
		Name:     name,
		Data:     data,
	})
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
		return lang
	}

	// 3. Check cookie
	if lang, err := c.Cookie("lang"); err == nil && lang != "" {
		return lang
	}

	// 4. Check Accept-Language header
	if lang := c.GetHeader("Accept-Language"); lang != "" {
		// Simple parsing - just get first language
		if len(lang) >= 2 {
			code := lang[:2]
			// Check if we support this language
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

func (r *Renderer) loadTemplates(theme, templateName string, translator *i18n.Translator) (*template.Template, error) {
	themeDir := filepath.Join(r.basePath, theme)
	files, err := r.resolveTemplateFiles(themeDir, templateName)
	if err != nil {
		return nil, err
	}
	funcMap := template.FuncMap{
		"hook": r.hook,
		"t":    translator.T,
		"T":    translator.T,
		"dict": templateDict,
		"safe": templateSafe,
		"eq":   templateEq,
		"ne":   templateNe,
		"lt":   templateLt,
		"gt":   templateGt,
		"add":  templateAdd,
		"sub":  templateSub,
	}
	return template.New("themes").Funcs(funcMap).ParseFiles(files...)
}

// Template helper functions
func templateDict(values ...any) map[string]any {
	if len(values)%2 != 0 {
		return nil
	}
	dict := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			continue
		}
		dict[key] = values[i+1]
	}
	return dict
}

func templateSafe(s string) template.HTML {
	return template.HTML(s)
}

func templateEq(a, b any) bool {
	return a == b
}

func templateNe(a, b any) bool {
	return a != b
}

func templateLt(a, b int) bool {
	return a < b
}

func templateGt(a, b int) bool {
	return a > b
}

func templateAdd(a, b int) int {
	return a + b
}

func templateSub(a, b int) int {
	return a - b
}

func (r *Renderer) resolveTemplateFiles(themeDir, templateName string) ([]string, error) {
	var files []string

	basePath := filepath.Join(themeDir, "layouts", "base.html")
	if fileExists(basePath) {
		files = append(files, basePath)
	}

	pagePath := filepath.Join(themeDir, "pages", templateName)
	if fileExists(pagePath) {
		files = append(files, pagePath)
		return files, nil
	}

	altPath := filepath.Join(themeDir, templateName)
	if fileExists(altPath) {
		files = append(files, altPath)
		return files, nil
	}

	return templateFiles(themeDir)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (r *Renderer) hook(name string, data any) template.HTML {
	var output string
	for _, provider := range r.providers {
		html, ok, err := provider.HookHTML(name, data)
		if err != nil || !ok || html == "" {
			continue
		}
		output += html
	}
	return template.HTML(output)
}

func contextValue(c *gin.Context, key string) any {
	if value, ok := c.Get(key); ok {
		return value
	}
	return nil
}
