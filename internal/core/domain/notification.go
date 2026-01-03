package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// EmailTemplateType represents the type of email template
type EmailTemplateType string

const (
	EmailTypeWelcome          EmailTemplateType = "welcome"
	EmailTypePasswordReset    EmailTemplateType = "password_reset"
	EmailTypeEmailVerify      EmailTemplateType = "email_verification"
	EmailTypeInvoiceCreated   EmailTemplateType = "invoice_created"
	EmailTypeInvoicePaid      EmailTemplateType = "invoice_paid"
	EmailTypePaymentReceipt   EmailTemplateType = "payment_receipt"
	EmailTypePaymentFailed    EmailTemplateType = "payment_failed"
	EmailTypePaymentReminder  EmailTemplateType = "payment_reminder"
	EmailTypeOverdueNotice    EmailTemplateType = "overdue_notice"
	EmailTypeServiceActivated EmailTemplateType = "service_activated"
	EmailTypeServiceSuspended EmailTemplateType = "service_suspended"
	EmailTypeServiceRenewal   EmailTemplateType = "service_renewal"
	EmailTypeServiceExpiring  EmailTemplateType = "service_expiring"
	EmailTypeTicketOpened     EmailTemplateType = "ticket_opened"
	EmailTypeTicketReply      EmailTemplateType = "ticket_reply"
	EmailTypeTicketClosed     EmailTemplateType = "ticket_closed"
	EmailTypeOrderConfirm     EmailTemplateType = "order_confirmation"
	EmailTypeQuoteSent        EmailTemplateType = "quote_sent"
	EmailTypeAffiliateApproved EmailTemplateType = "affiliate_approved"
	EmailTypeAffiliateCommission EmailTemplateType = "affiliate_commission"
	EmailTypeDomainExpiring   EmailTemplateType = "domain_expiring"
	EmailTypeDomainRenewed    EmailTemplateType = "domain_renewed"
	EmailTypeNewsletter       EmailTemplateType = "newsletter"
	EmailTypeAnnouncement     EmailTemplateType = "announcement"
	EmailTypeCustom           EmailTemplateType = "custom"
)

// EmailTemplateVariables defines available variables for templates
type EmailTemplateVariables struct {
	CustomerName       bool `json:"customer_name"`
	CustomerEmail      bool `json:"customer_email"`
	CustomerCompany    bool `json:"customer_company"`
	InvoiceNumber      bool `json:"invoice_number"`
	InvoiceTotal       bool `json:"invoice_total"`
	InvoiceDueDate     bool `json:"invoice_due_date"`
	InvoiceLink        bool `json:"invoice_link"`
	ServiceName        bool `json:"service_name"`
	ServiceDueDate     bool `json:"service_due_date"`
	TicketID           bool `json:"ticket_id"`
	TicketSubject      bool `json:"ticket_subject"`
	TicketReply        bool `json:"ticket_reply"`
	OrderNumber        bool `json:"order_number"`
	DomainName         bool `json:"domain_name"`
	PasswordResetLink  bool `json:"password_reset_link"`
	VerificationLink   bool `json:"verification_link"`
	CompanyName        bool `json:"company_name"`
	SupportEmail       bool `json:"support_email"`
	SupportURL         bool `json:"support_url"`
}

// Value implements driver.Valuer
func (v EmailTemplateVariables) Value() (driver.Value, error) {
	return json.Marshal(v)
}

// Scan implements sql.Scanner
func (v *EmailTemplateVariables) Scan(value interface{}) error {
	if value == nil {
		*v = EmailTemplateVariables{}
		return nil
	}
	data, err := normalizeJSONBytes(value)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		*v = EmailTemplateVariables{}
		return nil
	}
	return json.Unmarshal(data, v)
}

// SMTPConfig represents SMTP server configuration
type SMTPConfig struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	Host        string    `gorm:"size:255;not null"`
	Port        int       `gorm:"not null;default:587"`
	Username    string    `gorm:"size:255"`
	Password    string    `gorm:"size:255"` // Encrypted
	Encryption  string    `gorm:"size:10;not null;default:'tls'"` // none, ssl, tls
	FromEmail   string    `gorm:"size:255;not null"`
	FromName    string    `gorm:"size:100;not null"`
	ReplyTo     string    `gorm:"size:255"`
	Default     bool      `gorm:"not null;default:false"`
	Active      bool      `gorm:"not null;default:true"`
	DailyLimit  int       `gorm:"not null;default:0"` // 0 = unlimited
	SentToday   int       `gorm:"not null;default:0"`
	LastSent    *time.Time
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// CanSend checks if the SMTP config can send emails
func (s *SMTPConfig) CanSend() bool {
	if !s.Active {
		return false
	}
	if s.DailyLimit > 0 && s.SentToday >= s.DailyLimit {
		return false
	}
	return true
}

// EmailQueue represents a queued email
type EmailQueue struct {
	ID           uint64    `gorm:"primaryKey"`
	TemplateID   *uint64   `gorm:"index"`
	SMTPConfigID *uint64   `gorm:"index"`
	ToEmail      string    `gorm:"size:255;not null"`
	ToName       string    `gorm:"size:100"`
	FromEmail    string    `gorm:"size:255"`
	FromName     string    `gorm:"size:100"`
	ReplyTo      string    `gorm:"size:255"`
	Subject      string    `gorm:"size:500;not null"`
	BodyHTML     string    `gorm:"type:text"`
	BodyPlain    string    `gorm:"type:text"`
	CC           string    `gorm:"size:500"`
	BCC          string    `gorm:"size:500"`
	Headers      JSONMap   `gorm:"type:jsonb"`
	Attachments  JSONMap   `gorm:"type:jsonb"` // File paths
	Priority     int       `gorm:"not null;default:5"` // 1-10, lower is higher
	Status       string    `gorm:"size:32;not null;default:'pending'"` // pending, sending, sent, failed
	Attempts     int       `gorm:"not null;default:0"`
	MaxAttempts  int       `gorm:"not null;default:3"`
	LastError    string    `gorm:"type:text"`
	ScheduledAt  *time.Time
	SentAt       *time.Time
	RelatedType  string    `gorm:"size:50;index"` // invoice, ticket, order, etc.
	RelatedID    *uint64   `gorm:"index"`
	CustomerID   *uint64   `gorm:"index"`
	CreatedAt    time.Time `gorm:"not null;index"`
	UpdatedAt    time.Time `gorm:"not null"`

	Template   *EmailTemplate `gorm:"foreignKey:TemplateID"`
	SMTPConfig *SMTPConfig    `gorm:"foreignKey:SMTPConfigID"`
	Customer   *User          `gorm:"foreignKey:CustomerID"`
}

// CanRetry checks if the email can be retried
func (e *EmailQueue) CanRetry() bool {
	return e.Status == "failed" && e.Attempts < e.MaxAttempts
}

// NotificationChannel represents a notification channel type
type NotificationChannel string

const (
	NotificationChannelEmail   NotificationChannel = "email"
	NotificationChannelSMS     NotificationChannel = "sms"
	NotificationChannelWebhook NotificationChannel = "webhook"
	NotificationChannelSlack   NotificationChannel = "slack"
	NotificationChannelInApp   NotificationChannel = "in_app"
)

// NotificationPreference represents user notification preferences
type NotificationPreference struct {
	ID               uint64              `gorm:"primaryKey"`
	UserID           uint64              `gorm:"not null;uniqueIndex:idx_user_notification"`
	NotificationType string              `gorm:"size:50;not null;uniqueIndex:idx_user_notification"`
	Channel          NotificationChannel `gorm:"size:32;not null;uniqueIndex:idx_user_notification"`
	Enabled          bool                `gorm:"not null;default:true"`
	CreatedAt        time.Time           `gorm:"not null"`
	UpdatedAt        time.Time           `gorm:"not null"`

	User User `gorm:"foreignKey:UserID"`
}

// SMSConfig represents SMS provider configuration
type SMSConfig struct {
	ID          uint64    `gorm:"primaryKey"`
	Provider    string    `gorm:"size:50;not null"` // twilio, nexmo, etc.
	AccountSID  string    `gorm:"size:255"`
	AuthToken   string    `gorm:"size:255"` // Encrypted
	FromNumber  string    `gorm:"size:20"`
	APIKey      string    `gorm:"size:255"` // Encrypted
	APISecret   string    `gorm:"size:255"` // Encrypted
	Config      JSONMap   `gorm:"type:jsonb"`
	Active      bool      `gorm:"not null;default:true"`
	Default     bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// SMSMessage represents an SMS message
type SMSMessage struct {
	ID          uint64    `gorm:"primaryKey"`
	ConfigID    uint64    `gorm:"not null;index"`
	ToNumber    string    `gorm:"size:20;not null"`
	FromNumber  string    `gorm:"size:20"`
	Message     string    `gorm:"size:1600;not null"`
	Status      string    `gorm:"size:32;not null;default:'pending'"` // pending, sent, delivered, failed
	ProviderID  string    `gorm:"size:100"` // Message ID from provider
	ErrorCode   string    `gorm:"size:50"`
	ErrorMsg    string    `gorm:"size:500"`
	Segments    int       `gorm:"not null;default:1"`
	Cost        string    `gorm:"size:20"`
	CustomerID  *uint64   `gorm:"index"`
	RelatedType string    `gorm:"size:50;index"`
	RelatedID   *uint64   `gorm:"index"`
	SentAt      *time.Time
	CreatedAt   time.Time `gorm:"not null;index"`
	UpdatedAt   time.Time `gorm:"not null"`

	Config   SMSConfig `gorm:"foreignKey:ConfigID"`
	Customer *User     `gorm:"foreignKey:CustomerID"`
}

// WebhookConfig represents a webhook configuration
type WebhookConfig struct {
	ID            uint64    `gorm:"primaryKey"`
	CustomerID    *uint64   `gorm:"index"` // null = system webhook
	Name          string    `gorm:"size:100;not null"`
	URL           string    `gorm:"size:500;not null"`
	Secret        string    `gorm:"size:100"` // For signature verification
	Events        JSONMap   `gorm:"type:jsonb;not null"` // Array of event types
	Headers       JSONMap   `gorm:"type:jsonb"` // Custom headers
	Active        bool      `gorm:"not null;default:true"`
	VerifySSL     bool      `gorm:"not null;default:true"`
	Timeout       int       `gorm:"not null;default:30"` // Seconds
	RetryAttempts int       `gorm:"not null;default:3"`
	LastTriggered *time.Time
	FailureCount  int       `gorm:"not null;default:0"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Customer *User `gorm:"foreignKey:CustomerID"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID           uint64    `gorm:"primaryKey"`
	WebhookID    uint64    `gorm:"not null;index"`
	EventType    string    `gorm:"size:100;not null;index"`
	Payload      string    `gorm:"type:text;not null"`
	RequestHeaders JSONMap `gorm:"type:jsonb"`
	ResponseCode int       `gorm:"not null;default:0"`
	ResponseBody string    `gorm:"type:text"`
	ResponseTime int       `gorm:"not null;default:0"` // Milliseconds
	Status       string    `gorm:"size:32;not null"` // pending, success, failed
	ErrorMsg     string    `gorm:"type:text"`
	Attempts     int       `gorm:"not null;default:1"`
	NextRetryAt  *time.Time
	DeliveredAt  *time.Time
	CreatedAt    time.Time `gorm:"not null;index"`

	Webhook WebhookConfig `gorm:"foreignKey:WebhookID"`
}

// IsSuccess checks if the delivery was successful
func (w *WebhookDelivery) IsSuccess() bool {
	return w.ResponseCode >= 200 && w.ResponseCode < 300
}

// SlackConfig represents Slack integration configuration
type SlackConfig struct {
	ID           uint64    `gorm:"primaryKey"`
	WorkspaceID  string    `gorm:"size:100"`
	WorkspaceName string   `gorm:"size:100"`
	WebhookURL   string    `gorm:"size:500"`
	BotToken     string    `gorm:"size:255"` // Encrypted
	ChannelID    string    `gorm:"size:100"`
	ChannelName  string    `gorm:"size:100"`
	Events       JSONMap   `gorm:"type:jsonb"` // Events to send to Slack
	Active       bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// AdminNotificationSetting represents admin notification settings
type AdminNotificationSetting struct {
	ID               uint64    `gorm:"primaryKey"`
	AdminID          uint64    `gorm:"not null;index"`
	NotificationType string    `gorm:"size:50;not null;uniqueIndex:idx_admin_notification"`
	EmailEnabled     bool      `gorm:"not null;default:true"`
	SlackEnabled     bool      `gorm:"not null;default:false"`
	SMSEnabled       bool      `gorm:"not null;default:false"`
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`

	Admin User `gorm:"foreignKey:AdminID"`
}

// NotificationEvent represents a notification event to be processed
type NotificationEvent struct {
	ID          uint64    `gorm:"primaryKey"`
	EventType   string    `gorm:"size:100;not null;index"`
	Payload     JSONMap   `gorm:"type:jsonb;not null"`
	CustomerID  *uint64   `gorm:"index"`
	RelatedType string    `gorm:"size:50;index"`
	RelatedID   *uint64   `gorm:"index"`
	Status      string    `gorm:"size:32;not null;default:'pending'"` // pending, processed, failed
	ProcessedAt *time.Time
	ErrorMsg    string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null;index"`

	Customer *User `gorm:"foreignKey:CustomerID"`
}

// NewsletterSubscription represents a newsletter subscription
type NewsletterSubscription struct {
	ID            uint64    `gorm:"primaryKey"`
	Email         string    `gorm:"size:255;uniqueIndex;not null"`
	CustomerID    *uint64   `gorm:"index"`
	FirstName     string    `gorm:"size:100"`
	LastName      string    `gorm:"size:100"`
	Status        string    `gorm:"size:32;not null;default:'subscribed'"` // subscribed, unsubscribed, bounced
	Source        string    `gorm:"size:50"` // website, checkout, import
	IPAddress     string    `gorm:"size:45"`
	ConfirmedAt   *time.Time
	UnsubscribedAt *time.Time
	UnsubscribeReason string `gorm:"type:text"`
	Tags          JSONMap   `gorm:"type:jsonb"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Customer *User `gorm:"foreignKey:CustomerID"`
}

// Newsletter represents a newsletter campaign
type Newsletter struct {
	ID            uint64    `gorm:"primaryKey"`
	Subject       string    `gorm:"size:255;not null"`
	BodyHTML      string    `gorm:"type:text;not null"`
	BodyPlain     string    `gorm:"type:text"`
	FromEmail     string    `gorm:"size:255"`
	FromName      string    `gorm:"size:100"`
	Status        string    `gorm:"size:32;not null;default:'draft'"` // draft, scheduled, sending, sent
	TargetGroups  JSONMap   `gorm:"type:jsonb"` // Customer groups to target
	TotalRecipients int     `gorm:"not null;default:0"`
	SentCount     int       `gorm:"not null;default:0"`
	OpenCount     int       `gorm:"not null;default:0"`
	ClickCount    int       `gorm:"not null;default:0"`
	ScheduledAt   *time.Time
	SentAt        *time.Time
	CompletedAt   *time.Time
	CreatedBy     uint64    `gorm:"not null"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Creator User `gorm:"foreignKey:CreatedBy"`
}

// NewsletterRecipient represents a recipient of a newsletter
type NewsletterRecipient struct {
	ID           uint64    `gorm:"primaryKey"`
	NewsletterID uint64    `gorm:"not null;index"`
	Email        string    `gorm:"size:255;not null"`
	Status       string    `gorm:"size:32;not null;default:'pending'"` // pending, sent, opened, clicked, bounced
	SentAt       *time.Time
	OpenedAt     *time.Time
	ClickedAt    *time.Time
	BouncedAt    *time.Time
	CreatedAt    time.Time `gorm:"not null"`

	Newsletter Newsletter `gorm:"foreignKey:NewsletterID"`
}
