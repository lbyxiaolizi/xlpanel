package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/openhost/openhost/internal/core/domain"
	"github.com/openhost/openhost/internal/core/service/affiliate"
)

// AffiliateHandler handles affiliate API endpoints
type AffiliateHandler struct {
	service *affiliate.Service
}

// NewAffiliateHandler creates a new affiliate handler
func NewAffiliateHandler(service *affiliate.Service) *AffiliateHandler {
	return &AffiliateHandler{service: service}
}

// GetAffiliate gets the current user's affiliate account
// @Summary Get affiliate account
// @Description Get the current user's affiliate account information
// @Tags Affiliates
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/affiliate [get]
func (h *AffiliateHandler) GetAffiliate(c *gin.Context) {
	// Get customer ID from context (set by auth middleware)
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	aff, err := h.service.GetAffiliateByCustomer(customerID.(uint64))
	if err != nil {
		if err == affiliate.ErrAffiliateNotFound {
			c.JSON(http.StatusOK, gin.H{"affiliate": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"affiliate": aff})
}

// Apply applies for the affiliate program
// @Summary Apply for affiliate program
// @Description Submit an application to join the affiliate program
// @Tags Affiliates
// @Accept json
// @Produce json
// @Param request body ApplyAffiliateRequest true "Application request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/affiliate [post]
func (h *AffiliateHandler) Apply(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ApplyAffiliateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	aff, err := h.service.ApplyForAffiliate(customerID.(uint64), req.PayoutMethod, req.PayoutEmail)
	if err != nil {
		if err == affiliate.ErrAffiliateExists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You are already enrolled in the affiliate program"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Application submitted successfully",
		"affiliate": aff,
	})
}

// GetCommissions lists commissions for the current affiliate
// @Summary List commissions
// @Description Get a list of commissions for the current affiliate
// @Tags Affiliates
// @Produce json
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset results"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/affiliate/commissions [get]
func (h *AffiliateHandler) GetCommissions(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	aff, err := h.service.GetAffiliateByCustomer(customerID.(uint64))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not an affiliate"})
		return
	}

	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	commissions, total, err := h.service.ListCommissions(aff.ID, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"commissions": commissions,
		"total":       total,
	})
}

// GetWithdrawals lists withdrawals for the current affiliate
// @Summary List withdrawals
// @Description Get a list of withdrawals for the current affiliate
// @Tags Affiliates
// @Produce json
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset results"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/affiliate/withdrawals [get]
func (h *AffiliateHandler) GetWithdrawals(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	aff, err := h.service.GetAffiliateByCustomer(customerID.(uint64))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not an affiliate"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	withdrawals, total, err := h.service.ListWithdrawals(aff.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"withdrawals": withdrawals,
		"total":       total,
	})
}

// RequestWithdrawal requests a withdrawal
// @Summary Request withdrawal
// @Description Submit a withdrawal request
// @Tags Affiliates
// @Accept json
// @Produce json
// @Param request body WithdrawalRequest true "Withdrawal request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/affiliate/withdraw [post]
func (h *AffiliateHandler) RequestWithdrawal(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	aff, err := h.service.GetAffiliateByCustomer(customerID.(uint64))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not an affiliate"})
		return
	}

	var req WithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount := decimal.NewFromFloat(req.Amount)

	withdrawal, err := h.service.RequestWithdrawal(aff.ID, amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Withdrawal request submitted",
		"withdrawal": withdrawal,
	})
}

// UpdateSettings updates affiliate settings
// @Summary Update affiliate settings
// @Description Update payout settings for the affiliate
// @Tags Affiliates
// @Accept json
// @Produce json
// @Param request body UpdateAffiliateSettingsRequest true "Settings request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/affiliate/settings [put]
func (h *AffiliateHandler) UpdateSettings(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	aff, err := h.service.GetAffiliateByCustomer(customerID.(uint64))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not an affiliate"})
		return
	}

	var req UpdateAffiliateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateAffiliateSettings(aff.ID, aff.CommissionRate, aff.MinimumPayout, req.PayoutMethod, req.PayoutEmail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated"})
}

// GetBanners gets available promotional banners
// @Summary Get promotional banners
// @Description Get available banners for affiliates
// @Tags Affiliates
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/affiliate/banners [get]
func (h *AffiliateHandler) GetBanners(c *gin.Context) {
	banners, err := h.service.GetBanners()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"banners": banners})
}

// TrackClick tracks an affiliate referral click
// @Summary Track click
// @Description Track a click on an affiliate link
// @Tags Affiliates
// @Accept json
// @Produce json
// @Param code path string true "Referral code"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/ref/{code} [get]
func (h *AffiliateHandler) TrackClick(c *gin.Context) {
	code := c.Param("code")

	aff, err := h.service.GetAffiliateByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid referral code"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	referrer := c.GetHeader("Referer")
	landingPage := c.Request.URL.String()

	if err := h.service.TrackClick(aff.ID, ipAddress, userAgent, referrer, landingPage, nil); err != nil {
		// Log error but don't fail
	}

	// Create referral
	referral, _ := h.service.CreateReferral(aff.ID, ipAddress, userAgent, referrer, landingPage)

	c.JSON(http.StatusOK, gin.H{
		"referral_id": referral.ID,
		"affiliate":   aff.CustomerID,
	})
}

// Request/Response types
type ApplyAffiliateRequest struct {
	PayoutMethod string `json:"payout_method" binding:"required"`
	PayoutEmail  string `json:"payout_email"`
}

type WithdrawalRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type UpdateAffiliateSettingsRequest struct {
	PayoutMethod string `json:"payout_method" binding:"required"`
	PayoutEmail  string `json:"payout_email"`
}

// Admin handlers

// AdminListAffiliates lists all affiliates (admin only)
func (h *AffiliateHandler) AdminListAffiliates(c *gin.Context) {
	statusStr := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Convert string to AffiliateStatus
	var status domain.AffiliateStatus
	if statusStr != "" {
		status = domain.AffiliateStatus(statusStr)
	}

	affiliates, total, err := h.service.ListAffiliates(status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"affiliates": affiliates,
		"total":      total,
	})
}

// AdminApproveAffiliate approves an affiliate application
func (h *AffiliateHandler) AdminApproveAffiliate(c *gin.Context) {
	adminID, _ := c.Get("admin_id")
	affiliateID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid affiliate ID"})
		return
	}

	if err := h.service.ApproveAffiliate(affiliateID, adminID.(uint64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Affiliate approved"})
}

// AdminSuspendAffiliate suspends an affiliate
func (h *AffiliateHandler) AdminSuspendAffiliate(c *gin.Context) {
	affiliateID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid affiliate ID"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	if err := h.service.SuspendAffiliate(affiliateID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Affiliate suspended"})
}

// AdminProcessWithdrawal processes a withdrawal request
func (h *AffiliateHandler) AdminProcessWithdrawal(c *gin.Context) {
	adminID, _ := c.Get("admin_id")
	withdrawalID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid withdrawal ID"})
		return
	}

	var req struct {
		Status         string `json:"status" binding:"required"`
		TransactionRef string `json:"transaction_ref"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ProcessWithdrawal(withdrawalID, adminID.(uint64), req.TransactionRef, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal processed"})
}
