package domain

import "time"

type Customer struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Order struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	ProductCode string    `json:"product_code"`
	Quantity    int       `json:"quantity"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Invoice struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	DueAt      time.Time `json:"due_at"`
}

type Coupon struct {
	Code           string `json:"code"`
	PercentOff     int    `json:"percent_off"`
	Recurring      bool   `json:"recurring"`
	MaxRedemptions *int   `json:"max_redemptions,omitempty"`
}

type Ticket struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Subject    string    `json:"subject"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type IPAllocation struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	IPAddress  string    `json:"ip_address"`
	AssignedAt time.Time `json:"assigned_at"`
}

type VPSInstance struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Hostname   string    `json:"hostname"`
	PlanCode   string    `json:"plan_code"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
