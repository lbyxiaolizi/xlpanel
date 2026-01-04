package database

import (
	"fmt"

	"github.com/openhost/openhost/internal/core/domain"
	"github.com/openhost/openhost/internal/infrastructure/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Open(cfg config.DatabaseConfig) (*gorm.DB, error) {
	switch cfg.Type {
	case "sqlite":
		return gorm.Open(sqlite.Open(cfg.SQLite.Path), &gorm.Config{})
	case "postgres":
		dsn := postgresDSN(cfg.Postgres)
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// User & Auth
		&domain.User{},
		&domain.Session{},
		&domain.PasswordResetToken{},
		&domain.EmailVerificationToken{},
		&domain.APIKey{},
		&domain.LoginAttempt{},
		&domain.ContactEmail{},
		&domain.AdminNote{},
		&domain.AuditLog{},

		// Products & Catalog
		&domain.ProductGroup{},
		&domain.Product{},
		&domain.ProductPricing{},
		&domain.ConfigGroup{},
		&domain.ProductConfigGroup{},
		&domain.ConfigOption{},
		&domain.ConfigSubOption{},
		&domain.ProductAddon{},
		&domain.ProductAddonAssignment{},
		&domain.ServiceAddon{},
		&domain.ProductBundle{},
		&domain.ProductBundleItem{},
		&domain.ProductUpgrade{},
		&domain.ConfigurableOptionGroup{},
		&domain.ConfigurableOption{},
		&domain.ConfigurableSubOption{},
		&domain.ServiceUpgrade{},
		&domain.ProductStock{},
		&domain.ProductWelcomeEmail{},
		&domain.FreeTrialConfig{},

		// Orders & Services
		&domain.Order{},
		&domain.OrderItem{},
		&domain.Service{},
		&domain.Cart{},
		&domain.CartItem{},

		// Billing & Payments
		&domain.Invoice{},
		&domain.InvoiceItem{},
		&domain.Transaction{},
		&domain.PaymentMethod{},
		&domain.PaymentGatewayModule{},
		&domain.PaymentRequest{},
		&domain.GatewayWebhookLog{},
		&domain.PaymentSubscription{},
		&domain.AutoPayment{},
		&domain.Coupon{},
		&domain.CouponUsage{},
		&domain.TaxRule{},
		&domain.Credit{},
		&domain.PromoCode{},
		&domain.Quote{},
		&domain.QuoteItem{},
		&domain.OrderApproval{},
		&domain.OrderFraudCheck{},
		&domain.OrderNote{},
		&domain.OrderStatusLog{},
		&domain.RecurringInvoice{},
		&domain.RecurringInvoiceItem{},
		&domain.InvoiceMerge{},
		&domain.DraftInvoice{},
		&domain.DraftInvoiceItem{},
		&domain.CreditAdjustment{},
		&domain.Chargeback{},
		&domain.LateFee{},

		// Affiliate
		&domain.Affiliate{},
		&domain.AffiliateReferral{},
		&domain.AffiliateCommission{},
		&domain.AffiliateWithdrawal{},
		&domain.AffiliateTier{},
		&domain.AffiliateBanner{},
		&domain.AffiliateClick{},

		// IP Management
		&domain.Subnet{},
		&domain.IPAddress{},
		&domain.DomainRegister{},
		&domain.DomainTLD{},
		&domain.DomainTLDPricing{},
		&domain.CustomerDomain{},
		&domain.DomainContact{},
		&domain.DomainDNSRecord{},
		&domain.DomainEmailForward{},
		&domain.DomainNameserver{},
		&domain.DomainTransfer{},
		&domain.DomainRenewal{},
		&domain.WHOISPrivacy{},

		// Support
		&domain.Ticket{},
		&domain.TicketMessage{},
		&domain.TicketAttachment{},
		&domain.KnowledgeBaseCategory{},
		&domain.KnowledgeBaseArticle{},
		&domain.KBArticleAttachment{},
		&domain.KBArticleFeedback{},
		&domain.KBSearchLog{},

		// Servers & Provisioning
		&domain.Server{},
		&domain.ServerGroup{},
		&domain.SSLProviderModule{},
		&domain.SSLCertificateType{},
		&domain.SSLOrder{},
		&domain.ProvisioningServerModule{},
		&domain.ProvisioningServer{},
		&domain.ServerIPAddress{},
		&domain.ProvisioningServerGroupMember{},
		&domain.ServiceProvisioningData{},
		&domain.ProvisioningLog{},
		&domain.ResellersConfig{},

		// System
		&domain.Setting{},
		&domain.EmailTemplate{},
		&domain.EmailLog{},
		&domain.Currency{},
		&domain.Announcement{},
		&domain.PaymentGateway{},
		&domain.CronTask{},
		&domain.ActivityLog{},
		&domain.Notification{},
		&domain.Module{},
		&domain.ModuleHook{},
		&domain.ModuleLog{},
		&domain.ProvisioningModule{},
		&domain.AddonModule{},
		&domain.Theme{},
		&domain.Language{},
		&domain.Translation{},
		&domain.CustomField{},
		&domain.CustomFieldValue{},
		&domain.SystemConfig{},
		&domain.IPBan{},
		&domain.CountryRestriction{},
		&domain.BackupConfig{},
		&domain.BackupLog{},
		&domain.SystemHealth{},
		&domain.UsageStatistic{},
		&domain.UsageBillingRule{},
		&domain.UsageTier{},
		&domain.EmailQueue{},
		&domain.NotificationPreference{},
		&domain.SMSConfig{},
		&domain.SMSMessage{},
		&domain.WebhookConfig{},
		&domain.WebhookDelivery{},
		&domain.SlackConfig{},
		&domain.AdminNotificationSetting{},
		&domain.NotificationEvent{},
		&domain.NewsletterSubscription{},
		&domain.Newsletter{},
		&domain.NewsletterRecipient{},
		&domain.TaxClass{},
		&domain.TaxRate{},
		&domain.TaxExemption{},
		&domain.TaxReport{},
		&domain.CronJob{},
		&domain.CronJobLog{},
		&domain.AutomationRule{},
		&domain.AutomationLog{},
		&domain.SuspensionRule{},
		&domain.InvoiceSettings{},
		&domain.ServiceAutoSettings{},
		&domain.OrderAutoSettings{},
		&domain.TicketAutoSettings{},
		&domain.DataRetentionPolicy{},
		&domain.SystemTask{},
		&domain.DiscountRule{},

		// Sub-users
		&domain.SubUser{},
		&domain.SubUserInvite{},
		&domain.SubUserSession{},
	)
}

func postgresDSN(cfg config.PostgresConfig) string {
	sslmode := cfg.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		sslmode,
	)
}
