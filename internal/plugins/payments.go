package plugins

import "context"

type PaymentRequest struct {
	TenantID   string
	CustomerID string
	InvoiceID  string
	Amount     float64
	Currency   string
	Metadata   map[string]string
}

type PaymentResponse struct {
	TransactionID string
	Status        string
	Message       string
}

type PaymentPlugin interface {
	Name() string
	Provider() string
	Initialize(config map[string]string) error
	Charge(ctx context.Context, request PaymentRequest) (PaymentResponse, error)
}

type Registry struct {
	plugins map[string]PaymentPlugin
}

func NewRegistry() *Registry {
	return &Registry{plugins: make(map[string]PaymentPlugin)}
}

func (r *Registry) Register(plugin PaymentPlugin) {
	r.plugins[plugin.Provider()] = plugin
}

func (r *Registry) Get(provider string) (PaymentPlugin, bool) {
	plugin, ok := r.plugins[provider]
	return plugin, ok
}

func (r *Registry) List() []PaymentPlugin {
	result := make([]PaymentPlugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		result = append(result, plugin)
	}
	return result
}
