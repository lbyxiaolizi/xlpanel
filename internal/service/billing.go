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
	payments *infra.Repository[domain.Payment]
	metrics  *infra.MetricsRegistry
}

func NewBillingService(
	invoices *infra.Repository[domain.Invoice],
	coupons *infra.CouponRepository,
	payments *infra.Repository[domain.Payment],
	metrics *infra.MetricsRegistry,
) *BillingService {
	return &BillingService{
		invoices: invoices,
		coupons:  coupons,
		payments: payments,
		metrics:  metrics,
	}
}

func (s *BillingService) CreateInvoiceWithLines(
	tenantID string,
	customerID string,
	lines []domain.InvoiceLine,
	currency string,
) domain.Invoice {
	total := 0.0
	for i := range lines {
		if lines[i].Quantity <= 0 {
			lines[i].Quantity = 1
		}
		lines[i].Total = float64(lines[i].Quantity) * lines[i].UnitPrice
		total += lines[i].Total
	}
	invoice := domain.Invoice{
		ID:         core.NewID(),
		TenantID:   tenantID,
		CustomerID: customerID,
		Amount:     total,
		Currency:   currency,
		Status:     "unpaid",
		DueAt:      time.Now().UTC().Add(7 * 24 * time.Hour),
		Lines:      lines,
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

func (s *BillingService) ApplyCoupon(code string, amount float64) (float64, bool) {
	raw, ok := s.coupons.Get(code)
	if !ok {
		return amount, false
	}
	coupon, ok := raw.(domain.Coupon)
	if !ok {
		return amount, false
	}
	discount := float64(coupon.PercentOff) / 100.0
	newAmount := amount - amount*discount
	if newAmount < 0 {
		newAmount = 0
	}
	s.metrics.Increment("coupon_applied")
	return newAmount, true
}

func (s *BillingService) RecordPayment(payment domain.Payment) domain.Payment {
	payment.ID = core.NewID()
	payment.Status = "received"
	payment.ReceivedAt = time.Now().UTC()
	s.metrics.Increment("payment_received")
	return s.payments.Add(payment)
}

func (s *BillingService) ListInvoices() []domain.Invoice {
	return s.invoices.List()
}

func (s *BillingService) ListPayments() []domain.Payment {
	return s.payments.List()
}
