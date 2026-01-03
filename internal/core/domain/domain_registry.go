package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// DomainStatus represents the status of a domain
type DomainStatus string

const (
	DomainStatusPending      DomainStatus = "pending"
	DomainStatusActive       DomainStatus = "active"
	DomainStatusExpired      DomainStatus = "expired"
	DomainStatusTransferring DomainStatus = "transferring"
	DomainStatusRedemption   DomainStatus = "redemption"
	DomainStatusSuspended    DomainStatus = "suspended"
	DomainStatusCancelled    DomainStatus = "cancelled"
)

// DomainRegister represents a domain registrar integration
type DomainRegister struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	ModuleName  string    `gorm:"size:100;not null;uniqueIndex"`
	DisplayName string    `gorm:"size:100;not null"`
	Active      bool      `gorm:"not null;default:true"`
	Default     bool      `gorm:"not null;default:false"`
	Config      JSONMap   `gorm:"type:jsonb"`
	TestMode    bool      `gorm:"not null;default:false"`
	SortOrder   int       `gorm:"not null;default:0"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// DomainTLD represents a Top Level Domain configuration
type DomainTLD struct {
	ID                 uint64          `gorm:"primaryKey"`
	TLD                string          `gorm:"size:50;uniqueIndex;not null"` // .com, .net, etc
	RegistrarID        uint64          `gorm:"not null;index"`
	Active             bool            `gorm:"not null;default:true"`
	IDProtectionEnabled bool           `gorm:"not null;default:false"`
	IDProtectionCost   decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	EppRequired        bool            `gorm:"not null;default:true"`
	AutoRenew          bool            `gorm:"not null;default:true"`
	MinYears           int             `gorm:"not null;default:1"`
	MaxYears           int             `gorm:"not null;default:10"`
	GracePeriodDays    int             `gorm:"not null;default:30"`
	RedemptionDays     int             `gorm:"not null;default:30"`
	DNSManagement      bool            `gorm:"not null;default:true"`
	EmailForwarding    bool            `gorm:"not null;default:false"`
	RequiredFields     JSONMap         `gorm:"type:jsonb"` // Additional required fields
	SortOrder          int             `gorm:"not null;default:0"`
	CreatedAt          time.Time       `gorm:"not null"`
	UpdatedAt          time.Time       `gorm:"not null"`

	Registrar DomainRegister `gorm:"foreignKey:RegistrarID"`
}

// DomainTLDPricing represents pricing for a TLD
type DomainTLDPricing struct {
	ID           uint64          `gorm:"primaryKey"`
	TLDID        uint64          `gorm:"not null;uniqueIndex:idx_tld_currency_type"`
	Currency     string          `gorm:"size:3;not null;uniqueIndex:idx_tld_currency_type"`
	Type         string          `gorm:"size:32;not null;uniqueIndex:idx_tld_currency_type"` // register, transfer, renew
	Year1        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Year2        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Year3        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Year4        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Year5        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Year6        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Year7        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Year8        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Year9        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Year10       decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	CreatedAt    time.Time       `gorm:"not null"`
	UpdatedAt    time.Time       `gorm:"not null"`

	TLD DomainTLD `gorm:"foreignKey:TLDID"`
}

// GetPriceForYears returns the total price for the given number of years
func (p *DomainTLDPricing) GetPriceForYears(years int) decimal.Decimal {
	prices := map[int]decimal.Decimal{
		1: p.Year1, 2: p.Year2, 3: p.Year3, 4: p.Year4, 5: p.Year5,
		6: p.Year6, 7: p.Year7, 8: p.Year8, 9: p.Year9, 10: p.Year10,
	}
	if price, ok := prices[years]; ok {
		return price
	}
	return decimal.Zero
}

// CustomerDomain represents a domain owned by a customer
type CustomerDomain struct {
	ID               uint64          `gorm:"primaryKey"`
	CustomerID       uint64          `gorm:"not null;index"`
	TLDID            uint64          `gorm:"not null;index"`
	Domain           string          `gorm:"size:255;uniqueIndex;not null"`
	Status           DomainStatus    `gorm:"size:32;not null;default:'pending'"`
	RegistrarID      uint64          `gorm:"not null;index"`
	RegistrarDomainID string         `gorm:"size:255"` // ID at registrar
	RegistrationDate time.Time       `gorm:"not null"`
	ExpiryDate       time.Time       `gorm:"not null;index"`
	NextDueDate      time.Time       `gorm:"not null;index"`
	BillingYears     int             `gorm:"not null;default:1"`
	AutoRenew        bool            `gorm:"not null;default:true"`
	IDProtection     bool            `gorm:"not null;default:false"`
	RecurringAmount  decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Currency         string          `gorm:"size:3;not null;default:'USD'"`
	EPPCode          string          `gorm:"size:100"` // Transfer auth code (encrypted)
	DNSManagement    bool            `gorm:"not null;default:true"`
	EmailForwarding  bool            `gorm:"not null;default:false"`
	FirstPaymentAmount decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Notes            string          `gorm:"type:text"`
	AdminNotes       string          `gorm:"type:text"`
	CreatedAt        time.Time       `gorm:"not null"`
	UpdatedAt        time.Time       `gorm:"not null"`

	Customer  User           `gorm:"foreignKey:CustomerID"`
	TLD       DomainTLD      `gorm:"foreignKey:TLDID"`
	Registrar DomainRegister `gorm:"foreignKey:RegistrarID"`
}

// IsDueForRenewal checks if the domain is due for renewal
func (d *CustomerDomain) IsDueForRenewal() bool {
	return time.Now().After(d.NextDueDate)
}

// IsExpired checks if the domain has expired
func (d *CustomerDomain) IsExpired() bool {
	return time.Now().After(d.ExpiryDate)
}

// DomainContact represents contact information for a domain
type DomainContact struct {
	ID           uint64    `gorm:"primaryKey"`
	DomainID     uint64    `gorm:"not null;index"`
	Type         string    `gorm:"size:32;not null"` // registrant, admin, tech, billing
	FirstName    string    `gorm:"size:100;not null"`
	LastName     string    `gorm:"size:100;not null"`
	Organization string    `gorm:"size:255"`
	Email        string    `gorm:"size:255;not null"`
	Phone        string    `gorm:"size:32;not null"`
	PhoneCC      string    `gorm:"size:5;not null"` // Country code
	Fax          string    `gorm:"size:32"`
	FaxCC        string    `gorm:"size:5"`
	Address1     string    `gorm:"size:255;not null"`
	Address2     string    `gorm:"size:255"`
	City         string    `gorm:"size:100;not null"`
	State        string    `gorm:"size:100;not null"`
	PostalCode   string    `gorm:"size:20;not null"`
	Country      string    `gorm:"size:2;not null"` // ISO 3166-1 alpha-2
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`

	Domain CustomerDomain `gorm:"foreignKey:DomainID"`
}

// DomainDNSRecord represents a DNS record for a domain
type DomainDNSRecord struct {
	ID        uint64    `gorm:"primaryKey"`
	DomainID  uint64    `gorm:"not null;index"`
	Type      string    `gorm:"size:16;not null"` // A, AAAA, CNAME, MX, TXT, NS, SRV, CAA
	Hostname  string    `gorm:"size:255;not null"`
	Address   string    `gorm:"size:500;not null"` // Value/Target
	Priority  int       `gorm:"not null;default:0"` // For MX, SRV
	TTL       int       `gorm:"not null;default:3600"`
	Weight    int       `gorm:"not null;default:0"` // For SRV
	Port      int       `gorm:"not null;default:0"` // For SRV
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Domain CustomerDomain `gorm:"foreignKey:DomainID"`
}

// DomainEmailForward represents an email forwarding rule
type DomainEmailForward struct {
	ID          uint64    `gorm:"primaryKey"`
	DomainID    uint64    `gorm:"not null;index"`
	Prefix      string    `gorm:"size:100;not null"` // email prefix or * for catch-all
	Destination string    `gorm:"size:255;not null"` // Forward to address
	Active      bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	Domain CustomerDomain `gorm:"foreignKey:DomainID"`
}

// DomainNameserver represents a nameserver for a domain
type DomainNameserver struct {
	ID        uint64    `gorm:"primaryKey"`
	DomainID  uint64    `gorm:"not null;index"`
	Hostname  string    `gorm:"size:255;not null"`
	IPAddress string    `gorm:"size:45"` // For glue records
	SortOrder int       `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Domain CustomerDomain `gorm:"foreignKey:DomainID"`
}

// DomainTransfer represents a domain transfer request
type DomainTransfer struct {
	ID            uint64       `gorm:"primaryKey"`
	CustomerID    uint64       `gorm:"not null;index"`
	DomainName    string       `gorm:"size:255;not null"`
	TLDID         uint64       `gorm:"not null;index"`
	RegistrarID   uint64       `gorm:"not null;index"`
	EPPCode       string       `gorm:"size:100;not null"` // Encrypted
	Status        string       `gorm:"size:32;not null;default:'pending'"` // pending, processing, completed, failed, cancelled
	StatusMessage string       `gorm:"type:text"`
	OrderID       *uint64      `gorm:"index"`
	DomainID      *uint64      `gorm:"index"` // Created after successful transfer
	RequestedAt   time.Time    `gorm:"not null"`
	ProcessedAt   *time.Time
	CompletedAt   *time.Time
	CreatedAt     time.Time    `gorm:"not null"`
	UpdatedAt     time.Time    `gorm:"not null"`

	Customer  User            `gorm:"foreignKey:CustomerID"`
	TLD       DomainTLD       `gorm:"foreignKey:TLDID"`
	Registrar DomainRegister  `gorm:"foreignKey:RegistrarID"`
	Order     *Order          `gorm:"foreignKey:OrderID"`
	Domain    *CustomerDomain `gorm:"foreignKey:DomainID"`
}

// DomainRenewal represents a domain renewal record
type DomainRenewal struct {
	ID              uint64          `gorm:"primaryKey"`
	DomainID        uint64          `gorm:"not null;index"`
	InvoiceID       *uint64         `gorm:"index"`
	Type            string          `gorm:"size:32;not null"` // auto, manual, grace, redemption
	Years           int             `gorm:"not null;default:1"`
	Amount          decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Currency        string          `gorm:"size:3;not null"`
	OldExpiryDate   time.Time       `gorm:"not null"`
	NewExpiryDate   time.Time       `gorm:"not null"`
	Status          string          `gorm:"size:32;not null;default:'pending'"` // pending, completed, failed
	StatusMessage   string          `gorm:"type:text"`
	ProcessedAt     *time.Time
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`

	Domain  CustomerDomain `gorm:"foreignKey:DomainID"`
	Invoice *Invoice       `gorm:"foreignKey:InvoiceID"`
}

// WHOISPrivacy represents WHOIS privacy protection for a domain
type WHOISPrivacy struct {
	ID          uint64    `gorm:"primaryKey"`
	DomainID    uint64    `gorm:"not null;uniqueIndex"`
	Active      bool      `gorm:"not null;default:true"`
	ExpiryDate  time.Time `gorm:"not null"`
	ServiceID   *uint64   `gorm:"index"` // If billed as separate service
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	Domain  CustomerDomain `gorm:"foreignKey:DomainID"`
	Service *Service       `gorm:"foreignKey:ServiceID"`
}

// DomainAvailability represents a domain availability check result
type DomainAvailability struct {
	Domain      string          `json:"domain"`
	Available   bool            `json:"available"`
	Premium     bool            `json:"premium"`
	PremiumPrice decimal.Decimal `json:"premium_price,omitempty"`
	Currency    string          `json:"currency,omitempty"`
	Reason      string          `json:"reason,omitempty"`
	Suggestions []string        `json:"suggestions,omitempty"`
}

// DomainSearchResult represents domain search results with pricing
type DomainSearchResult struct {
	Domain        string          `json:"domain"`
	TLD           string          `json:"tld"`
	Available     bool            `json:"available"`
	Premium       bool            `json:"premium"`
	RegisterPrice decimal.Decimal `json:"register_price"`
	RenewPrice    decimal.Decimal `json:"renew_price"`
	TransferPrice decimal.Decimal `json:"transfer_price"`
	Currency      string          `json:"currency"`
}
