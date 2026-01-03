package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/openhost/openhost/internal/core/domain"
	ticketSvc "github.com/openhost/openhost/internal/core/service/ticket"
)

// TicketHandler handles ticket API endpoints
type TicketHandler struct {
	ticketService *ticketSvc.Service
}

// NewTicketHandler creates a new ticket handler
func NewTicketHandler(ticketService *ticketSvc.Service) *TicketHandler {
	return &TicketHandler{ticketService: ticketService}
}

// ListTickets godoc
// @Summary List tickets
// @Description Returns the current user's tickets
// @Tags tickets
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (open, closed, on_hold)"
// @Param limit query int false "Number of results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} PaginatedResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/tickets [get]
func (h *TicketHandler) ListTickets(c *gin.Context) {
	userID := GetCurrentUserID(c)
	limit, offset := PaginationParams(c)
	status := domain.TicketStatus(c.Query("status"))

	tickets, total, err := h.ticketService.ListTickets(&userID, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch tickets"})
		return
	}

	var response []TicketResponse
	for _, t := range tickets {
		response = append(response, toTicketResponse(&t))
	}

	c.JSON(http.StatusOK, NewPaginatedResponse(response, total, limit, offset))
}

// GetTicket godoc
// @Summary Get ticket details
// @Description Returns details of a specific ticket with all messages
// @Tags tickets
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Success 200 {object} TicketDetailResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/tickets/{id} [get]
func (h *TicketHandler) GetTicket(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ticket ID"})
		return
	}

	user := GetCurrentUser(c)

	var ticket *domain.Ticket
	if user.IsStaff() {
		ticket, err = h.ticketService.GetTicket(ticketID)
	} else {
		ticket, err = h.ticketService.GetTicketForCustomer(ticketID, user.ID)
	}

	if err != nil {
		if err == ticketSvc.ErrTicketNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Ticket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch ticket"})
		return
	}

	c.JSON(http.StatusOK, toTicketDetailResponse(ticket))
}

// CreateTicket godoc
// @Summary Create ticket
// @Description Creates a new support ticket
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTicketRequest true "Ticket data"
// @Success 201 {object} TicketResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/tickets [post]
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	user := GetCurrentUser(c)

	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	priority := domain.TicketPriority(req.Priority)
	if priority == "" {
		priority = domain.TicketPriorityNormal
	}

	ticket, err := h.ticketService.CreateTicket(&user.ID, req.Subject, req.Body, user.Email, priority, "web")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create ticket"})
		return
	}

	c.JSON(http.StatusCreated, toTicketResponse(ticket))
}

// ReplyToTicket godoc
// @Summary Reply to ticket
// @Description Adds a reply to an existing ticket
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body ReplyTicketRequest true "Reply data"
// @Success 201 {object} TicketMessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/tickets/{id}/reply [post]
func (h *TicketHandler) ReplyToTicket(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ticket ID"})
		return
	}

	user := GetCurrentUser(c)

	// Verify ownership (unless staff)
	if !user.IsStaff() {
		_, err := h.ticketService.GetTicketForCustomer(ticketID, user.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Ticket not found"})
			return
		}
	}

	var req ReplyTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	message, err := h.ticketService.AddReply(ticketID, user.Email, req.Body, user.IsStaff(), nil)
	if err != nil {
		if err == ticketSvc.ErrTicketNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Ticket not found"})
			return
		}
		if err == ticketSvc.ErrTicketClosed {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Cannot reply to closed ticket"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add reply"})
		return
	}

	c.JSON(http.StatusCreated, toTicketMessageResponse(message))
}

// CloseTicket godoc
// @Summary Close ticket
// @Description Closes a support ticket
// @Tags tickets
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/tickets/{id}/close [post]
func (h *TicketHandler) CloseTicket(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ticket ID"})
		return
	}

	user := GetCurrentUser(c)

	// Verify ownership (unless staff)
	if !user.IsStaff() {
		_, err := h.ticketService.GetTicketForCustomer(ticketID, user.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Ticket not found"})
			return
		}
	}

	if err := h.ticketService.CloseTicket(ticketID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to close ticket"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Ticket closed"})
}

// GetTicketStats godoc
// @Summary Get ticket stats
// @Description Returns ticket statistics for the current user
// @Tags tickets
// @Produce json
// @Security BearerAuth
// @Success 200 {object} CustomerTicketStatsResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/tickets/stats [get]
func (h *TicketHandler) GetTicketStats(c *gin.Context) {
	userID := GetCurrentUserID(c)

	stats, err := h.ticketService.GetCustomerTicketStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, CustomerTicketStatsResponse{
		Total: stats.Total,
		Open:  stats.Open,
	})
}

// Admin endpoints

// AdminListTickets godoc
// @Summary List all tickets (Admin)
// @Description Returns all tickets in the system
// @Tags admin/tickets
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status"
// @Param limit query int false "Number of results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} PaginatedResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/admin/tickets [get]
func (h *TicketHandler) AdminListTickets(c *gin.Context) {
	limit, offset := PaginationParams(c)
	status := domain.TicketStatus(c.Query("status"))

	tickets, total, err := h.ticketService.ListTickets(nil, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch tickets"})
		return
	}

	var response []TicketResponse
	for _, t := range tickets {
		response = append(response, toTicketResponse(&t))
	}

	c.JSON(http.StatusOK, NewPaginatedResponse(response, total, limit, offset))
}

// AdminGetTicketStats godoc
// @Summary Get ticket stats (Admin)
// @Description Returns overall ticket statistics
// @Tags admin/tickets
// @Produce json
// @Security BearerAuth
// @Success 200 {object} TicketStatsResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/admin/tickets/stats [get]
func (h *TicketHandler) AdminGetTicketStats(c *gin.Context) {
	stats, err := h.ticketService.GetTicketStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, TicketStatsResponse{
		Open:         stats.Open,
		Closed:       stats.Closed,
		OnHold:       stats.OnHold,
		HighPriority: stats.HighPriority,
		Unresponded:  stats.Unresponded,
	})
}

// AdminUpdateTicketStatus godoc
// @Summary Update ticket status (Admin)
// @Description Updates the status of a ticket
// @Tags admin/tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body UpdateTicketStatusRequest true "Status data"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/admin/tickets/{id}/status [put]
func (h *TicketHandler) AdminUpdateTicketStatus(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ticket ID"})
		return
	}

	var req UpdateTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.ticketService.UpdateTicketStatus(ticketID, domain.TicketStatus(req.Status)); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update ticket status"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Ticket status updated"})
}

// AdminUpdateTicketPriority godoc
// @Summary Update ticket priority (Admin)
// @Description Updates the priority of a ticket
// @Tags admin/tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body UpdateTicketPriorityRequest true "Priority data"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/admin/tickets/{id}/priority [put]
func (h *TicketHandler) AdminUpdateTicketPriority(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ticket ID"})
		return
	}

	var req UpdateTicketPriorityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.ticketService.UpdateTicketPriority(ticketID, domain.TicketPriority(req.Priority)); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update ticket priority"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Ticket priority updated"})
}

// AdminDeleteTicket godoc
// @Summary Delete ticket (Admin)
// @Description Deletes a ticket and all its messages
// @Tags admin/tickets
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/admin/tickets/{id} [delete]
func (h *TicketHandler) AdminDeleteTicket(c *gin.Context) {
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ticket ID"})
		return
	}

	if err := h.ticketService.DeleteTicket(ticketID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete ticket"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Ticket deleted"})
}

// Helper functions

func toTicketResponse(t *domain.Ticket) TicketResponse {
	return TicketResponse{
		ID:        t.ID,
		Subject:   t.Subject,
		Status:    string(t.Status),
		Priority:  string(t.Priority),
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toTicketDetailResponse(t *domain.Ticket) TicketDetailResponse {
	var messages []TicketMessageResponse
	for _, m := range t.Messages {
		messages = append(messages, toTicketMessageResponse(&m))
	}

	return TicketDetailResponse{
		ID:        t.ID,
		Subject:   t.Subject,
		Status:    string(t.Status),
		Priority:  string(t.Priority),
		Source:    t.Source,
		Messages:  messages,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toTicketMessageResponse(m *domain.TicketMessage) TicketMessageResponse {
	var attachments []TicketAttachmentResponse
	for _, a := range m.Attachments {
		attachments = append(attachments, TicketAttachmentResponse{
			ID:          a.ID,
			FileName:    a.FileName,
			ContentType: a.ContentType,
			SizeBytes:   a.SizeBytes,
		})
	}

	return TicketMessageResponse{
		ID:          m.ID,
		SenderEmail: m.SenderEmail,
		Body:        m.Body,
		IsStaff:     m.IsStaff,
		Attachments: attachments,
		CreatedAt:   m.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// Request/Response types

type TicketResponse struct {
	ID        uint64 `json:"id"`
	Subject   string `json:"subject"`
	Status    string `json:"status"`
	Priority  string `json:"priority"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type TicketDetailResponse struct {
	ID        uint64                  `json:"id"`
	Subject   string                  `json:"subject"`
	Status    string                  `json:"status"`
	Priority  string                  `json:"priority"`
	Source    string                  `json:"source"`
	Messages  []TicketMessageResponse `json:"messages"`
	CreatedAt string                  `json:"created_at"`
	UpdatedAt string                  `json:"updated_at"`
}

type TicketMessageResponse struct {
	ID          uint64                     `json:"id"`
	SenderEmail string                     `json:"sender_email"`
	Body        string                     `json:"body"`
	IsStaff     bool                       `json:"is_staff"`
	Attachments []TicketAttachmentResponse `json:"attachments,omitempty"`
	CreatedAt   string                     `json:"created_at"`
}

type TicketAttachmentResponse struct {
	ID          uint64 `json:"id"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type TicketStatsResponse struct {
	Open         int64 `json:"open"`
	Closed       int64 `json:"closed"`
	OnHold       int64 `json:"on_hold"`
	HighPriority int64 `json:"high_priority"`
	Unresponded  int64 `json:"unresponded"`
}

type CustomerTicketStatsResponse struct {
	Total int64 `json:"total"`
	Open  int64 `json:"open"`
}

type CreateTicketRequest struct {
	Subject  string `json:"subject" binding:"required"`
	Body     string `json:"body" binding:"required"`
	Priority string `json:"priority"`
}

type ReplyTicketRequest struct {
	Body string `json:"body" binding:"required"`
}

type UpdateTicketStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type UpdateTicketPriorityRequest struct {
	Priority string `json:"priority" binding:"required"`
}
