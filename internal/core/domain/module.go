package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// ModuleType represents the type of module
type ModuleType string

const (
	ModuleTypeProvisioning  ModuleType = "provisioning"
	ModuleTypePayment       ModuleType = "payment"
	ModuleTypeRegistrar     ModuleType = "registrar"
	ModuleTypeFraud         ModuleType = "fraud"
	ModuleTypeNotification  ModuleType = "notification"
	ModuleTypeReporting     ModuleType = "reporting"
	ModuleTypeAddon         ModuleType = "addon"
)

// ModuleStatus represents the status of a module
type ModuleStatus string

const (
	ModuleStatusActive   ModuleStatus = "active"
	ModuleStatusInactive ModuleStatus = "inactive"
	ModuleStatusError    ModuleStatus = "error"
)

// Module represents a system module/plugin
type Module struct {
	ID          uint64       `gorm:"primaryKey"`
	Name        string       `gorm:"size:100;not null"`
	Slug        string       `gorm:"size:100;uniqueIndex;not null"`
	Type        ModuleType   `gorm:"size:32;not null"`
	Description string       `gorm:"type:text"`
	Version     string       `gorm:"size:20;not null"`
	Author      string       `gorm:"size:100"`
	AuthorURL   string       `gorm:"size:255"`
	Status      ModuleStatus `gorm:"size:32;not null;default:'inactive'"`
	Config      JSONMap      `gorm:"type:jsonb"`
	ConfigSchema string      `gorm:"type:text"` // JSON schema for config
	LogoURL     string       `gorm:"size:500"`
	DocumentationURL string  `gorm:"size:500"`
	Dependencies JSONMap     `gorm:"type:jsonb"` // Required modules
	Permissions JSONMap      `gorm:"type:jsonb"` // Required permissions
	Events      JSONMap      `gorm:"type:jsonb"` // Events this module listens to
	InstallDate *time.Time
	LastError   string       `gorm:"type:text"`
	LastErrorAt *time.Time
	CreatedAt   time.Time    `gorm:"not null"`
	UpdatedAt   time.Time    `gorm:"not null"`
}

// IsActive checks if the module is active
func (m *Module) IsActive() bool {
	return m.Status == ModuleStatusActive
}

// ModuleHook represents a hook that a module can register
type ModuleHook struct {
	ID          uint64    `gorm:"primaryKey"`
	ModuleID    uint64    `gorm:"not null;index"`
	HookPoint   string    `gorm:"size:100;not null;index"` // The event/hook name
	Priority    int       `gorm:"not null;default:0"` // Higher = runs later
	Method      string    `gorm:"size:100;not null"` // Method to call
	Active      bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	Module Module `gorm:"foreignKey:ModuleID"`
}

// ModuleLog represents a log entry from a module
type ModuleLog struct {
	ID         uint64    `gorm:"primaryKey"`
	ModuleID   uint64    `gorm:"not null;index"`
	Level      string    `gorm:"size:20;not null;index"` // debug, info, warning, error
	Message    string    `gorm:"type:text;not null"`
	Context    JSONMap   `gorm:"type:jsonb"`
	RelatedType string   `gorm:"size:50;index"`
	RelatedID  *uint64   `gorm:"index"`
	CreatedAt  time.Time `gorm:"not null;index"`

	Module Module `gorm:"foreignKey:ModuleID"`
}

// ProvisioningModule represents a server provisioning module
type ProvisioningModule struct {
	ID            uint64    `gorm:"primaryKey"`
	ModuleID      uint64    `gorm:"not null;uniqueIndex"`
	SupportsCreate bool     `gorm:"not null;default:true"`
	SupportsSuspend bool    `gorm:"not null;default:true"`
	SupportsUnsuspend bool  `gorm:"not null;default:true"`
	SupportsTerminate bool  `gorm:"not null;default:true"`
	SupportsUpgrade bool    `gorm:"not null;default:false"`
	SupportsRename bool     `gorm:"not null;default:false"`
	SupportsUsage bool      `gorm:"not null;default:false"`
	SupportsSSO bool        `gorm:"not null;default:false"`
	SupportsBackup bool     `gorm:"not null;default:false"`
	RequiredFields JSONMap  `gorm:"type:jsonb"` // Required fields for provisioning
	ClientAreaFunctions JSONMap `gorm:"type:jsonb"` // Functions available in client area
	AdminAreaFunctions JSONMap `gorm:"type:jsonb"` // Functions available in admin area
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Module Module `gorm:"foreignKey:ModuleID"`
}

// AddonModule represents a product addon module
type AddonModule struct {
	ID             uint64    `gorm:"primaryKey"`
	ModuleID       uint64    `gorm:"not null;uniqueIndex"`
	SupportsProvision bool   `gorm:"not null;default:false"`
	SupportsTerminate bool   `gorm:"not null;default:false"`
	SupportsRenew  bool      `gorm:"not null;default:false"`
	ConfigOptions  JSONMap   `gorm:"type:jsonb"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`

	Module Module `gorm:"foreignKey:ModuleID"`
}

// Theme represents a frontend theme
type Theme struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	Slug        string    `gorm:"size:100;uniqueIndex;not null"`
	Description string    `gorm:"type:text"`
	Version     string    `gorm:"size:20"`
	Author      string    `gorm:"size:100"`
	AuthorURL   string    `gorm:"size:255"`
	Screenshot  string    `gorm:"size:500"`
	Type        string    `gorm:"size:32;not null"` // client, admin, both
	Active      bool      `gorm:"not null;default:false"`
	Default     bool      `gorm:"not null;default:false"`
	CustomCSS   string    `gorm:"type:text"`
	CustomJS    string    `gorm:"type:text"`
	Settings    JSONMap   `gorm:"type:jsonb"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// Language represents a language pack
type Language struct {
	ID          uint64    `gorm:"primaryKey"`
	Code        string    `gorm:"size:10;uniqueIndex;not null"` // ISO 639-1
	Name        string    `gorm:"size:100;not null"`
	NativeName  string    `gorm:"size:100"`
	Flag        string    `gorm:"size:10"` // Flag emoji
	Direction   string    `gorm:"size:3;not null;default:'ltr'"` // ltr, rtl
	Active      bool      `gorm:"not null;default:true"`
	Default     bool      `gorm:"not null;default:false"`
	Completeness int      `gorm:"not null;default:0"` // Percentage
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// Translation represents a translation string
type Translation struct {
	ID           uint64    `gorm:"primaryKey"`
	LanguageID   uint64    `gorm:"not null;uniqueIndex:idx_lang_key"`
	Key          string    `gorm:"size:255;not null;uniqueIndex:idx_lang_key"`
	Value        string    `gorm:"type:text;not null"`
	Category     string    `gorm:"size:50;index"`
	CustomValue  string    `gorm:"type:text"` // Custom override
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`

	Language Language `gorm:"foreignKey:LanguageID"`
}

// CustomField represents a custom field definition
type CustomField struct {
	ID           uint64    `gorm:"primaryKey"`
	Type         string    `gorm:"size:32;not null"` // client, product, service, order
	Name         string    `gorm:"size:100;not null"`
	FieldName    string    `gorm:"size:100;not null"` // Internal field name
	FieldType    string    `gorm:"size:32;not null"` // text, textarea, select, checkbox, radio, date
	Description  string    `gorm:"type:text"`
	Options      JSONMap   `gorm:"type:jsonb"` // For select/radio
	Required     bool      `gorm:"not null;default:false"`
	ShowOnOrder  bool      `gorm:"not null;default:false"`
	ShowOnInvoice bool     `gorm:"not null;default:false"`
	AdminOnly    bool      `gorm:"not null;default:false"`
	ProductIDs   JSONMap   `gorm:"type:jsonb"` // Restrict to products
	Validation   string    `gorm:"size:255"` // Regex validation
	DefaultValue string    `gorm:"size:500"`
	SortOrder    int       `gorm:"not null;default:0"`
	Active       bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// CustomFieldValue represents a custom field value
type CustomFieldValue struct {
	ID           uint64    `gorm:"primaryKey"`
	CustomFieldID uint64   `gorm:"not null;index"`
	EntityType   string    `gorm:"size:32;not null;index"` // client, service, order
	EntityID     uint64    `gorm:"not null;index"`
	Value        string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`

	CustomField CustomField `gorm:"foreignKey:CustomFieldID"`
}

// SystemConfig represents system configuration settings
type SystemConfig struct {
	ID        uint64    `gorm:"primaryKey"`
	Key       string    `gorm:"size:100;uniqueIndex;not null"`
	Value     string    `gorm:"type:text"`
	Type      string    `gorm:"size:32;not null;default:'string'"` // string, int, bool, json, array
	Category  string    `gorm:"size:50;index"`
	Label     string    `gorm:"size:200"`
	HelpText  string    `gorm:"type:text"`
	Protected bool      `gorm:"not null;default:false"`
	Encrypted bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// IPBan represents an IP ban
type IPBan struct {
	ID        uint64     `gorm:"primaryKey"`
	IPAddress string     `gorm:"size:45;not null;index"`
	IPRange   string     `gorm:"size:50"` // CIDR notation
	Reason    string     `gorm:"size:500"`
	BannedBy  uint64     `gorm:"not null"`
	ExpiresAt *time.Time
	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`

	Admin User `gorm:"foreignKey:BannedBy"`
}

// IsExpired checks if the ban has expired
func (b *IPBan) IsExpired() bool {
	if b.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*b.ExpiresAt)
}

// CountryRestriction represents a country restriction
type CountryRestriction struct {
	ID           uint64    `gorm:"primaryKey"`
	CountryCode  string    `gorm:"size:2;not null;uniqueIndex:idx_country_type"`
	Type         string    `gorm:"size:32;not null;uniqueIndex:idx_country_type"` // registration, ordering, payment
	Allowed      bool      `gorm:"not null;default:true"`
	Reason       string    `gorm:"size:500"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	ID           uint64    `gorm:"primaryKey"`
	Name         string    `gorm:"size:100;not null"`
	Type         string    `gorm:"size:32;not null"` // database, files, full
	Destination  string    `gorm:"size:32;not null"` // local, s3, ftp, sftp
	Config       JSONMap   `gorm:"type:jsonb"`
	Schedule     string    `gorm:"size:50"` // Cron expression
	Retention    int       `gorm:"not null;default:7"` // Days
	Active       bool      `gorm:"not null;default:true"`
	LastRun      *time.Time
	LastStatus   string    `gorm:"size:32"`
	LastError    string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// BackupLog represents a backup execution log
type BackupLog struct {
	ID           uint64    `gorm:"primaryKey"`
	BackupConfigID uint64  `gorm:"not null;index"`
	Status       string    `gorm:"size:32;not null"` // running, success, failed
	StartedAt    time.Time `gorm:"not null"`
	CompletedAt  *time.Time
	FileSize     int64     `gorm:"not null;default:0"`
	FilePath     string    `gorm:"size:500"`
	ErrorMsg     string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"not null"`

	BackupConfig BackupConfig `gorm:"foreignKey:BackupConfigID"`
}

// SystemHealth represents system health metrics
type SystemHealth struct {
	ID            uint64          `gorm:"primaryKey"`
	Component     string          `gorm:"size:50;not null;index"` // database, redis, mail, etc.
	Status        string          `gorm:"size:32;not null"` // healthy, degraded, down
	ResponseTime  int             `gorm:"not null;default:0"` // Milliseconds
	Message       string          `gorm:"size:500"`
	Metrics       JSONMap         `gorm:"type:jsonb"`
	CheckedAt     time.Time       `gorm:"not null;index"`
	CreatedAt     time.Time       `gorm:"not null"`
}

// UsageStatistic represents a usage statistic for a service
type UsageStatistic struct {
	ID            uint64          `gorm:"primaryKey"`
	ServiceID     uint64          `gorm:"not null;index"`
	Type          string          `gorm:"size:32;not null;index"` // bandwidth, disk, cpu, memory
	Date          time.Time       `gorm:"not null;uniqueIndex:idx_service_type_date"`
	Value         decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Unit          string          `gorm:"size:20;not null"` // GB, MB, %, etc.
	Peak          decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Average       decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	CreatedAt     time.Time       `gorm:"not null"`
	UpdatedAt     time.Time       `gorm:"not null"`

	Service Service `gorm:"foreignKey:ServiceID"`
}

// UsageBillingRule represents a usage-based billing rule
type UsageBillingRule struct {
	ID             uint64          `gorm:"primaryKey"`
	ProductID      uint64          `gorm:"not null;index"`
	UsageType      string          `gorm:"size:32;not null"` // bandwidth, disk, cpu, etc.
	IncludedAmount decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Unit           string          `gorm:"size:20;not null"`
	OverageRate    decimal.Decimal `gorm:"type:numeric(20,8);not null"` // Per unit
	OverageCap     decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"` // Max overage charge
	BillingMethod  string          `gorm:"size:32;not null"` // tiered, per_unit, flat
	Tiers          UsageTiers      `gorm:"type:jsonb"`
	Active         bool            `gorm:"not null;default:true"`
	CreatedAt      time.Time       `gorm:"not null"`
	UpdatedAt      time.Time       `gorm:"not null"`

	Product Product `gorm:"foreignKey:ProductID"`
}

// UsageTiers represents usage billing tiers
type UsageTiers []UsageTier

// UsageTier represents a single usage tier
type UsageTier struct {
	UpTo   decimal.Decimal `json:"up_to"`
	Rate   decimal.Decimal `json:"rate"`
	Flat   decimal.Decimal `json:"flat,omitempty"` // Flat fee for this tier
}

// Value implements driver.Valuer for UsageTiers
func (t UsageTiers) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan implements sql.Scanner for UsageTiers
func (t *UsageTiers) Scan(value interface{}) error {
	if value == nil {
		*t = UsageTiers{}
		return nil
	}
	data, err := normalizeJSONBytes(value)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		*t = UsageTiers{}
		return nil
	}
	return json.Unmarshal(data, t)
}
