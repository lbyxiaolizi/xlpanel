package affiliate

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
)

var (
	ErrAffiliateNotFound      = errors.New("affiliate not found")
	ErrAffiliateExists        = errors.New("customer is already an affiliate")
	ErrAffiliateInactive      = errors.New("affiliate account is not active")
	ErrInsufficientBalance    = errors.New("insufficient balance for withdrawal")
	ErrWithdrawalBelowMinimum = errors.New("withdrawal amount is below minimum")
	ErrCommissionNotFound     = errors.New("commission not found")
)

// Service provides affiliate management operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new affiliate service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ApplyForAffiliate creates an affiliate application for a customer
func (s *Service) ApplyForAffiliate(customerID uint64, payoutMethod, payoutEmail string) (*domain.Affiliate, error) {
	// Check if already an affiliate
	var existing domain.Affiliate
	if err := s.db.Where("customer_id = ?", customerID).First(&existing).Error; err == nil {
		return nil, ErrAffiliateExists
	}

	// Generate referral code
	referralCode, err := s.generateReferralCode()
	if err != nil {
		return nil, err
	}

	affiliate := &domain.Affiliate{
		CustomerID:      customerID,
		Status:          domain.AffiliateStatusPending,
		PayoutMethod:    payoutMethod,
		PayoutEmail:     payoutEmail,
		CommissionRate:  decimal.NewFromInt(10), // Default 10%
		MinimumPayout:   decimal.NewFromInt(50),
		Currency:        "USD",
		ReferralCode:    referralCode,
		ReferralURL:     fmt.Sprintf("/ref/%s", referralCode),
	}

	if err := s.db.Create(affiliate).Error; err != nil {
		return nil, err
	}

	return affiliate, nil
}

// ApproveAffiliate approves an affiliate application
func (s *Service) ApproveAffiliate(affiliateID, approvedBy uint64) error {
	now := time.Now()
	return s.db.Model(&domain.Affiliate{}).Where("id = ?", affiliateID).
		Updates(map[string]interface{}{
			"status":      domain.AffiliateStatusActive,
			"approved_at": &now,
			"approved_by": &approvedBy,
		}).Error
}

// SuspendAffiliate suspends an affiliate account
func (s *Service) SuspendAffiliate(affiliateID uint64, reason string) error {
	return s.db.Model(&domain.Affiliate{}).Where("id = ?", affiliateID).
		Updates(map[string]interface{}{
			"status": domain.AffiliateStatusSuspended,
			"notes":  reason,
		}).Error
}

// GetAffiliate retrieves an affiliate by ID
func (s *Service) GetAffiliate(id uint64) (*domain.Affiliate, error) {
	var affiliate domain.Affiliate
	if err := s.db.Preload("Customer").First(&affiliate, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAffiliateNotFound
		}
		return nil, err
	}
	return &affiliate, nil
}

// GetAffiliateByCustomer retrieves an affiliate by customer ID
func (s *Service) GetAffiliateByCustomer(customerID uint64) (*domain.Affiliate, error) {
	var affiliate domain.Affiliate
	if err := s.db.Where("customer_id = ?", customerID).First(&affiliate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAffiliateNotFound
		}
		return nil, err
	}
	return &affiliate, nil
}

// GetAffiliateByCode retrieves an affiliate by referral code
func (s *Service) GetAffiliateByCode(code string) (*domain.Affiliate, error) {
	var affiliate domain.Affiliate
	if err := s.db.Where("referral_code = ? AND status = ?", code, domain.AffiliateStatusActive).First(&affiliate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAffiliateNotFound
		}
		return nil, err
	}
	return &affiliate, nil
}

// TrackClick records a click on an affiliate link
func (s *Service) TrackClick(affiliateID uint64, ipAddress, userAgent, referrerURL, landingPage string, bannerID *uint64) error {
	click := &domain.AffiliateClick{
		AffiliateID: affiliateID,
		BannerID:    bannerID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		ReferrerURL: referrerURL,
		LandingPage: landingPage,
	}

	if err := s.db.Create(click).Error; err != nil {
		return err
	}

	// Update click count
	return s.db.Model(&domain.Affiliate{}).Where("id = ?", affiliateID).
		Update("clicks", gorm.Expr("clicks + 1")).Error
}

// CreateReferral creates a referral record when a visitor arrives via affiliate link
func (s *Service) CreateReferral(affiliateID uint64, ipAddress, userAgent, referrerURL, landingPage string) (*domain.AffiliateReferral, error) {
	referral := &domain.AffiliateReferral{
		AffiliateID: affiliateID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		ReferrerURL: referrerURL,
		LandingPage: landingPage,
	}

	if err := s.db.Create(referral).Error; err != nil {
		return nil, err
	}

	return referral, nil
}

// ConvertReferral marks a referral as converted when the visitor signs up
func (s *Service) ConvertReferral(referralID, customerID uint64) error {
	now := time.Now()
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Update referral
		if err := tx.Model(&domain.AffiliateReferral{}).Where("id = ?", referralID).
			Updates(map[string]interface{}{
				"customer_id":   customerID,
				"signed_up_at": &now,
			}).Error; err != nil {
			return err
		}

		// Get affiliate
		var referral domain.AffiliateReferral
		if err := tx.First(&referral, referralID).Error; err != nil {
			return err
		}

		// Update affiliate stats
		return tx.Model(&domain.Affiliate{}).Where("id = ?", referral.AffiliateID).
			Update("signups", gorm.Expr("signups + 1")).Error
	})
}

// RecordCommission records a commission for an affiliate
func (s *Service) RecordCommission(affiliateID uint64, referralID, invoiceID, orderID *uint64, commissionType string, baseAmount, rate decimal.Decimal, currency, description string) (*domain.AffiliateCommission, error) {
	amount := baseAmount.Mul(rate).Div(decimal.NewFromInt(100))

	commission := &domain.AffiliateCommission{
		AffiliateID: affiliateID,
		ReferralID:  referralID,
		InvoiceID:   invoiceID,
		OrderID:     orderID,
		Type:        commissionType,
		Amount:      amount,
		Currency:    currency,
		Status:      "pending",
		BaseAmount:  baseAmount,
		Rate:        rate,
		Description: description,
	}

	if err := s.db.Create(commission).Error; err != nil {
		return nil, err
	}

	return commission, nil
}

// ApproveCommission approves a pending commission
func (s *Service) ApproveCommission(commissionID, approvedBy uint64) error {
	var commission domain.AffiliateCommission
	if err := s.db.First(&commission, commissionID).Error; err != nil {
		return ErrCommissionNotFound
	}

	now := time.Now()
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Update commission
		if err := tx.Model(&commission).Updates(map[string]interface{}{
			"status":      "approved",
			"approved_at": &now,
			"approved_by": &approvedBy,
		}).Error; err != nil {
			return err
		}

		// Add to affiliate balance
		return tx.Model(&domain.Affiliate{}).Where("id = ?", commission.AffiliateID).
			Updates(map[string]interface{}{
				"balance":      gorm.Expr("balance + ?", commission.Amount),
				"total_earned": gorm.Expr("total_earned + ?", commission.Amount),
				"conversions":  gorm.Expr("conversions + 1"),
			}).Error
	})
}

// RequestWithdrawal creates a withdrawal request
func (s *Service) RequestWithdrawal(affiliateID uint64, amount decimal.Decimal) (*domain.AffiliateWithdrawal, error) {
	affiliate, err := s.GetAffiliate(affiliateID)
	if err != nil {
		return nil, err
	}

	if affiliate.Status != domain.AffiliateStatusActive {
		return nil, ErrAffiliateInactive
	}

	if affiliate.Balance.LessThan(amount) {
		return nil, ErrInsufficientBalance
	}

	if amount.LessThan(affiliate.MinimumPayout) {
		return nil, ErrWithdrawalBelowMinimum
	}

	withdrawal := &domain.AffiliateWithdrawal{
		AffiliateID:   affiliateID,
		Amount:        amount,
		Currency:      affiliate.Currency,
		Status:        domain.AffiliateWithdrawalPending,
		PayoutMethod:  affiliate.PayoutMethod,
		PayoutDetails: affiliate.PayoutDetails,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(withdrawal).Error; err != nil {
			return err
		}

		// Deduct from balance
		return tx.Model(affiliate).Update("balance", affiliate.Balance.Sub(amount)).Error
	})
	if err != nil {
		return nil, err
	}
	return withdrawal, nil
}

// ProcessWithdrawal processes a withdrawal request
func (s *Service) ProcessWithdrawal(withdrawalID, processedBy uint64, transactionRef, status string) error {
	now := time.Now()

	var withdrawalStatus domain.AffiliateWithdrawalStatus
	switch status {
	case "completed":
		withdrawalStatus = domain.AffiliateWithdrawalCompleted
	case "rejected":
		withdrawalStatus = domain.AffiliateWithdrawalRejected
	default:
		withdrawalStatus = domain.AffiliateWithdrawalProcessing
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.AffiliateWithdrawal{}).Where("id = ?", withdrawalID).
			Updates(map[string]interface{}{
				"status":          withdrawalStatus,
				"processed_by":    &processedBy,
				"processed_at":    &now,
				"transaction_ref": transactionRef,
			}).Error; err != nil {
			return err
		}

		// If rejected, refund the balance
		if status == "rejected" {
			var withdrawal domain.AffiliateWithdrawal
			if err := tx.First(&withdrawal, withdrawalID).Error; err != nil {
				return err
			}
			return tx.Model(&domain.Affiliate{}).Where("id = ?", withdrawal.AffiliateID).
				Update("balance", gorm.Expr("balance + ?", withdrawal.Amount)).Error
		}

		// If completed, update total withdrawn
		if status == "completed" {
			var withdrawal domain.AffiliateWithdrawal
			if err := tx.First(&withdrawal, withdrawalID).Error; err != nil {
				return err
			}
			return tx.Model(&domain.Affiliate{}).Where("id = ?", withdrawal.AffiliateID).
				Update("total_withdrawn", gorm.Expr("total_withdrawn + ?", withdrawal.Amount)).Error
		}

		return nil
	})

	return err
}

// GetAffiliateStats returns statistics for an affiliate
func (s *Service) GetAffiliateStats(affiliateID uint64, from, to time.Time) (*AffiliateStats, error) {
	var stats AffiliateStats

	// Get clicks
	s.db.Model(&domain.AffiliateClick{}).
		Where("affiliate_id = ? AND created_at BETWEEN ? AND ?", affiliateID, from, to).
		Count(&stats.Clicks)

	// Get signups
	s.db.Model(&domain.AffiliateReferral{}).
		Where("affiliate_id = ? AND signed_up_at BETWEEN ? AND ?", affiliateID, from, to).
		Count(&stats.Signups)

	// Get conversions and earnings
	var result struct {
		Conversions int64
		Earnings    decimal.Decimal
	}
	s.db.Model(&domain.AffiliateCommission{}).
		Select("COUNT(*) as conversions, COALESCE(SUM(amount), 0) as earnings").
		Where("affiliate_id = ? AND status = 'approved' AND created_at BETWEEN ? AND ?", affiliateID, from, to).
		Scan(&result)
	stats.Conversions = result.Conversions
	stats.Earnings = result.Earnings

	// Calculate conversion rate
	if stats.Clicks > 0 {
		stats.ConversionRate = decimal.NewFromInt(stats.Conversions).Div(decimal.NewFromInt(stats.Clicks)).Mul(decimal.NewFromInt(100))
	}

	return &stats, nil
}

// ListCommissions lists commissions for an affiliate
func (s *Service) ListCommissions(affiliateID uint64, status string, limit, offset int) ([]domain.AffiliateCommission, int64, error) {
	var commissions []domain.AffiliateCommission
	var total int64

	query := s.db.Model(&domain.AffiliateCommission{}).Where("affiliate_id = ?", affiliateID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&commissions).Error; err != nil {
		return nil, 0, err
	}

	return commissions, total, nil
}

// ListWithdrawals lists withdrawals for an affiliate
func (s *Service) ListWithdrawals(affiliateID uint64, limit, offset int) ([]domain.AffiliateWithdrawal, int64, error) {
	var withdrawals []domain.AffiliateWithdrawal
	var total int64

	query := s.db.Model(&domain.AffiliateWithdrawal{}).Where("affiliate_id = ?", affiliateID)
	query.Count(&total)

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&withdrawals).Error; err != nil {
		return nil, 0, err
	}

	return withdrawals, total, nil
}

// ListAffiliates lists all affiliates (admin)
func (s *Service) ListAffiliates(status domain.AffiliateStatus, limit, offset int) ([]domain.Affiliate, int64, error) {
	var affiliates []domain.Affiliate
	var total int64

	query := s.db.Model(&domain.Affiliate{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)

	if err := query.Preload("Customer").Order("created_at DESC").Limit(limit).Offset(offset).Find(&affiliates).Error; err != nil {
		return nil, 0, err
	}

	return affiliates, total, nil
}

// UpdateAffiliateSettings updates affiliate settings
func (s *Service) UpdateAffiliateSettings(affiliateID uint64, commissionRate, minimumPayout decimal.Decimal, payoutMethod, payoutEmail string) error {
	return s.db.Model(&domain.Affiliate{}).Where("id = ?", affiliateID).
		Updates(map[string]interface{}{
			"commission_rate": commissionRate,
			"minimum_payout":  minimumPayout,
			"payout_method":   payoutMethod,
			"payout_email":    payoutEmail,
		}).Error
}

// GetBanners retrieves active affiliate banners
func (s *Service) GetBanners() ([]domain.AffiliateBanner, error) {
	var banners []domain.AffiliateBanner
	if err := s.db.Where("active = ?", true).Order("sort_order ASC").Find(&banners).Error; err != nil {
		return nil, err
	}
	return banners, nil
}

// generateReferralCode generates a unique referral code
func (s *Service) generateReferralCode() (string, error) {
	for i := 0; i < 10; i++ {
		bytes := make([]byte, 4)
		if _, err := rand.Read(bytes); err != nil {
			return "", err
		}
		code := hex.EncodeToString(bytes)

		// Check if code exists
		var count int64
		s.db.Model(&domain.Affiliate{}).Where("referral_code = ?", code).Count(&count)
		if count == 0 {
			return code, nil
		}
	}
	return "", errors.New("failed to generate unique referral code")
}

// AffiliateStats represents affiliate statistics
type AffiliateStats struct {
	Clicks         int64           `json:"clicks"`
	Signups        int64           `json:"signups"`
	Conversions    int64           `json:"conversions"`
	Earnings       decimal.Decimal `json:"earnings"`
	ConversionRate decimal.Decimal `json:"conversion_rate"`
}
