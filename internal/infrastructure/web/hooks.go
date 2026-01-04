// Package web provides the hook system for extending OpenHost templates.
// Hooks allow plugins and modules to inject HTML content at specific points
// in templates, enabling customization without modifying theme files.
package web

import (
	"errors"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

// Hook names for common extension points
const (
	// Layout hooks
	HookHeadStart     = "head_start"
	HookHeadEnd       = "head_end"
	HookBodyStart     = "body_start"
	HookBodyEnd       = "body_end"
	HookNavbarStart   = "navbar_start"
	HookNavbarEnd     = "navbar_end"
	HookSidebarStart  = "sidebar_start"
	HookSidebarEnd    = "sidebar_end"
	HookFooterStart   = "footer_start"
	HookFooterEnd     = "footer_end"

	// Dashboard hooks
	HookDashboardWidgets     = "dashboard_widgets"
	HookClientSidebar        = "client_sidebar"
	HookAdminSidebar         = "admin_sidebar"
	HookClientDashboardTop   = "client_dashboard_top"
	HookClientDashboardBottom = "client_dashboard_bottom"
	HookAdminDashboardTop    = "admin_dashboard_top"
	HookAdminDashboardBottom = "admin_dashboard_bottom"

	// Service hooks
	HookServiceDetails = "service_details"
	HookServiceActions = "service_actions"
	HookServiceSidebar = "service_sidebar"

	// Invoice hooks
	HookInvoiceItems   = "invoice_items"
	HookInvoiceSummary = "invoice_summary"
	HookPaymentMethods = "payment_methods"

	// Product hooks
	HookProductDetails = "product_details"
	HookProductConfig  = "product_config"
	HookOrderSummary   = "order_summary"

	// Profile hooks
	HookProfileSidebar = "profile_sidebar"
	HookProfileTabs    = "profile_tabs"
)

// HookProvider interface for components that can inject HTML into hooks
type HookProvider interface {
	// HookHTML returns HTML content to inject at the specified hook point
	// Returns (html, handled, error) where handled indicates if the hook was processed
	HookHTML(hook string, data any) (string, bool, error)

	// Priority returns the priority of this provider (lower = earlier execution)
	Priority() int
}

// HookFunc is a function type for simple hook handlers
type HookFunc func(data any) (string, error)

// FuncHookProvider wraps a function as a HookProvider
type FuncHookProvider struct {
	hookName string
	handler  HookFunc
	priority int
}

// NewFuncHookProvider creates a new function-based hook provider
func NewFuncHookProvider(hookName string, handler HookFunc, priority int) *FuncHookProvider {
	return &FuncHookProvider{
		hookName: hookName,
		handler:  handler,
		priority: priority,
	}
}

// HookHTML implements HookProvider
func (p *FuncHookProvider) HookHTML(hook string, data any) (string, bool, error) {
	if hook != p.hookName {
		return "", false, nil
	}
	html, err := p.handler(data)
	return html, true, err
}

// Priority implements HookProvider
func (p *FuncHookProvider) Priority() int {
	return p.priority
}

// HookRegistry manages registered hook providers
type HookRegistry struct {
	mu        sync.RWMutex
	providers []HookProvider
	sorted    bool
}

// DefaultHookRegistry is the global hook registry
var DefaultHookRegistry = NewHookRegistry()

// NewHookRegistry creates a new hook registry
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		providers: make([]HookProvider, 0),
	}
}

// Register adds a hook provider to the registry
func (r *HookRegistry) Register(provider HookProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers = append(r.providers, provider)
	r.sorted = false
}

// RegisterFunc registers a function as a hook handler
func (r *HookRegistry) RegisterFunc(hookName string, handler HookFunc, priority int) {
	r.Register(NewFuncHookProvider(hookName, handler, priority))
}

// Unregister removes a hook provider from the registry
func (r *HookRegistry) Unregister(provider HookProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.providers {
		if p == provider {
			r.providers = append(r.providers[:i], r.providers[i+1:]...)
			return
		}
	}
}

// Execute executes all providers for a hook and returns combined HTML
func (r *HookRegistry) Execute(hook string, data any) template.HTML {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Sort providers by priority if needed
	if !r.sorted {
		r.sortProviders()
	}

	var output strings.Builder
	for _, provider := range r.providers {
		html, handled, err := provider.HookHTML(hook, data)
		if err != nil || !handled || html == "" {
			continue
		}
		output.WriteString(html)
	}
	return template.HTML(output.String())
}

// sortProviders sorts providers by priority (lower = earlier)
func (r *HookRegistry) sortProviders() {
	// Note: This should be called with write lock held
	// Using simple bubble sort for small number of providers
	n := len(r.providers)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if r.providers[j].Priority() > r.providers[j+1].Priority() {
				r.providers[j], r.providers[j+1] = r.providers[j+1], r.providers[j]
			}
		}
	}
	r.sorted = true
}

// GetProviders returns all registered providers
func (r *HookRegistry) GetProviders() []HookProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]HookProvider, len(r.providers))
	copy(result, r.providers)
	return result
}

// Clear removes all registered providers
func (r *HookRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers = make([]HookProvider, 0)
	r.sorted = false
}

// Global convenience functions

// RegisterHook registers a hook handler with the default registry
func RegisterHook(hookName string, handler HookFunc, priority int) {
	DefaultHookRegistry.RegisterFunc(hookName, handler, priority)
}

// RegisterHookProvider registers a hook provider with the default registry
func RegisterHookProvider(provider HookProvider) {
	DefaultHookRegistry.Register(provider)
}

// ExecuteHook executes a hook using the default registry
func ExecuteHook(hook string, data any) template.HTML {
	return DefaultHookRegistry.Execute(hook, data)
}

// CompositeHookProvider combines multiple hook providers
type CompositeHookProvider struct {
	providers []HookProvider
}

// NewCompositeHookProvider creates a new composite provider
func NewCompositeHookProvider(providers ...HookProvider) *CompositeHookProvider {
	return &CompositeHookProvider{providers: providers}
}

// HookHTML implements HookProvider
func (c *CompositeHookProvider) HookHTML(hook string, data any) (string, bool, error) {
	var output strings.Builder
	handled := false

	for _, provider := range c.providers {
		html, h, err := provider.HookHTML(hook, data)
		if err != nil {
			return "", false, err
		}
		if h {
			handled = true
			output.WriteString(html)
		}
	}

	return output.String(), handled, nil
}

// Priority implements HookProvider
func (c *CompositeHookProvider) Priority() int {
	if len(c.providers) > 0 {
		return c.providers[0].Priority()
	}
	return 0
}

// Add adds a provider to the composite
func (c *CompositeHookProvider) Add(provider HookProvider) {
	c.providers = append(c.providers, provider)
}

// templateFiles finds all template files in a theme directory
func templateFiles(themeDir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(themeDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".html", ".tmpl", ".gotmpl":
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, errors.New("no templates found in theme directory")
	}
	return files, nil
}
