package service

import (
	"time"

	"xlpanel/internal/core"
	"xlpanel/internal/domain"
	"xlpanel/internal/infra"
)

type BillingService struct {
	invoices *infra.Repository[domain.Invoice]
	coupons  *infra.CouponRepository
	metrics  *infra.MetricsRegistry
}

func NewBillingService(
	invoices *infra.Repository[domain.Invoice],
	coupons *infra.CouponRepository,
	metrics *infra.MetricsRegistry,
) *BillingService {
	return &BillingService{
		invoices: invoices,
		coupons:  coupons,
		metrics:  metrics,
	}
}

func (s *BillingService) CreateInvoice(customerID string, amount float64, currency string) domain.Invoice {
	invoice := domain.Invoice{
		ID:         core.NewID(),
		CustomerID: customerID,
		Amount:     amount,
		Currency:   currency,
		Status:     "unpaid",
		DueAt:      time.Now().UTC().Add(7 * 24 * time.Hour),
	}
	s.metrics.Increment("invoice_created")
	return s.invoices.Add(invoice)
}

func (s *BillingService) AddCoupon(coupon domain.Coupon) domain.Coupon {
	s.metrics.Increment("coupon_created")
	s.coupons.Add(coupon.Code, coupon)
	return coupon
}

func (s *BillingService) ListCoupons() []domain.Coupon {
	raw := s.coupons.List()
	result := make([]domain.Coupon, 0, len(raw))
	for _, item := range raw {
		if coupon, ok := item.(domain.Coupon); ok {
			result = append(result, coupon)
		}
	}
	return result
}
