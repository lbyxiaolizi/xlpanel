package domain

import "time"

type TicketStatus string

type TicketPriority string

const (
	TicketStatusOpen   TicketStatus = "open"
	TicketStatusClosed TicketStatus = "closed"
	TicketStatusOnHold TicketStatus = "on_hold"
)

const (
	TicketPriorityLow    TicketPriority = "low"
	TicketPriorityNormal TicketPriority = "normal"
	TicketPriorityHigh   TicketPriority = "high"
)

type Ticket struct {
	ID         uint64          `gorm:"primaryKey"`
	CustomerID *uint64         `gorm:"index"`
	Subject    string          `gorm:"size:255;not null"`
	Status     TicketStatus    `gorm:"size:32;not null;index"`
	Priority   TicketPriority  `gorm:"size:32;not null"`
	Source     string          `gorm:"size:32;not null"`
	Messages   []TicketMessage `gorm:"foreignKey:TicketID"`
	CreatedAt  time.Time       `gorm:"not null"`
	UpdatedAt  time.Time       `gorm:"not null"`
}

type TicketMessage struct {
	ID          uint64             `gorm:"primaryKey"`
	TicketID    uint64             `gorm:"not null;index"`
	SenderEmail string             `gorm:"size:255;not null"`
	Body        string             `gorm:"type:text;not null"`
	IsStaff     bool               `gorm:"not null;default:false"`
	Attachments []TicketAttachment `gorm:"foreignKey:TicketMessageID"`
	CreatedAt   time.Time          `gorm:"not null"`
	UpdatedAt   time.Time          `gorm:"not null"`
}

type TicketAttachment struct {
	ID              uint64    `gorm:"primaryKey"`
	TicketMessageID uint64    `gorm:"not null;index"`
	FileName        string    `gorm:"size:255;not null"`
	ContentType     string    `gorm:"size:128;not null"`
	SizeBytes       int64     `gorm:"not null"`
	Data            []byte    `gorm:"type:bytea;not null"`
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}
