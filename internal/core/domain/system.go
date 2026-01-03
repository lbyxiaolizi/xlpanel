package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// ServerType represents the type of server module
type ServerType string

const (
	ServerTypeCPanel     ServerType = "cpanel"
	ServerTypePlesk      ServerType = "plesk"
	ServerTypeDirectAdmin ServerType = "directadmin"
	ServerTypeVirtualizor ServerType = "virtualizor"
	ServerTypeProxmox    ServerType = "proxmox"
	ServerTypeCustom     ServerType = "custom"
)

// ServerStatus represents the status of a server
type ServerStatus string

const (
	ServerStatusActive   ServerStatus = "active"
	ServerStatusInactive ServerStatus = "inactive"
	ServerStatusOffline  ServerStatus = "offline"
)

// Server represents a server/node for service provisioning
type Server struct {
	ID             uint64          `gorm:"primaryKey"`
	Name           string          `gorm:"size:100;not null"`
	Type           ServerType      `gorm:"size:50;not null"`
	Hostname       string          `gorm:"size:255;not null"`
	IPAddress      string          `gorm:"size:45;not null"`
	Port           int             `gorm:"not null;default:443"`
	Username       string          `gorm:"size:100"`
	Password       string          `gorm:"size:255"` // Encrypted
	AccessHash     string          `gorm:"size:500"` // API key/token
	Secure         bool            `gorm:"not null;default:true"`
	Status         ServerStatus    `gorm:"size:32;not null;default:'active'"`
	MaxAccounts    int             `gorm:"not null;default:0"` // 0 = unlimited
	CurrentAccounts int            `gorm:"not null;default:0"`
	Priority       int             `gorm:"not null;default:0"` // For load balancing
	AssignedIPs    int             `gorm:"not null;default:0"`
	Location       string          `gorm:"size:100"`
	ModuleConfig   JSONMap         `gorm:"type:jsonb"`
	MonthlyBandwidth decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"` // GB
	UsedBandwidth  decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`  // GB
	DiskSpace      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`  // GB
	UsedDiskSpace  decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`  // GB
	Memory         decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`  // GB
	UsedMemory     decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`  // GB
	LastCheck      *time.Time
	Notes          string          `gorm:"type:text"`
	CreatedAt      time.Time       `gorm:"not null"`
	UpdatedAt      time.Time       `gorm:"not null"`

	// Relations
	ServerGroup *ServerGroup `gorm:"foreignKey:ServerGroupID"`
	ServerGroupID *uint64    `gorm:"index"`
	Services    []Service    `gorm:"foreignKey:ServerID"`
}

// HasCapacity checks if the server can accept more accounts
func (s *Server) HasCapacity() bool {
	if s.MaxAccounts == 0 {
		return true
	}
	return s.CurrentAccounts < s.MaxAccounts
}

// IsOnline checks if the server is online
func (s *Server) IsOnline() bool {
	return s.Status == ServerStatusActive
}

// ServerGroup represents a group of servers
type ServerGroup struct {
	ID         uint64   `gorm:"primaryKey"`
	Name       string   `gorm:"size:100;not null"`
	FillType   string   `gorm:"size:32;not null;default:'fill'"` // fill, round-robin, least-used
	Active     bool     `gorm:"not null;default:true"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`

	Servers []Server `gorm:"foreignKey:ServerGroupID"`
}

// Setting represents a system setting
type Setting struct {
	ID        uint64    `gorm:"primaryKey"`
	Key       string    `gorm:"size:100;uniqueIndex;not null"`
	Value     string    `gorm:"type:text"`
	Type      string    `gorm:"size:32;not null;default:'string'"` // string, int, bool, json
	Group     string    `gorm:"size:50;index"`
	Label     string    `gorm:"size:200"`
	HelpText  string    `gorm:"type:text"`
	Options   JSONMap   `gorm:"type:jsonb"` // For select options
	Protected bool      `gorm:"not null;default:false"` // Hidden from non-admin
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	Type        string    `gorm:"size:50;not null;uniqueIndex:idx_template_type_lang"`
	Language    string    `gorm:"size:10;not null;default:'en';uniqueIndex:idx_template_type_lang"`
	Subject     string    `gorm:"size:500;not null"`
	BodyHTML    string    `gorm:"type:text"`
	BodyPlain   string    `gorm:"type:text"`
	Variables   JSONMap   `gorm:"type:jsonb"` // Available variables
	Active      bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// EmailLog represents a sent email log
type EmailLog struct {
	ID          uint64    `gorm:"primaryKey"`
	CustomerID  *uint64   `gorm:"index"`
	TemplateID  *uint64   `gorm:"index"`
	ToEmail     string    `gorm:"size:255;not null;index"`
	FromEmail   string    `gorm:"size:255"`
	Subject     string    `gorm:"size:500;not null"`
	Body        string    `gorm:"type:text"`
	Status      string    `gorm:"size:32;not null"` // sent, failed, queued
	ErrorMsg    string    `gorm:"type:text"`
	RelatedType string    `gorm:"size:50;index"`
	RelatedID   *uint64   `gorm:"index"`
	SentAt      *time.Time
	CreatedAt   time.Time `gorm:"not null"`

	Customer *User          `gorm:"foreignKey:CustomerID"`
	Template *EmailTemplate `gorm:"foreignKey:TemplateID"`
}

// Currency represents a supported currency
type Currency struct {
	ID           uint64          `gorm:"primaryKey"`
	Code         string          `gorm:"size:3;uniqueIndex;not null"` // ISO 4217
	Name         string          `gorm:"size:100;not null"`
	Symbol       string          `gorm:"size:10;not null"`
	SymbolPos    string          `gorm:"size:10;not null;default:'left'"` // left, right
	DecimalPlaces int            `gorm:"not null;default:2"`
	ExchangeRate decimal.Decimal `gorm:"type:numeric(20,8);not null;default:1"`
	IsDefault    bool            `gorm:"not null;default:false"`
	Active       bool            `gorm:"not null;default:true"`
	CreatedAt    time.Time       `gorm:"not null"`
	UpdatedAt    time.Time       `gorm:"not null"`
}

// FormatAmount formats an amount in this currency
func (c *Currency) FormatAmount(amount decimal.Decimal) string {
	formatted := amount.StringFixed(int32(c.DecimalPlaces))
	if c.SymbolPos == "right" {
		return formatted + c.Symbol
	}
	return c.Symbol + formatted
}

// Announcement represents a system announcement
type Announcement struct {
	ID           uint64    `gorm:"primaryKey"`
	Title        string    `gorm:"size:255;not null"`
	Body         string    `gorm:"type:text;not null"`
	Published    bool      `gorm:"not null;default:false"`
	PublishedAt  *time.Time
	Type         string    `gorm:"size:32;not null;default:'general'"` // general, maintenance, security
	Priority     int       `gorm:"not null;default:0"`
	ExpiresAt    *time.Time
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// IsActive checks if the announcement is active and visible
func (a *Announcement) IsActive() bool {
	if !a.Published {
		return false
	}
	now := time.Now()
	if a.ExpiresAt != nil && now.After(*a.ExpiresAt) {
		return false
	}
	return true
}

// PaymentGateway represents a configured payment gateway
type PaymentGateway struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	ModuleName  string    `gorm:"size:100;not null;uniqueIndex"`
	DisplayName string    `gorm:"size:100;not null"`
	Type        string    `gorm:"size:32;not null"` // card, bank, crypto, wallet
	Active      bool      `gorm:"not null;default:true"`
	Visible     bool      `gorm:"not null;default:true"`
	Config      JSONMap   `gorm:"type:jsonb"`
	SortOrder   int       `gorm:"not null;default:0"`
	TestMode    bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// CronTask represents a scheduled cron task
type CronTask struct {
	ID          uint64     `gorm:"primaryKey"`
	Name        string     `gorm:"size:100;not null"`
	TaskType    string     `gorm:"size:100;not null;uniqueIndex"`
	Schedule    string     `gorm:"size:50;not null"` // Cron expression
	Active      bool       `gorm:"not null;default:true"`
	LastRun     *time.Time
	NextRun     *time.Time
	LastStatus  string     `gorm:"size:32"`
	LastError   string     `gorm:"type:text"`
	LastRunTime int        `gorm:"not null;default:0"` // Seconds
	CreatedAt   time.Time  `gorm:"not null"`
	UpdatedAt   time.Time  `gorm:"not null"`
}

// ActivityLog represents user activity for tracking
type ActivityLog struct {
	ID          uint64    `gorm:"primaryKey"`
	UserID      *uint64   `gorm:"index"`
	Action      string    `gorm:"size:100;not null;index"`
	Description string    `gorm:"type:text"`
	IPAddress   string    `gorm:"size:45"`
	UserAgent   string    `gorm:"size:512"`
	EntityType  string    `gorm:"size:50;index"`
	EntityID    *uint64
	Metadata    JSONMap   `gorm:"type:jsonb"`
	CreatedAt   time.Time `gorm:"not null;index"`

	User *User `gorm:"foreignKey:UserID"`
}

// Notification represents a user notification
type Notification struct {
	ID         uint64    `gorm:"primaryKey"`
	UserID     uint64    `gorm:"not null;index"`
	Type       string    `gorm:"size:50;not null"`
	Title      string    `gorm:"size:255;not null"`
	Message    string    `gorm:"type:text"`
	Link       string    `gorm:"size:500"`
	Read       bool      `gorm:"not null;default:false"`
	ReadAt     *time.Time
	CreatedAt  time.Time `gorm:"not null;index"`

	User User `gorm:"foreignKey:UserID"`
}
