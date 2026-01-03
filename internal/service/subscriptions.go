package service

import (
	"errors"
	"time"

	"xlpanel/internal/core"
	"xlpanel/internal/domain"
	"xlpanel/internal/infra"
)

type SubscriptionService struct {
	subscriptions *infra.Repository[domain.Subscription]
	catalog       *CatalogService
	billing       *BillingService
	metrics       *infra.MetricsRegistry
}

func NewSubscriptionService(
	subscriptions *infra.Repository[domain.Subscription],
	catalog *CatalogService,
	billing *BillingService,
	metrics *infra.MetricsRegistry,
) *SubscriptionService {
	return &SubscriptionService{
		subscriptions: subscriptions,
		catalog:       catalog,
		billing:       billing,
		metrics:       metrics,
	}
}

func (s *SubscriptionService) CreateSubscription(sub domain.Subscription) (domain.Subscription, error) {
	product, ok := s.catalog.GetProduct(sub.ProductCode)
	if !ok {
		return domain.Subscription{}, errors.New("product not found")
	}
	if product.BillingDays <= 0 {
		product.BillingDays = 30
	}
	sub.ID = core.NewID()
	sub.Status = "active"
	sub.CreatedAt = time.Now().UTC()
	sub.NextBillAt = time.Now().UTC().Add(time.Duration(product.BillingDays) * 24 * time.Hour)

	s.metrics.Increment("subscription_created")
	return s.subscriptions.Add(sub), nil
}

func (s *SubscriptionService) ListSubscriptions() []domain.Subscription {
	return s.subscriptions.List()
}

func (s *SubscriptionService) GenerateInvoice(subscriptionID string) (domain.Invoice, error) {
	subscription, ok := s.subscriptions.Get(subscriptionID)
	if !ok {
		return domain.Invoice{}, errors.New("subscription not found")
	}
	product, ok := s.catalog.GetProduct(subscription.ProductCode)
	if !ok {
		return domain.Invoice{}, errors.New("product not found")
	}
	line := domain.InvoiceLine{
		Description: product.Name,
		Quantity:    1,
		UnitPrice:   product.UnitPrice,
		Total:       product.UnitPrice,
	}
	invoice := s.billing.CreateInvoiceWithLines(
		subscription.TenantID,
		subscription.CustomerID,
		[]domain.InvoiceLine{line},
		product.Currency,
	)

	if product.BillingDays <= 0 {
		product.BillingDays = 30
	}
	subscription.NextBillAt = time.Now().UTC().Add(time.Duration(product.BillingDays) * 24 * time.Hour)
	s.subscriptions.Add(subscription)
	s.metrics.Increment("subscription_invoiced")
	return invoice, nil
}
