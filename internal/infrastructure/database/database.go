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
		&domain.ProductGroup{},
		&domain.Product{},
		&domain.ConfigGroup{},
		&domain.ProductConfigGroup{},
		&domain.ConfigOption{},
		&domain.ConfigSubOption{},
		&domain.Order{},
		&domain.Service{},
		&domain.Subnet{},
		&domain.IPAddress{},
		&domain.Ticket{},
		&domain.TicketMessage{},
		&domain.TicketAttachment{},
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
