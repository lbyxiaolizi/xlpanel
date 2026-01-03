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
		&domain.ConfigGroup{},
		&domain.ProductConfigGroup{},
		&domain.ConfigOption{},
		&domain.ConfigSubOption{},
		
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
		&domain.Coupon{},
		&domain.CouponUsage{},
		&domain.TaxRule{},
		&domain.Credit{},
		
		// IP Management
		&domain.Subnet{},
		&domain.IPAddress{},
		
		// Support
		&domain.Ticket{},
		&domain.TicketMessage{},
		&domain.TicketAttachment{},
		
		// Servers & Provisioning
		&domain.Server{},
		&domain.ServerGroup{},
		
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
