package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/openhost/openhost/internal/core/domain"
	"github.com/openhost/openhost/internal/core/service/payment"
)

// PaymentHandler handles payment API endpoints
type PaymentHandler struct {
	service *payment.Service
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(service *payment.Service) *PaymentHandler {
	return &PaymentHandler{service: service}
}

// ListGateways lists available payment gateways
// @Summary List payment gateways
// @Description Get a list of available payment gateways for the customer
// @Tags Payments
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/payments/gateways [get]
func (h *PaymentHandler) ListGateways(c *gin.Context) {
	gateways, err := h.service.ListActiveGateways()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"gateways": gateways})
}

// CreatePaymentRequest creates a payment request for an invoice
// @Summary Create payment request
// @Description Create a payment request for an invoice using a specific gateway
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body CreatePaymentRequestBody true "Payment request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/payments [post]
func (h *PaymentHandler) CreatePaymentRequest(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreatePaymentRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount := decimal.NewFromFloat(req.Amount)

	result, err := h.service.CreatePaymentRequest(
		customerID.(uint64),
		req.InvoiceID,
		req.GatewayID,
		amount,
		req.Currency,
		c.ClientIP(),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment request created",
		"request": result,
	})
}

// ProcessPayment processes a pending payment request
// @Summary Process payment
// @Description Process a pending payment request
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path int true "Payment Request ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/payments/{id}/process [post]
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	requestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment request ID"})
		return
	}

	result, err := h.service.ProcessPayment(requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": result.Success,
		"result":  result,
	})
}

// ProcessCallback processes a payment gateway callback
// @Summary Process callback
// @Description Process a callback/webhook from a payment gateway
// @Tags Payments
// @Accept json
// @Produce json
// @Param gateway path string true "Gateway slug"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/payments/callback/{gateway} [post]
func (h *PaymentHandler) ProcessCallback(c *gin.Context) {
	gateway := c.Param("gateway")

	// Read the raw body
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Get signature from headers
	signature := c.GetHeader("X-Signature")

	// Process the webhook
	if err := h.service.ProcessWebhook(gateway, body, signature); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// PayWithCredit pays an invoice using customer credit balance
// @Summary Pay with credit
// @Description Pay an invoice using customer credit balance
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body PayWithCreditRequest true "Credit payment request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/payments/credit [post]
func (h *PaymentHandler) PayWithCredit(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req PayWithCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount := decimal.NewFromFloat(req.Amount)

	transaction, err := h.service.PayWithCredit(customerID.(uint64), req.InvoiceID, amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Payment successful",
		"transaction": transaction,
	})
}

// SavePaymentMethod saves a payment method for the customer
// @Summary Save payment method
// @Description Save a payment method for future use
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body SavePaymentMethodRequest true "Payment method request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/payments/methods [post]
func (h *PaymentHandler) SavePaymentMethod(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req SavePaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	method, err := h.service.SavePaymentMethod(
		customerID.(uint64),
		domain.PaymentMethodCard,
		req.Gateway,
		req.Token,
		req.Label,
		req.Last4,
		req.Brand,
		req.ExpiryMonth,
		req.ExpiryYear,
		req.SetDefault,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment method saved",
		"method":  method,
	})
}

// SetDefaultPaymentMethod sets a payment method as default
// @Summary Set default payment method
// @Description Set a payment method as the default
// @Tags Payments
// @Produce json
// @Param id path int true "Payment method ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/payments/methods/{id}/default [post]
func (h *PaymentHandler) SetDefaultPaymentMethod(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	methodID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid method ID"})
		return
	}

	if err := h.service.SetDefaultPaymentMethod(customerID.(uint64), methodID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default payment method updated"})
}

// DeletePaymentMethod deletes a saved payment method
// @Summary Delete payment method
// @Description Delete a saved payment method
// @Tags Payments
// @Produce json
// @Param id path int true "Payment method ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/payments/methods/{id} [delete]
func (h *PaymentHandler) DeletePaymentMethod(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	methodID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid method ID"})
		return
	}

	if err := h.service.DeletePaymentMethod(customerID.(uint64), methodID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment method deleted"})
}

// SetupAutoPayment sets up automatic payment for a customer
// @Summary Setup auto payment
// @Description Configure automatic payment for invoices
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body SetupAutoPaymentRequest true "Auto payment setup request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/payments/auto [post]
func (h *PaymentHandler) SetupAutoPayment(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req SetupAutoPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	maxAmount := decimal.NewFromFloat(req.MaxAmount)

	config, err := h.service.SetupAutoPayment(customerID.(uint64), req.PaymentMethodID, maxAmount, req.DaysBefore)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Auto payment configured",
		"config":  config,
	})
}

// GetAutoPaymentConfig gets auto payment configuration
// @Summary Get auto payment config
// @Description Get the auto payment configuration for the customer
// @Tags Payments
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/payments/auto [get]
func (h *PaymentHandler) GetAutoPaymentConfig(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	config, err := h.service.GetAutoPaymentConfig(customerID.(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": config})
}

// Admin handlers

// AdminAddCredit adds credit to a customer account
// @Summary Admin: Add credit
// @Description Add credit to a customer's account (admin only)
// @Tags Admin Payments
// @Accept json
// @Produce json
// @Param request body AdminAddCreditRequest true "Add credit request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/payments/credit [post]
func (h *PaymentHandler) AdminAddCredit(c *gin.Context) {
	adminID, _ := c.Get("admin_id")

	var req AdminAddCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount := decimal.NewFromFloat(req.Amount)
	staffID := adminID.(uint64)

	adjustment, err := h.service.AddCredit(req.CustomerID, amount, req.Currency, req.Reason, &staffID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Credit added",
		"adjustment": adjustment,
	})
}

// AdminRefundPayment refunds a payment
// @Summary Admin: Refund payment
// @Description Refund a payment (admin only)
// @Tags Admin Payments
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Param request body RefundRequest true "Refund request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/payments/{id}/refund [post]
func (h *PaymentHandler) AdminRefundPayment(c *gin.Context) {
	adminID, _ := c.Get("admin_id")
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	var req RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount := decimal.NewFromFloat(req.Amount)

	refund, err := h.service.ProcessRefund(transactionID, amount, req.Reason, adminID.(uint64))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Refund processed",
		"refund":  refund,
	})
}

// Request/Response types
type CreatePaymentRequestBody struct {
	InvoiceID uint64  `json:"invoice_id" binding:"required"`
	GatewayID uint64  `json:"gateway_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Currency  string  `json:"currency" binding:"required,len=3"`
}

type PayWithCreditRequest struct {
	InvoiceID uint64  `json:"invoice_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
}

type SavePaymentMethodRequest struct {
	Gateway     string `json:"gateway" binding:"required"`
	Token       string `json:"token" binding:"required"`
	Label       string `json:"label"`
	Last4       string `json:"last4"`
	Brand       string `json:"brand"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	SetDefault  bool   `json:"set_default"`
}

type SetupAutoPaymentRequest struct {
	PaymentMethodID uint64  `json:"payment_method_id" binding:"required"`
	MaxAmount       float64 `json:"max_amount" binding:"required,gt=0"`
	DaysBefore      int     `json:"days_before" binding:"required,gt=0"`
}

type AdminAddCreditRequest struct {
	CustomerID uint64  `json:"customer_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Currency   string  `json:"currency" binding:"required,len=3"`
	Reason     string  `json:"reason" binding:"required"`
}

type RefundRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Reason string  `json:"reason"`
}
