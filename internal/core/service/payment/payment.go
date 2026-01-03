package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
)

var (
	ErrGatewayNotFound        = errors.New("payment gateway not found")
	ErrGatewayInactive        = errors.New("payment gateway is inactive")
	ErrInvalidAmount          = errors.New("invalid payment amount")
	ErrPaymentFailed          = errors.New("payment failed")
	ErrRefundFailed           = errors.New("refund failed")
	ErrSubscriptionNotFound   = errors.New("subscription not found")
	ErrInsufficientBalance    = errors.New("insufficient credit balance")
)

// PaymentProcessor defines the interface for payment gateway implementations
type PaymentProcessor interface {
	Name() string
	ProcessPayment(request *PaymentRequest) (*PaymentResult, error)
	ProcessRefund(transactionID string, amount decimal.Decimal) (*RefundResult, error)
	CreateSubscription(request *SubscriptionRequest) (*SubscriptionResult, error)
	CancelSubscription(subscriptionID string) error
	ValidateWebhook(payload []byte, signature string) bool
	GetPaymentURL(request *PaymentRequest) (string, error)
	TokenizeCard(cardDetails *CardDetails) (string, error)
}

// PaymentRequest represents a payment request to a gateway
type PaymentRequest struct {
	CustomerID      uint64
	InvoiceID       uint64
	Amount          decimal.Decimal
	Currency        string
	Description     string
	PaymentMethodID *uint64
	CardToken       string
	SaveCard        bool
	ReturnURL       string
	CancelURL       string
	IPAddress       string
	Metadata        map[string]string
}

// PaymentResult represents the result of a payment
type PaymentResult struct {
	Success       bool
	TransactionID string
	GatewayRef    string
	Amount        decimal.Decimal
	Fee           decimal.Decimal
	Status        string
	Message       string
	RedirectURL   string
	CardToken     string
}

// RefundResult represents the result of a refund
type RefundResult struct {
	Success       bool
	RefundID      string
	Amount        decimal.Decimal
	Status        string
	Message       string
}

// SubscriptionRequest represents a subscription creation request
type SubscriptionRequest struct {
	CustomerID      uint64
	ServiceID       uint64
	Amount          decimal.Decimal
	Currency        string
	Interval        string
	IntervalCount   int
	PaymentMethodID uint64
	TrialDays       int
}

// SubscriptionResult represents the result of subscription creation
type SubscriptionResult struct {
	Success        bool
	SubscriptionID string
	Status         string
	CurrentPeriodEnd time.Time
	Message        string
}

// CardDetails represents card information for tokenization
type CardDetails struct {
	Number      string
	ExpiryMonth int
	ExpiryYear  int
	CVV         string
	Name        string
}

// Service provides payment operations
type Service struct {
	db         *gorm.DB
	processors map[string]PaymentProcessor
}

// NewService creates a new payment service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db:         db,
		processors: make(map[string]PaymentProcessor),
	}
}

// RegisterProcessor registers a payment processor
func (s *Service) RegisterProcessor(name string, processor PaymentProcessor) {
	s.processors[name] = processor
}

// GetGateway retrieves a payment gateway by ID
func (s *Service) GetGateway(id uint64) (*domain.PaymentGatewayModule, error) {
	var gateway domain.PaymentGatewayModule
	if err := s.db.First(&gateway, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGatewayNotFound
		}
		return nil, err
	}
	return &gateway, nil
}

// ListActiveGateways returns all active payment gateways
func (s *Service) ListActiveGateways() ([]domain.PaymentGatewayModule, error) {
	var gateways []domain.PaymentGatewayModule
	if err := s.db.Where("active = ? AND visible = ?", true, true).
		Order("sort_order ASC").Find(&gateways).Error; err != nil {
		return nil, err
	}
	return gateways, nil
}

// CreatePaymentRequest creates a new payment request
func (s *Service) CreatePaymentRequest(customerID, invoiceID, gatewayID uint64, amount decimal.Decimal, currency, ipAddress string) (*domain.PaymentRequest, error) {
	gateway, err := s.GetGateway(gatewayID)
	if err != nil {
		return nil, err
	}
	if !gateway.Active {
		return nil, ErrGatewayInactive
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	request := &domain.PaymentRequest{
		CustomerID: customerID,
		InvoiceID:  invoiceID,
		GatewayID:  gatewayID,
		Amount:     amount,
		Currency:   currency,
		Status:     "pending",
		IPAddress:  ipAddress,
		ExpiresAt:  &expiresAt,
	}

	if err := s.db.Create(request).Error; err != nil {
		return nil, err
	}

	return request, nil
}

// ProcessPayment processes a payment through the appropriate gateway
func (s *Service) ProcessPayment(requestID uint64) (*PaymentResult, error) {
	var request domain.PaymentRequest
	if err := s.db.Preload("Gateway").First(&request, requestID).Error; err != nil {
		return nil, err
	}

	processor, ok := s.processors[request.Gateway.Slug]
	if !ok {
		return nil, fmt.Errorf("processor not registered: %s", request.Gateway.Slug)
	}

	result, err := processor.ProcessPayment(&PaymentRequest{
		CustomerID:  request.CustomerID,
		InvoiceID:   request.InvoiceID,
		Amount:      request.Amount,
		Currency:    request.Currency,
		IPAddress:   request.IPAddress,
	})

	now := time.Now()
	if err != nil {
		s.db.Model(&request).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": err.Error(),
			"processed_at":  &now,
		})
		return nil, err
	}

	// Update request status
	updates := map[string]interface{}{
		"status":       result.Status,
		"gateway_ref":  result.GatewayRef,
		"processed_at": &now,
	}

	if result.Success {
		// Create transaction
		transaction := &domain.Transaction{
			CustomerID:     request.CustomerID,
			InvoiceID:      &request.InvoiceID,
			Type:           domain.TransactionTypePayment,
			Status:         domain.TransactionStatusCompleted,
			Currency:       request.Currency,
			Amount:         result.Amount,
			Fee:            result.Fee,
			Gateway:        request.Gateway.Slug,
			GatewayTransID: result.TransactionID,
			IPAddress:      request.IPAddress,
		}
		if err := s.db.Create(transaction).Error; err != nil {
			return nil, err
		}
		updates["transaction_id"] = transaction.ID
	}

	s.db.Model(&request).Updates(updates)

	return result, nil
}

// PayWithCredit pays an invoice using customer credit balance
func (s *Service) PayWithCredit(customerID, invoiceID uint64, amount decimal.Decimal) (*domain.Transaction, error) {
	var customer domain.User
	if err := s.db.First(&customer, customerID).Error; err != nil {
		return nil, err
	}

	if customer.Credit.LessThan(amount) {
		return nil, ErrInsufficientBalance
	}

	var invoice domain.Invoice
	if err := s.db.First(&invoice, invoiceID).Error; err != nil {
		return nil, err
	}

	var transaction *domain.Transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Deduct credit
		if err := tx.Model(&customer).Update("credit", customer.Credit.Sub(amount)).Error; err != nil {
			return err
		}

		// Create credit adjustment
		adjustment := &domain.CreditAdjustment{
			CustomerID:    customerID,
			Type:          "subtract",
			Amount:        amount,
			Currency:      invoice.Currency,
			Reason:        fmt.Sprintf("Payment for invoice %s", invoice.InvoiceNumber),
			RelatedType:   "invoice",
			RelatedID:     &invoiceID,
			BalanceBefore: customer.Credit,
			BalanceAfter:  customer.Credit.Sub(amount),
		}
		if err := tx.Create(adjustment).Error; err != nil {
			return err
		}

		// Create transaction
		transaction = &domain.Transaction{
			CustomerID:  customerID,
			InvoiceID:   &invoiceID,
			Type:        domain.TransactionTypeCredit,
			Status:      domain.TransactionStatusCompleted,
			Currency:    invoice.Currency,
			Amount:      amount,
			Gateway:     "credit_balance",
			Description: fmt.Sprintf("Credit balance payment for invoice %s", invoice.InvoiceNumber),
		}
		if err := tx.Create(transaction).Error; err != nil {
			return err
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
		return tx.Model(&invoice).Updates(updates).Error
	})

	return transaction, err
}

// AddCredit adds credit to a customer account
func (s *Service) AddCredit(customerID uint64, amount decimal.Decimal, currency, reason string, staffID *uint64) (*domain.CreditAdjustment, error) {
	var customer domain.User
	if err := s.db.First(&customer, customerID).Error; err != nil {
		return nil, err
	}

	adjustment := &domain.CreditAdjustment{
		CustomerID:    customerID,
		Type:          "add",
		Amount:        amount,
		Currency:      currency,
		Reason:        reason,
		StaffID:       staffID,
		BalanceBefore: customer.Credit,
		BalanceAfter:  customer.Credit.Add(amount),
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&customer).Update("credit", customer.Credit.Add(amount)).Error; err != nil {
			return err
		}
		return tx.Create(adjustment).Error
	})

	return adjustment, err
}

// ProcessRefund processes a refund for a transaction
func (s *Service) ProcessRefund(transactionID uint64, amount decimal.Decimal, reason string, staffID uint64) (*domain.Transaction, error) {
	var original domain.Transaction
	if err := s.db.First(&original, transactionID).Error; err != nil {
		return nil, err
	}

	if !original.IsRefundable() {
		return nil, errors.New("transaction is not refundable")
	}

	if amount.GreaterThan(original.RemainingRefundable()) {
		return nil, ErrInvalidAmount
	}

	var refund *domain.Transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		refund = &domain.Transaction{
			CustomerID:     original.CustomerID,
			InvoiceID:      original.InvoiceID,
			Type:           domain.TransactionTypeRefund,
			Status:         domain.TransactionStatusCompleted,
			Currency:       original.Currency,
			Amount:         amount.Neg(),
			Gateway:        original.Gateway,
			RefundTransID:  &original.ID,
			Description:    fmt.Sprintf("Refund: %s", reason),
		}

		// Update original transaction's refunded amount
		if err := tx.Model(&original).Update("refunded_amount", original.RefundedAmount.Add(amount)).Error; err != nil {
			return err
		}
		return tx.Create(refund).Error
	})

	return refund, err
}

// CreateSubscription creates a recurring payment subscription
func (s *Service) CreateSubscription(request *SubscriptionRequest, gatewayID uint64) (*domain.PaymentSubscription, error) {
	gateway, err := s.GetGateway(gatewayID)
	if err != nil {
		return nil, err
	}
	if !gateway.SupportsRecurring {
		return nil, errors.New("gateway does not support recurring payments")
	}

	processor, ok := s.processors[gateway.Slug]
	if !ok {
		return nil, fmt.Errorf("processor not registered: %s", gateway.Slug)
	}

	result, err := processor.CreateSubscription(request)
	if err != nil {
		return nil, err
	}

	subscription := &domain.PaymentSubscription{
		CustomerID:         request.CustomerID,
		ServiceID:          &request.ServiceID,
		GatewayID:          gatewayID,
		PaymentMethodID:    request.PaymentMethodID,
		GatewaySubID:       result.SubscriptionID,
		Amount:             request.Amount,
		Currency:           request.Currency,
		Interval:           request.Interval,
		IntervalCount:      request.IntervalCount,
		Status:             domain.SubscriptionActive,
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   result.CurrentPeriodEnd,
	}

	if err := s.db.Create(subscription).Error; err != nil {
		return nil, err
	}

	return subscription, nil
}

// CancelSubscription cancels a payment subscription
func (s *Service) CancelSubscription(subscriptionID uint64, immediately bool) error {
	var subscription domain.PaymentSubscription
	if err := s.db.Preload("Gateway").First(&subscription, subscriptionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSubscriptionNotFound
		}
		return err
	}

	processor, ok := s.processors[subscription.Gateway.Slug]
	if !ok {
		return fmt.Errorf("processor not registered: %s", subscription.Gateway.Slug)
	}

	if err := processor.CancelSubscription(subscription.GatewaySubID); err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"cancelled_at": &now,
	}

	if immediately {
		updates["status"] = domain.SubscriptionCancelled
		updates["ended_at"] = &now
	} else {
		updates["cancel_at_period_end"] = true
	}

	return s.db.Model(&subscription).Updates(updates).Error
}

// ProcessWebhook processes a payment gateway webhook
func (s *Service) ProcessWebhook(gatewaySlug string, payload []byte, signature string) error {
	var gateway domain.PaymentGatewayModule
	if err := s.db.Where("slug = ?", gatewaySlug).First(&gateway).Error; err != nil {
		return ErrGatewayNotFound
	}

	processor, ok := s.processors[gatewaySlug]
	if !ok {
		return fmt.Errorf("processor not registered: %s", gatewaySlug)
	}

	if !processor.ValidateWebhook(payload, signature) {
		return errors.New("invalid webhook signature")
	}

	// Log the webhook
	log := &domain.GatewayWebhookLog{
		GatewayID: gateway.ID,
		Payload:   string(payload),
		Status:    "received",
	}
	s.db.Create(log)

	return nil
}

// VerifyWebhookSignature verifies a webhook signature using HMAC-SHA256
func VerifyWebhookSignature(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// SavePaymentMethod saves a payment method for a customer
func (s *Service) SavePaymentMethod(customerID uint64, methodType domain.PaymentMethodType, gateway, gatewayMethodID, label, last4, brand string, expiryMonth, expiryYear int, isDefault bool) (*domain.PaymentMethod, error) {
	method := &domain.PaymentMethod{
		CustomerID:      customerID,
		Type:            methodType,
		Gateway:         gateway,
		GatewayMethodID: gatewayMethodID,
		Label:           label,
		Last4:           last4,
		Brand:           brand,
		ExpiryMonth:     expiryMonth,
		ExpiryYear:      expiryYear,
		IsDefault:       isDefault,
		Active:          true,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if isDefault {
			// Clear default from other methods
			tx.Model(&domain.PaymentMethod{}).Where("customer_id = ?", customerID).
				Update("is_default", false)
		}
		return tx.Create(method).Error
	})

	return method, err
}

// SetDefaultPaymentMethod sets a payment method as default
func (s *Service) SetDefaultPaymentMethod(customerID, methodID uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Clear default from all methods
		if err := tx.Model(&domain.PaymentMethod{}).Where("customer_id = ?", customerID).
			Update("is_default", false).Error; err != nil {
			return err
		}
		// Set new default
		return tx.Model(&domain.PaymentMethod{}).Where("id = ? AND customer_id = ?", methodID, customerID).
			Update("is_default", true).Error
	})
}

// DeletePaymentMethod deletes a payment method
func (s *Service) DeletePaymentMethod(customerID, methodID uint64) error {
	return s.db.Where("id = ? AND customer_id = ?", methodID, customerID).
		Delete(&domain.PaymentMethod{}).Error
}

// GetAutoPaymentConfig gets auto-payment configuration for a customer
func (s *Service) GetAutoPaymentConfig(customerID uint64) (*domain.AutoPayment, error) {
	var config domain.AutoPayment
	if err := s.db.Where("customer_id = ?", customerID).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// SetupAutoPayment configures automatic payment for a customer
func (s *Service) SetupAutoPayment(customerID, paymentMethodID uint64, maxAmount decimal.Decimal, daysBefore int) (*domain.AutoPayment, error) {
	config := &domain.AutoPayment{
		CustomerID:      customerID,
		PaymentMethodID: paymentMethodID,
		Active:          true,
		MaxAmount:       maxAmount,
		DaysBefore:      daysBefore,
	}

	err := s.db.Create(config).Error
	if err != nil {
		// Try update if exists
		err = s.db.Model(&domain.AutoPayment{}).Where("customer_id = ?", customerID).
			Updates(map[string]interface{}{
				"payment_method_id": paymentMethodID,
				"active":            true,
				"max_amount":        maxAmount,
				"days_before":       daysBefore,
			}).Error
	}

	return config, err
}
