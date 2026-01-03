package domain

import "time"

type Customer struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Order struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	CustomerID  string    `json:"customer_id"`
	ProductCode string    `json:"product_code"`
	Quantity    int       `json:"quantity"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Invoice struct {
	ID         string        `json:"id"`
	TenantID   string        `json:"tenant_id"`
	CustomerID string        `json:"customer_id"`
	Amount     float64       `json:"amount"`
	Currency   string        `json:"currency"`
	Status     string        `json:"status"`
	DueAt      time.Time     `json:"due_at"`
	Lines      []InvoiceLine `json:"lines"`
}

type Coupon struct {
	Code           string `json:"code"`
	PercentOff     int    `json:"percent_off"`
	Recurring      bool   `json:"recurring"`
	MaxRedemptions *int   `json:"max_redemptions,omitempty"`
}

type Ticket struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	CustomerID string    `json:"customer_id"`
	Subject    string    `json:"subject"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type IPAllocation struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	CustomerID string    `json:"customer_id"`
	IPAddress  string    `json:"ip_address"`
	AssignedAt time.Time `json:"assigned_at"`
}

type VPSInstance struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	CustomerID string    `json:"customer_id"`
	Hostname   string    `json:"hostname"`
	PlanCode   string    `json:"plan_code"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Contact   string    `json:"contact"`
	CreatedAt time.Time `json:"created_at"`
}

type Product struct {
	Code        string  `json:"code"`
	TenantID    string  `json:"tenant_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	UnitPrice   float64 `json:"unit_price"`
	Currency    string  `json:"currency"`
	Recurring   bool    `json:"recurring"`
	BillingDays int     `json:"billing_days"`
}

type Subscription struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	CustomerID  string    `json:"customer_id"`
	ProductCode string    `json:"product_code"`
	Status      string    `json:"status"`
	NextBillAt  time.Time `json:"next_bill_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type InvoiceLine struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

type Payment struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	CustomerID string    `json:"customer_id"`
	InvoiceID  string    `json:"invoice_id"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Method     string    `json:"method"`
	Status     string    `json:"status"`
	ReceivedAt time.Time `json:"received_at"`
}

// User represents a user account with authentication capabilities
type User struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"` // Never expose password hash in JSON
	Name           string    `json:"name"`
	Role           string    `json:"role"` // admin, customer
	TwoFactorSecret string   `json:"-"`     // TOTP secret, never expose
	TwoFactorEnabled bool    `json:"two_factor_enabled"`
	CreatedAt      time.Time `json:"created_at"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
}

// Session represents an authenticated user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// CartItem represents an item in a shopping cart
type CartItem struct {
	ProductCode string  `json:"product_code"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// Cart represents a shopping cart for a customer
type Cart struct {
	ID         string     `json:"id"`
	CustomerID string     `json:"customer_id"`
	TenantID   string     `json:"tenant_id"`
	Items      []CartItem `json:"items"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
