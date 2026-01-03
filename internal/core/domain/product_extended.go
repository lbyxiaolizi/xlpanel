package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// ProductType represents the type of product
type ProductType string

const (
	ProductTypeHosting      ProductType = "hosting"
	ProductTypeVPS          ProductType = "vps"
	ProductTypeDedicated    ProductType = "dedicated"
	ProductTypeResellerHosting ProductType = "reseller"
	ProductTypeDomain       ProductType = "domain"
	ProductTypeSSL          ProductType = "ssl"
	ProductTypeLicense      ProductType = "license"
	ProductTypeOther        ProductType = "other"
)

// ProductVisibility represents product visibility settings
type ProductVisibility string

const (
	ProductVisibilityPublic   ProductVisibility = "public"
	ProductVisibilityHidden   ProductVisibility = "hidden"
	ProductVisibilityCustomerOnly ProductVisibility = "customer_only"
)

// ProductAddon represents an addon that can be added to products
type ProductAddon struct {
	ID             uint64          `gorm:"primaryKey"`
	Name           string          `gorm:"size:255;not null"`
	Description    string          `gorm:"type:text"`
	Type           string          `gorm:"size:32;not null"` // recurring, onetime
	ModuleID       *uint64         `gorm:"index"`
	SetupFee       decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Monthly        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Quarterly      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	SemiAnnually   decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Annually       decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Biennially     decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Triennially    decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	OneTimePrice   decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Weight         int             `gorm:"not null;default:0"` // Sort order
	ShowOnOrder    bool            `gorm:"not null;default:true"`
	Active         bool            `gorm:"not null;default:true"`
	SuspendParent  bool            `gorm:"not null;default:false"` // Suspend if parent suspended
	ProvisionAutomatically bool   `gorm:"not null;default:true"`
	AllowQuantity  bool            `gorm:"not null;default:false"`
	MaxQuantity    int             `gorm:"not null;default:0"` // 0 = unlimited
	ModuleConfig   JSONMap         `gorm:"type:jsonb"`
	CreatedAt      time.Time       `gorm:"not null"`
	UpdatedAt      time.Time       `gorm:"not null"`

	Module   *Module   `gorm:"foreignKey:ModuleID"`
	Products []Product `gorm:"many2many:product_addon_assignments"`
}

// ProductAddonAssignment represents the assignment of an addon to a product
type ProductAddonAssignment struct {
	ProductID uint64    `gorm:"primaryKey"`
	AddonID   uint64    `gorm:"primaryKey"`
	Required  bool      `gorm:"not null;default:false"`
	SortOrder int       `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"not null"`

	Product Product      `gorm:"foreignKey:ProductID"`
	Addon   ProductAddon `gorm:"foreignKey:AddonID"`
}

// ServiceAddon represents an addon attached to a customer service
type ServiceAddon struct {
	ID             uint64          `gorm:"primaryKey"`
	ServiceID      uint64          `gorm:"not null;index"`
	AddonID        uint64          `gorm:"not null;index"`
	Quantity       int             `gorm:"not null;default:1"`
	Status         ServiceStatus   `gorm:"size:64;not null;default:'active'"`
	BillingCycle   string          `gorm:"size:32"`
	RecurringAmount decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	NextDueDate    time.Time       `gorm:"not null;index"`
	SetupFeeApplied bool           `gorm:"not null;default:false"`
	Notes          string          `gorm:"type:text"`
	CreatedAt      time.Time       `gorm:"not null"`
	UpdatedAt      time.Time       `gorm:"not null"`

	Service Service      `gorm:"foreignKey:ServiceID"`
	Addon   ProductAddon `gorm:"foreignKey:AddonID"`
}

// ProductBundle represents a bundle of products
type ProductBundle struct {
	ID            uint64          `gorm:"primaryKey"`
	Name          string          `gorm:"size:255;not null"`
	Description   string          `gorm:"type:text"`
	SetupFee      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Monthly       decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Quarterly     decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	SemiAnnually  decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Annually      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Biennially    decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Triennially   decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	AllowCustomize bool           `gorm:"not null;default:false"` // Allow changing products
	ShowSavings   bool            `gorm:"not null;default:true"`
	Active        bool            `gorm:"not null;default:true"`
	SortOrder     int             `gorm:"not null;default:0"`
	CreatedAt     time.Time       `gorm:"not null"`
	UpdatedAt     time.Time       `gorm:"not null"`

	Items []ProductBundleItem `gorm:"foreignKey:BundleID"`
}

// ProductBundleItem represents a product in a bundle
type ProductBundleItem struct {
	ID            uint64          `gorm:"primaryKey"`
	BundleID      uint64          `gorm:"not null;index"`
	ProductID     uint64          `gorm:"not null;index"`
	Quantity      int             `gorm:"not null;default:1"`
	Optional      bool            `gorm:"not null;default:false"`
	Discount      decimal.Decimal `gorm:"type:numeric(10,4);not null;default:0"` // Percentage
	SortOrder     int             `gorm:"not null;default:0"`
	CreatedAt     time.Time       `gorm:"not null"`
	UpdatedAt     time.Time       `gorm:"not null"`

	Bundle  ProductBundle `gorm:"foreignKey:BundleID"`
	Product Product       `gorm:"foreignKey:ProductID"`
}

// ProductUpgrade represents upgrade/downgrade paths for a product
type ProductUpgrade struct {
	ID              uint64          `gorm:"primaryKey"`
	SourceProductID uint64          `gorm:"not null;index"`
	TargetProductID uint64          `gorm:"not null;index"`
	Type            string          `gorm:"size:32;not null"` // upgrade, downgrade, crossgrade
	Enabled         bool            `gorm:"not null;default:true"`
	ProrationCredit bool            `gorm:"not null;default:true"` // Credit remaining days
	ChargeSetupFee  bool            `gorm:"not null;default:false"`
	ClearExisting   bool            `gorm:"not null;default:false"` // Clear config options
	SortOrder       int             `gorm:"not null;default:0"`
	CreatedAt       time.Time       `gorm:"not null"`
	UpdatedAt       time.Time       `gorm:"not null"`

	SourceProduct Product `gorm:"foreignKey:SourceProductID"`
	TargetProduct Product `gorm:"foreignKey:TargetProductID"`
}

// ConfigurableOptionType represents the type of configurable option
type ConfigurableOptionType string

const (
	OptionTypeDropdown  ConfigurableOptionType = "dropdown"
	OptionTypeRadio     ConfigurableOptionType = "radio"
	OptionTypeCheckbox  ConfigurableOptionType = "checkbox"
	OptionTypeQuantity  ConfigurableOptionType = "quantity"
	OptionTypeText      ConfigurableOptionType = "text"
)

// ConfigurableOptionGroup represents a group of configurable options
type ConfigurableOptionGroup struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:255;not null"`
	Description string    `gorm:"type:text"`
	SortOrder   int       `gorm:"not null;default:0"`
	Active      bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	Options  []ConfigurableOption `gorm:"foreignKey:GroupID"`
	Products []Product            `gorm:"many2many:product_configurable_option_groups"`
}

// ConfigurableOption represents a configurable option
type ConfigurableOption struct {
	ID          uint64                 `gorm:"primaryKey"`
	GroupID     uint64                 `gorm:"not null;index"`
	Name        string                 `gorm:"size:255;not null"`
	OptionType  ConfigurableOptionType `gorm:"size:32;not null"`
	Description string                 `gorm:"type:text"`
	Required    bool                   `gorm:"not null;default:false"`
	Hidden      bool                   `gorm:"not null;default:false"`
	SortOrder   int                    `gorm:"not null;default:0"`
	Active      bool                   `gorm:"not null;default:true"`
	CreatedAt   time.Time              `gorm:"not null"`
	UpdatedAt   time.Time              `gorm:"not null"`

	Group      ConfigurableOptionGroup  `gorm:"foreignKey:GroupID"`
	SubOptions []ConfigurableSubOption  `gorm:"foreignKey:OptionID"`
}

// ConfigurableSubOption represents a sub-option for a configurable option
type ConfigurableSubOption struct {
	ID             uint64          `gorm:"primaryKey"`
	OptionID       uint64          `gorm:"not null;index"`
	Name           string          `gorm:"size:255;not null"`
	SetupFee       decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Monthly        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Quarterly      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	SemiAnnually   decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Annually       decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Biennially     decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Triennially    decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	MinQuantity    int             `gorm:"not null;default:0"`
	MaxQuantity    int             `gorm:"not null;default:0"` // 0 = unlimited
	SortOrder      int             `gorm:"not null;default:0"`
	Hidden         bool            `gorm:"not null;default:false"`
	Active         bool            `gorm:"not null;default:true"`
	StockControl   bool            `gorm:"not null;default:false"`
	StockQuantity  int             `gorm:"not null;default:0"`
	OutOfStockMsg  string          `gorm:"size:255"`
	CreatedAt      time.Time       `gorm:"not null"`
	UpdatedAt      time.Time       `gorm:"not null"`

	Option ConfigurableOption `gorm:"foreignKey:OptionID"`
}

// ServiceUpgrade represents an upgrade request for a service
type ServiceUpgrade struct {
	ID                uint64          `gorm:"primaryKey"`
	ServiceID         uint64          `gorm:"not null;index"`
	OldProductID      uint64          `gorm:"not null;index"`
	NewProductID      uint64          `gorm:"not null;index"`
	OldConfig         JSONMap         `gorm:"type:jsonb"`
	NewConfig         JSONMap         `gorm:"type:jsonb"`
	Type              string          `gorm:"size:32;not null"` // upgrade, downgrade, crossgrade
	CreditAmount      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	NewPrice          decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	SetupFee          decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	ProratedCharge    decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Status            string          `gorm:"size:32;not null;default:'pending'"` // pending, approved, completed, cancelled
	InvoiceID         *uint64         `gorm:"index"`
	ProcessedAt       *time.Time
	RequestedBy       uint64          `gorm:"not null"`
	Notes             string          `gorm:"type:text"`
	CreatedAt         time.Time       `gorm:"not null"`
	UpdatedAt         time.Time       `gorm:"not null"`

	Service    Service  `gorm:"foreignKey:ServiceID"`
	OldProduct Product  `gorm:"foreignKey:OldProductID"`
	NewProduct Product  `gorm:"foreignKey:NewProductID"`
	Invoice    *Invoice `gorm:"foreignKey:InvoiceID"`
	Requester  User     `gorm:"foreignKey:RequestedBy"`
}

// ProductStock represents stock tracking for a product
type ProductStock struct {
	ID             uint64    `gorm:"primaryKey"`
	ProductID      uint64    `gorm:"not null;uniqueIndex"`
	Quantity       int       `gorm:"not null;default:0"`
	ReservedQty    int       `gorm:"not null;default:0"` // Reserved by pending orders
	LowStockAlert  int       `gorm:"not null;default:5"` // Alert when below
	OutOfStockMsg  string    `gorm:"size:500"`
	AllowBackorder bool      `gorm:"not null;default:false"`
	LastRestocked  *time.Time
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`

	Product Product `gorm:"foreignKey:ProductID"`
}

// AvailableQuantity returns the quantity available for sale
func (s *ProductStock) AvailableQuantity() int {
	return s.Quantity - s.ReservedQty
}

// IsLowStock checks if stock is below the alert threshold
func (s *ProductStock) IsLowStock() bool {
	return s.AvailableQuantity() <= s.LowStockAlert
}

// ProductPricing represents detailed pricing for a product
type ProductPricing struct {
	ID             uint64          `gorm:"primaryKey"`
	ProductID      uint64          `gorm:"not null;uniqueIndex:idx_product_currency"`
	Currency       string          `gorm:"size:3;not null;uniqueIndex:idx_product_currency"`
	SetupFee       decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Monthly        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:-1"` // -1 = disabled
	Quarterly      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:-1"`
	SemiAnnually   decimal.Decimal `gorm:"type:numeric(20,8);not null;default:-1"`
	Annually       decimal.Decimal `gorm:"type:numeric(20,8);not null;default:-1"`
	Biennially     decimal.Decimal `gorm:"type:numeric(20,8);not null;default:-1"`
	Triennially    decimal.Decimal `gorm:"type:numeric(20,8);not null;default:-1"`
	CreatedAt      time.Time       `gorm:"not null"`
	UpdatedAt      time.Time       `gorm:"not null"`

	Product Product `gorm:"foreignKey:ProductID"`
}

// GetPrice returns the price for a billing cycle
func (p *ProductPricing) GetPrice(cycle string) decimal.Decimal {
	switch cycle {
	case "monthly":
		return p.Monthly
	case "quarterly":
		return p.Quarterly
	case "semiannually", "semi-annually":
		return p.SemiAnnually
	case "annually", "yearly":
		return p.Annually
	case "biennially":
		return p.Biennially
	case "triennially":
		return p.Triennially
	default:
		return decimal.NewFromInt(-1)
	}
}

// IsCycleEnabled checks if a billing cycle is enabled
func (p *ProductPricing) IsCycleEnabled(cycle string) bool {
	return p.GetPrice(cycle).GreaterThanOrEqual(decimal.Zero)
}

// ProductWelcomeEmail represents a custom welcome email for a product
type ProductWelcomeEmail struct {
	ID        uint64    `gorm:"primaryKey"`
	ProductID uint64    `gorm:"not null;uniqueIndex"`
	Subject   string    `gorm:"size:255;not null"`
	Body      string    `gorm:"type:text;not null"`
	Active    bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Product Product `gorm:"foreignKey:ProductID"`
}

// FreeTrialConfig represents free trial configuration for a product
type FreeTrialConfig struct {
	ID                uint64    `gorm:"primaryKey"`
	ProductID         uint64    `gorm:"not null;uniqueIndex"`
	Enabled           bool      `gorm:"not null;default:false"`
	Days              int       `gorm:"not null;default:7"`
	RequirePayment    bool      `gorm:"not null;default:false"` // Require payment method
	LimitPerCustomer  int       `gorm:"not null;default:1"` // 0 = unlimited
	AutoActivate      bool      `gorm:"not null;default:true"`
	ConvertToService  bool      `gorm:"not null;default:true"` // Auto-convert after trial
	CreatedAt         time.Time `gorm:"not null"`
	UpdatedAt         time.Time `gorm:"not null"`

	Product Product `gorm:"foreignKey:ProductID"`
}
