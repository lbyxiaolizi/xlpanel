package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusActive     OrderStatus = "active"
	OrderStatusFraud      OrderStatus = "fraud"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusCompleted  OrderStatus = "completed"
)

// Order represents a customer order
type Order struct {
	ID            uint64          `gorm:"primaryKey"`
	OrderNumber   string          `gorm:"size:50;uniqueIndex;not null"`
	CustomerID    uint64          `gorm:"not null;index"`
	InvoiceID     *uint64         `gorm:"index"`
	Status        OrderStatus     `gorm:"size:64;not null;default:'pending'"`
	Currency      string          `gorm:"size:3;not null;default:'USD'"`
	Subtotal      decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Discount      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	TaxAmount     decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Total         decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	CouponID      *uint64         `gorm:"index"`
	IPAddress     string          `gorm:"size:45"`
	Notes         string          `gorm:"type:text"`
	AdminNotes    string          `gorm:"type:text"`
	FraudCheck    JSONMap         `gorm:"type:jsonb"`
	CreatedAt     time.Time       `gorm:"not null"`
	UpdatedAt     time.Time       `gorm:"not null"`

	// Relations
	Customer  User        `gorm:"foreignKey:CustomerID"`
	Invoice   *Invoice    `gorm:"foreignKey:InvoiceID"`
	Coupon    *Coupon     `gorm:"foreignKey:CouponID"`
	Items     []OrderItem `gorm:"foreignKey:OrderID"`
}

// OrderItem represents a line item in an order
type OrderItem struct {
	ID            uint64          `gorm:"primaryKey"`
	OrderID       uint64          `gorm:"not null;index"`
	ProductID     uint64          `gorm:"not null;index"`
	ServiceID     *uint64         `gorm:"index"`
	Description   string          `gorm:"size:500;not null"`
	Quantity      int             `gorm:"not null;default:1"`
	BillingCycle  string          `gorm:"size:32"`
	SetupFee      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	RecurringFee  decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Discount      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Total         decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	ConfigOptions JSONMap         `gorm:"type:jsonb"`
	Domain        string          `gorm:"size:255"`
	Hostname      string          `gorm:"size:255"`
	CreatedAt     time.Time       `gorm:"not null"`
	UpdatedAt     time.Time       `gorm:"not null"`

	Order   Order    `gorm:"foreignKey:OrderID"`
	Product Product  `gorm:"foreignKey:ProductID"`
	Service *Service `gorm:"foreignKey:ServiceID"`
}

// ServiceStatus represents the status of a service
type ServiceStatus string

const (
	ServiceStatusPending    ServiceStatus = "pending"
	ServiceStatusActive     ServiceStatus = "active"
	ServiceStatusSuspended  ServiceStatus = "suspended"
	ServiceStatusTerminated ServiceStatus = "terminated"
	ServiceStatusCancelled  ServiceStatus = "cancelled"
)

// Service represents a customer's active service/product
type Service struct {
	ID                uint64          `gorm:"primaryKey"`
	CustomerID        uint64          `gorm:"not null;index"`
	ProductID         uint64          `gorm:"not null;index"`
	OrderID           *uint64         `gorm:"index"`
	ServerID          *uint64         `gorm:"index"`
	Status            ServiceStatus   `gorm:"size:64;not null;default:'pending'"`
	Domain            string          `gorm:"size:255"`
	Hostname          string          `gorm:"size:255"`
	Username          string          `gorm:"size:100"`
	Password          string          `gorm:"size:255"` // Encrypted
	BillingCycle      string          `gorm:"size:32"`
	Currency          string          `gorm:"size:3;not null;default:'USD'"`
	RecurringAmount   decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	NextDueDate       time.Time       `gorm:"not null;index"`
	RegistrationDate  time.Time       `gorm:"not null"`
	TerminationDate   *time.Time
	SuspensionReason  string          `gorm:"size:500"`
	OverrideAutoSusp  bool            `gorm:"not null;default:false"`
	OverrideAutoTerm  bool            `gorm:"not null;default:false"`
	ExternalID        string          `gorm:"size:255;index"` // ID in external system
	TimesUsed         int             `gorm:"not null;default:0"`
	IPAddressID       *uint64         `gorm:"index"`
	ConfigSelection   JSONMap         `gorm:"type:jsonb;not null"`
	PluginConfig      PluginConfig    `gorm:"type:jsonb;not null"`
	Notes             string          `gorm:"type:text"`
	AdminNotes        string          `gorm:"type:text"`
	CreatedAt         time.Time       `gorm:"not null"`
	UpdatedAt         time.Time       `gorm:"not null"`

	// Relations
	Product   Product    `gorm:"foreignKey:ProductID"`
	Customer  User       `gorm:"foreignKey:CustomerID"`
	Order     *Order     `gorm:"foreignKey:OrderID"`
	Server    *Server    `gorm:"foreignKey:ServerID"`
	IPAddress *IPAddress `gorm:"foreignKey:IPAddressID"`
}

// IsActive checks if the service is active
func (s *Service) IsActive() bool {
	return s.Status == ServiceStatusActive
}

// IsSuspended checks if the service is suspended
func (s *Service) IsSuspended() bool {
	return s.Status == ServiceStatusSuspended
}

// IsDueForRenewal checks if the service is due for renewal
func (s *Service) IsDueForRenewal() bool {
	return time.Now().After(s.NextDueDate)
}

// Cart represents a shopping cart
type Cart struct {
	ID         uint64    `gorm:"primaryKey"`
	CustomerID *uint64   `gorm:"index"`
	SessionID  string    `gorm:"size:64;index"`
	Currency   string    `gorm:"size:3;not null;default:'USD'"`
	CouponID   *uint64   `gorm:"index"`
	ExpiresAt  time.Time `gorm:"not null;index"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`

	Customer *User      `gorm:"foreignKey:CustomerID"`
	Coupon   *Coupon    `gorm:"foreignKey:CouponID"`
	Items    []CartItem `gorm:"foreignKey:CartID"`
}

// CartItem represents an item in the shopping cart
type CartItem struct {
	ID            uint64          `gorm:"primaryKey"`
	CartID        uint64          `gorm:"not null;index"`
	ProductID     uint64          `gorm:"not null;index"`
	Quantity      int             `gorm:"not null;default:1"`
	BillingCycle  string          `gorm:"size:32"`
	ConfigOptions JSONMap         `gorm:"type:jsonb"`
	Domain        string          `gorm:"size:255"`
	Hostname      string          `gorm:"size:255"`
	SetupFee      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	RecurringFee  decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	Discount      decimal.Decimal `gorm:"type:numeric(20,8);not null;default:0"`
	Total         decimal.Decimal `gorm:"type:numeric(20,8);not null"`
	CreatedAt     time.Time       `gorm:"not null"`
	UpdatedAt     time.Time       `gorm:"not null"`

	Cart    Cart    `gorm:"foreignKey:CartID"`
	Product Product `gorm:"foreignKey:ProductID"`
}

type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	payload, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (m *JSONMap) Scan(value any) error {
	if m == nil {
		return errors.New("json map scan: nil receiver")
	}
	data, err := normalizeJSONBytes(value)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		*m = JSONMap{}
		return nil
	}
	return json.Unmarshal(data, m)
}

type PluginConfig struct {
	Region string            `json:"region,omitempty"`
	NodeID string            `json:"node_id,omitempty"`
	Values map[string]string `json:"values,omitempty"`
}

func (c PluginConfig) Value() (driver.Value, error) {
	payload, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *PluginConfig) Scan(value any) error {
	if c == nil {
		return errors.New("plugin config scan: nil receiver")
	}
	data, err := normalizeJSONBytes(value)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		*c = PluginConfig{}
		return nil
	}
	return json.Unmarshal(data, c)
}

func normalizeJSONBytes(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	switch data := value.(type) {
	case []byte:
		return data, nil
	case string:
		return []byte(data), nil
	default:
		return nil, errors.New("unsupported JSON value type")
	}
}
