package service

import (
	"context"
	"errors"

	"xlpanel/internal/infra"
	"xlpanel/internal/plugins"
)

type PaymentGateway struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Enabled  bool   `json:"enabled"`
}

type PaymentsService struct {
	metrics  *infra.MetricsRegistry
	gateways []PaymentGateway
	registry *plugins.Registry
}

func NewPaymentsService(metrics *infra.MetricsRegistry, registry *plugins.Registry) *PaymentsService {
	return &PaymentsService{metrics: metrics, gateways: []PaymentGateway{}, registry: registry}
}

func (s *PaymentsService) RegisterGateway(gateway PaymentGateway) PaymentGateway {
	s.gateways = append(s.gateways, gateway)
	s.metrics.Increment("gateway_registered")
	return gateway
}

func (s *PaymentsService) ListGateways() []PaymentGateway {
	return append([]PaymentGateway{}, s.gateways...)
}

func (s *PaymentsService) Charge(ctx context.Context, provider string, request plugins.PaymentRequest) (plugins.PaymentResponse, error) {
	plugin, ok := s.registry.Get(provider)
	if !ok {
		return plugins.PaymentResponse{}, errors.New("payment provider not found")
	}
	response, err := plugin.Charge(ctx, request)
	if err != nil {
		return plugins.PaymentResponse{}, err
	}
	s.metrics.Increment("payment_plugin_charge")
	return response, nil
}
