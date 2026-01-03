package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/openhost/openhost/internal/core/domain"
	invoiceSvc "github.com/openhost/openhost/internal/core/service/invoice"
)

// InvoiceHandler handles invoice API endpoints
type InvoiceHandler struct {
	invoiceService *invoiceSvc.Service
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(invoiceService *invoiceSvc.Service) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService}
}

// ListInvoices godoc
// @Summary List invoices
// @Description Returns the current user's invoices
// @Tags invoices
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (unpaid, paid, cancelled, refunded, overdue)"
// @Param limit query int false "Number of results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} PaginatedResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/invoices [get]
func (h *InvoiceHandler) ListInvoices(c *gin.Context) {
	userID := GetCurrentUserID(c)
	limit, offset := PaginationParams(c)
	status := domain.InvoiceStatus(c.Query("status"))

	invoices, total, err := h.invoiceService.ListInvoices(userID, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch invoices"})
		return
	}

	var response []InvoiceResponse
	for _, inv := range invoices {
		response = append(response, toInvoiceResponse(&inv))
	}

	c.JSON(http.StatusOK, NewPaginatedResponse(response, total, limit, offset))
}

// GetInvoice godoc
// @Summary Get invoice details
// @Description Returns details of a specific invoice
// @Tags invoices
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Success 200 {object} InvoiceDetailResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/invoices/{id} [get]
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid invoice ID"})
		return
	}

	inv, err := h.invoiceService.GetInvoice(invoiceID)
	if err != nil {
		if err == invoiceSvc.ErrInvoiceNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Invoice not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch invoice"})
		return
	}

	// Verify ownership (unless admin)
	user := GetCurrentUser(c)
	if inv.CustomerID != user.ID && !user.IsAdmin() {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Invoice not found"})
		return
	}

	c.JSON(http.StatusOK, toInvoiceDetailResponse(inv))
}

// GetUnpaidInvoices godoc
// @Summary Get unpaid invoices
// @Description Returns all unpaid invoices for the current user
// @Tags invoices
// @Produce json
// @Security BearerAuth
// @Success 200 {array} InvoiceResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/invoices/unpaid [get]
func (h *InvoiceHandler) GetUnpaidInvoices(c *gin.Context) {
	userID := GetCurrentUserID(c)

	invoices, err := h.invoiceService.GetUnpaidInvoices(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch invoices"})
		return
	}

	var response []InvoiceResponse
	for _, inv := range invoices {
		response = append(response, toInvoiceResponse(&inv))
	}

	c.JSON(http.StatusOK, response)
}

// Admin endpoints

// AdminListInvoices godoc
// @Summary List all invoices (Admin)
// @Description Returns all invoices in the system
// @Tags admin/invoices
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status"
// @Param customer_id query int false "Filter by customer"
// @Param limit query int false "Number of results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} PaginatedResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/admin/invoices [get]
func (h *InvoiceHandler) AdminListInvoices(c *gin.Context) {
	limit, offset := PaginationParams(c)
	status := domain.InvoiceStatus(c.Query("status"))

	// For admin, list all invoices
	invoices, total, err := h.invoiceService.ListInvoices(0, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch invoices"})
		return
	}

	var response []InvoiceResponse
	for _, inv := range invoices {
		response = append(response, toInvoiceResponse(&inv))
	}

	c.JSON(http.StatusOK, NewPaginatedResponse(response, total, limit, offset))
}

// AdminCancelInvoice godoc
// @Summary Cancel invoice (Admin)
// @Description Cancels an unpaid invoice
// @Tags admin/invoices
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invoice ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/admin/invoices/{id}/cancel [post]
func (h *InvoiceHandler) AdminCancelInvoice(c *gin.Context) {
	invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid invoice ID"})
		return
	}

	if err := h.invoiceService.CancelInvoice(invoiceID); err != nil {
		if err == invoiceSvc.ErrInvoiceNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Invoice not found"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Invoice cancelled"})
}

// Helper functions

func toInvoiceResponse(inv *domain.Invoice) InvoiceResponse {
	return InvoiceResponse{
		ID:            inv.ID,
		InvoiceNumber: inv.InvoiceNumber,
		Status:        string(inv.Status),
		Currency:      inv.Currency,
		Total:         inv.Total.String(),
		Balance:       inv.Balance.String(),
		DueDate:       inv.DueDate.Format("2006-01-02"),
		CreatedAt:     inv.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toInvoiceDetailResponse(inv *domain.Invoice) InvoiceDetailResponse {
	var items []InvoiceItemResponse
	for _, item := range inv.LineItems {
		items = append(items, InvoiceItemResponse{
			ID:          item.ID,
			Type:        item.Type,
			Description: item.Description,
			Quantity:    item.Quantity.String(),
			UnitPrice:   item.UnitPrice.String(),
			Discount:    item.Discount.String(),
			Total:       item.Total.String(),
		})
	}

	resp := InvoiceDetailResponse{
		ID:            inv.ID,
		InvoiceNumber: inv.InvoiceNumber,
		Status:        string(inv.Status),
		Currency:      inv.Currency,
		Subtotal:      inv.Subtotal.String(),
		Discount:      inv.Discount.String(),
		TaxAmount:     inv.TaxAmount.String(),
		Total:         inv.Total.String(),
		AmountPaid:    inv.AmountPaid.String(),
		Balance:       inv.Balance.String(),
		DueDate:       inv.DueDate.Format("2006-01-02"),
		Items:         items,
		Notes:         inv.Notes,
		CreatedAt:     inv.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if inv.PaidAt != nil {
		paidAt := inv.PaidAt.Format("2006-01-02T15:04:05Z")
		resp.PaidAt = &paidAt
	}

	return resp
}

// Response types

type InvoiceResponse struct {
	ID            uint64 `json:"id"`
	InvoiceNumber string `json:"invoice_number"`
	Status        string `json:"status"`
	Currency      string `json:"currency"`
	Total         string `json:"total"`
	Balance       string `json:"balance"`
	DueDate       string `json:"due_date"`
	CreatedAt     string `json:"created_at"`
}

type InvoiceDetailResponse struct {
	ID            uint64                `json:"id"`
	InvoiceNumber string                `json:"invoice_number"`
	Status        string                `json:"status"`
	Currency      string                `json:"currency"`
	Subtotal      string                `json:"subtotal"`
	Discount      string                `json:"discount"`
	TaxAmount     string                `json:"tax_amount"`
	Total         string                `json:"total"`
	AmountPaid    string                `json:"amount_paid"`
	Balance       string                `json:"balance"`
	DueDate       string                `json:"due_date"`
	PaidAt        *string               `json:"paid_at,omitempty"`
	Items         []InvoiceItemResponse `json:"items"`
	Notes         string                `json:"notes,omitempty"`
	CreatedAt     string                `json:"created_at"`
}

type InvoiceItemResponse struct {
	ID          uint64 `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Quantity    string `json:"quantity"`
	UnitPrice   string `json:"unit_price"`
	Discount    string `json:"discount"`
	Total       string `json:"total"`
}
