package domain

import (
	"time"
)

// SSLProviderModule represents an SSL certificate provider module
type SSLProviderModule struct {
	ID            uint64    `gorm:"primaryKey"`
	Name          string    `gorm:"size:100;not null"`
	Slug          string    `gorm:"size:100;uniqueIndex;not null"`
	ProviderType  string    `gorm:"size:50;not null"` // comodo, digicert, letsencrypt, etc.
	Config        JSONMap   `gorm:"type:jsonb"`
	TestMode      bool      `gorm:"not null;default:false"`
	Active        bool      `gorm:"not null;default:true"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

// SSLCertificateType represents a type of SSL certificate
type SSLCertificateType struct {
	ID              uint64    `gorm:"primaryKey"`
	ProviderID      uint64    `gorm:"not null;index"`
	Name            string    `gorm:"size:255;not null"`
	Type            string    `gorm:"size:50;not null"` // dv, ov, ev, wildcard
	ValidationLevel string    `gorm:"size:32;not null"` // domain, organization, extended
	Warranty        int       `gorm:"not null;default:0"` // Warranty amount in USD
	SAN             bool      `gorm:"not null;default:false"` // Supports Subject Alternative Names
	MaxSAN          int       `gorm:"not null;default:0"`
	Wildcard        bool      `gorm:"not null;default:false"`
	IDN             bool      `gorm:"not null;default:false"` // Supports IDN
	SGC             bool      `gorm:"not null;default:false"` // Server Gated Cryptography
	IssuanceTime    string    `gorm:"size:50"` // estimated issuance time
	Description     string    `gorm:"type:text"`
	Features        JSONMap   `gorm:"type:jsonb"`
	Active          bool      `gorm:"not null;default:true"`
	SortOrder       int       `gorm:"not null;default:0"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`

	Provider SSLProviderModule `gorm:"foreignKey:ProviderID"`
}

// SSLOrder represents an SSL certificate order
type SSLOrder struct {
	ID                uint64    `gorm:"primaryKey"`
	CustomerID        uint64    `gorm:"not null;index"`
	ServiceID         *uint64   `gorm:"index"`
	CertTypeID        uint64    `gorm:"not null;index"`
	OrderID           *uint64   `gorm:"index"`
	Status            string    `gorm:"size:32;not null;default:'pending'"` // pending, processing, issued, cancelled, expired
	Domain            string    `gorm:"size:255;not null"`
	AdditionalDomains JSONMap   `gorm:"type:jsonb"` // SANs
	Years             int       `gorm:"not null;default:1"`
	CSR               string    `gorm:"type:text"` // Certificate Signing Request
	PrivateKey        string    `gorm:"type:text"` // Encrypted
	Certificate       string    `gorm:"type:text"`
	CACertificate     string    `gorm:"type:text"`
	ValidationMethod  string    `gorm:"size:32"` // email, dns, http
	ValidationEmail   string    `gorm:"size:255"`
	ValidationStatus  string    `gorm:"size:32"`
	ApproverEmail     string    `gorm:"size:255"`
	ProviderOrderID   string    `gorm:"size:255"`
	IssuedAt          *time.Time
	ExpiresAt         *time.Time
	RenewalReminder   bool      `gorm:"not null;default:true"`
	AutoRenew         bool      `gorm:"not null;default:false"`
	Notes             string    `gorm:"type:text"`
	CreatedAt         time.Time `gorm:"not null"`
	UpdatedAt         time.Time `gorm:"not null"`

	Customer User               `gorm:"foreignKey:CustomerID"`
	Service  *Service           `gorm:"foreignKey:ServiceID"`
	CertType SSLCertificateType `gorm:"foreignKey:CertTypeID"`
	Order    *Order             `gorm:"foreignKey:OrderID"`
}

// ProvisioningServerModule represents a server provisioning module
type ProvisioningServerModule struct {
	ID              uint64    `gorm:"primaryKey"`
	Name            string    `gorm:"size:100;not null"`
	Slug            string    `gorm:"size:100;uniqueIndex;not null"`
	ModuleType      string    `gorm:"size:50;not null"` // cpanel, plesk, directadmin, virtualizor, etc.
	Config          JSONMap   `gorm:"type:jsonb"`
	TestMode        bool      `gorm:"not null;default:false"`
	SupportsCreate  bool      `gorm:"not null;default:true"`
	SupportsSuspend bool      `gorm:"not null;default:true"`
	SupportsUnsuspend bool    `gorm:"not null;default:true"`
	SupportsTerminate bool    `gorm:"not null;default:true"`
	SupportsUpgrade bool      `gorm:"not null;default:false"`
	SupportsUsage   bool      `gorm:"not null;default:false"`
	SupportsSSO     bool      `gorm:"not null;default:false"`
	Active          bool      `gorm:"not null;default:true"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}

// ProvisioningServer represents a provisioning server
type ProvisioningServer struct {
	ID            uint64    `gorm:"primaryKey"`
	ModuleID      uint64    `gorm:"not null;index"`
	Name          string    `gorm:"size:100;not null"`
	Hostname      string    `gorm:"size:255;not null"`
	IPAddress     string    `gorm:"size:45"`
	Port          int       `gorm:"not null;default:0"`
	Username      string    `gorm:"size:100"`
	Password      string    `gorm:"size:255"` // Encrypted
	AccessHash    string    `gorm:"type:text"` // For WHM
	SecureAPI     bool      `gorm:"not null;default:true"`
	MaxAccounts   int       `gorm:"not null;default:0"` // 0 = unlimited
	CurrentAccounts int     `gorm:"not null;default:0"`
	Status        string    `gorm:"size:32;not null;default:'active'"` // active, inactive, full, error
	LastCheck     *time.Time
	LastError     string    `gorm:"type:text"`
	Datacenter    string    `gorm:"size:100"`
	Location      string    `gorm:"size:100"`
	AssignedIPs   JSONMap   `gorm:"type:jsonb"` // Available IPs
	NameserverOne string    `gorm:"size:255"`
	NameserverTwo string    `gorm:"size:255"`
	NameserverThree string  `gorm:"size:255"`
	NameserverFour string   `gorm:"size:255"`
	Config        JSONMap   `gorm:"type:jsonb"` // Additional config
	SortOrder     int       `gorm:"not null;default:0"`
	Active        bool      `gorm:"not null;default:true"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Module ProvisioningServerModule `gorm:"foreignKey:ModuleID"`
}

// IsAvailable checks if the server can accept new accounts
func (s *ProvisioningServer) IsAvailable() bool {
	if !s.Active || s.Status != "active" {
		return false
	}
	if s.MaxAccounts > 0 && s.CurrentAccounts >= s.MaxAccounts {
		return false
	}
	return true
}

// ServerIPAddress represents an IP address on a server
type ServerIPAddress struct {
	ID         uint64    `gorm:"primaryKey"`
	ServerID   uint64    `gorm:"not null;index"`
	IPAddress  string    `gorm:"size:45;not null;uniqueIndex"`
	IPv6       bool      `gorm:"not null;default:false"`
	Type       string    `gorm:"size:32;not null;default:'shared'"` // dedicated, shared
	ServiceID  *uint64   `gorm:"index"` // Assigned to a service
	Gateway    string    `gorm:"size:45"`
	Subnet     string    `gorm:"size:45"`
	Notes      string    `gorm:"type:text"`
	Active     bool      `gorm:"not null;default:true"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`

	Server  ProvisioningServer `gorm:"foreignKey:ServerID"`
	Service *Service           `gorm:"foreignKey:ServiceID"`
}

// IsAvailable checks if the IP is available for assignment
func (ip *ServerIPAddress) IsAvailable() bool {
	return ip.Active && ip.ServiceID == nil
}

// ProvisioningServerGroupMember represents a provisioning server in a group
type ProvisioningServerGroupMember struct {
	ID            uint64    `gorm:"primaryKey"`
	GroupID       uint64    `gorm:"not null;uniqueIndex:idx_prov_group_server"`
	ServerID      uint64    `gorm:"not null;uniqueIndex:idx_prov_group_server"`
	Weight        int       `gorm:"not null;default:1"` // For weighted assignment
	Active        bool      `gorm:"not null;default:true"`
	CreatedAt     time.Time `gorm:"not null"`

	Group  ServerGroup        `gorm:"foreignKey:GroupID"`
	Server ProvisioningServer `gorm:"foreignKey:ServerID"`
}

// ServiceProvisioningData represents additional provisioning data for a service
type ServiceProvisioningData struct {
	ID           uint64    `gorm:"primaryKey"`
	ServiceID    uint64    `gorm:"not null;uniqueIndex"`
	ServerID     uint64    `gorm:"not null;index"`
	Username     string    `gorm:"size:100"`
	Password     string    `gorm:"size:255"` // Encrypted
	Domain       string    `gorm:"size:255"`
	Package      string    `gorm:"size:100"` // Server-side package name
	DiskLimit    int64     `gorm:"not null;default:0"` // MB
	BandwidthLimit int64   `gorm:"not null;default:0"` // MB
	DiskUsage    int64     `gorm:"not null;default:0"` // MB
	BandwidthUsage int64   `gorm:"not null;default:0"` // MB
	LastUsageSync *time.Time
	SSHPort      int       `gorm:"not null;default:22"`
	HomeDir      string    `gorm:"size:255"`
	IPAddress    string    `gorm:"size:45"`
	ControlPanel string    `gorm:"size:255"` // URL to control panel
	CustomData   JSONMap   `gorm:"type:jsonb"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`

	Service Service            `gorm:"foreignKey:ServiceID"`
	Server  ProvisioningServer `gorm:"foreignKey:ServerID"`
}

// ProvisioningLog represents a provisioning action log
type ProvisioningLog struct {
	ID          uint64    `gorm:"primaryKey"`
	ServiceID   uint64    `gorm:"not null;index"`
	ServerID    uint64    `gorm:"not null;index"`
	Action      string    `gorm:"size:50;not null"` // create, suspend, unsuspend, terminate, upgrade
	Status      string    `gorm:"size:32;not null"` // success, failed, pending
	Request     string    `gorm:"type:text"`
	Response    string    `gorm:"type:text"`
	ErrorMsg    string    `gorm:"type:text"`
	Duration    int       `gorm:"not null;default:0"` // Milliseconds
	TriggeredBy *uint64   `gorm:"index"` // Admin/system
	CreatedAt   time.Time `gorm:"not null;index"`

	Service Service            `gorm:"foreignKey:ServiceID"`
	Server  ProvisioningServer `gorm:"foreignKey:ServerID"`
	Admin   *User              `gorm:"foreignKey:TriggeredBy"`
}

// ResellersConfig represents reseller account configuration
type ResellersConfig struct {
	ID                uint64    `gorm:"primaryKey"`
	CustomerID        uint64    `gorm:"not null;uniqueIndex"`
	Enabled           bool      `gorm:"not null;default:false"`
	MaxServices       int       `gorm:"not null;default:0"` // 0 = unlimited
	MaxClients        int       `gorm:"not null;default:0"`
	MaxDiskSpace      int64     `gorm:"not null;default:0"` // MB
	MaxBandwidth      int64     `gorm:"not null;default:0"` // MB
	DiscountPercent   int       `gorm:"not null;default:0"`
	AllowedProducts   JSONMap   `gorm:"type:jsonb"` // Products they can resell
	BrandingEnabled   bool      `gorm:"not null;default:false"`
	CustomDomain      string    `gorm:"size:255"`
	LogoURL           string    `gorm:"size:500"`
	CompanyName       string    `gorm:"size:255"`
	SupportEmail      string    `gorm:"size:255"`
	CreatedAt         time.Time `gorm:"not null"`
	UpdatedAt         time.Time `gorm:"not null"`

	Customer User `gorm:"foreignKey:CustomerID"`
}
