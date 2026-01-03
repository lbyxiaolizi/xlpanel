package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/openhost/openhost/internal/core/service/notification"
)

// NotificationHandler handles notification API endpoints
type NotificationHandler struct {
	service *notification.Service
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(service *notification.Service) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// GetUnreadNotifications gets unread notifications for the current user
// @Summary Get unread notifications
// @Description Get a list of unread notifications for the current user
// @Tags Notifications
// @Produce json
// @Param limit query int false "Limit results"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/notifications [get]
func (h *NotificationHandler) GetUnreadNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	notifications, err := h.service.GetUnreadNotifications(userID.(uint64), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// MarkAsRead marks a notification as read
// @Summary Mark as read
// @Description Mark a notification as read
// @Tags Notifications
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/notifications/{id}/read [post]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	if err := h.service.MarkNotificationRead(notificationID, userID.(uint64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllAsRead marks all notifications as read
// @Summary Mark all as read
// @Description Mark all notifications as read
// @Tags Notifications
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/notifications/read-all [post]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.service.MarkAllNotificationsRead(userID.(uint64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// Admin handlers

// AdminSendNotification sends a notification to a user
// @Summary Admin: Send notification
// @Description Send a notification to a user (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Param request body AdminSendNotificationRequest true "Notification request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/notifications/send [post]
func (h *NotificationHandler) AdminSendNotification(c *gin.Context) {
	var req AdminSendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SendNotification(
		req.UserID,
		req.Type,
		req.Title,
		req.Message,
		req.Link,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification sent"})
}

// AdminListEmailTemplates lists email templates
// @Summary Admin: List email templates
// @Description Get a list of email templates (admin only)
// @Tags Admin Notifications
// @Produce json
// @Param language query string false "Filter by language"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/email-templates [get]
func (h *NotificationHandler) AdminListEmailTemplates(c *gin.Context) {
	language := c.Query("language")

	templates, err := h.service.GetEmailTemplates(language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// AdminCreateEmailTemplate creates an email template
// @Summary Admin: Create email template
// @Description Create an email template (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Param request body CreateEmailTemplateRequest true "Template request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/email-templates [post]
func (h *NotificationHandler) AdminCreateEmailTemplate(c *gin.Context) {
	var req CreateEmailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.service.CreateEmailTemplate(
		req.Name,
		req.Type,
		req.Language,
		req.Subject,
		req.BodyHTML,
		req.BodyPlain,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Template created",
		"template": template,
	})
}

// AdminUpdateEmailTemplate updates an email template
// @Summary Admin: Update email template
// @Description Update an email template (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Param id path int true "Template ID"
// @Param request body UpdateEmailTemplateRequest true "Template request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/email-templates/{id} [put]
func (h *NotificationHandler) AdminUpdateEmailTemplate(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	var req UpdateEmailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateEmailTemplate(templateID, req.Subject, req.BodyHTML, req.BodyPlain, req.Active); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template updated"})
}

// AdminTestEmail sends a test email
// @Summary Admin: Send test email
// @Description Send a test email to a specific address (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Param request body TestEmailRequest true "Test email request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/email-templates/test [post]
func (h *NotificationHandler) AdminTestEmail(c *gin.Context) {
	var req TestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SendEmailDirect(req.To, req.Subject, req.BodyHTML, req.BodyPlain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test email sent"})
}

// AdminCreateWebhook creates a webhook
// @Summary Admin: Create webhook
// @Description Create a webhook configuration (admin only)
// @Tags Admin Notifications
// @Accept json
// @Produce json
// @Param request body CreateWebhookRequest true "Webhook request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/webhooks [post]
func (h *NotificationHandler) AdminCreateWebhook(c *gin.Context) {
	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	webhook, err := h.service.CreateWebhook(
		req.CustomerID,
		req.Name,
		req.URL,
		req.Secret,
		req.Events,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook created",
		"webhook": webhook,
	})
}

// Request/Response types
type AdminSendNotificationRequest struct {
	UserID  uint64 `json:"user_id" binding:"required"`
	Type    string `json:"type" binding:"required"`
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`
	Link    string `json:"link"`
}

type CreateEmailTemplateRequest struct {
	Name      string `json:"name" binding:"required"`
	Type      string `json:"type" binding:"required"`
	Language  string `json:"language" binding:"required"`
	Subject   string `json:"subject" binding:"required"`
	BodyHTML  string `json:"body_html" binding:"required"`
	BodyPlain string `json:"body_plain"`
}

type UpdateEmailTemplateRequest struct {
	Subject   string `json:"subject" binding:"required"`
	BodyHTML  string `json:"body_html" binding:"required"`
	BodyPlain string `json:"body_plain"`
	Active    bool   `json:"active"`
}

type TestEmailRequest struct {
	To        string `json:"to" binding:"required,email"`
	Subject   string `json:"subject" binding:"required"`
	BodyHTML  string `json:"body_html" binding:"required"`
	BodyPlain string `json:"body_plain"`
}

type CreateWebhookRequest struct {
	CustomerID *uint64  `json:"customer_id"`
	Name       string   `json:"name" binding:"required"`
	URL        string   `json:"url" binding:"required,url"`
	Secret     string   `json:"secret"`
	Events     []string `json:"events" binding:"required"`
}
