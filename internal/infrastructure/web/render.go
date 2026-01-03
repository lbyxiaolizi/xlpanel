package web

import (
	"html/template"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

const (
	ContextUserKey     = "user"
	ContextCSRFKey     = "csrf_token"
	ContextCurrencyKey = "currency"
	ContextThemeKey    = "theme"
)

type Renderer struct {
	theme     string
	basePath  string
	providers []HookProvider
}

var defaultRenderer = NewRenderer("default")

func NewRenderer(theme string, providers ...HookProvider) *Renderer {
	if theme == "" {
		theme = "default"
	}
	return &Renderer{
		theme:     theme,
		basePath:  "./themes",
		providers: providers,
	}
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

	activeTheme := r.theme
	if override, ok := c.Get(ContextThemeKey); ok {
		if theme, ok := override.(string); ok && theme != "" {
			activeTheme = theme
		}
	}

	tmpl, err := r.loadTemplates(activeTheme, templateName)
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

func (r *Renderer) loadTemplates(theme, templateName string) (*template.Template, error) {
	themeDir := filepath.Join(r.basePath, theme)
	files, err := r.resolveTemplateFiles(themeDir, templateName)
	if err != nil {
		return nil, err
	}
	funcMap := template.FuncMap{
		"hook": r.hook,
	}
	return template.New("themes").Funcs(funcMap).ParseFiles(files...)
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
