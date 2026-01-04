package invoice

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
	"github.com/openhost/openhost/internal/core/service/tax"
)

var (
	ErrInvoiceNotFound    = errors.New("invoice not found")
	ErrInvoiceAlreadyPaid = errors.New("invoice is already paid")
	ErrInvalidAmount      = errors.New("invalid payment amount")
	ErrInvoiceCancelled   = errors.New("invoice is cancelled")
)

// Service provides invoice management operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new invoice service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateInvoice creates a new invoice
func (s *Service) CreateInvoice(customerID uint64, currency string, dueDate time.Time, items []InvoiceItemRequest) (*domain.Invoice, error) {
	// Generate invoice number
	invoiceNumber := s.generateInvoiceNumber()

	invoice := &domain.Invoice{
		CustomerID:    customerID,
		InvoiceNumber: invoiceNumber,
		Status:        domain.InvoiceStatusUnpaid,
		Currency:      currency,
		DueDate:       dueDate,
	}

	// Calculate totals
	subtotal := decimal.Zero
	taxableSubtotal := decimal.Zero
	for _, item := range items {
		itemTotal := item.UnitPrice.Mul(item.Quantity)
		subtotal = subtotal.Add(itemTotal)
		if item.Taxable {
			taxableSubtotal = taxableSubtotal.Add(itemTotal.Sub(item.Discount))
		}

		invoice.LineItems = append(invoice.LineItems, domain.InvoiceItem{
			ServiceID:   item.ServiceID,
			Type:        item.Type,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Discount:    item.Discount,
			Total:       itemTotal.Sub(item.Discount),
			Taxable:     item.Taxable,
			PeriodStart: item.PeriodStart,
			PeriodEnd:   item.PeriodEnd,
		})
	}

	taxAmount, err := tax.NewCalculator(s.db).CalculateForCustomer(customerID, taxableSubtotal)
	if err != nil {
		return nil, err
	}

	invoice.Subtotal = subtotal
	invoice.TaxAmount = taxAmount
	invoice.Total = subtotal.Add(taxAmount).Sub(invoice.Discount)
	invoice.Balance = invoice.Total

	if err := s.db.Create(invoice).Error; err != nil {
		return nil, err
	}

	return invoice, nil
}

// CreateInvoiceFromOrder creates an invoice from an order
func (s *Service) CreateInvoiceFromOrder(order *domain.Order, dueDate time.Time) (*domain.Invoice, error) {
	invoiceNumber := s.generateInvoiceNumber()

	invoice := &domain.Invoice{
		CustomerID:    order.CustomerID,
		InvoiceNumber: invoiceNumber,
		Status:        domain.InvoiceStatusUnpaid,
		Currency:      order.Currency,
		DueDate:       dueDate,
		Subtotal:      order.Subtotal,
		Discount:      order.Discount,
		TaxAmount:     order.TaxAmount,
		Total:         order.Total,
		Balance:       order.Total,
	}

	// Create line items from order items
	for _, orderItem := range order.Items {
		invoiceItem := domain.InvoiceItem{
			ServiceID:   orderItem.ServiceID,
			Type:        "service",
			Description: orderItem.Description,
			Quantity:    decimal.NewFromInt(int64(orderItem.Quantity)),
			UnitPrice:   orderItem.SetupFee.Add(orderItem.RecurringFee),
			Discount:    orderItem.Discount,
			Total:       orderItem.Total,
			Taxable:     true,
		}
		invoice.LineItems = append(invoice.LineItems, invoiceItem)
	}

	if err := s.db.Create(invoice).Error; err != nil {
		return nil, err
	}

	// Update order with invoice ID
	s.db.Model(order).Update("invoice_id", invoice.ID)

	return invoice, nil
}

// CreateServiceRenewalInvoice creates a renewal invoice for a service
func (s *Service) CreateServiceRenewalInvoice(service *domain.Service, dueDate time.Time) (*domain.Invoice, error) {
	invoiceNumber := s.generateInvoiceNumber()

	// Calculate period
	periodStart := service.NextDueDate
	periodEnd := s.addBillingPeriod(periodStart, service.BillingCycle)

	invoice := &domain.Invoice{
		CustomerID:    service.CustomerID,
		InvoiceNumber: invoiceNumber,
		Status:        domain.InvoiceStatusUnpaid,
		Currency:      service.Currency,
		DueDate:       dueDate,
		Subtotal:      service.RecurringAmount,
		Total:         service.RecurringAmount,
		Balance:       service.RecurringAmount,
		LineItems: []domain.InvoiceItem{
			{
				ServiceID:   &service.ID,
				Type:        "renewal",
				Description: fmt.Sprintf("%s - %s to %s", service.Product.Name, periodStart.Format("Jan 2, 2006"), periodEnd.Format("Jan 2, 2006")),
				Quantity:    decimal.NewFromInt(1),
				UnitPrice:   service.RecurringAmount,
				Total:       service.RecurringAmount,
				Taxable:     true,
				PeriodStart: &periodStart,
				PeriodEnd:   &periodEnd,
			},
		},
	}

	taxAmount, err := tax.NewCalculator(s.db).CalculateForCustomer(service.CustomerID, service.RecurringAmount)
	if err != nil {
		return nil, err
	}
	invoice.TaxAmount = taxAmount
	invoice.Total = invoice.Subtotal.Add(taxAmount)
	invoice.Balance = invoice.Total

	if err := s.db.Create(invoice).Error; err != nil {
		return nil, err
	}

	return invoice, nil
}

// GetInvoice retrieves an invoice by ID
func (s *Service) GetInvoice(id uint64) (*domain.Invoice, error) {
	var invoice domain.Invoice
	if err := s.db.Preload("LineItems").Preload("Customer").First(&invoice, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvoiceNotFound
		}
		return nil, err
	}
	return &invoice, nil
}

// GetInvoiceByNumber retrieves an invoice by invoice number
func (s *Service) GetInvoiceByNumber(invoiceNumber string) (*domain.Invoice, error) {
	var invoice domain.Invoice
	if err := s.db.Preload("LineItems").Preload("Customer").
		Where("invoice_number = ?", invoiceNumber).First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvoiceNotFound
		}
		return nil, err
	}
	return &invoice, nil
}

// ListInvoices returns invoices for a customer
func (s *Service) ListInvoices(customerID uint64, status domain.InvoiceStatus, limit, offset int) ([]domain.Invoice, int64, error) {
	var invoices []domain.Invoice
	var total int64

	query := s.db.Model(&domain.Invoice{}).Where("customer_id = ?", customerID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	if err := query.Preload("LineItems").Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

// GetUnpaidInvoices returns all unpaid invoices for a customer
func (s *Service) GetUnpaidInvoices(customerID uint64) ([]domain.Invoice, error) {
	var invoices []domain.Invoice
	if err := s.db.Where("customer_id = ? AND status IN ?", customerID,
		[]domain.InvoiceStatus{domain.InvoiceStatusUnpaid, domain.InvoiceStatusOverdue}).
		Find(&invoices).Error; err != nil {
		return nil, err
	}
	return invoices, nil
}

// AddPayment records a payment against an invoice
func (s *Service) AddPayment(invoiceID uint64, amount decimal.Decimal, gateway, transactionID string) (*domain.Transaction, error) {
	var invoice domain.Invoice
	if err := s.db.First(&invoice, invoiceID).Error; err != nil {
		return nil, ErrInvoiceNotFound
	}

	if invoice.Status == domain.InvoiceStatusPaid {
		return nil, ErrInvoiceAlreadyPaid
	}
	if invoice.Status == domain.InvoiceStatusCancelled {
		return nil, ErrInvoiceCancelled
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	// Create transaction
	transaction := &domain.Transaction{
		CustomerID:     invoice.CustomerID,
		InvoiceID:      &invoice.ID,
		Type:           domain.TransactionTypePayment,
		Status:         domain.TransactionStatusCompleted,
		Currency:       invoice.Currency,
		Amount:         amount,
		Gateway:        gateway,
		GatewayTransID: transactionID,
		Description:    fmt.Sprintf("Payment for invoice %s", invoice.InvoiceNumber),
	}

	if err := s.db.Create(transaction).Error; err != nil {
		return nil, err
	}

	// Update invoice
	newAmountPaid := invoice.AmountPaid.Add(amount)
	newBalance := invoice.Total.Sub(newAmountPaid)

	updates := map[string]interface{}{
		"amount_paid": newAmountPaid,
		"balance":     newBalance,
	}

	if newBalance.LessThanOrEqual(decimal.Zero) {
		now := time.Now()
		updates["status"] = domain.InvoiceStatusPaid
		updates["paid_at"] = &now
		updates["balance"] = decimal.Zero
	}

	if err := s.db.Model(&invoice).Updates(updates).Error; err != nil {
		return nil, err
	}

	return transaction, nil
}

// CancelInvoice cancels an invoice
func (s *Service) CancelInvoice(invoiceID uint64) error {
	var invoice domain.Invoice
	if err := s.db.First(&invoice, invoiceID).Error; err != nil {
		return ErrInvoiceNotFound
	}

	if invoice.Status == domain.InvoiceStatusPaid {
		return errors.New("cannot cancel a paid invoice")
	}

	return s.db.Model(&invoice).Update("status", domain.InvoiceStatusCancelled).Error
}

// RefundInvoice creates a refund for a paid invoice
func (s *Service) RefundInvoice(invoiceID uint64, amount decimal.Decimal, reason string) (*domain.Transaction, error) {
	var invoice domain.Invoice
	if err := s.db.First(&invoice, invoiceID).Error; err != nil {
		return nil, ErrInvoiceNotFound
	}

	if invoice.Status != domain.InvoiceStatusPaid {
		return nil, errors.New("can only refund paid invoices")
	}

	if amount.GreaterThan(invoice.AmountPaid) {
		return nil, ErrInvalidAmount
	}

	// Create refund transaction
	transaction := &domain.Transaction{
		CustomerID:  invoice.CustomerID,
		InvoiceID:   &invoice.ID,
		Type:        domain.TransactionTypeRefund,
		Status:      domain.TransactionStatusCompleted,
		Currency:    invoice.Currency,
		Amount:      amount.Neg(), // Negative amount for refund
		Description: fmt.Sprintf("Refund for invoice %s: %s", invoice.InvoiceNumber, reason),
	}

	if err := s.db.Create(transaction).Error; err != nil {
		return nil, err
	}

	// Update invoice if fully refunded
	if amount.Equal(invoice.AmountPaid) {
		s.db.Model(&invoice).Update("status", domain.InvoiceStatusRefunded)
	}

	return transaction, nil
}

// MarkOverdueInvoices marks unpaid invoices past due date as overdue
func (s *Service) MarkOverdueInvoices() error {
	return s.db.Model(&domain.Invoice{}).
		Where("status = ? AND due_date < ?", domain.InvoiceStatusUnpaid, time.Now()).
		Update("status", domain.InvoiceStatusOverdue).Error
}

// generateInvoiceNumber generates a unique invoice number
func (s *Service) generateInvoiceNumber() string {
	return fmt.Sprintf("INV-%d-%d", time.Now().Year(), time.Now().UnixNano()%100000)
}

// addBillingPeriod adds a billing period to a date
func (s *Service) addBillingPeriod(from time.Time, billingCycle string) time.Time {
	switch billingCycle {
	case "monthly":
		return from.AddDate(0, 1, 0)
	case "quarterly":
		return from.AddDate(0, 3, 0)
	case "semi-annually", "semiannually":
		return from.AddDate(0, 6, 0)
	case "annually", "yearly":
		return from.AddDate(1, 0, 0)
	case "biennially":
		return from.AddDate(2, 0, 0)
	case "triennially":
		return from.AddDate(3, 0, 0)
	default:
		return from.AddDate(0, 1, 0)
	}
}

// InvoiceItemRequest represents a request to add an invoice item
type InvoiceItemRequest struct {
	ServiceID   *uint64
	Type        string
	Description string
	Quantity    decimal.Decimal
	UnitPrice   decimal.Decimal
	Discount    decimal.Decimal
	Taxable     bool
	PeriodStart *time.Time
	PeriodEnd   *time.Time
}
