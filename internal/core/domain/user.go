package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// UserRole defines the role of a user in the system
type UserRole string

const (
	UserRoleCustomer UserRole = "customer"
	UserRoleAdmin    UserRole = "admin"
	UserRoleStaff    UserRole = "staff"
)

// UserStatus defines the status of a user account
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending   UserStatus = "pending"
)

// User represents a system user (customer, admin, or staff)
type User struct {
	ID            uint64          `gorm:"primaryKey"`
	Email         string          `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash  string          `gorm:"size:255;not null"`
	FirstName     string          `gorm:"size:100;not null"`
	LastName      string          `gorm:"size:100;not null"`
	Company       string          `gorm:"size:255"`
	Role          UserRole        `gorm:"size:32;not null;default:'customer'"`
	Status        UserStatus      `gorm:"size:32;not null;default:'active'"`
	Phone         string          `gorm:"size:32"`
	Address1      string          `gorm:"size:255"`
	Address2      string          `gorm:"size:255"`
	City          string          `gorm:"size:100"`
	State         string          `gorm:"size:100"`
	PostalCode    string          `gorm:"size:20"`
	Country       string          `gorm:"size:2"` // ISO 3166-1 alpha-2
	Language      string          `gorm:"size:10;default:'en'"`
	Currency      string          `gorm:"size:3;default:'USD'"` // ISO 4217
	TaxID         string          `gorm:"size:50"`
	Credit        decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	TwoFactorAuth bool            `gorm:"not null;default:false"`
	TwoFactorKey  string          `gorm:"size:64"`
	LastLoginAt   *time.Time
	LastLoginIP   string    `gorm:"size:45"`
	EmailVerified bool      `gorm:"not null;default:false"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	// Relations
	Services       []Service       `gorm:"foreignKey:CustomerID"`
	Orders         []Order         `gorm:"foreignKey:CustomerID"`
	Invoices       []Invoice       `gorm:"foreignKey:CustomerID"`
	Tickets        []Ticket        `gorm:"foreignKey:CustomerID"`
	PaymentMethods []PaymentMethod `gorm:"foreignKey:CustomerID"`
	Transactions   []Transaction   `gorm:"foreignKey:CustomerID"`
}

// FullName returns the user's full name
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Email
	}
	if u.LastName == "" {
		return u.FirstName
	}
	if u.FirstName == "" {
		return u.LastName
	}
	return u.FirstName + " " + u.LastName
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// IsStaff checks if the user has staff or admin role
func (u *User) IsStaff() bool {
	return u.Role == UserRoleStaff || u.Role == UserRoleAdmin
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// Session represents a user session
type Session struct {
	ID        string    `gorm:"primaryKey;size:64"`
	UserID    uint64    `gorm:"not null;index"`
	UserAgent string    `gorm:"size:512"`
	IPAddress string    `gorm:"size:45"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	User User `gorm:"foreignKey:UserID"`
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        uint64    `gorm:"primaryKey"`
	UserID    uint64    `gorm:"not null;index"`
	Token     string    `gorm:"size:64;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time `gorm:"not null"`

	User User `gorm:"foreignKey:UserID"`
}

// IsValid checks if the token is still valid
func (t *PasswordResetToken) IsValid() bool {
	return t.UsedAt == nil && time.Now().Before(t.ExpiresAt)
}

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	ID        uint64    `gorm:"primaryKey"`
	UserID    uint64    `gorm:"not null;index"`
	Token     string    `gorm:"size:64;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time `gorm:"not null"`

	User User `gorm:"foreignKey:UserID"`
}

// IsValid checks if the token is still valid
func (t *EmailVerificationToken) IsValid() bool {
	return t.UsedAt == nil && time.Now().Before(t.ExpiresAt)
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          uint64    `gorm:"primaryKey"`
	UserID      *uint64   `gorm:"index"`
	Action      string    `gorm:"size:100;not null;index"`
	EntityType  string    `gorm:"size:50;index"`
	EntityID    *uint64   `gorm:"index"`
	OldValues   JSONMap   `gorm:"type:jsonb"`
	NewValues   JSONMap   `gorm:"type:jsonb"`
	IPAddress   string    `gorm:"size:45"`
	UserAgent   string    `gorm:"size:512"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null;index"`

	User *User `gorm:"foreignKey:UserID"`
}

// AdminNote represents a staff note on a customer account
type AdminNote struct {
	ID         uint64    `gorm:"primaryKey"`
	CustomerID uint64    `gorm:"not null;index"`
	StaffID    uint64    `gorm:"not null;index"`
	Note       string    `gorm:"type:text;not null"`
	Sticky     bool      `gorm:"not null;default:false"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`

	Customer User `gorm:"foreignKey:CustomerID"`
	Staff    User `gorm:"foreignKey:StaffID"`
}

// UserPreferences stores user-specific preferences
type UserPreferences struct {
	EmailNotifications  bool   `json:"email_notifications"`
	NewsletterOptIn     bool   `json:"newsletter_opt_in"`
	TwoFactorEnabled    bool   `json:"two_factor_enabled"`
	DefaultPaymentID    uint64 `json:"default_payment_id,omitempty"`
	DarkMode            bool   `json:"dark_mode"`
	Language            string `json:"language"`
	Timezone            string `json:"timezone"`
	DateFormat          string `json:"date_format"`
	InvoiceEmailEnabled bool   `json:"invoice_email_enabled"`
}

// Value implements driver.Valuer for UserPreferences
func (p UserPreferences) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner for UserPreferences
func (p *UserPreferences) Scan(value interface{}) error {
	if value == nil {
		*p = UserPreferences{}
		return nil
	}
	data, err := normalizeJSONBytes(value)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		*p = UserPreferences{}
		return nil
	}
	return json.Unmarshal(data, p)
}

// ContactEmail represents an additional contact email
type ContactEmail struct {
	ID         uint64    `gorm:"primaryKey"`
	UserID     uint64    `gorm:"not null;index"`
	Email      string    `gorm:"size:255;not null"`
	Label      string    `gorm:"size:100"`
	IsPrimary  bool      `gorm:"not null;default:false"`
	ReceiveAll bool      `gorm:"not null;default:false"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`

	User User `gorm:"foreignKey:UserID"`
}

// LoginAttempt tracks login attempts for security
type LoginAttempt struct {
	ID        uint64    `gorm:"primaryKey"`
	Email     string    `gorm:"size:255;not null;index"`
	IPAddress string    `gorm:"size:45;not null;index"`
	Success   bool      `gorm:"not null"`
	UserAgent string    `gorm:"size:512"`
	FailReason string   `gorm:"size:100"`
	CreatedAt time.Time `gorm:"not null;index"`
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID          uint64     `gorm:"primaryKey"`
	UserID      uint64     `gorm:"not null;index"`
	Name        string     `gorm:"size:100;not null"`
	KeyHash     string     `gorm:"size:64;uniqueIndex;not null"`
	Permissions JSONMap    `gorm:"type:jsonb"`
	LastUsedAt  *time.Time
	ExpiresAt   *time.Time
	Active      bool       `gorm:"not null;default:true"`
	CreatedAt   time.Time  `gorm:"not null"`
	UpdatedAt   time.Time  `gorm:"not null"`

	User User `gorm:"foreignKey:UserID"`
}

// IsValid checks if the API key is still valid
func (k *APIKey) IsValid() bool {
	if !k.Active {
		return false
	}
	if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
		return false
	}
	return true
}

// StaffPermissions defines staff member permissions
type StaffPermissions struct {
	ManageProducts   bool `json:"manage_products"`
	ManageOrders     bool `json:"manage_orders"`
	ManageInvoices   bool `json:"manage_invoices"`
	ManageCustomers  bool `json:"manage_customers"`
	ManageTickets    bool `json:"manage_tickets"`
	ManageSettings   bool `json:"manage_settings"`
	ManageStaff      bool `json:"manage_staff"`
	ViewReports      bool `json:"view_reports"`
	ManagePlugins    bool `json:"manage_plugins"`
	ManageServers    bool `json:"manage_servers"`
	RefundInvoices   bool `json:"refund_invoices"`
	SuspendServices  bool `json:"suspend_services"`
	TerminateServices bool `json:"terminate_services"`
}

// Value implements driver.Valuer for StaffPermissions
func (p StaffPermissions) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner for StaffPermissions
func (p *StaffPermissions) Scan(value interface{}) error {
	if value == nil {
		*p = StaffPermissions{}
		return nil
	}
	data, ok := value.([]byte)
	if !ok {
		return errors.New("invalid staff permissions type")
	}
	return json.Unmarshal(data, p)
}
