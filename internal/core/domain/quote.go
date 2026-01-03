package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// QuoteStatus represents the status of a quote
type QuoteStatus string

const (
	QuoteStatusDraft     QuoteStatus = "draft"
	QuoteStatusSent      QuoteStatus = "sent"
	QuoteStatusViewed    QuoteStatus = "viewed"
	QuoteStatusAccepted  QuoteStatus = "accepted"
	QuoteStatusDeclined  QuoteStatus = "declined"
	QuoteStatusExpired   QuoteStatus = "expired"
	QuoteStatusConverted QuoteStatus = "converted"
)

// Quote represents a price quote for a customer
type Quote struct {
	ID            uint64          `gorm:"primaryKey"`
	QuoteNumber   string          `gorm:"size:50;uniqueIndex;not null"`
	CustomerID    uint64          `gorm:"not null;index"`
	StaffID       uint64          `gorm:"not null;index"`
	Status        QuoteStatus     `gorm:"size:32;not null;default:'draft'"`
	Subject       string          `gorm:"size:255;not null"`
	Currency      string          `gorm:"size:3;not null;default:'USD'"`
	Subtotal      decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Discount      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	TaxRate       decimal.Decimal `gorm:"type:numeric(10,4);not null;default:0"`
	TaxAmount     decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Total         decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	ProposalText  string          `gorm:"type:text"`
	Notes         string          `gorm:"type:text"` // Internal notes
	CustomerNotes string          `gorm:"type:text"` // Notes visible to customer
	ValidUntil    time.Time       `gorm:"not null"`
	SentAt        *time.Time
	ViewedAt      *time.Time
	AcceptedAt    *time.Time
	DeclinedAt    *time.Time
	OrderID       *uint64   `gorm:"index"` // Created order when accepted
	InvoiceID     *uint64   `gorm:"index"` // Created invoice when accepted
	Metadata      JSONMap   `gorm:"type:jsonb"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Customer  User        `gorm:"foreignKey:CustomerID"`
	Staff     User        `gorm:"foreignKey:StaffID"`
	Order     *Order      `gorm:"foreignKey:OrderID"`
	Invoice   *Invoice    `gorm:"foreignKey:InvoiceID"`
	LineItems []QuoteItem `gorm:"foreignKey:QuoteID"`
}

// IsExpired checks if the quote has expired
func (q *Quote) IsExpired() bool {
	return time.Now().After(q.ValidUntil)
}

// CanAccept checks if the quote can be accepted
func (q *Quote) CanAccept() bool {
	return q.Status == QuoteStatusSent || q.Status == QuoteStatusViewed
}

// QuoteItem represents a line item on a quote
type QuoteItem struct {
	ID          uint64          `gorm:"primaryKey"`
	QuoteID     uint64          `gorm:"not null;index"`
	ProductID   *uint64         `gorm:"index"`
	Type        string          `gorm:"size:50;not null"`
	Description string          `gorm:"size:500;not null"`
	Quantity    decimal.Decimal `gorm:"type:numeric(20,8);not null;default:1"`
	UnitPrice   decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Discount    decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Total       decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Taxable     bool            `gorm:"not null;default:true"`
	BillingCycle string         `gorm:"size:32"`
	SetupFee    decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	SortOrder   int             `gorm:"not null;default:0"`
	CreatedAt   time.Time       `gorm:"not null"`
	UpdatedAt   time.Time       `gorm:"not null"`

	Quote   Quote    `gorm:"foreignKey:QuoteID"`
	Product *Product `gorm:"foreignKey:ProductID"`
}

// CalculateTotal calculates and updates the line item total
func (i *QuoteItem) CalculateTotal() {
	subtotal := i.UnitPrice.Mul(i.Quantity).Add(i.SetupFee)
	i.Total = subtotal.Sub(i.Discount)
}

// OrderApprovalStatus represents the status of an order approval
type OrderApprovalStatus string

const (
	OrderApprovalPending  OrderApprovalStatus = "pending"
	OrderApprovalApproved OrderApprovalStatus = "approved"
	OrderApprovalRejected OrderApprovalStatus = "rejected"
)

// OrderApproval represents an order that requires manual approval
type OrderApproval struct {
	ID          uint64              `gorm:"primaryKey"`
	OrderID     uint64              `gorm:"not null;uniqueIndex"`
	Status      OrderApprovalStatus `gorm:"size:32;not null;default:'pending'"`
	Reason      string              `gorm:"type:text"` // Why approval was needed
	ReviewedBy  *uint64             `gorm:"index"`
	ReviewedAt  *time.Time
	Notes       string    `gorm:"type:text"`
	AutoApprove *time.Time // Auto-approve after this time
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	Order    Order `gorm:"foreignKey:OrderID"`
	Reviewer *User `gorm:"foreignKey:ReviewedBy"`
}

// OrderFraudCheck represents fraud check results for an order
type OrderFraudCheck struct {
	ID            uint64          `gorm:"primaryKey"`
	OrderID       uint64          `gorm:"not null;uniqueIndex"`
	Score         decimal.Decimal `gorm:"type:numeric(10,4);not null;default:0"` // 0-100
	Result        string          `gorm:"size:32;not null"` // pass, review, fail
	RulesFailed   JSONMap         `gorm:"type:jsonb"`
	IPCountry     string          `gorm:"size:2"`
	BillingCountry string         `gorm:"size:2"`
	CountryMatch  bool            `gorm:"not null"`
	ProxyDetected bool            `gorm:"not null;default:false"`
	HighRiskEmail bool            `gorm:"not null;default:false"`
	HighRiskIP    bool            `gorm:"not null;default:false"`
	VPNDetected   bool            `gorm:"not null;default:false"`
	Metadata      JSONMap         `gorm:"type:jsonb"`
	ReviewedBy    *uint64         `gorm:"index"`
	ReviewedAt    *time.Time
	ReviewNotes   string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Order    Order `gorm:"foreignKey:OrderID"`
	Reviewer *User `gorm:"foreignKey:ReviewedBy"`
}

// IsFailed checks if the fraud check failed
func (f *OrderFraudCheck) IsFailed() bool {
	return f.Result == "fail"
}

// NeedsReview checks if the order needs manual review
func (f *OrderFraudCheck) NeedsReview() bool {
	return f.Result == "review"
}

// OrderNote represents an internal note on an order
type OrderNote struct {
	ID        uint64    `gorm:"primaryKey"`
	OrderID   uint64    `gorm:"not null;index"`
	StaffID   uint64    `gorm:"not null;index"`
	Note      string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Order Order `gorm:"foreignKey:OrderID"`
	Staff User  `gorm:"foreignKey:StaffID"`
}

// OrderStatusLog represents order status change history
type OrderStatusLog struct {
	ID        uint64      `gorm:"primaryKey"`
	OrderID   uint64      `gorm:"not null;index"`
	OldStatus OrderStatus `gorm:"size:64"`
	NewStatus OrderStatus `gorm:"size:64;not null"`
	ChangedBy *uint64     `gorm:"index"`
	Reason    string      `gorm:"type:text"`
	Automatic bool        `gorm:"not null;default:false"`
	CreatedAt time.Time   `gorm:"not null"`

	Order     Order `gorm:"foreignKey:OrderID"`
	ChangedByUser *User `gorm:"foreignKey:ChangedBy"`
}

// RecurringInvoice represents a recurring invoice template
type RecurringInvoice struct {
	ID            uint64          `gorm:"primaryKey"`
	CustomerID    uint64          `gorm:"not null;index"`
	Currency      string          `gorm:"size:3;not null;default:'USD'"`
	Subtotal      decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	TaxRate       decimal.Decimal `gorm:"type:numeric(10,4);not null;default:0"`
	TaxAmount     decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Total         decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	BillingCycle  string          `gorm:"size:32;not null"` // monthly, quarterly, etc
	NextRunDate   time.Time       `gorm:"not null;index"`
	LastRunDate   *time.Time
	StartDate     time.Time `gorm:"not null"`
	EndDate       *time.Time
	PaymentMethod *uint64   `gorm:"index"` // Auto-charge payment method
	AutoCharge    bool      `gorm:"not null;default:false"`
	EmailDaysBefore int     `gorm:"not null;default:3"` // Days before to send reminder
	Status        string    `gorm:"size:32;not null;default:'active'"` // active, paused, completed, cancelled
	Notes         string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Customer      User                   `gorm:"foreignKey:CustomerID"`
	PaymentMethodRel *PaymentMethod      `gorm:"foreignKey:PaymentMethod"`
	LineItems     []RecurringInvoiceItem `gorm:"foreignKey:RecurringInvoiceID"`
}

// RecurringInvoiceItem represents a line item on a recurring invoice
type RecurringInvoiceItem struct {
	ID                  uint64          `gorm:"primaryKey"`
	RecurringInvoiceID  uint64          `gorm:"not null;index"`
	ServiceID           *uint64         `gorm:"index"`
	Type                string          `gorm:"size:50;not null"`
	Description         string          `gorm:"size:500;not null"`
	Quantity            decimal.Decimal `gorm:"type:numeric(20,8);not null;default:1"`
	UnitPrice           decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Discount            decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Total               decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Taxable             bool            `gorm:"not null;default:true"`
	CreatedAt           time.Time       `gorm:"not null"`
	UpdatedAt           time.Time       `gorm:"not null"`

	RecurringInvoice RecurringInvoice `gorm:"foreignKey:RecurringInvoiceID"`
	Service          *Service         `gorm:"foreignKey:ServiceID"`
}

// InvoiceMerge represents a merged invoice
type InvoiceMerge struct {
	ID              uint64    `gorm:"primaryKey"`
	MergedInvoiceID uint64    `gorm:"not null;index"`
	SourceInvoiceID uint64    `gorm:"not null;index"`
	MergedBy        uint64    `gorm:"not null"`
	CreatedAt       time.Time `gorm:"not null"`

	MergedInvoice Invoice `gorm:"foreignKey:MergedInvoiceID"`
	SourceInvoice Invoice `gorm:"foreignKey:SourceInvoiceID"`
	Staff         User    `gorm:"foreignKey:MergedBy"`
}

// DraftInvoice represents a draft invoice that hasn't been sent
type DraftInvoice struct {
	ID            uint64          `gorm:"primaryKey"`
	CustomerID    uint64          `gorm:"not null;index"`
	Currency      string          `gorm:"size:3;not null;default:'USD'"`
	Subtotal      decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	TaxAmount     decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Total         decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	ProposedDueDate time.Time     `gorm:"not null"`
	Notes         string          `gorm:"type:text"`
	InternalNotes string          `gorm:"type:text"`
	AutoSendDate  *time.Time
	CreatedBy     uint64    `gorm:"not null"`
	InvoiceID     *uint64   `gorm:"index"` // Set when published
	PublishedAt   *time.Time
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Customer  User              `gorm:"foreignKey:CustomerID"`
	Creator   User              `gorm:"foreignKey:CreatedBy"`
	Invoice   *Invoice          `gorm:"foreignKey:InvoiceID"`
	LineItems []DraftInvoiceItem `gorm:"foreignKey:DraftInvoiceID"`
}

// DraftInvoiceItem represents a line item on a draft invoice
type DraftInvoiceItem struct {
	ID             uint64          `gorm:"primaryKey"`
	DraftInvoiceID uint64          `gorm:"not null;index"`
	Type           string          `gorm:"size:50;not null"`
	Description    string          `gorm:"size:500;not null"`
	Quantity       decimal.Decimal `gorm:"type:numeric(20,8);not null;default:1"`
	UnitPrice      decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Discount       decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Total          decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Taxable        bool            `gorm:"not null;default:true"`
	CreatedAt      time.Time       `gorm:"not null"`
	UpdatedAt      time.Time       `gorm:"not null"`

	DraftInvoice DraftInvoice `gorm:"foreignKey:DraftInvoiceID"`
}

// CreditAdjustment represents a credit balance adjustment
type CreditAdjustment struct {
	ID          uint64          `gorm:"primaryKey"`
	CustomerID  uint64          `gorm:"not null;index"`
	Type        string          `gorm:"size:32;not null"` // add, subtract
	Amount      decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Currency    string          `gorm:"size:3;not null"`
	Reason      string          `gorm:"size:500;not null"`
	RelatedType string          `gorm:"size:50"` // invoice, transaction, refund
	RelatedID   *uint64         `gorm:"index"`
	StaffID     *uint64         `gorm:"index"`
	BalanceBefore decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	BalanceAfter  decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	CreatedAt   time.Time `gorm:"not null"`

	Customer User  `gorm:"foreignKey:CustomerID"`
	Staff    *User `gorm:"foreignKey:StaffID"`
}

// Chargeback represents a chargeback/dispute record
type Chargeback struct {
	ID            uint64          `gorm:"primaryKey"`
	TransactionID uint64          `gorm:"not null;index"`
	InvoiceID     uint64          `gorm:"not null;index"`
	CustomerID    uint64          `gorm:"not null;index"`
	Amount        decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Currency      string          `gorm:"size:3;not null"`
	Status        string          `gorm:"size:32;not null;default:'open'"` // open, won, lost, resolved
	Reason        string          `gorm:"size:100"`
	ReasonCode    string          `gorm:"size:20"`
	Gateway       string          `gorm:"size:50;not null"`
	GatewayID     string          `gorm:"size:255"`
	Evidence      JSONMap         `gorm:"type:jsonb"` // Evidence submitted
	DueDate       *time.Time                          // Evidence due date
	ResolvedAt    *time.Time
	Resolution    string    `gorm:"type:text"`
	Notes         string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Transaction Transaction `gorm:"foreignKey:TransactionID"`
	Invoice     Invoice     `gorm:"foreignKey:InvoiceID"`
	Customer    User        `gorm:"foreignKey:CustomerID"`
}

// IsOpen checks if the chargeback is still open
func (c *Chargeback) IsOpen() bool {
	return c.Status == "open"
}

// LateFee represents a late fee applied to an invoice
type LateFee struct {
	ID        uint64          `gorm:"primaryKey"`
	InvoiceID uint64          `gorm:"not null;index"`
	Type      string          `gorm:"size:32;not null"` // percentage, fixed
	Amount    decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	DaysLate  int             `gorm:"not null"`
	CreatedAt time.Time       `gorm:"not null"`

	Invoice Invoice `gorm:"foreignKey:InvoiceID"`
}
