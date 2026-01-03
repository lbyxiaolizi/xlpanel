package notification

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
)

var (
	ErrTemplateNotFound = errors.New("email template not found")
	ErrSMTPNotConfigured = errors.New("SMTP not configured")
	ErrEmailSendFailed  = errors.New("failed to send email")
)

// Service provides notification operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new notification service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// SendEmail sends an email using a template
func (s *Service) SendEmail(templateType string, recipient string, data map[string]interface{}) error {
	// Get template
	var tmpl domain.EmailTemplate
	if err := s.db.Where("type = ? AND active = ?", templateType, true).First(&tmpl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTemplateNotFound
		}
		return err
	}

	// Get default SMTP config
	var smtp domain.SMTPConfig
	if err := s.db.Where("active = ? AND \"default\" = ?", true, true).First(&smtp).Error; err != nil {
		return ErrSMTPNotConfigured
	}

	// Parse and execute template
	subject, err := s.parseTemplate(tmpl.Subject, data)
	if err != nil {
		return fmt.Errorf("failed to parse subject: %w", err)
	}

	bodyHTML, err := s.parseTemplate(tmpl.BodyHTML, data)
	if err != nil {
		return fmt.Errorf("failed to parse HTML body: %w", err)
	}

	bodyPlain, err := s.parseTemplate(tmpl.BodyPlain, data)
	if err != nil {
		bodyPlain = "" // Plain text is optional
	}

	// Queue the email
	return s.QueueEmail(smtp.ID, recipient, "", subject, bodyHTML, bodyPlain, nil, nil)
}

// SendEmailDirect sends an email directly without using a template
func (s *Service) SendEmailDirect(to, subject, bodyHTML, bodyPlain string) error {
	var smtpConfig domain.SMTPConfig
	if err := s.db.Where("active = ? AND \"default\" = ?", true, true).First(&smtpConfig).Error; err != nil {
		return ErrSMTPNotConfigured
	}

	return s.QueueEmail(smtpConfig.ID, to, "", subject, bodyHTML, bodyPlain, nil, nil)
}

// QueueEmail adds an email to the send queue
func (s *Service) QueueEmail(smtpConfigID uint64, toEmail, toName, subject, bodyHTML, bodyPlain string, customerID *uint64, relatedID *uint64) error {
	email := &domain.EmailQueue{
		SMTPConfigID: &smtpConfigID,
		ToEmail:      toEmail,
		ToName:       toName,
		Subject:      subject,
		BodyHTML:     bodyHTML,
		BodyPlain:    bodyPlain,
		CustomerID:   customerID,
		Status:       "pending",
		Priority:     5,
		MaxAttempts:  3,
	}

	return s.db.Create(email).Error
}

// ProcessEmailQueue processes pending emails in the queue
func (s *Service) ProcessEmailQueue(batchSize int) error {
	var emails []domain.EmailQueue
	if err := s.db.Where("status = ? AND (scheduled_at IS NULL OR scheduled_at <= ?)", "pending", time.Now()).
		Order("priority ASC, created_at ASC").
		Limit(batchSize).
		Find(&emails).Error; err != nil {
		return err
	}

	for _, email := range emails {
		if err := s.sendQueuedEmail(&email); err != nil {
			// Update with error
			s.db.Model(&email).Updates(map[string]interface{}{
				"status":      "failed",
				"last_error":  err.Error(),
				"attempts":    email.Attempts + 1,
			})
		} else {
			// Mark as sent
			now := time.Now()
			s.db.Model(&email).Updates(map[string]interface{}{
				"status":  "sent",
				"sent_at": &now,
			})
		}
	}

	return nil
}

// sendQueuedEmail sends a single queued email
func (s *Service) sendQueuedEmail(email *domain.EmailQueue) error {
	var smtpConfig domain.SMTPConfig
	if email.SMTPConfigID != nil {
		if err := s.db.First(&smtpConfig, *email.SMTPConfigID).Error; err != nil {
			return err
		}
	} else {
		if err := s.db.Where("active = ? AND \"default\" = ?", true, true).First(&smtpConfig).Error; err != nil {
			return ErrSMTPNotConfigured
		}
	}

	if !smtpConfig.CanSend() {
		return errors.New("SMTP daily limit reached")
	}

	// Build message
	fromEmail := smtpConfig.FromEmail
	if email.FromEmail != "" {
		fromEmail = email.FromEmail
	}
	fromName := smtpConfig.FromName
	if email.FromName != "" {
		fromName = email.FromName
	}

	message := s.buildMIMEMessage(fromEmail, fromName, email.ToEmail, email.ToName, email.Subject, email.BodyHTML, email.BodyPlain)

	// Send email
	if err := s.sendSMTP(&smtpConfig, fromEmail, email.ToEmail, message); err != nil {
		return err
	}

	// Update SMTP sent count
	s.db.Model(&smtpConfig).Updates(map[string]interface{}{
		"sent_today": smtpConfig.SentToday + 1,
		"last_sent":  time.Now(),
	})

	// Log the email
	s.logEmail(email, &smtpConfig, "sent", "")

	return nil
}

// sendSMTP sends an email via SMTP
func (s *Service) sendSMTP(config *domain.SMTPConfig, from, to string, message []byte) error {
	var auth smtp.Auth
	if config.Username != "" && config.Password != "" {
		auth = smtp.PlainAuth("", config.Username, config.Password, config.Host)
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	switch config.Encryption {
	case "tls":
		return s.sendTLS(addr, auth, from, to, message, config)
	case "ssl":
		return s.sendSSL(addr, auth, from, to, message, config)
	default:
		return smtp.SendMail(addr, auth, from, []string{to}, message)
	}
}

// sendTLS sends email using STARTTLS
func (s *Service) sendTLS(addr string, auth smtp.Auth, from, to string, message []byte, config *domain.SMTPConfig) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()

	if err = c.StartTLS(&tls.Config{ServerName: config.Host}); err != nil {
		return err
	}

	if auth != nil {
		if err = c.Auth(auth); err != nil {
			return err
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}

	if err = c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(message)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}

// sendSSL sends email using SSL/TLS
func (s *Service) sendSSL(addr string, auth smtp.Auth, from, to string, message []byte, config *domain.SMTPConfig) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: config.Host})
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return err
	}
	defer c.Close()

	if auth != nil {
		if err = c.Auth(auth); err != nil {
			return err
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}

	if err = c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(message)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}

// buildMIMEMessage builds a MIME email message
func (s *Service) buildMIMEMessage(fromEmail, fromName, toEmail, toName, subject, bodyHTML, bodyPlain string) []byte {
	var buf bytes.Buffer

	boundary := "OPENHOST_BOUNDARY_" + time.Now().Format("20060102150405")

	// Headers
	if fromName != "" {
		buf.WriteString(fmt.Sprintf("From: %s <%s>\r\n", fromName, fromEmail))
	} else {
		buf.WriteString(fmt.Sprintf("From: %s\r\n", fromEmail))
	}

	if toName != "" {
		buf.WriteString(fmt.Sprintf("To: %s <%s>\r\n", toName, toEmail))
	} else {
		buf.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	}

	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")

	if bodyHTML != "" && bodyPlain != "" {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		buf.WriteString("\r\n")

		// Plain text part
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(bodyPlain)
		buf.WriteString("\r\n")

		// HTML part
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(bodyHTML)
		buf.WriteString("\r\n")

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if bodyHTML != "" {
		buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(bodyHTML)
	} else {
		buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(bodyPlain)
	}

	return buf.Bytes()
}

// parseTemplate parses and executes a template string
func (s *Service) parseTemplate(templateStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// logEmail logs a sent email
func (s *Service) logEmail(email *domain.EmailQueue, smtp *domain.SMTPConfig, status, errorMsg string) {
	log := &domain.EmailLog{
		CustomerID:  email.CustomerID,
		TemplateID:  email.TemplateID,
		ToEmail:     email.ToEmail,
		FromEmail:   smtp.FromEmail,
		Subject:     email.Subject,
		Body:        email.BodyHTML,
		Status:      status,
		ErrorMsg:    errorMsg,
		RelatedType: email.RelatedType,
		RelatedID:   email.RelatedID,
	}
	if status == "sent" {
		now := time.Now()
		log.SentAt = &now
	}
	s.db.Create(log)
}

// CreateWebhook creates a webhook configuration
func (s *Service) CreateWebhook(customerID *uint64, name, url, secret string, events []string) (*domain.WebhookConfig, error) {
	eventsMap := make(domain.JSONMap)
	eventsMap["events"] = events

	webhook := &domain.WebhookConfig{
		CustomerID:    customerID,
		Name:          name,
		URL:           url,
		Secret:        secret,
		Events:        eventsMap,
		Active:        true,
		VerifySSL:     true,
		Timeout:       30,
		RetryAttempts: 3,
	}

	if err := s.db.Create(webhook).Error; err != nil {
		return nil, err
	}

	return webhook, nil
}

// TriggerWebhooks triggers webhooks for an event
func (s *Service) TriggerWebhooks(eventType string, payload interface{}) error {
	var webhooks []domain.WebhookConfig
	if err := s.db.Where("active = ?", true).Find(&webhooks).Error; err != nil {
		return err
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	for _, webhook := range webhooks {
		// Check if webhook is subscribed to this event
		if events, ok := webhook.Events["events"].([]interface{}); ok {
			subscribed := false
			for _, e := range events {
				if e.(string) == eventType || e.(string) == "*" {
					subscribed = true
					break
				}
			}
			if !subscribed {
				continue
			}
		}

		// Queue webhook delivery
		go s.deliverWebhook(&webhook, eventType, payloadJSON)
	}

	return nil
}

// deliverWebhook delivers a webhook
func (s *Service) deliverWebhook(webhook *domain.WebhookConfig, eventType string, payload []byte) {
	delivery := &domain.WebhookDelivery{
		WebhookID: webhook.ID,
		EventType: eventType,
		Payload:   string(payload),
		Status:    "pending",
		Attempts:  0,
	}
	s.db.Create(delivery)

	// Try delivery with retries
	for attempt := 1; attempt <= webhook.RetryAttempts; attempt++ {
		delivery.Attempts = attempt

		req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(payload))
		if err != nil {
			delivery.Status = "failed"
			delivery.ErrorMsg = err.Error()
			s.db.Save(delivery)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-OpenHost-Event", eventType)
		req.Header.Set("X-OpenHost-Delivery", fmt.Sprintf("%d", delivery.ID))

		// Add signature if secret is set
		if webhook.Secret != "" {
			signature := s.signPayload(payload, webhook.Secret)
			req.Header.Set("X-OpenHost-Signature", signature)
		}

		client := &http.Client{
			Timeout: time.Duration(webhook.Timeout) * time.Second,
		}

		start := time.Now()
		resp, err := client.Do(req)
		responseTime := int(time.Since(start).Milliseconds())

		delivery.ResponseTime = responseTime

		if err != nil {
			delivery.Status = "failed"
			delivery.ErrorMsg = err.Error()
			s.db.Save(delivery)
			
			// Wait before retry
			time.Sleep(time.Duration(attempt*attempt) * time.Second)
			continue
		}
		defer resp.Body.Close()

		delivery.ResponseCode = resp.StatusCode

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			now := time.Now()
			delivery.Status = "success"
			delivery.DeliveredAt = &now
			s.db.Save(delivery)

			// Update webhook last triggered
			s.db.Model(webhook).Update("last_triggered", &now)
			return
		}

		delivery.Status = "failed"
		delivery.ErrorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		s.db.Save(delivery)

		// Wait before retry
		time.Sleep(time.Duration(attempt*attempt) * time.Second)
	}

	// Increment failure count
	s.db.Model(webhook).Update("failure_count", webhook.FailureCount+1)
}

// signPayload signs a payload for webhook verification
func (s *Service) signPayload(payload []byte, secret string) string {
	// Implementation would use HMAC-SHA256
	return fmt.Sprintf("sha256=%x", payload) // Simplified
}

// SendNotification sends a notification through the appropriate channel
func (s *Service) SendNotification(userID uint64, notificationType, title, message, link string) error {
	// Create in-app notification
	notification := &domain.Notification{
		UserID:  userID,
		Type:    notificationType,
		Title:   title,
		Message: message,
		Link:    link,
		Read:    false,
	}

	if err := s.db.Create(notification).Error; err != nil {
		return err
	}

	// Check notification preferences
	var prefs []domain.NotificationPreference
	s.db.Where("user_id = ? AND notification_type = ? AND enabled = ?", userID, notificationType, true).
		Find(&prefs)

	for _, pref := range prefs {
		switch pref.Channel {
		case domain.NotificationChannelEmail:
			// Get user email and send
			var user domain.User
			if err := s.db.First(&user, userID).Error; err == nil {
				s.SendEmailDirect(user.Email, title, message, "")
			}
		case domain.NotificationChannelWebhook:
			s.TriggerWebhooks("notification."+notificationType, map[string]interface{}{
				"user_id": userID,
				"title":   title,
				"message": message,
				"link":    link,
			})
		}
	}

	return nil
}

// GetUnreadNotifications gets unread notifications for a user
func (s *Service) GetUnreadNotifications(userID uint64, limit int) ([]domain.Notification, error) {
	var notifications []domain.Notification
	if err := s.db.Where("user_id = ? AND read = ?", userID, false).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

// MarkNotificationRead marks a notification as read
func (s *Service) MarkNotificationRead(notificationID, userID uint64) error {
	now := time.Now()
	return s.db.Model(&domain.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"read":    true,
			"read_at": &now,
		}).Error
}

// MarkAllNotificationsRead marks all notifications as read for a user
func (s *Service) MarkAllNotificationsRead(userID uint64) error {
	now := time.Now()
	return s.db.Model(&domain.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Updates(map[string]interface{}{
			"read":    true,
			"read_at": &now,
		}).Error
}

// CreateEmailTemplate creates an email template
func (s *Service) CreateEmailTemplate(name, templateType, language, subject, bodyHTML, bodyPlain string) (*domain.EmailTemplate, error) {
	tmpl := &domain.EmailTemplate{
		Name:      name,
		Type:      templateType,
		Language:  language,
		Subject:   subject,
		BodyHTML:  bodyHTML,
		BodyPlain: bodyPlain,
		Active:    true,
	}

	if err := s.db.Create(tmpl).Error; err != nil {
		return nil, err
	}

	return tmpl, nil
}

// GetEmailTemplates gets all email templates
func (s *Service) GetEmailTemplates(language string) ([]domain.EmailTemplate, error) {
	var templates []domain.EmailTemplate
	query := s.db.Model(&domain.EmailTemplate{})
	if language != "" {
		query = query.Where("language = ?", language)
	}
	if err := query.Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// UpdateEmailTemplate updates an email template
func (s *Service) UpdateEmailTemplate(id uint64, subject, bodyHTML, bodyPlain string, active bool) error {
	return s.db.Model(&domain.EmailTemplate{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"subject":    subject,
			"body_html":  bodyHTML,
			"body_plain": bodyPlain,
			"active":     active,
		}).Error
}

// Helper function to replace template variables
func (s *Service) ReplaceTemplateVariables(content string, variables map[string]string) string {
	result := content
	for key, value := range variables {
		result = strings.ReplaceAll(result, "{{."+key+"}}", value)
		result = strings.ReplaceAll(result, "{{ ."+key+" }}", value)
	}
	return result
}
