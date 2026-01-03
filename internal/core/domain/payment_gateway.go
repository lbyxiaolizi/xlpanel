package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// PaymentGatewayType represents the type of payment gateway
type PaymentGatewayType string

const (
	GatewayTypeCard     PaymentGatewayType = "card"
	GatewayTypeBank     PaymentGatewayType = "bank"
	GatewayTypeCrypto   PaymentGatewayType = "crypto"
	GatewayTypeWallet   PaymentGatewayType = "wallet"
	GatewayTypeManual   PaymentGatewayType = "manual"
)

// PaymentGatewayConfig represents gateway configuration
type PaymentGatewayConfig struct {
	APIKey           string `json:"api_key,omitempty"`
	APISecret        string `json:"api_secret,omitempty"`
	MerchantID       string `json:"merchant_id,omitempty"`
	PublicKey        string `json:"public_key,omitempty"`
	PrivateKey       string `json:"private_key,omitempty"`
	WebhookSecret    string `json:"webhook_secret,omitempty"`
	SandboxMode      bool   `json:"sandbox_mode,omitempty"`
	SupportedCurrencies []string `json:"supported_currencies,omitempty"`
	Extra            map[string]string `json:"extra,omitempty"`
}

// Value implements driver.Valuer for PaymentGatewayConfig
func (c PaymentGatewayConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements sql.Scanner for PaymentGatewayConfig
func (c *PaymentGatewayConfig) Scan(value interface{}) error {
	if value == nil {
		*c = PaymentGatewayConfig{}
		return nil
	}
	data, err := normalizeJSONBytes(value)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		*c = PaymentGatewayConfig{}
		return nil
	}
	return json.Unmarshal(data, c)
}

// PaymentGatewayModule represents a payment gateway module
type PaymentGatewayModule struct {
	ID                uint64               `gorm:"primaryKey"`
	Name              string               `gorm:"size:100;not null"`
	Slug              string               `gorm:"size:100;uniqueIndex;not null"`
	DisplayName       string               `gorm:"size:100;not null"`
	Description       string               `gorm:"type:text"`
	Type              PaymentGatewayType   `gorm:"size:32;not null"`
	Config            PaymentGatewayConfig `gorm:"type:jsonb"`
	LogoURL           string               `gorm:"size:500"`
	SupportsRefund    bool                 `gorm:"not null;default:false"`
	SupportsRecurring bool                 `gorm:"not null;default:false"`
	SupportsTokenize  bool                 `gorm:"not null;default:false"`
	SupportedCards    JSONMap              `gorm:"type:jsonb"` // Visa, MC, etc.
	MinAmount         decimal.Decimal      `gorm:"type:numeric(20,8);not null;default:0"`
	MaxAmount         decimal.Decimal      `gorm:"type:numeric(20,8);not null;default:0"` // 0 = unlimited
	FeePercent        decimal.Decimal      `gorm:"type:numeric(10,4);not null;default:0"`
	FeeFixed          decimal.Decimal      `gorm:"type:numeric(20,8);not null;default:0"`
	TestMode          bool                 `gorm:"not null;default:false"`
	RequiresSSL       bool                 `gorm:"not null;default:true"`
	Active            bool                 `gorm:"not null;default:true"`
	Visible           bool                 `gorm:"not null;default:true"`
	SortOrder         int                  `gorm:"not null;default:0"`
	CreatedAt         time.Time            `gorm:"not null"`
	UpdatedAt         time.Time            `gorm:"not null"`
}

// CalculateFee calculates the fee for a given amount
func (g *PaymentGatewayModule) CalculateFee(amount decimal.Decimal) decimal.Decimal {
	percentFee := amount.Mul(g.FeePercent).Div(decimal.NewFromInt(100))
	return percentFee.Add(g.FeeFixed)
}

// PaymentRequest represents a payment request/attempt
type PaymentRequest struct {
	ID              uint64          `gorm:"primaryKey"`
	CustomerID      uint64          `gorm:"not null;index"`
	InvoiceID       uint64          `gorm:"not null;index"`
	GatewayID       uint64          `gorm:"not null;index"`
	PaymentMethodID *uint64         `gorm:"index"`
	Amount          decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Currency        string          `gorm:"size:3;not null"`
	Status          string          `gorm:"size:32;not null;default:'pending'"` // pending, processing, completed, failed, cancelled
	GatewayRef      string          `gorm:"size:255;index"`
	PaymentURL      string          `gorm:"size:500"` // For hosted payment pages
	RedirectURL     string          `gorm:"size:500"`
	CallbackURL     string          `gorm:"size:500"`
	Metadata        JSONMap         `gorm:"type:jsonb"`
	IPAddress       string          `gorm:"size:45"`
	UserAgent       string          `gorm:"size:512"`
	ErrorCode       string          `gorm:"size:50"`
	ErrorMessage    string          `gorm:"type:text"`
	ProcessedAt     *time.Time
	ExpiresAt       *time.Time
	TransactionID   *uint64   `gorm:"index"` // Created transaction
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`

	Customer      User                  `gorm:"foreignKey:CustomerID"`
	Invoice       Invoice               `gorm:"foreignKey:InvoiceID"`
	Gateway       PaymentGatewayModule  `gorm:"foreignKey:GatewayID"`
	PaymentMethod *PaymentMethod        `gorm:"foreignKey:PaymentMethodID"`
	Transaction   *Transaction          `gorm:"foreignKey:TransactionID"`
}

// IsPending checks if the payment request is pending
func (p *PaymentRequest) IsPending() bool {
	return p.Status == "pending" || p.Status == "processing"
}

// GatewayWebhookLog represents a webhook received from a payment gateway
type GatewayWebhookLog struct {
	ID            uint64    `gorm:"primaryKey"`
	GatewayID     uint64    `gorm:"not null;index"`
	EventType     string    `gorm:"size:100;not null;index"`
	Payload       string    `gorm:"type:text;not null"`
	Headers       JSONMap   `gorm:"type:jsonb"`
	IPAddress     string    `gorm:"size:45"`
	Status        string    `gorm:"size:32;not null"` // received, processed, failed, ignored
	ErrorMessage  string    `gorm:"type:text"`
	ProcessedAt   *time.Time
	RelatedType   string    `gorm:"size:50;index"` // payment_request, transaction, subscription
	RelatedID     *uint64   `gorm:"index"`
	CreatedAt     time.Time `gorm:"not null;index"`

	Gateway PaymentGatewayModule `gorm:"foreignKey:GatewayID"`
}

// SubscriptionStatus represents a subscription status
type SubscriptionStatus string

const (
	SubscriptionActive    SubscriptionStatus = "active"
	SubscriptionPaused    SubscriptionStatus = "paused"
	SubscriptionCancelled SubscriptionStatus = "cancelled"
	SubscriptionExpired   SubscriptionStatus = "expired"
	SubscriptionPastDue   SubscriptionStatus = "past_due"
)

// PaymentSubscription represents a recurring payment subscription
type PaymentSubscription struct {
	ID              uint64             `gorm:"primaryKey"`
	CustomerID      uint64             `gorm:"not null;index"`
	ServiceID       *uint64            `gorm:"index"`
	GatewayID       uint64             `gorm:"not null;index"`
	PaymentMethodID uint64             `gorm:"not null;index"`
	GatewaySubID    string             `gorm:"size:255;uniqueIndex"` // Subscription ID at gateway
	Amount          decimal.Decimal    `gorm:"type:numeric(20,8);not null"`
	Currency        string             `gorm:"size:3;not null"`
	Interval        string             `gorm:"size:32;not null"` // monthly, quarterly, yearly
	IntervalCount   int                `gorm:"not null;default:1"`
	Status          SubscriptionStatus `gorm:"size:32;not null;default:'active'"`
	CurrentPeriodStart time.Time       `gorm:"not null"`
	CurrentPeriodEnd   time.Time       `gorm:"not null"`
	TrialEnd        *time.Time
	CancelAtPeriodEnd bool             `gorm:"not null;default:false"`
	CancelledAt     *time.Time
	EndedAt         *time.Time
	LastPaymentAt   *time.Time
	NextPaymentAt   *time.Time
	FailedPayments  int                `gorm:"not null;default:0"`
	Metadata        JSONMap            `gorm:"type:jsonb"`
	CreatedAt       time.Time          `gorm:"not null"`
	UpdatedAt       time.Time          `gorm:"not null"`

	Customer      User                 `gorm:"foreignKey:CustomerID"`
	Service       *Service             `gorm:"foreignKey:ServiceID"`
	Gateway       PaymentGatewayModule `gorm:"foreignKey:GatewayID"`
	PaymentMethod PaymentMethod        `gorm:"foreignKey:PaymentMethodID"`
}

// IsActive checks if the subscription is active
func (s *PaymentSubscription) IsActive() bool {
	return s.Status == SubscriptionActive
}

// AutoPayment represents an automatic payment configuration
type AutoPayment struct {
	ID              uint64    `gorm:"primaryKey"`
	CustomerID      uint64    `gorm:"not null;uniqueIndex"`
	PaymentMethodID uint64    `gorm:"not null;index"`
	Active          bool      `gorm:"not null;default:true"`
	MaxAmount       decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"` // 0 = unlimited
	DaysBefore      int       `gorm:"not null;default:3"` // Days before due date to charge
	LastAttempt     *time.Time
	LastSuccess     *time.Time
	ConsecutiveFails int      `gorm:"not null;default:0"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`

	Customer      User          `gorm:"foreignKey:CustomerID"`
	PaymentMethod PaymentMethod `gorm:"foreignKey:PaymentMethodID"`
}

// BankAccount represents a bank account for bank transfers
type BankAccount struct {
	ID           uint64    `gorm:"primaryKey"`
	Name         string    `gorm:"size:100;not null"`
	BankName     string    `gorm:"size:255;not null"`
	AccountName  string    `gorm:"size:255;not null"`
	AccountNumber string   `gorm:"size:50"`
	RoutingNumber string   `gorm:"size:50"`
	IBAN         string    `gorm:"size:50"`
	BIC          string    `gorm:"size:20"`
	Currency     string    `gorm:"size:3;not null"`
	Country      string    `gorm:"size:2;not null"`
	Instructions string    `gorm:"type:text"`
	Active       bool      `gorm:"not null;default:true"`
	SortOrder    int       `gorm:"not null;default:0"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// ManualPayment represents a manual/offline payment record
type ManualPayment struct {
	ID            uint64          `gorm:"primaryKey"`
	CustomerID    uint64          `gorm:"not null;index"`
	InvoiceID     *uint64         `gorm:"index"`
	BankAccountID *uint64         `gorm:"index"`
	Amount        decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Currency      string          `gorm:"size:3;not null"`
	Method        string          `gorm:"size:50;not null"` // bank_transfer, check, cash
	Reference     string          `gorm:"size:255"`
	PaymentDate   time.Time       `gorm:"not null"`
	Status        string          `gorm:"size:32;not null;default:'pending'"` // pending, verified, rejected
	Notes         string          `gorm:"type:text"`
	Attachments   JSONMap         `gorm:"type:jsonb"` // Receipt images
	VerifiedBy    *uint64         `gorm:"index"`
	VerifiedAt    *time.Time
	TransactionID *uint64   `gorm:"index"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Customer    User         `gorm:"foreignKey:CustomerID"`
	Invoice     *Invoice     `gorm:"foreignKey:InvoiceID"`
	BankAccount *BankAccount `gorm:"foreignKey:BankAccountID"`
	Verifier    *User        `gorm:"foreignKey:VerifiedBy"`
	Transaction *Transaction `gorm:"foreignKey:TransactionID"`
}

// PaymentReminder represents a payment reminder configuration
type PaymentReminder struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	DaysOffset  int       `gorm:"not null"` // Negative = before, Positive = after due date
	Type        string    `gorm:"size:32;not null"` // reminder, overdue, final
	TemplateID  uint64    `gorm:"not null;index"`
	Active      bool      `gorm:"not null;default:true"`
	SortOrder   int       `gorm:"not null;default:0"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	Template EmailTemplate `gorm:"foreignKey:TemplateID"`
}

// PaymentReminderLog represents a sent payment reminder
type PaymentReminderLog struct {
	ID         uint64    `gorm:"primaryKey"`
	InvoiceID  uint64    `gorm:"not null;index"`
	ReminderID uint64    `gorm:"not null;index"`
	SentAt     time.Time `gorm:"not null"`
	EmailLogID *uint64   `gorm:"index"`
	CreatedAt  time.Time `gorm:"not null"`

	Invoice  Invoice         `gorm:"foreignKey:InvoiceID"`
	Reminder PaymentReminder `gorm:"foreignKey:ReminderID"`
	EmailLog *EmailLog       `gorm:"foreignKey:EmailLogID"`
}

// CurrencyExchangeRate represents an exchange rate between currencies
type CurrencyExchangeRate struct {
	ID           uint64          `gorm:"primaryKey"`
	FromCurrency string          `gorm:"size:3;not null;uniqueIndex:idx_currency_pair"`
	ToCurrency   string          `gorm:"size:3;not null;uniqueIndex:idx_currency_pair"`
	Rate         decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Source       string          `gorm:"size:50"` // manual, api, etc.
	ValidFrom    time.Time       `gorm:"not null"`
	ValidUntil   *time.Time
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// Convert converts an amount from one currency to another
func (r *CurrencyExchangeRate) Convert(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(r.Rate)
}
