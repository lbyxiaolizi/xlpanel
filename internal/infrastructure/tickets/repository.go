package tickets

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AutoMigrate() error {
	if r.db == nil {
		return errors.New("db is required")
	}
	return r.db.AutoMigrate(&domain.Ticket{}, &domain.TicketMessage{}, &domain.TicketAttachment{})
}

func (r *Repository) FindTicketByID(id uint64) (domain.Ticket, error) {
	var ticket domain.Ticket
	if r.db == nil {
		return ticket, errors.New("db is required")
	}
	if err := r.db.First(&ticket, id).Error; err != nil {
		return ticket, err
	}
	return ticket, nil
}

func (r *Repository) CreateTicket(ticket *domain.Ticket) error {
	if r.db == nil {
		return errors.New("db is required")
	}
	if ticket == nil {
		return errors.New("ticket is required")
	}
	if err := r.db.Create(ticket).Error; err != nil {
		return fmt.Errorf("create ticket: %w", err)
	}
	return nil
}

func (r *Repository) CreateMessage(message *domain.TicketMessage) error {
	if r.db == nil {
		return errors.New("db is required")
	}
	if message == nil {
		return errors.New("ticket message is required")
	}
	if err := r.db.Create(message).Error; err != nil {
		return fmt.Errorf("create message: %w", err)
	}
	return nil
}
