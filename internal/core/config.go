package core

import (
	"os"
	"strconv"
	"time"
)

type AppConfig struct {
	DefaultTheme       string
	AllowThemeOverride bool
	Address            string
	ReadTimeout        time.Duration
	ReadHeaderTimeout  time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	ShutdownTimeout    time.Duration
	MaxBodyBytes       int64
	APIKey             string
	// JWT configuration
	JWTSecret      string
	JWTExpiration  time.Duration
	// SMTP configuration
	SMTPHost       string
	SMTPPort       int
	SMTPUsername   string
	SMTPPassword   string
	SMTPFrom       string
	SMTPFromName   string
}

var DefaultConfig = AppConfig{
	DefaultTheme:       "classic",
	AllowThemeOverride: true,
	Address:            ":8080",
	ReadTimeout:        15 * time.Second,
	ReadHeaderTimeout:  5 * time.Second,
	WriteTimeout:       30 * time.Second,
	IdleTimeout:        60 * time.Second,
	ShutdownTimeout:    10 * time.Second,
	MaxBodyBytes:       1 << 20,
	JWTSecret:          "change-me-in-production",
	JWTExpiration:      24 * time.Hour,
	SMTPPort:           587,
	SMTPFromName:       "XLPanel",
}

func LoadConfigFromEnv() AppConfig {
	cfg := DefaultConfig
	if value := os.Getenv("XLPANEL_DEFAULT_THEME"); value != "" {
		cfg.DefaultTheme = value
	}
	if value := os.Getenv("XLPANEL_ALLOW_THEME_OVERRIDE"); value != "" {
		cfg.AllowThemeOverride = value == "1" || value == "true" || value == "TRUE"
	}
	if value := os.Getenv("XLPANEL_ADDR"); value != "" {
		cfg.Address = value
	}
	if value := os.Getenv("XLPANEL_READ_TIMEOUT"); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			cfg.ReadTimeout = parsed
		}
	}
	if value := os.Getenv("XLPANEL_READ_HEADER_TIMEOUT"); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			cfg.ReadHeaderTimeout = parsed
		}
	}
	if value := os.Getenv("XLPANEL_WRITE_TIMEOUT"); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			cfg.WriteTimeout = parsed
		}
	}
	if value := os.Getenv("XLPANEL_IDLE_TIMEOUT"); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			cfg.IdleTimeout = parsed
		}
	}
	if value := os.Getenv("XLPANEL_SHUTDOWN_TIMEOUT"); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			cfg.ShutdownTimeout = parsed
		}
	}
	if value := os.Getenv("XLPANEL_MAX_BODY_BYTES"); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			cfg.MaxBodyBytes = parsed
		}
	}
	if value := os.Getenv("XLPANEL_API_KEY"); value != "" {
		cfg.APIKey = value
	}
	if value := os.Getenv("XLPANEL_JWT_SECRET"); value != "" {
		cfg.JWTSecret = value
	}
	if value := os.Getenv("XLPANEL_JWT_EXPIRATION"); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			cfg.JWTExpiration = parsed
		}
	}
	if value := os.Getenv("XLPANEL_SMTP_HOST"); value != "" {
		cfg.SMTPHost = value
	}
	if value := os.Getenv("XLPANEL_SMTP_PORT"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			cfg.SMTPPort = parsed
		}
	}
	if value := os.Getenv("XLPANEL_SMTP_USERNAME"); value != "" {
		cfg.SMTPUsername = value
	}
	if value := os.Getenv("XLPANEL_SMTP_PASSWORD"); value != "" {
		cfg.SMTPPassword = value
	}
	if value := os.Getenv("XLPANEL_SMTP_FROM"); value != "" {
		cfg.SMTPFrom = value
	}
	if value := os.Getenv("XLPANEL_SMTP_FROM_NAME"); value != "" {
		cfg.SMTPFromName = value
	}
	return cfg
}
