package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// TaxClass represents a tax class (e.g., Standard, Reduced, Zero-rated)
type TaxClass struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null;uniqueIndex"`
	Description string    `gorm:"type:text"`
	Default     bool      `gorm:"not null;default:false"`
	SortOrder   int       `gorm:"not null;default:0"`
	Active      bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// TaxRate represents a tax rate for a specific region
type TaxRate struct {
	ID          uint64          `gorm:"primaryKey"`
	TaxClassID  uint64          `gorm:"not null;index"`
	Country     string          `gorm:"size:2;not null;index"` // ISO 3166-1 alpha-2
	State       string          `gorm:"size:100;index"`        // State/Province/Region
	City        string          `gorm:"size:100"`
	PostalCode  string          `gorm:"size:20"`
	Name        string          `gorm:"size:100;not null"` // Tax name (e.g., VAT, GST)
	Rate        decimal.Decimal `gorm:"type:numeric(10,4);not null"` // Percentage rate
	Compound    bool            `gorm:"not null;default:false"` // Applied on top of other taxes
	Priority    int             `gorm:"not null;default:1"`
	Inclusive   bool            `gorm:"not null;default:false"` // Price includes tax
	Active      bool            `gorm:"not null;default:true"`
	StartsAt    *time.Time
	EndsAt      *time.Time
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	TaxClass TaxClass `gorm:"foreignKey:TaxClassID"`
}

// IsValidForDate checks if the tax rate is valid for a given date
func (t *TaxRate) IsValidForDate(date time.Time) bool {
	if !t.Active {
		return false
	}
	if t.StartsAt != nil && date.Before(*t.StartsAt) {
		return false
	}
	if t.EndsAt != nil && date.After(*t.EndsAt) {
		return false
	}
	return true
}

// TaxExemption represents a tax exemption for a customer
type TaxExemption struct {
	ID              uint64    `gorm:"primaryKey"`
	CustomerID      uint64    `gorm:"not null;index"`
	TaxClassID      *uint64   `gorm:"index"` // null = all tax classes
	Country         string    `gorm:"size:2"` // null = all countries
	ExemptionNumber string    `gorm:"size:100"`
	ExemptionType   string    `gorm:"size:50;not null"` // vat_registered, non_profit, reseller, government
	Notes           string    `gorm:"type:text"`
	VerifiedAt      *time.Time
	VerifiedBy      *uint64   `gorm:"index"`
	ExpiresAt       *time.Time
	Active          bool      `gorm:"not null;default:true"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`

	Customer User      `gorm:"foreignKey:CustomerID"`
	TaxClass *TaxClass `gorm:"foreignKey:TaxClassID"`
	Verifier *User     `gorm:"foreignKey:VerifiedBy"`
}

// IsValid checks if the exemption is currently valid
func (e *TaxExemption) IsValid() bool {
	if !e.Active {
		return false
	}
	if e.ExpiresAt != nil && time.Now().After(*e.ExpiresAt) {
		return false
	}
	return true
}

// TaxReport represents a tax report for a period
type TaxReport struct {
	ID              uint64          `gorm:"primaryKey"`
	Name            string          `gorm:"size:100;not null"`
	Country         string          `gorm:"size:2;not null"`
	PeriodStart     time.Time       `gorm:"not null"`
	PeriodEnd       time.Time       `gorm:"not null"`
	TotalSales      decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	TaxableSales    decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	ExemptSales     decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	TaxCollected    decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	TaxDue          decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Currency        string          `gorm:"size:3;not null"`
	Status          string          `gorm:"size:32;not null;default:'draft'"` // draft, submitted, filed, paid
	FiledAt         *time.Time
	FiledBy         *uint64   `gorm:"index"`
	ReferenceNumber string    `gorm:"size:100"`
	Notes           string    `gorm:"type:text"`
	ReportData      JSONMap   `gorm:"type:jsonb"` // Detailed breakdown
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`

	Filer *User `gorm:"foreignKey:FiledBy"`
}

// CronJob represents a scheduled job
type CronJob struct {
	ID           uint64    `gorm:"primaryKey"`
	Name         string    `gorm:"size:100;not null;uniqueIndex"`
	Description  string    `gorm:"type:text"`
	Schedule     string    `gorm:"size:50;not null"` // Cron expression
	Handler      string    `gorm:"size:100;not null"` // Function/command to run
	Parameters   JSONMap   `gorm:"type:jsonb"`
	Timeout      int       `gorm:"not null;default:300"` // Seconds
	Active       bool      `gorm:"not null;default:true"`
	RunOnStartup bool      `gorm:"not null;default:false"`
	LastRunAt    *time.Time
	NextRunAt    *time.Time `gorm:"index"`
	LastStatus   string     `gorm:"size:32"`
	LastDuration int        `gorm:"not null;default:0"` // Milliseconds
	FailCount    int        `gorm:"not null;default:0"`
	MaxFails     int        `gorm:"not null;default:5"` // Disable after this many consecutive fails
	CreatedAt    time.Time  `gorm:"not null"`
	UpdatedAt    time.Time  `gorm:"not null"`
}

// CronJobLog represents a cron job execution log
type CronJobLog struct {
	ID        uint64    `gorm:"primaryKey"`
	CronJobID uint64    `gorm:"not null;index"`
	StartedAt time.Time `gorm:"not null"`
	EndedAt   *time.Time
	Duration  int       `gorm:"not null;default:0"` // Milliseconds
	Status    string    `gorm:"size:32;not null"`   // running, success, failed, timeout
	Output    string    `gorm:"type:text"`
	Error     string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null;index"`

	CronJob CronJob `gorm:"foreignKey:CronJobID"`
}

// AutomationRule represents an automation/hook rule
type AutomationRule struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	Description string    `gorm:"type:text"`
	Trigger     string    `gorm:"size:100;not null;index"` // Event that triggers this rule
	Conditions  JSONMap   `gorm:"type:jsonb"` // Conditions that must be met
	Actions     JSONMap   `gorm:"type:jsonb;not null"` // Actions to perform
	Priority    int       `gorm:"not null;default:0"`
	Active      bool      `gorm:"not null;default:true"`
	LastRun     *time.Time
	RunCount    int64     `gorm:"not null;default:0"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// AutomationLog represents an automation rule execution log
type AutomationLog struct {
	ID        uint64    `gorm:"primaryKey"`
	RuleID    uint64    `gorm:"not null;index"`
	Trigger   string    `gorm:"size:100;not null"`
	TriggerData JSONMap `gorm:"type:jsonb"`
	Status    string    `gorm:"size:32;not null"` // success, failed, skipped
	Result    JSONMap   `gorm:"type:jsonb"`
	Error     string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null;index"`

	Rule AutomationRule `gorm:"foreignKey:RuleID"`
}

// SuspensionRule represents auto-suspension configuration
type SuspensionRule struct {
	ID              uint64    `gorm:"primaryKey"`
	Name            string    `gorm:"size:100;not null"`
	DaysOverdue     int       `gorm:"not null"` // Days after due date to suspend
	SendReminder    bool      `gorm:"not null;default:true"`
	ReminderDays    int       `gorm:"not null;default:0"` // Days before suspension to send reminder
	ExcludeGroups   JSONMap   `gorm:"type:jsonb"` // Customer groups to exclude
	TerminationDays int       `gorm:"not null;default:0"` // Days after suspension to terminate (0 = never)
	Active          bool      `gorm:"not null;default:true"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

// InvoiceSettings represents invoice generation settings
type InvoiceSettings struct {
	ID                   uint64    `gorm:"primaryKey"`
	DaysBeforeDue        int       `gorm:"not null;default:14"` // Generate invoice X days before due
	DueDateDays          int       `gorm:"not null;default:7"` // Invoice due X days after generation
	NumberFormat         string    `gorm:"size:100;not null;default:'INV-{YEAR}-{NUMBER}'"` 
	NextNumber           int       `gorm:"not null;default:1"`
	ResetYearly          bool      `gorm:"not null;default:true"` // Reset numbering each year
	DefaultPaymentMethod string    `gorm:"size:50"`
	FooterText           string    `gorm:"type:text"`
	TermsConditions      string    `gorm:"type:text"`
	LogoPath             string    `gorm:"size:255"`
	AutoSend             bool      `gorm:"not null;default:true"`
	AutoCharge           bool      `gorm:"not null;default:false"`
	MergeInvoices        bool      `gorm:"not null;default:true"` // Merge same-day invoices
	CreatedAt            time.Time `gorm:"not null"`
	UpdatedAt            time.Time `gorm:"not null"`
}

// ServiceAutoSettings represents automatic service management settings
type ServiceAutoSettings struct {
	ID                     uint64    `gorm:"primaryKey"`
	AutoProvision          bool      `gorm:"not null;default:true"` // Auto-provision on payment
	AutoSuspendOverdue     bool      `gorm:"not null;default:true"`
	SuspendGraceDays       int       `gorm:"not null;default:3"`
	AutoTerminate          bool      `gorm:"not null;default:false"`
	TerminateGraceDays     int       `gorm:"not null;default:30"`
	AutoUnsuspendPayment   bool      `gorm:"not null;default:true"` // Auto-unsuspend when paid
	SendSuspensionNotice   bool      `gorm:"not null;default:true"`
	SendTerminationNotice  bool      `gorm:"not null;default:true"`
	SendRenewalReminder    bool      `gorm:"not null;default:true"`
	RenewalReminderDays    int       `gorm:"not null;default:7"`
	CreatedAt              time.Time `gorm:"not null"`
	UpdatedAt              time.Time `gorm:"not null"`
}

// OrderAutoSettings represents automatic order processing settings
type OrderAutoSettings struct {
	ID                    uint64          `gorm:"primaryKey"`
	AutoActivateOnPayment bool            `gorm:"not null;default:true"`
	RequireApproval       bool            `gorm:"not null;default:false"`
	ApprovalThreshold     decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"` // Orders above this amount require approval
	FraudCheckEnabled     bool            `gorm:"not null;default:false"`
	FraudCheckModule      string          `gorm:"size:100"`
	AutoCancelUnpaidHours int             `gorm:"not null;default:0"` // Cancel unpaid orders after X hours (0 = never)
	SendConfirmationEmail bool            `gorm:"not null;default:true"`
	CreatedAt             time.Time       `gorm:"not null"`
	UpdatedAt             time.Time       `gorm:"not null"`
}

// TicketAutoSettings represents automatic ticket settings
type TicketAutoSettings struct {
	ID                  uint64    `gorm:"primaryKey"`
	AutoCloseInactive   bool      `gorm:"not null;default:true"`
	AutoCloseHours      int       `gorm:"not null;default:72"`
	AutoAssign          bool      `gorm:"not null;default:false"`
	AssignmentMethod    string    `gorm:"size:32;default:'round_robin'"` // round_robin, least_tickets, manual
	SendRatingRequest   bool      `gorm:"not null;default:true"`
	AllowReopenHours    int       `gorm:"not null;default:48"` // Allow reopen within X hours of close
	SendNewTicketAlert  bool      `gorm:"not null;default:true"`
	SendReplyAlert      bool      `gorm:"not null;default:true"`
	CreatedAt           time.Time `gorm:"not null"`
	UpdatedAt           time.Time `gorm:"not null"`
}

// DataRetentionPolicy represents data retention settings for GDPR compliance
type DataRetentionPolicy struct {
	ID                    uint64    `gorm:"primaryKey"`
	DataType              string    `gorm:"size:50;not null;uniqueIndex"` // invoices, tickets, logs, etc.
	RetentionDays         int       `gorm:"not null;default:0"` // 0 = keep forever
	AnonymizeInstead      bool      `gorm:"not null;default:false"` // Anonymize instead of delete
	RequireInactive       bool      `gorm:"not null;default:true"` // Only delete for inactive customers
	InactiveDays          int       `gorm:"not null;default:365"`
	ExcludeStatutoryData  bool      `gorm:"not null;default:true"` // Exclude data required by law
	Active                bool      `gorm:"not null;default:false"`
	LastRunAt             *time.Time
	CreatedAt             time.Time `gorm:"not null"`
	UpdatedAt             time.Time `gorm:"not null"`
}

// SystemTask represents a system background task
type SystemTask struct {
	ID          uint64    `gorm:"primaryKey"`
	Type        string    `gorm:"size:50;not null;index"` // invoice_generation, suspension, etc.
	EntityType  string    `gorm:"size:50;index"` // service, invoice, ticket, etc.
	EntityID    *uint64   `gorm:"index"`
	Status      string    `gorm:"size:32;not null;default:'pending'"` // pending, running, completed, failed
	Priority    int       `gorm:"not null;default:5"` // 1-10
	Data        JSONMap   `gorm:"type:jsonb"`
	Result      JSONMap   `gorm:"type:jsonb"`
	Error       string    `gorm:"type:text"`
	ScheduledAt time.Time `gorm:"not null;index"`
	StartedAt   *time.Time
	CompletedAt *time.Time
	Attempts    int       `gorm:"not null;default:0"`
	MaxAttempts int       `gorm:"not null;default:3"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// DiscountRule represents a general discount rule
type DiscountRule struct {
	ID            uint64          `gorm:"primaryKey"`
	Name          string          `gorm:"size:100;not null"`
	Type          string          `gorm:"size:32;not null"` // percentage, fixed, free_setup, free_domain
	Value         decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	AppliesTo     string          `gorm:"size:32;not null"` // product, category, order, customer
	TargetIDs     JSONMap         `gorm:"type:jsonb"` // Specific products/categories
	CustomerGroups JSONMap        `gorm:"type:jsonb"` // Customer groups that qualify
	MinQuantity   int             `gorm:"not null;default:0"`
	MaxQuantity   int             `gorm:"not null;default:0"` // 0 = unlimited
	MinOrderAmount decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	BillingCycles  JSONMap        `gorm:"type:jsonb"` // Applicable billing cycles
	Stackable     bool            `gorm:"not null;default:false"` // Can stack with other discounts
	Priority      int             `gorm:"not null;default:0"`
	StartsAt      *time.Time
	EndsAt        *time.Time
	Active        bool      `gorm:"not null;default:true"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}
