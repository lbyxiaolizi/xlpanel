package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// AffiliateStatus represents the status of an affiliate
type AffiliateStatus string

const (
	AffiliateStatusPending  AffiliateStatus = "pending"
	AffiliateStatusActive   AffiliateStatus = "active"
	AffiliateStatusSuspended AffiliateStatus = "suspended"
	AffiliateStatusRejected AffiliateStatus = "rejected"
)

// Affiliate represents an affiliate/referrer account
type Affiliate struct {
	ID               uint64          `gorm:"primaryKey"`
	CustomerID       uint64          `gorm:"not null;uniqueIndex"`
	Status           AffiliateStatus `gorm:"size:32;not null;default:'pending'"`
	PayoutMethod     string          `gorm:"size:50"` // paypal, bank, credit
	PayoutEmail      string          `gorm:"size:255"`
	PayoutDetails    JSONMap         `gorm:"type:jsonb"`
	CommissionRate   decimal.Decimal `gorm:"type:numeric(10,4);not null;default:10"` // Percentage
	MinimumPayout    decimal.Decimal `gorm:"type:numeric(20,8);not null;default:50"`
	Currency         string          `gorm:"size:3;not null;default:'USD'"`
	Balance          decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	TotalEarned      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	TotalWithdrawn   decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	ReferralCode     string          `gorm:"size:50;uniqueIndex;not null"`
	ReferralURL      string          `gorm:"size:500"`
	Clicks           int64           `gorm:"not null;default:0"`
	Signups          int64           `gorm:"not null;default:0"`
	Conversions      int64           `gorm:"not null;default:0"`
	ConversionRate   decimal.Decimal `gorm:"type:numeric(10,4);not null;default:0"`
	Notes            string          `gorm:"type:text"`
	TermsAcceptedAt  *time.Time
	ApprovedAt       *time.Time
	ApprovedBy       *uint64
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`

	Customer User  `gorm:"foreignKey:CustomerID"`
	Approver *User `gorm:"foreignKey:ApprovedBy"`
}

// IsActive checks if the affiliate is active
func (a *Affiliate) IsActive() bool {
	return a.Status == AffiliateStatusActive
}

// CanWithdraw checks if the affiliate can make a withdrawal
func (a *Affiliate) CanWithdraw() bool {
	return a.IsActive() && a.Balance.GreaterThanOrEqual(a.MinimumPayout)
}

// AffiliateReferral represents a referral from an affiliate
type AffiliateReferral struct {
	ID            uint64    `gorm:"primaryKey"`
	AffiliateID   uint64    `gorm:"not null;index"`
	CustomerID    *uint64   `gorm:"index"` // Set when customer signs up
	IPAddress     string    `gorm:"size:45"`
	UserAgent     string    `gorm:"size:512"`
	ReferrerURL   string    `gorm:"size:500"`
	LandingPage   string    `gorm:"size:500"`
	SignedUpAt    *time.Time
	ConvertedAt   *time.Time // When they made their first purchase
	CreatedAt     time.Time `gorm:"not null;index"`

	Affiliate Affiliate `gorm:"foreignKey:AffiliateID"`
	Customer  *User     `gorm:"foreignKey:CustomerID"`
}

// AffiliateCommission represents a commission earned by an affiliate
type AffiliateCommission struct {
	ID            uint64          `gorm:"primaryKey"`
	AffiliateID   uint64          `gorm:"not null;index"`
	ReferralID    *uint64         `gorm:"index"`
	InvoiceID     *uint64         `gorm:"index"`
	OrderID       *uint64         `gorm:"index"`
	Type          string          `gorm:"size:32;not null"` // signup, purchase, renewal, recurring
	Amount        decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Currency      string          `gorm:"size:3;not null"`
	Status        string          `gorm:"size:32;not null;default:'pending'"` // pending, approved, paid, cancelled
	BaseAmount    decimal.Decimal `gorm:"type:numeric(20,8);not null"` // Original order amount
	Rate          decimal.Decimal `gorm:"type:numeric(10,4);not null"` // Commission rate applied
	Description   string          `gorm:"size:500"`
	ApprovedAt    *time.Time
	ApprovedBy    *uint64
	PaidAt        *time.Time
	WithdrawalID  *uint64   `gorm:"index"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Affiliate  Affiliate           `gorm:"foreignKey:AffiliateID"`
	Referral   *AffiliateReferral  `gorm:"foreignKey:ReferralID"`
	Invoice    *Invoice            `gorm:"foreignKey:InvoiceID"`
	Order      *Order              `gorm:"foreignKey:OrderID"`
	Approver   *User               `gorm:"foreignKey:ApprovedBy"`
	Withdrawal *AffiliateWithdrawal `gorm:"foreignKey:WithdrawalID"`
}

// IsPending checks if the commission is pending
func (c *AffiliateCommission) IsPending() bool {
	return c.Status == "pending"
}

// AffiliateWithdrawalStatus represents the status of a withdrawal
type AffiliateWithdrawalStatus string

const (
	AffiliateWithdrawalPending   AffiliateWithdrawalStatus = "pending"
	AffiliateWithdrawalApproved  AffiliateWithdrawalStatus = "approved"
	AffiliateWithdrawalProcessing AffiliateWithdrawalStatus = "processing"
	AffiliateWithdrawalCompleted AffiliateWithdrawalStatus = "completed"
	AffiliateWithdrawalRejected  AffiliateWithdrawalStatus = "rejected"
)

// AffiliateWithdrawal represents a withdrawal request from an affiliate
type AffiliateWithdrawal struct {
	ID            uint64                    `gorm:"primaryKey"`
	AffiliateID   uint64                    `gorm:"not null;index"`
	Amount        decimal.Decimal           `gorm:"type:numeric(20,8);not null"`
	Currency      string                    `gorm:"size:3;not null"`
	Status        AffiliateWithdrawalStatus `gorm:"size:32;not null;default:'pending'"`
	PayoutMethod  string                    `gorm:"size:50;not null"`
	PayoutDetails JSONMap                   `gorm:"type:jsonb"`
	TransactionRef string                   `gorm:"size:255"`
	ProcessedBy   *uint64
	ProcessedAt   *time.Time
	Notes         string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Affiliate  Affiliate             `gorm:"foreignKey:AffiliateID"`
	Processor  *User                 `gorm:"foreignKey:ProcessedBy"`
	Commissions []AffiliateCommission `gorm:"foreignKey:WithdrawalID"`
}

// AffiliateTier represents a tier in the affiliate program
type AffiliateTier struct {
	ID               uint64          `gorm:"primaryKey"`
	Name             string          `gorm:"size:100;not null"`
	MinSales         decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	CommissionRate   decimal.Decimal `gorm:"type:numeric(10,4);not null"`
	RecurringRate    decimal.Decimal `gorm:"type:numeric(10,4);not null;default:0"` // For recurring commissions
	BonusAmount      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Description      string          `gorm:"type:text"`
	Benefits         JSONMap         `gorm:"type:jsonb"`
	Active           bool            `gorm:"not null;default:true"`
	SortOrder        int             `gorm:"not null;default:0"`
	CreatedAt        time.Time       `gorm:"not null"`
	UpdatedAt        time.Time       `gorm:"not null"`
}

// AffiliateBanner represents a promotional banner for affiliates
type AffiliateBanner struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	ImageURL    string    `gorm:"size:500;not null"`
	Width       int       `gorm:"not null"`
	Height      int       `gorm:"not null"`
	TargetURL   string    `gorm:"size:500"`
	HTMLCode    string    `gorm:"type:text"`
	Active      bool      `gorm:"not null;default:true"`
	Clicks      int64     `gorm:"not null;default:0"`
	SortOrder   int       `gorm:"not null;default:0"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// AffiliateClick represents a click tracked for an affiliate
type AffiliateClick struct {
	ID          uint64    `gorm:"primaryKey"`
	AffiliateID uint64    `gorm:"not null;index"`
	BannerID    *uint64   `gorm:"index"`
	IPAddress   string    `gorm:"size:45;not null"`
	UserAgent   string    `gorm:"size:512"`
	ReferrerURL string    `gorm:"size:500"`
	LandingPage string    `gorm:"size:500"`
	CreatedAt   time.Time `gorm:"not null;index"`

	Affiliate Affiliate        `gorm:"foreignKey:AffiliateID"`
	Banner    *AffiliateBanner `gorm:"foreignKey:BannerID"`
}

// AffiliateSettings represents affiliate program settings
type AffiliateSettings struct {
	Enabled           bool            `json:"enabled"`
	RequireApproval   bool            `json:"require_approval"`
	DefaultRate       decimal.Decimal `json:"default_rate"`
	MinimumPayout     decimal.Decimal `json:"minimum_payout"`
	RecurringEnabled  bool            `json:"recurring_enabled"`
	RecurringLifetime int             `json:"recurring_lifetime"` // Months, 0 = forever
	CookieDays        int             `json:"cookie_days"`
	AllowSelfReferral bool            `json:"allow_self_referral"`
	PayoutMethods     []string        `json:"payout_methods"`
	TermsAndConditions string         `json:"terms_and_conditions"`
}

// PromoCode represents a promotional code (similar to coupon but for marketing)
type PromoCode struct {
	ID              uint64          `gorm:"primaryKey"`
	Code            string          `gorm:"size:50;uniqueIndex;not null"`
	CampaignID      *uint64         `gorm:"index"`
	AffiliateID     *uint64         `gorm:"index"`
	Type            CouponType      `gorm:"size:32;not null"`
	Amount          decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Currency        string          `gorm:"size:3"`
	MaxUses         int             `gorm:"not null;default:0"`
	CurrentUses     int             `gorm:"not null;default:0"`
	MaxUsesPerUser  int             `gorm:"not null;default:1"`
	MinOrderAmount  decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	ApplicableTo    JSONMap         `gorm:"type:jsonb"` // Product IDs or categories
	StartsAt        *time.Time
	ExpiresAt       *time.Time
	Active          bool      `gorm:"not null;default:true"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`

	Campaign  *MarketingCampaign `gorm:"foreignKey:CampaignID"`
	Affiliate *Affiliate         `gorm:"foreignKey:AffiliateID"`
}

// MarketingCampaign represents a marketing campaign
type MarketingCampaign struct {
	ID            uint64          `gorm:"primaryKey"`
	Name          string          `gorm:"size:100;not null"`
	Description   string          `gorm:"type:text"`
	Type          string          `gorm:"size:32;not null"` // email, social, affiliate, ppc
	Budget        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Spent         decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Revenue       decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Conversions   int             `gorm:"not null;default:0"`
	Clicks        int             `gorm:"not null;default:0"`
	Impressions   int             `gorm:"not null;default:0"`
	TargetURL     string          `gorm:"size:500"`
	UTMSource     string          `gorm:"size:100"`
	UTMMedium     string          `gorm:"size:100"`
	UTMCampaign   string          `gorm:"size:100"`
	Status        string          `gorm:"size:32;not null;default:'draft'"` // draft, active, paused, completed
	StartsAt      *time.Time
	EndsAt        *time.Time
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

// Upsell represents an upsell opportunity
type Upsell struct {
	ID              uint64    `gorm:"primaryKey"`
	Name            string    `gorm:"size:100;not null"`
	Description     string    `gorm:"type:text"`
	TriggerType     string    `gorm:"size:32;not null"` // cart, order, service, renewal
	TriggerProducts JSONMap   `gorm:"type:jsonb"` // Products that trigger this upsell
	OfferProducts   JSONMap   `gorm:"type:jsonb"` // Products being offered
	DiscountType    string    `gorm:"size:32"`    // percentage, fixed, free_setup
	DiscountAmount  decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Priority        int       `gorm:"not null;default:0"`
	MaxDisplays     int       `gorm:"not null;default:0"` // 0 = unlimited
	DisplayCount    int       `gorm:"not null;default:0"`
	Conversions     int       `gorm:"not null;default:0"`
	Active          bool      `gorm:"not null;default:true"`
	StartsAt        *time.Time
	ExpiresAt       *time.Time
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

// CrossSell represents a cross-sell recommendation
type CrossSell struct {
	ID             uint64    `gorm:"primaryKey"`
	ProductID      uint64    `gorm:"not null;index"`
	RelatedProduct uint64    `gorm:"not null;index"`
	RelationType   string    `gorm:"size:32;not null"` // related, alternative, addon, required
	Priority       int       `gorm:"not null;default:0"`
	Active         bool      `gorm:"not null;default:true"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`

	Product Product `gorm:"foreignKey:ProductID"`
	Related Product `gorm:"foreignKey:RelatedProduct"`
}
