package ticket

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
)

var (
	ErrTicketNotFound    = errors.New("ticket not found")
	ErrTicketClosed      = errors.New("ticket is closed")
	ErrMessageNotFound   = errors.New("message not found")
	ErrUnauthorized      = errors.New("not authorized to access this ticket")
)

// Service provides ticket management operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new ticket service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateTicket creates a new support ticket
func (s *Service) CreateTicket(customerID *uint64, subject, body, senderEmail string, priority domain.TicketPriority, source string) (*domain.Ticket, error) {
	if priority == "" {
		priority = domain.TicketPriorityNormal
	}
	if source == "" {
		source = "web"
	}

	ticket := &domain.Ticket{
		CustomerID: customerID,
		Subject:    subject,
		Status:     domain.TicketStatusOpen,
		Priority:   priority,
		Source:     source,
	}

	if err := s.db.Create(ticket).Error; err != nil {
		return nil, err
	}

	// Create initial message
	message := &domain.TicketMessage{
		TicketID:    ticket.ID,
		SenderEmail: senderEmail,
		Body:        body,
		IsStaff:     false,
	}

	if err := s.db.Create(message).Error; err != nil {
		return nil, err
	}

	ticket.Messages = append(ticket.Messages, *message)
	return ticket, nil
}

// GetTicket retrieves a ticket by ID
func (s *Service) GetTicket(id uint64) (*domain.Ticket, error) {
	var ticket domain.Ticket
	if err := s.db.Preload("Messages.Attachments").First(&ticket, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	return &ticket, nil
}

// GetTicketForCustomer retrieves a ticket ensuring customer ownership
func (s *Service) GetTicketForCustomer(ticketID, customerID uint64) (*domain.Ticket, error) {
	var ticket domain.Ticket
	if err := s.db.Preload("Messages.Attachments").
		Where("id = ? AND customer_id = ?", ticketID, customerID).
		First(&ticket).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}
	return &ticket, nil
}

// ListTickets returns tickets with optional filters
func (s *Service) ListTickets(customerID *uint64, status domain.TicketStatus, limit, offset int) ([]domain.Ticket, int64, error) {
	var tickets []domain.Ticket
	var total int64

	query := s.db.Model(&domain.Ticket{})
	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	if err := query.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// AddReply adds a reply to a ticket
func (s *Service) AddReply(ticketID uint64, senderEmail, body string, isStaff bool, attachments []AttachmentData) (*domain.TicketMessage, error) {
	var ticket domain.Ticket
	if err := s.db.First(&ticket, ticketID).Error; err != nil {
		return nil, ErrTicketNotFound
	}

	if ticket.Status == domain.TicketStatusClosed && !isStaff {
		return nil, ErrTicketClosed
	}

	message := &domain.TicketMessage{
		TicketID:    ticketID,
		SenderEmail: senderEmail,
		Body:        body,
		IsStaff:     isStaff,
	}

	if err := s.db.Create(message).Error; err != nil {
		return nil, err
	}

	// Add attachments
	for _, att := range attachments {
		attachment := &domain.TicketAttachment{
			TicketMessageID: message.ID,
			FileName:        att.FileName,
			ContentType:     att.ContentType,
			SizeBytes:       int64(len(att.Data)),
			Data:            att.Data,
		}
		if err := s.db.Create(attachment).Error; err != nil {
			return nil, err
		}
		message.Attachments = append(message.Attachments, *attachment)
	}

	// Update ticket status if reply from staff
	if isStaff && ticket.Status == domain.TicketStatusOpen {
		s.db.Model(&ticket).Updates(map[string]interface{}{
			"status":     domain.TicketStatusOpen,
			"updated_at": time.Now(),
		})
	}

	// Reopen ticket if customer replies to closed ticket (staff-initiated)
	if !isStaff && ticket.Status == domain.TicketStatusClosed {
		s.db.Model(&ticket).Update("status", domain.TicketStatusOpen)
	}

	return message, nil
}

// UpdateTicketStatus updates the status of a ticket
func (s *Service) UpdateTicketStatus(ticketID uint64, status domain.TicketStatus) error {
	return s.db.Model(&domain.Ticket{}).Where("id = ?", ticketID).
		Update("status", status).Error
}

// UpdateTicketPriority updates the priority of a ticket
func (s *Service) UpdateTicketPriority(ticketID uint64, priority domain.TicketPriority) error {
	return s.db.Model(&domain.Ticket{}).Where("id = ?", ticketID).
		Update("priority", priority).Error
}

// CloseTicket closes a ticket
func (s *Service) CloseTicket(ticketID uint64) error {
	return s.UpdateTicketStatus(ticketID, domain.TicketStatusClosed)
}

// ReopenTicket reopens a closed ticket
func (s *Service) ReopenTicket(ticketID uint64) error {
	return s.UpdateTicketStatus(ticketID, domain.TicketStatusOpen)
}

// PutTicketOnHold puts a ticket on hold
func (s *Service) PutTicketOnHold(ticketID uint64) error {
	return s.UpdateTicketStatus(ticketID, domain.TicketStatusOnHold)
}

// DeleteTicket deletes a ticket and all its messages
func (s *Service) DeleteTicket(ticketID uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete attachments
		var messages []domain.TicketMessage
		if err := tx.Where("ticket_id = ?", ticketID).Find(&messages).Error; err != nil {
			return err
		}
		for _, msg := range messages {
			if err := tx.Delete(&domain.TicketAttachment{}, "ticket_message_id = ?", msg.ID).Error; err != nil {
				return err
			}
		}

		// Delete messages
		if err := tx.Delete(&domain.TicketMessage{}, "ticket_id = ?", ticketID).Error; err != nil {
			return err
		}

		// Delete ticket
		return tx.Delete(&domain.Ticket{}, ticketID).Error
	})
}

// GetAttachment retrieves an attachment by ID
func (s *Service) GetAttachment(attachmentID uint64) (*domain.TicketAttachment, error) {
	var attachment domain.TicketAttachment
	if err := s.db.First(&attachment, attachmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("attachment not found")
		}
		return nil, err
	}
	return &attachment, nil
}

// GetTicketStats returns ticket statistics
func (s *Service) GetTicketStats() (*TicketStats, error) {
	stats := &TicketStats{}

	// Count by status
	s.db.Model(&domain.Ticket{}).Where("status = ?", domain.TicketStatusOpen).Count(&stats.Open)
	s.db.Model(&domain.Ticket{}).Where("status = ?", domain.TicketStatusClosed).Count(&stats.Closed)
	s.db.Model(&domain.Ticket{}).Where("status = ?", domain.TicketStatusOnHold).Count(&stats.OnHold)

	// Count by priority for open tickets
	s.db.Model(&domain.Ticket{}).Where("status = ? AND priority = ?", domain.TicketStatusOpen, domain.TicketPriorityHigh).Count(&stats.HighPriority)

	// Unresponded (open tickets without staff reply)
	subquery := s.db.Model(&domain.TicketMessage{}).
		Select("ticket_id").
		Where("is_staff = true").
		Group("ticket_id")
	s.db.Model(&domain.Ticket{}).
		Where("status = ?", domain.TicketStatusOpen).
		Where("id NOT IN (?)", subquery).
		Count(&stats.Unresponded)

	return stats, nil
}

// GetCustomerTicketStats returns ticket statistics for a specific customer
func (s *Service) GetCustomerTicketStats(customerID uint64) (*CustomerTicketStats, error) {
	stats := &CustomerTicketStats{}

	s.db.Model(&domain.Ticket{}).Where("customer_id = ?", customerID).Count(&stats.Total)
	s.db.Model(&domain.Ticket{}).Where("customer_id = ? AND status = ?", customerID, domain.TicketStatusOpen).Count(&stats.Open)

	return stats, nil
}

// SearchTickets searches tickets by subject or content
func (s *Service) SearchTickets(query string, customerID *uint64, limit, offset int) ([]domain.Ticket, int64, error) {
	var tickets []domain.Ticket
	var total int64

	searchQuery := s.db.Model(&domain.Ticket{}).
		Where("subject ILIKE ?", "%"+query+"%")

	if customerID != nil {
		searchQuery = searchQuery.Where("customer_id = ?", *customerID)
	}

	searchQuery.Count(&total)

	if err := searchQuery.Order("updated_at DESC").
		Limit(limit).Offset(offset).Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// AssignTicketToCustomer assigns a ticket to a customer (for email-created tickets)
func (s *Service) AssignTicketToCustomer(ticketID, customerID uint64) error {
	return s.db.Model(&domain.Ticket{}).Where("id = ?", ticketID).
		Update("customer_id", customerID).Error
}

// AttachmentData represents attachment data for creating attachments
type AttachmentData struct {
	FileName    string
	ContentType string
	Data        []byte
}

// TicketStats represents ticket statistics
type TicketStats struct {
	Open         int64 `json:"open"`
	Closed       int64 `json:"closed"`
	OnHold       int64 `json:"on_hold"`
	HighPriority int64 `json:"high_priority"`
	Unresponded  int64 `json:"unresponded"`
}

// CustomerTicketStats represents ticket statistics for a customer
type CustomerTicketStats struct {
	Total int64 `json:"total"`
	Open  int64 `json:"open"`
}
