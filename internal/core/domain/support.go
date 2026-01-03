package domain

import (
	"time"
)

// KnowledgeBaseCategory represents a category in the knowledge base
type KnowledgeBaseCategory struct {
	ID          uint64    `gorm:"primaryKey"`
	ParentID    *uint64   `gorm:"index"`
	Name        string    `gorm:"size:100;not null"`
	Slug        string    `gorm:"size:100;uniqueIndex;not null"`
	Description string    `gorm:"type:text"`
	IconClass   string    `gorm:"size:50"`
	SortOrder   int       `gorm:"not null;default:0"`
	Active      bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	Parent   *KnowledgeBaseCategory   `gorm:"foreignKey:ParentID"`
	Children []KnowledgeBaseCategory  `gorm:"foreignKey:ParentID"`
	Articles []KnowledgeBaseArticle   `gorm:"foreignKey:CategoryID"`
}

// KnowledgeBaseArticle represents an article in the knowledge base
type KnowledgeBaseArticle struct {
	ID          uint64    `gorm:"primaryKey"`
	CategoryID  uint64    `gorm:"not null;index"`
	Title       string    `gorm:"size:255;not null"`
	Slug        string    `gorm:"size:255;uniqueIndex;not null"`
	Content     string    `gorm:"type:text;not null"`
	Excerpt     string    `gorm:"size:500"`
	AuthorID    uint64    `gorm:"not null;index"`
	Status      string    `gorm:"size:32;not null;default:'draft'"` // draft, published, archived
	ViewCount   int64     `gorm:"not null;default:0"`
	HelpfulYes  int64     `gorm:"not null;default:0"`
	HelpfulNo   int64     `gorm:"not null;default:0"`
	Featured    bool      `gorm:"not null;default:false"`
	AllowComments bool    `gorm:"not null;default:true"`
	MetaTitle   string    `gorm:"size:255"`
	MetaDescription string `gorm:"size:500"`
	Tags        JSONMap   `gorm:"type:jsonb"` // Array of tags
	PublishedAt *time.Time
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	Category KnowledgeBaseCategory    `gorm:"foreignKey:CategoryID"`
	Author   User                     `gorm:"foreignKey:AuthorID"`
	Attachments []KBArticleAttachment `gorm:"foreignKey:ArticleID"`
}

// IsPublished checks if the article is published
func (a *KnowledgeBaseArticle) IsPublished() bool {
	return a.Status == "published"
}

// KBArticleAttachment represents an attachment to a KB article
type KBArticleAttachment struct {
	ID          uint64    `gorm:"primaryKey"`
	ArticleID   uint64    `gorm:"not null;index"`
	FileName    string    `gorm:"size:255;not null"`
	FilePath    string    `gorm:"size:500;not null"`
	FileSize    int64     `gorm:"not null"`
	ContentType string    `gorm:"size:100;not null"`
	Downloads   int64     `gorm:"not null;default:0"`
	CreatedAt   time.Time `gorm:"not null"`

	Article KnowledgeBaseArticle `gorm:"foreignKey:ArticleID"`
}

// KBArticleFeedback represents feedback on a KB article
type KBArticleFeedback struct {
	ID        uint64    `gorm:"primaryKey"`
	ArticleID uint64    `gorm:"not null;index"`
	CustomerID *uint64  `gorm:"index"`
	Helpful   bool      `gorm:"not null"`
	Comment   string    `gorm:"type:text"`
	IPAddress string    `gorm:"size:45"`
	CreatedAt time.Time `gorm:"not null"`

	Article  KnowledgeBaseArticle `gorm:"foreignKey:ArticleID"`
	Customer *User                `gorm:"foreignKey:CustomerID"`
}

// KBSearchLog represents a search query in the knowledge base
type KBSearchLog struct {
	ID          uint64    `gorm:"primaryKey"`
	Query       string    `gorm:"size:255;not null;index"`
	ResultCount int       `gorm:"not null"`
	CustomerID  *uint64   `gorm:"index"`
	IPAddress   string    `gorm:"size:45"`
	CreatedAt   time.Time `gorm:"not null;index"`

	Customer *User `gorm:"foreignKey:CustomerID"`
}

// DownloadCategory represents a category for downloads
type DownloadCategory struct {
	ID          uint64    `gorm:"primaryKey"`
	ParentID    *uint64   `gorm:"index"`
	Name        string    `gorm:"size:100;not null"`
	Description string    `gorm:"type:text"`
	SortOrder   int       `gorm:"not null;default:0"`
	Active      bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	Parent    *DownloadCategory  `gorm:"foreignKey:ParentID"`
	Children  []DownloadCategory `gorm:"foreignKey:ParentID"`
	Downloads []Download         `gorm:"foreignKey:CategoryID"`
}

// Download represents a downloadable file
type Download struct {
	ID            uint64    `gorm:"primaryKey"`
	CategoryID    uint64    `gorm:"not null;index"`
	Name          string    `gorm:"size:255;not null"`
	Description   string    `gorm:"type:text"`
	Version       string    `gorm:"size:50"`
	FileName      string    `gorm:"size:255;not null"`
	FilePath      string    `gorm:"size:500;not null"`
	FileSize      int64     `gorm:"not null"`
	ContentType   string    `gorm:"size:100;not null"`
	Downloads     int64     `gorm:"not null;default:0"`
	ClientsOnly   bool      `gorm:"not null;default:false"` // Require login
	ProductIDs    JSONMap   `gorm:"type:jsonb"` // Restrict to specific products
	Active        bool      `gorm:"not null;default:true"`
	SortOrder     int       `gorm:"not null;default:0"`
	Changelog     string    `gorm:"type:text"`
	UploadedBy    uint64    `gorm:"not null"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Category   DownloadCategory `gorm:"foreignKey:CategoryID"`
	Uploader   User             `gorm:"foreignKey:UploadedBy"`
}

// DownloadLog represents a download log entry
type DownloadLog struct {
	ID         uint64    `gorm:"primaryKey"`
	DownloadID uint64    `gorm:"not null;index"`
	CustomerID *uint64   `gorm:"index"`
	IPAddress  string    `gorm:"size:45"`
	UserAgent  string    `gorm:"size:512"`
	CreatedAt  time.Time `gorm:"not null;index"`

	Download Download `gorm:"foreignKey:DownloadID"`
	Customer *User    `gorm:"foreignKey:CustomerID"`
}

// TicketDepartment represents a department for ticket routing
type TicketDepartment struct {
	ID                uint64    `gorm:"primaryKey"`
	Name              string    `gorm:"size:100;not null"`
	Description       string    `gorm:"type:text"`
	Email             string    `gorm:"size:255"` // Email pipe address
	ClientsOnly       bool      `gorm:"not null;default:true"`
	PipesEnabled      bool      `gorm:"not null;default:false"`
	AutoClose         bool      `gorm:"not null;default:false"`
	AutoCloseHours    int       `gorm:"not null;default:72"`
	SLAResponseHours  int       `gorm:"not null;default:24"`
	SLAResolveHours   int       `gorm:"not null;default:72"`
	DefaultPriority   string    `gorm:"size:32;not null;default:'normal'"`
	Hidden            bool      `gorm:"not null;default:false"`
	SortOrder         int       `gorm:"not null;default:0"`
	Active            bool      `gorm:"not null;default:true"`
	CreatedAt         time.Time `gorm:"not null"`
	UpdatedAt         time.Time `gorm:"not null"`

	Tickets []Ticket `gorm:"foreignKey:DepartmentID"`
}

// TicketPredefinedReply represents a canned reply for tickets
type TicketPredefinedReply struct {
	ID           uint64    `gorm:"primaryKey"`
	Name         string    `gorm:"size:100;not null"`
	Reply        string    `gorm:"type:text;not null"`
	CategoryID   *uint64   `gorm:"index"`
	DepartmentID *uint64   `gorm:"index"`
	SortOrder    int       `gorm:"not null;default:0"`
	Active       bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`

	Category   *TicketPredefinedReplyCategory `gorm:"foreignKey:CategoryID"`
	Department *TicketDepartment              `gorm:"foreignKey:DepartmentID"`
}

// TicketPredefinedReplyCategory represents a category for canned replies
type TicketPredefinedReplyCategory struct {
	ID        uint64    `gorm:"primaryKey"`
	Name      string    `gorm:"size:100;not null"`
	SortOrder int       `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Replies []TicketPredefinedReply `gorm:"foreignKey:CategoryID"`
}

// TicketTag represents a tag that can be applied to tickets
type TicketTag struct {
	ID        uint64    `gorm:"primaryKey"`
	Name      string    `gorm:"size:50;uniqueIndex;not null"`
	Color     string    `gorm:"size:7"` // Hex color
	CreatedAt time.Time `gorm:"not null"`
}

// TicketTagAssignment represents a tag assigned to a ticket
type TicketTagAssignment struct {
	ID        uint64    `gorm:"primaryKey"`
	TicketID  uint64    `gorm:"not null;uniqueIndex:idx_ticket_tag"`
	TagID     uint64    `gorm:"not null;uniqueIndex:idx_ticket_tag"`
	CreatedAt time.Time `gorm:"not null"`

	Ticket Ticket    `gorm:"foreignKey:TicketID"`
	Tag    TicketTag `gorm:"foreignKey:TagID"`
}

// TicketWatcher represents a user watching a ticket for updates
type TicketWatcher struct {
	ID        uint64    `gorm:"primaryKey"`
	TicketID  uint64    `gorm:"not null;uniqueIndex:idx_ticket_watcher"`
	UserID    uint64    `gorm:"not null;uniqueIndex:idx_ticket_watcher"`
	CreatedAt time.Time `gorm:"not null"`

	Ticket Ticket `gorm:"foreignKey:TicketID"`
	User   User   `gorm:"foreignKey:UserID"`
}

// TicketSLA represents SLA tracking for a ticket
type TicketSLA struct {
	ID                uint64     `gorm:"primaryKey"`
	TicketID          uint64     `gorm:"not null;uniqueIndex"`
	ResponseDue       time.Time  `gorm:"not null"`
	ResolveDue        time.Time  `gorm:"not null"`
	FirstResponseAt   *time.Time
	ResolvedAt        *time.Time
	ResponseBreached  bool       `gorm:"not null;default:false"`
	ResolveBreached   bool       `gorm:"not null;default:false"`
	PausedAt          *time.Time
	TotalPausedMins   int        `gorm:"not null;default:0"`
	CreatedAt         time.Time  `gorm:"not null"`
	UpdatedAt         time.Time  `gorm:"not null"`

	Ticket Ticket `gorm:"foreignKey:TicketID"`
}

// IsResponseBreached checks if the SLA response time was breached
func (s *TicketSLA) IsResponseBreached() bool {
	if s.FirstResponseAt == nil {
		return time.Now().After(s.ResponseDue)
	}
	return s.FirstResponseAt.After(s.ResponseDue)
}

// TicketEscalation represents a ticket escalation rule
type TicketEscalation struct {
	ID            uint64    `gorm:"primaryKey"`
	Name          string    `gorm:"size:100;not null"`
	DepartmentID  *uint64   `gorm:"index"`
	Priority      string    `gorm:"size:32"`
	TriggerMins   int       `gorm:"not null"` // Minutes without response
	Action        string    `gorm:"size:32;not null"` // notify, reassign, change_priority
	ActionValue   string    `gorm:"size:255"`
	NotifyUserIDs JSONMap   `gorm:"type:jsonb"` // Users to notify
	Active        bool      `gorm:"not null;default:true"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Department *TicketDepartment `gorm:"foreignKey:DepartmentID"`
}

// NetworkStatusPage represents network/service status information
type NetworkStatusPage struct {
	ID          uint64    `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	Description string    `gorm:"type:text"`
	Status      string    `gorm:"size:32;not null;default:'operational'"` // operational, degraded, partial_outage, major_outage, maintenance
	Active      bool      `gorm:"not null;default:true"`
	SortOrder   int       `gorm:"not null;default:0"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`

	Incidents []NetworkIncident `gorm:"foreignKey:ComponentID"`
}

// NetworkIncident represents a network/service incident
type NetworkIncident struct {
	ID            uint64    `gorm:"primaryKey"`
	ComponentID   uint64    `gorm:"not null;index"`
	Title         string    `gorm:"size:255;not null"`
	Status        string    `gorm:"size:32;not null"` // investigating, identified, monitoring, resolved
	Impact        string    `gorm:"size:32;not null"` // none, minor, major, critical
	Message       string    `gorm:"type:text"`
	ScheduledAt   *time.Time // For maintenance
	ScheduledEnd  *time.Time
	ResolvedAt    *time.Time
	CreatedBy     uint64    `gorm:"not null"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Component NetworkStatusPage        `gorm:"foreignKey:ComponentID"`
	Creator   User                     `gorm:"foreignKey:CreatedBy"`
	Updates   []NetworkIncidentUpdate  `gorm:"foreignKey:IncidentID"`
}

// NetworkIncidentUpdate represents an update to an incident
type NetworkIncidentUpdate struct {
	ID         uint64    `gorm:"primaryKey"`
	IncidentID uint64    `gorm:"not null;index"`
	Status     string    `gorm:"size:32;not null"`
	Message    string    `gorm:"type:text;not null"`
	CreatedBy  uint64    `gorm:"not null"`
	CreatedAt  time.Time `gorm:"not null"`

	Incident NetworkIncident `gorm:"foreignKey:IncidentID"`
	Creator  User            `gorm:"foreignKey:CreatedBy"`
}

// ServiceUptime represents uptime statistics for a service
type ServiceUptime struct {
	ID        uint64    `gorm:"primaryKey"`
	ServiceID uint64    `gorm:"not null;index"`
	Date      time.Time `gorm:"not null;uniqueIndex:idx_service_date"`
	UptimePct float64   `gorm:"not null;default:100"`
	Checks    int       `gorm:"not null;default:0"`
	Failures  int       `gorm:"not null;default:0"`
	AvgResponseMs int   `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Service Service `gorm:"foreignKey:ServiceID"`
}

// ServiceMonitor represents a service monitoring configuration
type ServiceMonitor struct {
	ID            uint64    `gorm:"primaryKey"`
	ServiceID     uint64    `gorm:"not null;uniqueIndex"`
	Type          string    `gorm:"size:32;not null"` // http, https, ping, port, dns
	Target        string    `gorm:"size:255;not null"` // URL, IP, or hostname
	Port          int       `gorm:"not null;default:0"`
	Interval      int       `gorm:"not null;default:300"` // Seconds
	Timeout       int       `gorm:"not null;default:30"`  // Seconds
	ExpectedCode  int       `gorm:"not null;default:200"`
	ExpectedText  string    `gorm:"size:255"`
	LastStatus    string    `gorm:"size:32"` // up, down, unknown
	LastCheck     *time.Time
	LastDowntime  *time.Time
	ConsecutiveFails int    `gorm:"not null;default:0"`
	AlertThreshold int     `gorm:"not null;default:3"` // Consecutive failures before alert
	Active        bool      `gorm:"not null;default:true"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`

	Service Service `gorm:"foreignKey:ServiceID"`
}

// MonitoringAlert represents an alert from service monitoring
type MonitoringAlert struct {
	ID          uint64    `gorm:"primaryKey"`
	MonitorID   uint64    `gorm:"not null;index"`
	ServiceID   uint64    `gorm:"not null;index"`
	Type        string    `gorm:"size:32;not null"` // down, recovered, slow
	Message     string    `gorm:"type:text"`
	AlertedAt   time.Time `gorm:"not null"`
	RecoveredAt *time.Time
	Acknowledged bool     `gorm:"not null;default:false"`
	AcknowledgedBy *uint64 `gorm:"index"`
	AcknowledgedAt *time.Time
	CreatedAt   time.Time `gorm:"not null"`

	Monitor        ServiceMonitor `gorm:"foreignKey:MonitorID"`
	Service        Service        `gorm:"foreignKey:ServiceID"`
	Acknowledger   *User          `gorm:"foreignKey:AcknowledgedBy"`
}
