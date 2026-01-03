package service

import (
	"xlpanel/internal/domain"
	"xlpanel/internal/infra"
)

type OrderService struct {
	orders  *infra.Repository[domain.Order]
	metrics *infra.MetricsRegistry
}

func NewOrderService(orders *infra.Repository[domain.Order], metrics *infra.MetricsRegistry) *OrderService {
	return &OrderService{orders: orders, metrics: metrics}
}

func (s *OrderService) CreateOrder(order domain.Order) domain.Order {
	s.metrics.Increment("order_created")
	return s.orders.Add(order)
}

func (s *OrderService) ListOrders() []domain.Order {
	return s.orders.List()
}
