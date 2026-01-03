package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/openhost/openhost/internal/infrastructure/config"
	"github.com/openhost/openhost/internal/infrastructure/database"
	"github.com/openhost/openhost/internal/infrastructure/web"
)

const (
	defaultBaseURL  = "http://localhost:6421"
	defaultSQLite   = "./data/openhost.db"
	defaultPGPort   = 5432
	minPasswordSize = 8
)

type installForm struct {
	AppName        string
	BaseURL        string
	AdminEmail     string
	AdminPassword  string
	DatabaseType   string
	SQLitePath     string
	PostgresHost   string
	PostgresPort   string
	PostgresUser   string
	PostgresPass   string
	PostgresDBName string
	PostgresSSL    string
}

func InstallForm(c *gin.Context) {
	installed, err := config.Exists(config.DefaultPath)
	if err != nil {
		renderInstall(c, installViewData{
			Errors: []string{"无法读取安装状态，请检查文件权限。"},
		})
		return
	}
	if installed {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	data := installViewData{
		Installed: installed,
		Form: installForm{
			DatabaseType: "sqlite",
			SQLitePath:   defaultSQLite,
			BaseURL:      defaultBaseURL,
		},
	}
	renderInstall(c, data)
}

func InstallSubmit(c *gin.Context) {
	installed, err := config.Exists(config.DefaultPath)
	if err != nil {
		renderInstall(c, installViewData{
			Errors: []string{"无法读取安装状态，请检查文件权限。"},
		})
		return
	}
	if installed {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	form := installForm{
		AppName:        strings.TrimSpace(c.PostForm("app_name")),
		BaseURL:        strings.TrimSpace(c.PostForm("base_url")),
		AdminEmail:     strings.TrimSpace(c.PostForm("admin_email")),
		AdminPassword:  c.PostForm("admin_password"),
		DatabaseType:   strings.TrimSpace(c.PostForm("db_type")),
		SQLitePath:     strings.TrimSpace(c.PostForm("sqlite_path")),
		PostgresHost:   strings.TrimSpace(c.PostForm("pg_host")),
		PostgresPort:   strings.TrimSpace(c.PostForm("pg_port")),
		PostgresUser:   strings.TrimSpace(c.PostForm("pg_user")),
		PostgresPass:   c.PostForm("pg_password"),
		PostgresDBName: strings.TrimSpace(c.PostForm("pg_database")),
		PostgresSSL:    strings.TrimSpace(c.PostForm("pg_sslmode")),
	}

	data := installViewData{Form: form}
	errors := validateInstallForm(&form)
	if len(errors) > 0 {
		data.Errors = errors
		renderInstall(c, data)
		return
	}

	configPayload, err := buildConfig(form)
	if err != nil {
		data.Errors = []string{err.Error()}
		renderInstall(c, data)
		return
	}

	if err := ensureDatabaseReady(configPayload.Database); err != nil {
		data.Errors = []string{err.Error()}
		renderInstall(c, data)
		return
	}

	if err := config.Save(config.DefaultPath, configPayload); err != nil {
		data.Errors = []string{err.Error()}
		renderInstall(c, data)
		return
	}
	c.Redirect(http.StatusSeeOther, "/dashboard")
}

type installViewData struct {
	Title          string
	Description    string
	Year           int
	Installed      bool
	Success        bool
	SuccessMessage string
	Errors         []string
	Form           installForm
}

func renderInstall(c *gin.Context, data installViewData) {
	if data.Title == "" {
		data.Title = "安装向导"
	}
	if data.Description == "" {
		data.Description = "OpenHost 安装向导"
	}
	if data.Year == 0 {
		data.Year = time.Now().Year()
	}
	web.Render(c, "install.html", gin.H{
		"Title":          data.Title,
		"Description":    data.Description,
		"Year":           data.Year,
		"Installed":      data.Installed,
		"Success":        data.Success,
		"SuccessMessage": data.SuccessMessage,
		"Errors":         data.Errors,
		"Form":           data.Form,
	})
}

func validateInstallForm(form *installForm) []string {
	var errors []string
	if form.DatabaseType == "" {
		form.DatabaseType = "sqlite"
	}
	if form.AppName == "" {
		errors = append(errors, "请输入站点名称。")
	}
	if form.BaseURL == "" {
		form.BaseURL = defaultBaseURL
	}
	if form.AdminEmail == "" {
		errors = append(errors, "请输入管理员邮箱。")
	}
	if len(form.AdminPassword) < minPasswordSize {
		errors = append(errors, "管理员密码长度至少 8 位。")
	}
	switch form.DatabaseType {
	case "sqlite":
		if form.SQLitePath == "" {
			form.SQLitePath = defaultSQLite
		}
	case "postgres":
		if form.PostgresHost == "" {
			errors = append(errors, "请输入 PostgreSQL 主机地址。")
		}
		if form.PostgresUser == "" {
			errors = append(errors, "请输入 PostgreSQL 用户名。")
		}
		if form.PostgresDBName == "" {
			errors = append(errors, "请输入 PostgreSQL 数据库名称。")
		}
	default:
		errors = append(errors, "请选择正确的数据库类型。")
	}
	return errors
}

func buildConfig(form installForm) (config.Config, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(form.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return config.Config{}, err
	}
	cfg := config.Config{
		App: config.AppConfig{
			Name:    form.AppName,
			BaseURL: form.BaseURL,
		},
		Database: config.DatabaseConfig{
			Type: form.DatabaseType,
		},
		Admin: config.AdminConfig{
			Email:        form.AdminEmail,
			PasswordHash: string(passwordHash),
		},
	}
	if cfg.Database.Type == "sqlite" {
		cfg.Database.SQLite = config.SQLiteConfig{Path: form.SQLitePath}
	}
	if cfg.Database.Type == "postgres" {
		cfg.Database.Postgres = config.PostgresConfig{
			Host:     form.PostgresHost,
			Port:     parsePort(form.PostgresPort),
			User:     form.PostgresUser,
			Password: form.PostgresPass,
			Database: form.PostgresDBName,
			SSLMode:  form.PostgresSSL,
		}
	}
	return cfg, nil
}

func parsePort(port string) int {
	if port == "" {
		return defaultPGPort
	}
	parsed, err := strconv.Atoi(port)
	if err != nil || parsed <= 0 {
		return defaultPGPort
	}
	return parsed
}

func ensureDatabaseReady(cfg config.DatabaseConfig) error {
	if cfg.Type == "sqlite" {
		dir := filepath.Dir(cfg.SQLite.Path)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}
	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	if err := database.AutoMigrate(db); err != nil {
		return err
	}
	return nil
}
