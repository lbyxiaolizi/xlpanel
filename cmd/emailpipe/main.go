package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jhillyerd/enmime"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
	"github.com/openhost/openhost/internal/infrastructure/tickets"
)

var ticketIDRegex = regexp.MustCompile(`\[Ticket #(\d+)\]`)

func main() {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_DSN"))
	if dsn == "" {
		log.Fatal("DATABASE_DSN is required")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	repo := tickets.NewRepository(db)
	if os.Getenv("EMAILPIPE_AUTO_MIGRATE") == "true" {
		if err := repo.AutoMigrate(); err != nil {
			log.Fatalf("auto migrate: %v", err)
		}
	}

	if err := processEmail(os.Stdin, repo); err != nil {
		log.Fatalf("process email: %v", err)
	}
}

func processEmail(reader io.Reader, repo *tickets.Repository) error {
	envelope, err := enmime.ReadEnvelope(reader)
	if err != nil {
		return fmt.Errorf("parse email: %w", err)
	}

	subject := strings.TrimSpace(envelope.GetHeader("Subject"))
	if subject == "" {
		subject = "(no subject)"
	}

	ticketID, err := extractTicketID(subject)
	if err != nil {
		return err
	}

	body := strings.TrimSpace(envelope.Text)
	if body == "" {
		body = strings.TrimSpace(envelope.HTML)
	}
	if body == "" {
		body = "(no content)"
	}

	sender := strings.TrimSpace(envelope.GetHeader("From"))
	if sender == "" {
		sender = "unknown"
	}

	var ticket domain.Ticket
	if ticketID != nil {
		ticket, err = repo.FindTicketByID(*ticketID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ticket, err = createTicket(repo, subject)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("load ticket: %w", err)
			}
		}
	} else {
		ticket, err = createTicket(repo, subject)
		if err != nil {
			return err
		}
	}

	message := domain.TicketMessage{
		TicketID:    ticket.ID,
		SenderEmail: sender,
		Body:        body,
		IsStaff:     false,
	}

	for _, attachment := range envelope.Attachments {
		data := attachment.Content
		message.Attachments = append(message.Attachments, domain.TicketAttachment{
			FileName:    attachment.FileName,
			ContentType: attachment.ContentType,
			SizeBytes:   int64(len(data)),
			Data:        data,
		})
	}

	if err := repo.CreateMessage(&message); err != nil {
		return err
	}

	return nil
}

func extractTicketID(subject string) (*uint64, error) {
	matches := ticketIDRegex.FindStringSubmatch(subject)
	if len(matches) < 2 {
		return nil, nil
	}
	id, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse ticket id: %w", err)
	}
	return &id, nil
}

func createTicket(repo *tickets.Repository, subject string) (domain.Ticket, error) {
	ticket := domain.Ticket{
		Subject:  subject,
		Status:   domain.TicketStatusOpen,
		Priority: domain.TicketPriorityNormal,
		Source:   "email",
	}
	if err := repo.CreateTicket(&ticket); err != nil {
		return domain.Ticket{}, err
	}
	return ticket, nil
}
