package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const DefaultPath = "./config/openhost.json"

const filePerm = 0o600

const dirPerm = 0o750

type Config struct {
	App      AppConfig      `json:"app"`
	Database DatabaseConfig `json:"database"`
	Admin    AdminConfig    `json:"admin"`
}

type AppConfig struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
}

type DatabaseConfig struct {
	Type     string         `json:"type"`
	SQLite   SQLiteConfig   `json:"sqlite"`
	Postgres PostgresConfig `json:"postgres"`
}

type SQLiteConfig struct {
	Path string `json:"path"`
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslmode"`
}

type AdminConfig struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

func Exists(path string) (bool, error) {
	if path == "" {
		path = DefaultPath
	}
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Save(path string, cfg Config) error {
	if path == "" {
		path = DefaultPath
	}
	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	file, err := os.CreateTemp(dir, "openhost-config-*.json")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpName := file.Name()
	defer os.Remove(tmpName)
	if _, err := file.Write(payload); err != nil {
		_ = file.Close()
		return fmt.Errorf("write config: %w", err)
	}
	if err := file.Chmod(filePerm); err != nil {
		_ = file.Close()
		return fmt.Errorf("chmod config: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close config: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("move config: %w", err)
	}
	return nil
}
