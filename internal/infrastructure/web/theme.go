// Package web provides the theme management and configuration system for OpenHost.
// It handles loading theme manifests, managing theme assets, and providing
// a decoupled theme system that allows easy theme switching and customization.
package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ThemeConfig represents the theme.json configuration file
type ThemeConfig struct {
	Name               string            `json:"name"`
	Slug               string            `json:"slug"`
	Version            string            `json:"version"`
	Author             string            `json:"author"`
	AuthorURL          string            `json:"author_url"`
	Description        string            `json:"description"`
	License            string            `json:"license"`
	Screenshot         string            `json:"screenshot"`
	Type               string            `json:"type"` // "public", "admin", "both"
	MinOpenHostVersion string            `json:"min_openhost_version"`
	Supports           ThemeSupports     `json:"supports"`
	Settings           ThemeSettings     `json:"settings"`
	Layouts            map[string]string `json:"layouts"`
	Pages              map[string]string `json:"pages"`
	Assets             ThemeAssets       `json:"assets"`
	Hooks              []string          `json:"hooks"`
}

// ThemeSupports defines what features the theme supports
type ThemeSupports struct {
	DarkMode     bool `json:"dark_mode"`
	RTL          bool `json:"rtl"`
	CustomCSS    bool `json:"custom_css"`
	CustomJS     bool `json:"custom_js"`
	CustomHeader bool `json:"custom_header"`
	CustomFooter bool `json:"custom_footer"`
	ColorSchemes bool `json:"color_schemes"`
}

// ThemeSettings contains configurable theme settings
type ThemeSettings struct {
	PrimaryColor   string       `json:"primary_color"`
	SecondaryColor string       `json:"secondary_color"`
	AccentColor    string       `json:"accent_color"`
	LogoURL        string       `json:"logo_url"`
	FaviconURL     string       `json:"favicon_url"`
	FooterText     string       `json:"footer_text"`
	ShowLangSwitch bool         `json:"show_language_switcher"`
	ShowDarkToggle bool         `json:"show_dark_mode_toggle"`
	SocialLinks    SocialLinks  `json:"social_links"`
	CustomCSS      string       `json:"custom_css"`
	CustomJS       string       `json:"custom_js"`
	CustomHeader   string       `json:"custom_header"`
	CustomFooter   string       `json:"custom_footer"`
}

// SocialLinks contains social media URLs
type SocialLinks struct {
	GitHub   string `json:"github"`
	Twitter  string `json:"twitter"`
	Discord  string `json:"discord"`
	Facebook string `json:"facebook"`
	Telegram string `json:"telegram"`
	WeChat   string `json:"wechat"`
}

// ThemeAssets defines the CSS and JS files for the theme
type ThemeAssets struct {
	CSS []string `json:"css"`
	JS  []string `json:"js"`
}

// ThemeManager manages theme loading and caching
type ThemeManager struct {
	mu         sync.RWMutex
	basePath   string
	themes     map[string]*ThemeConfig
	active     string
	fallback   string
}

// DefaultThemeManager is the global theme manager instance
var DefaultThemeManager = NewThemeManager("./themes")

// NewThemeManager creates a new theme manager
func NewThemeManager(basePath string) *ThemeManager {
	return &ThemeManager{
		basePath: basePath,
		themes:   make(map[string]*ThemeConfig),
		active:   "default",
		fallback: "default",
	}
}

// LoadTheme loads a theme configuration from its directory
func (tm *ThemeManager) LoadTheme(name string) (*ThemeConfig, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check cache first
	if config, ok := tm.themes[name]; ok {
		return config, nil
	}

	// Load theme.json
	configPath := filepath.Join(tm.basePath, name, "theme.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		// If theme.json doesn't exist, create default config
		if os.IsNotExist(err) {
			config := tm.createDefaultConfig(name)
			tm.themes[name] = config
			return config, nil
		}
		return nil, fmt.Errorf("failed to read theme config: %w", err)
	}

	var config ThemeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse theme config: %w", err)
	}

	// Set defaults if not specified
	if config.Slug == "" {
		config.Slug = name
	}
	if config.Type == "" {
		config.Type = "both"
	}

	tm.themes[name] = &config
	return &config, nil
}

// createDefaultConfig creates a default theme configuration
func (tm *ThemeManager) createDefaultConfig(name string) *ThemeConfig {
	return &ThemeConfig{
		Name:        name,
		Slug:        name,
		Version:     "1.0.0",
		Description: "Default theme",
		Type:        "both",
		Supports: ThemeSupports{
			DarkMode:     true,
			RTL:          false,
			CustomCSS:    true,
			CustomJS:     true,
			CustomHeader: true,
			CustomFooter: true,
			ColorSchemes: false,
		},
		Settings: ThemeSettings{
			PrimaryColor:   "#6366f1",
			SecondaryColor: "#8b5cf6",
			AccentColor:    "#ec4899",
			ShowLangSwitch: true,
			ShowDarkToggle: true,
		},
		Layouts: map[string]string{
			"public": "layouts/base.html",
			"client": "layouts/base.html",
			"admin":  "layouts/base.html",
			"auth":   "layouts/base.html",
		},
		Assets: ThemeAssets{
			CSS: []string{"assets/css/main.css"},
			JS:  []string{"assets/js/main.js"},
		},
	}
}

// GetTheme returns a loaded theme, loading it if necessary
func (tm *ThemeManager) GetTheme(name string) (*ThemeConfig, error) {
	tm.mu.RLock()
	if config, ok := tm.themes[name]; ok {
		tm.mu.RUnlock()
		return config, nil
	}
	tm.mu.RUnlock()

	return tm.LoadTheme(name)
}

// SetActiveTheme sets the currently active theme
func (tm *ThemeManager) SetActiveTheme(name string) error {
	// Verify theme exists
	themePath := filepath.Join(tm.basePath, name)
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		return fmt.Errorf("theme '%s' does not exist", name)
	}

	tm.mu.Lock()
	tm.active = name
	tm.mu.Unlock()

	// Pre-load the theme
	_, err := tm.LoadTheme(name)
	return err
}

// GetActiveTheme returns the name of the currently active theme
func (tm *ThemeManager) GetActiveTheme() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.active
}

// SetFallbackTheme sets the fallback theme to use when active theme fails
func (tm *ThemeManager) SetFallbackTheme(name string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.fallback = name
}

// GetFallbackTheme returns the fallback theme name
func (tm *ThemeManager) GetFallbackTheme() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.fallback
}

// ListThemes returns all available themes
func (tm *ThemeManager) ListThemes() ([]string, error) {
	entries, err := os.ReadDir(tm.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read themes directory: %w", err)
	}

	var themes []string
	for _, entry := range entries {
		if entry.IsDir() {
			themes = append(themes, entry.Name())
		}
	}
	return themes, nil
}

// GetThemeAssetPath returns the full path to a theme asset
func (tm *ThemeManager) GetThemeAssetPath(themeName, assetPath string) string {
	return filepath.Join(tm.basePath, themeName, assetPath)
}

// ThemeExists checks if a theme exists
func (tm *ThemeManager) ThemeExists(name string) bool {
	themePath := filepath.Join(tm.basePath, name)
	info, err := os.Stat(themePath)
	return err == nil && info.IsDir()
}

// ReloadTheme forces a reload of a theme configuration
func (tm *ThemeManager) ReloadTheme(name string) (*ThemeConfig, error) {
	tm.mu.Lock()
	delete(tm.themes, name)
	tm.mu.Unlock()

	return tm.LoadTheme(name)
}

// ClearCache clears all cached theme configurations
func (tm *ThemeManager) ClearCache() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.themes = make(map[string]*ThemeConfig)
}
