package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusUnpaid    InvoiceStatus = "unpaid"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
	InvoiceStatusRefunded  InvoiceStatus = "refunded"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
)

// Invoice represents a billing invoice
type Invoice struct {
	ID            uint64          `gorm:"primaryKey"`
	CustomerID    uint64          `gorm:"not null;index"`
	InvoiceNumber string          `gorm:"size:50;uniqueIndex;not null"`
	Status        InvoiceStatus   `gorm:"size:32;not null;default:'unpaid'"`
	Currency      string          `gorm:"size:3;not null;default:'USD'"`
	Subtotal      decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Discount      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	TaxRate       decimal.Decimal `gorm:"type:numeric(10,4);not null;default:0"`
	TaxAmount     decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Total         decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	AmountPaid    decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Balance       decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Notes         string          `gorm:"type:text"`
	PaymentMethod string          `gorm:"size:50"`
	DueDate       time.Time       `gorm:"not null"`
	PaidAt        *time.Time
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	// Relations
	Customer  User          `gorm:"foreignKey:CustomerID"`
	LineItems []InvoiceItem `gorm:"foreignKey:InvoiceID"`
}

// IsPaid checks if the invoice is fully paid
func (i *Invoice) IsPaid() bool {
	return i.Status == InvoiceStatusPaid
}

// IsOverdue checks if the invoice is overdue
func (i *Invoice) IsOverdue() bool {
	if i.Status == InvoiceStatusPaid || i.Status == InvoiceStatusCancelled {
		return false
	}
	return time.Now().After(i.DueDate)
}

// CalculateBalance calculates and updates the balance
func (i *Invoice) CalculateBalance() {
	i.Balance = i.Total.Sub(i.AmountPaid)
}

// InvoiceItem represents a line item on an invoice
type InvoiceItem struct {
	ID          uint64          `gorm:"primaryKey"`
	InvoiceID   uint64          `gorm:"not null;index"`
	ServiceID   *uint64         `gorm:"index"`
	Type        string          `gorm:"size:50;not null"`
	Description string          `gorm:"size:500;not null"`
	Quantity    decimal.Decimal `gorm:"type:numeric(20,8);not null;default:1"`
	UnitPrice   decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Discount    decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Total       decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Taxable     bool            `gorm:"not null;default:true"`
	PeriodStart *time.Time
	PeriodEnd   *time.Time
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	Invoice Invoice  `gorm:"foreignKey:InvoiceID"`
	Service *Service `gorm:"foreignKey:ServiceID"`
}

// CalculateTotal calculates and updates the line item total
func (i *InvoiceItem) CalculateTotal() {
	subtotal := i.UnitPrice.Mul(i.Quantity)
	i.Total = subtotal.Sub(i.Discount)
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypePayment    TransactionType = "payment"
	TransactionTypeRefund     TransactionType = "refund"
	TransactionTypeCredit     TransactionType = "credit"
	TransactionTypeDebit      TransactionType = "debit"
	TransactionTypeChargeback TransactionType = "chargeback"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusRefunded  TransactionStatus = "refunded"
	TransactionStatusDisputed  TransactionStatus = "disputed"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID                uint64            `gorm:"primaryKey"`
	CustomerID        uint64            `gorm:"not null;index"`
	InvoiceID         *uint64           `gorm:"index"`
	PaymentMethodID   *uint64           `gorm:"index"`
	Type              TransactionType   `gorm:"size:32;not null"`
	Status            TransactionStatus `gorm:"size:32;not null"`
	Currency          string            `gorm:"size:3;not null"`
	Amount            decimal.Decimal   `gorm:"type:numeric(20,8);not null"`
	Fee               decimal.Decimal   `gorm:"type:numeric(20,8);not null;default:0"`
	Gateway           string            `gorm:"size:50"`
	GatewayTransID    string            `gorm:"size:255"`
	Description       string            `gorm:"size:500"`
	RefundedAmount    decimal.Decimal   `gorm:"type:numeric(20,8);not null;default:0"`
	RefundTransID     *uint64           `gorm:"index"`
	IPAddress         string            `gorm:"size:45"`
	Metadata          JSONMap           `gorm:"type:jsonb"`
	CreatedAt         time.Time         `gorm:"not null;index"`
	UpdatedAt         time.Time         `gorm:"not null"`

	Customer      User           `gorm:"foreignKey:CustomerID"`
	Invoice       *Invoice       `gorm:"foreignKey:InvoiceID"`
	PaymentMethod *PaymentMethod `gorm:"foreignKey:PaymentMethodID"`
}

// IsRefundable checks if the transaction can be refunded
func (t *Transaction) IsRefundable() bool {
	if t.Type != TransactionTypePayment {
		return false
	}
	if t.Status != TransactionStatusCompleted {
		return false
	}
	return t.RefundedAmount.LessThan(t.Amount)
}

// RemainingRefundable returns the amount that can still be refunded
func (t *Transaction) RemainingRefundable() decimal.Decimal {
	return t.Amount.Sub(t.RefundedAmount)
}

// PaymentMethodType represents the type of payment method
type PaymentMethodType string

const (
	PaymentMethodCard       PaymentMethodType = "card"
	PaymentMethodPayPal     PaymentMethodType = "paypal"
	PaymentMethodBankWire   PaymentMethodType = "bank_wire"
	PaymentMethodCrypto     PaymentMethodType = "crypto"
	PaymentMethodAlipay     PaymentMethodType = "alipay"
	PaymentMethodWechatPay  PaymentMethodType = "wechat_pay"
)

// PaymentMethod represents a saved payment method
type PaymentMethod struct {
	ID              uint64            `gorm:"primaryKey"`
	CustomerID      uint64            `gorm:"not null;index"`
	Type            PaymentMethodType `gorm:"size:32;not null"`
	Gateway         string            `gorm:"size:50;not null"`
	GatewayMethodID string            `gorm:"size:255"`
	Label           string            `gorm:"size:100"`
	Last4           string            `gorm:"size:4"`
	Brand           string            `gorm:"size:32"`
	ExpiryMonth     int               `gorm:""`
	ExpiryYear      int               `gorm:""`
	IsDefault       bool              `gorm:"not null;default:false"`
	Active          bool              `gorm:"not null;default:true"`
	Metadata        JSONMap           `gorm:"type:jsonb"`
	CreatedAt       time.Time         `gorm:"not null"`
	UpdatedAt       time.Time         `gorm:"not null"`

	Customer User `gorm:"foreignKey:CustomerID"`
}

// DisplayName returns a user-friendly display name for the payment method
func (p *PaymentMethod) DisplayName() string {
	if p.Label != "" {
		return p.Label
	}
	switch p.Type {
	case PaymentMethodCard:
		if p.Brand != "" && p.Last4 != "" {
			return p.Brand + " ending in " + p.Last4
		}
		if p.Last4 != "" {
			return "Card ending in " + p.Last4
		}
		return "Credit Card"
	case PaymentMethodPayPal:
		return "PayPal"
	case PaymentMethodBankWire:
		return "Bank Transfer"
	default:
		return string(p.Type)
	}
}

// IsExpired checks if a card payment method has expired
func (p *PaymentMethod) IsExpired() bool {
	if p.Type != PaymentMethodCard {
		return false
	}
	if p.ExpiryMonth == 0 || p.ExpiryYear == 0 {
		return false
	}
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())
	
	if p.ExpiryYear < currentYear {
		return true
	}
	if p.ExpiryYear == currentYear && p.ExpiryMonth < currentMonth {
		return true
	}
	return false
}

// CouponStatus represents the status of a coupon
type CouponStatus string

const (
	CouponStatusActive   CouponStatus = "active"
	CouponStatusInactive CouponStatus = "inactive"
	CouponStatusExpired  CouponStatus = "expired"
)

// CouponType represents the type of discount
type CouponType string

const (
	CouponTypePercentage  CouponType = "percentage"
	CouponTypeFixed       CouponType = "fixed"
	CouponTypeOverride    CouponType = "override"
	CouponTypeFreeSetup   CouponType = "free_setup"
)

// Coupon represents a promotional coupon/discount code
type Coupon struct {
	ID              uint64          `gorm:"primaryKey"`
	Code            string          `gorm:"size:50;uniqueIndex;not null"`
	Description     string          `gorm:"size:500"`
	Type            CouponType      `gorm:"size:32;not null"`
	Amount          decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Currency        string          `gorm:"size:3"`
	Status          CouponStatus    `gorm:"size:32;not null;default:'active'"`
	MaxUses         int             `gorm:"not null;default:0"` // 0 = unlimited
	CurrentUses     int             `gorm:"not null;default:0"`
	MaxUsesPerUser  int             `gorm:"not null;default:0"` // 0 = unlimited
	BillingCycles   int             `gorm:"not null;default:0"` // 0 = forever, >0 = number of cycles
	MinOrderAmount  decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	AppliesToNew    bool            `gorm:"not null;default:true"`
	AppliesToRenew  bool            `gorm:"not null;default:false"`
	ProductIDs      JSONMap         `gorm:"type:jsonb"` // List of product IDs if restricted
	StartsAt        *time.Time
	ExpiresAt       *time.Time
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

// IsValid checks if the coupon is currently valid
func (c *Coupon) IsValid() bool {
	if c.Status != CouponStatusActive {
		return false
	}
	now := time.Now()
	if c.StartsAt != nil && now.Before(*c.StartsAt) {
		return false
	}
	if c.ExpiresAt != nil && now.After(*c.ExpiresAt) {
		return false
	}
	if c.MaxUses > 0 && c.CurrentUses >= c.MaxUses {
		return false
	}
	return true
}

// CouponUsage tracks usage of coupons per customer
type CouponUsage struct {
	ID         uint64    `gorm:"primaryKey"`
	CouponID   uint64    `gorm:"not null;index;uniqueIndex:idx_coupon_customer_invoice"`
	CustomerID uint64    `gorm:"not null;index;uniqueIndex:idx_coupon_customer_invoice"`
	InvoiceID  uint64    `gorm:"not null;index;uniqueIndex:idx_coupon_customer_invoice"`
	ServiceID  *uint64   `gorm:"index"`
	CyclesUsed int       `gorm:"not null;default:1"`
	CreatedAt  time.Time `gorm:"not null"`

	Coupon   Coupon  `gorm:"foreignKey:CouponID"`
	Customer User    `gorm:"foreignKey:CustomerID"`
	Invoice  Invoice `gorm:"foreignKey:InvoiceID"`
}

// TaxRule represents a tax rule for a specific region
type TaxRule struct {
	ID          uint64          `gorm:"primaryKey"`
	Name        string          `gorm:"size:100;not null"`
	Country     string          `gorm:"size:2;not null;index"`
	State       string          `gorm:"size:100;index"`
	Rate        decimal.Decimal `gorm:"type:numeric(10,4);not null"`
	TaxType     string          `gorm:"size:32;not null;default:'vat'"` // vat, gst, sales_tax
	IsInclusive bool            `gorm:"not null;default:false"`
	Priority    int             `gorm:"not null;default:0"`
	Active      bool            `gorm:"not null;default:true"`
	CreatedAt   time.Time       `gorm:"not null"`
	UpdatedAt   time.Time       `gorm:"not null"`
}

// Credit represents a credit adjustment on a customer account
type Credit struct {
	ID          uint64          `gorm:"primaryKey"`
	CustomerID  uint64          `gorm:"not null;index"`
	Type        string          `gorm:"size:32;not null"` // add, deduct
	Amount      decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Currency    string          `gorm:"size:3;not null"`
	Description string          `gorm:"size:500"`
	RelatedID   *uint64         `gorm:"index"` // Can link to invoice, transaction, etc.
	RelatedType string          `gorm:"size:50"`
	AdminID     *uint64         `gorm:"index"`
	CreatedAt   time.Time       `gorm:"not null"`

	Customer User  `gorm:"foreignKey:CustomerID"`
	Admin    *User `gorm:"foreignKey:AdminID"`
}
