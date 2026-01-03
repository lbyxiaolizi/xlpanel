package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// SubUserRole defines the role of a sub-user
type SubUserRole string

const (
	SubUserRoleBilling   SubUserRole = "billing"
	SubUserRoleTechnical SubUserRole = "technical"
	SubUserRoleAdmin     SubUserRole = "admin"
)

// SubUserPermissions defines what a sub-user can access
type SubUserPermissions struct {
	ViewServices     bool `json:"view_services"`
	ManageServices   bool `json:"manage_services"`
	ViewInvoices     bool `json:"view_invoices"`
	PayInvoices      bool `json:"pay_invoices"`
	ViewTickets      bool `json:"view_tickets"`
	CreateTickets    bool `json:"create_tickets"`
	ReplyTickets     bool `json:"reply_tickets"`
	ViewProfile      bool `json:"view_profile"`
	EditProfile      bool `json:"edit_profile"`
	ManagePayments   bool `json:"manage_payments"`
	PlaceOrders      bool `json:"place_orders"`
	ManageSubUsers   bool `json:"manage_sub_users"`
	ViewAffiliates   bool `json:"view_affiliates"`
	ManageDomains    bool `json:"manage_domains"`
	AccessAPI        bool `json:"access_api"`
}

// Value implements driver.Valuer for SubUserPermissions
func (p SubUserPermissions) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner for SubUserPermissions
func (p *SubUserPermissions) Scan(value interface{}) error {
	if value == nil {
		*p = SubUserPermissions{}
		return nil
	}
	data, err := normalizeJSONBytes(value)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		*p = SubUserPermissions{}
		return nil
	}
	return json.Unmarshal(data, p)
}

// SubUser represents a sub-user/contact under a main customer account
type SubUser struct {
	ID            uint64             `gorm:"primaryKey"`
	CustomerID    uint64             `gorm:"not null;index"`
	Email         string             `gorm:"size:255;not null;uniqueIndex"`
	PasswordHash  string             `gorm:"size:255;not null"`
	FirstName     string             `gorm:"size:100;not null"`
	LastName      string             `gorm:"size:100;not null"`
	Phone         string             `gorm:"size:32"`
	Role          SubUserRole        `gorm:"size:32;not null;default:'technical'"`
	Permissions   SubUserPermissions `gorm:"type:jsonb;not null"`
	Active        bool               `gorm:"not null;default:true"`
	TwoFactorAuth bool               `gorm:"not null;default:false"`
	TwoFactorKey  string             `gorm:"size:64"`
	LastLoginAt   *time.Time
	LastLoginIP   string    `gorm:"size:45"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Customer User `gorm:"foreignKey:CustomerID"`
}

// FullName returns the sub-user's full name
func (s *SubUser) FullName() string {
	if s.FirstName == "" && s.LastName == "" {
		return s.Email
	}
	return s.FirstName + " " + s.LastName
}

// IsActive checks if the sub-user is active
func (s *SubUser) IsActive() bool {
	return s.Active
}

// SubUserSession represents a sub-user session
type SubUserSession struct {
	ID        string    `gorm:"primaryKey;size:64"`
	SubUserID uint64    `gorm:"not null;index"`
	UserAgent string    `gorm:"size:512"`
	IPAddress string    `gorm:"size:45"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	SubUser SubUser `gorm:"foreignKey:SubUserID"`
}

// IsExpired checks if the session has expired
func (s *SubUserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SubUserInvite represents an invitation for a new sub-user
type SubUserInvite struct {
	ID          uint64             `gorm:"primaryKey"`
	CustomerID  uint64             `gorm:"not null;index"`
	Email       string             `gorm:"size:255;not null"`
	Token       string             `gorm:"size:64;uniqueIndex;not null"`
	Role        SubUserRole        `gorm:"size:32;not null"`
	Permissions SubUserPermissions `gorm:"type:jsonb;not null"`
	InvitedBy   uint64             `gorm:"not null"`
	ExpiresAt   time.Time          `gorm:"not null"`
	UsedAt      *time.Time
	CreatedAt   time.Time `gorm:"not null"`

	Customer User `gorm:"foreignKey:CustomerID"`
}

// IsValid checks if the invite is still valid
func (i *SubUserInvite) IsValid() bool {
	return i.UsedAt == nil && time.Now().Before(i.ExpiresAt)
}

// CustomerGroup represents a group of customers for targeted actions
type CustomerGroup struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null;uniqueIndex"`
	Description string    `gorm:"type:text"`
	Color       string    `gorm:"size:7"` // Hex color
	Discount    int       `gorm:"not null;default:0"` // Percentage discount
	SortOrder   int       `gorm:"not null;default:0"`
	Active      bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// CustomerGroupMembership represents customer membership in groups
type CustomerGroupMembership struct {
	ID              uint64    `gorm:"primaryKey"`
	CustomerID      uint64    `gorm:"not null;uniqueIndex:idx_customer_group"`
	CustomerGroupID uint64    `gorm:"not null;uniqueIndex:idx_customer_group"`
	CreatedAt       time.Time `gorm:"not null"`

	Customer      User          `gorm:"foreignKey:CustomerID"`
	CustomerGroup CustomerGroup `gorm:"foreignKey:CustomerGroupID"`
}

// CustomerRiskProfile represents a customer's risk assessment
type CustomerRiskProfile struct {
	ID             uint64    `gorm:"primaryKey"`
	CustomerID     uint64    `gorm:"not null;uniqueIndex"`
	RiskScore      int       `gorm:"not null;default:0"` // 0-100
	RiskLevel      string    `gorm:"size:32;not null;default:'low'"` // low, medium, high
	FraudFlag      bool      `gorm:"not null;default:false"`
	Notes          string    `gorm:"type:text"`
	ChargebackCount int      `gorm:"not null;default:0"`
	DisputeCount   int       `gorm:"not null;default:0"`
	FailedPayments int       `gorm:"not null;default:0"`
	ReviewRequired bool      `gorm:"not null;default:false"`
	LastReviewedAt *time.Time
	LastReviewedBy *uint64
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`

	Customer User  `gorm:"foreignKey:CustomerID"`
	Reviewer *User `gorm:"foreignKey:LastReviewedBy"`
}

// UpdateRiskScore recalculates the risk score based on history
func (r *CustomerRiskProfile) UpdateRiskScore() {
	score := 0
	score += r.ChargebackCount * 30
	score += r.DisputeCount * 15
	score += r.FailedPayments * 5
	if r.FraudFlag {
		score += 50
	}
	if score > 100 {
		score = 100
	}
	r.RiskScore = score
	
	switch {
	case score >= 70:
		r.RiskLevel = "high"
	case score >= 40:
		r.RiskLevel = "medium"
	default:
		r.RiskLevel = "low"
	}
}

// GDPRRequest represents a GDPR data request
type GDPRRequest struct {
	ID          uint64    `gorm:"primaryKey"`
	CustomerID  uint64    `gorm:"not null;index"`
	Type        string    `gorm:"size:32;not null"` // export, delete
	Status      string    `gorm:"size:32;not null;default:'pending'"` // pending, processing, completed, rejected
	RequestIP   string    `gorm:"size:45"`
	ProcessedBy *uint64
	ProcessedAt *time.Time
	DownloadURL string    `gorm:"size:500"`
	ExpiresAt   *time.Time
	Notes       string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	Customer    User  `gorm:"foreignKey:CustomerID"`
	ProcessedUser *User `gorm:"foreignKey:ProcessedBy"`
}

// TwoFactorBackupCode represents a backup code for 2FA
type TwoFactorBackupCode struct {
	ID         uint64    `gorm:"primaryKey"`
	UserID     uint64    `gorm:"not null;index"`
	UserType   string    `gorm:"size:32;not null"` // user, subuser
	CodeHash   string    `gorm:"size:64;not null"`
	UsedAt     *time.Time
	CreatedAt  time.Time `gorm:"not null"`
}

// SecurityQuestion represents a security question for account recovery
type SecurityQuestion struct {
	ID           uint64    `gorm:"primaryKey"`
	UserID       uint64    `gorm:"not null;index"`
	Question     string    `gorm:"size:255;not null"`
	AnswerHash   string    `gorm:"size:64;not null"`
	SortOrder    int       `gorm:"not null;default:0"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`

	User User `gorm:"foreignKey:UserID"`
}

// AccountMergeRequest represents a request to merge customer accounts
type AccountMergeRequest struct {
	ID              uint64    `gorm:"primaryKey"`
	SourceCustomerID uint64   `gorm:"not null;index"`
	TargetCustomerID uint64   `gorm:"not null;index"`
	RequestedBy     uint64    `gorm:"not null"`
	Status          string    `gorm:"size:32;not null;default:'pending'"` // pending, approved, rejected, completed
	ApprovedBy      *uint64
	ApprovedAt      *time.Time
	CompletedAt     *time.Time
	Notes           string    `gorm:"type:text"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`

	SourceCustomer User  `gorm:"foreignKey:SourceCustomerID"`
	TargetCustomer User  `gorm:"foreignKey:TargetCustomerID"`
	Requester      User  `gorm:"foreignKey:RequestedBy"`
	Approver       *User `gorm:"foreignKey:ApprovedBy"`
}

// ContactType represents a type of contact for a customer
type ContactType struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null;uniqueIndex"`
	Description string    `gorm:"type:text"`
	Default     bool      `gorm:"not null;default:false"`
	SortOrder   int       `gorm:"not null;default:0"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// CustomerContact represents an additional contact for a customer
type CustomerContact struct {
	ID            uint64    `gorm:"primaryKey"`
	CustomerID    uint64    `gorm:"not null;index"`
	ContactTypeID uint64    `gorm:"not null;index"`
	FirstName     string    `gorm:"size:100;not null"`
	LastName      string    `gorm:"size:100;not null"`
	Email         string    `gorm:"size:255;not null"`
	Phone         string    `gorm:"size:32"`
	Notes         string    `gorm:"type:text"`
	IsPrimary     bool      `gorm:"not null;default:false"`
	ReceiveCopy   bool      `gorm:"not null;default:false"` // Receive copy of all emails
	Active        bool      `gorm:"not null;default:true"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Customer    User        `gorm:"foreignKey:CustomerID"`
	ContactType ContactType `gorm:"foreignKey:ContactTypeID"`
}

// LoginHistory represents a login history entry
type LoginHistory struct {
	ID        uint64    `gorm:"primaryKey"`
	UserID    uint64    `gorm:"not null;index"`
	UserType  string    `gorm:"size:32;not null"` // user, subuser, admin
	IPAddress string    `gorm:"size:45;not null"`
	UserAgent string    `gorm:"size:512"`
	Location  string    `gorm:"size:255"`
	Success   bool      `gorm:"not null"`
	FailReason string   `gorm:"size:100"`
	CreatedAt time.Time `gorm:"not null;index"`
}

// SessionManager handles session cleanup and validation
type SessionManager interface {
	ValidateSession(sessionID string) (*User, error)
	CreateSession(userID uint64, ipAddress, userAgent string) (*Session, error)
	DestroySession(sessionID string) error
	CleanupExpiredSessions() error
}

// AccountStatus represents detailed account status information
type AccountStatus struct {
	CustomerID       uint64          `json:"customer_id"`
	Status           UserStatus      `json:"status"`
	ServicesActive   int             `json:"services_active"`
	ServicesSuspended int            `json:"services_suspended"`
	UnpaidInvoices   int             `json:"unpaid_invoices"`
	OverdueInvoices  int             `json:"overdue_invoices"`
	TotalOwed        string          `json:"total_owed"`
	CreditBalance    string          `json:"credit_balance"`
	RiskLevel        string          `json:"risk_level"`
	TwoFactorEnabled bool            `json:"two_factor_enabled"`
	EmailVerified    bool            `json:"email_verified"`
}
