package service

import "xlpanel/internal/infra"

type PaymentGateway struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Enabled  bool   `json:"enabled"`
}

type PaymentsService struct {
	metrics  *infra.MetricsRegistry
	gateways []PaymentGateway
}

func NewPaymentsService(metrics *infra.MetricsRegistry) *PaymentsService {
	return &PaymentsService{metrics: metrics, gateways: []PaymentGateway{}}
}

func (s *PaymentsService) RegisterGateway(gateway PaymentGateway) PaymentGateway {
	s.gateways = append(s.gateways, gateway)
	s.metrics.Increment("gateway_registered")
	return gateway
}

func (s *PaymentsService) ListGateways() []PaymentGateway {
	return append([]PaymentGateway{}, s.gateways...)
}
