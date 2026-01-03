package api

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"xlpanel/internal/domain"
	"xlpanel/internal/i18n"
)

// Auth request/response types
type loginRequest struct {
	TenantID      string `json:"tenant_id"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	TwoFactorCode string `json:"two_factor_code,omitempty"`
}

type registerRequest struct {
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginResponse struct {
	Token       string       `json:"token,omitempty"`
	User        *domain.User `json:"user,omitempty"`
	Requires2FA bool         `json:"requires_2fa,omitempty"`
}

// handleLoginPage renders the login page
func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	translator := i18n.NewTranslator()
	
	data := map[string]interface{}{
		"Title": translator.T(lang, "auth.login"),
		"Lang":  lang,
		"T": func(key string) string {
			return translator.T(lang, key)
		},
	}

	s.renderTemplate(w, r, "login.html", data)
}

// handleRegisterPage renders the registration page
func (s *Server) handleRegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	translator := i18n.NewTranslator()
	
	data := map[string]interface{}{
		"Title": translator.T(lang, "auth.register"),
		"Lang":  lang,
		"T": func(key string) string {
			return translator.T(lang, key)
		},
	}

	s.renderTemplate(w, r, "register.html", data)
}

// handleLogin processes login requests
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.TenantID == "" {
		req.TenantID = "default"
	}

	// Attempt login
	token, user, err := s.authService.Login(req.TenantID, req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Check if 2FA is required
	if user.TwoFactorEnabled {
		if req.TwoFactorCode == "" {
			writeJSON(w, http.StatusOK, loginResponse{Requires2FA: true})
			return
		}
		
		// Verify 2FA code
		if err := s.authService.Verify2FA(user.ID, req.TwoFactorCode); err != nil {
			writeError(w, http.StatusUnauthorized, "invalid 2FA code")
			return
		}
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token: token,
		User:  user,
	})
}

// handleRegister processes registration requests
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.TenantID == "" {
		req.TenantID = "default"
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "email, password, and name are required")
		return
	}

	// Register user
	user, err := s.authService.Register(req.TenantID, req.Email, req.Password, req.Name, "customer")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Send welcome email
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}
	if err := s.emailService.SendWelcomeEmail(user.Email, user.Name, lang); err != nil {
		log.Printf("Failed to send welcome email: %v", err)
	}

	writeJSON(w, http.StatusCreated, user)
}

// handleLogout logs out a user
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// In a real implementation, you would invalidate the session/token here
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// handle2FAEnable enables 2FA for a user
func (s *Server) handle2FAEnable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	secret, err := s.authService.Enable2FA(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Send 2FA setup email
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}
	if err := s.emailService.Send2FAEmail(user.Email, user.Name, secret, lang); err != nil {
		log.Printf("Failed to send 2FA email: %v", err)
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"secret":  secret,
		"message": "2FA enabled",
	})
}

// handle2FAVerify verifies a 2FA code
func (s *Server) handle2FAVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.authService.Verify2FA(user.ID, req.Code); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid 2FA code")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "2FA verified"})
}

// handle2FADisable disables 2FA for a user
func (s *Server) handle2FADisable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := s.authService.Disable2FA(user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "2FA disabled"})
}

// getUserFromRequest extracts the user from the JWT token in the request
func (s *Server) getUserFromRequest(r *http.Request) *domain.User {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil
	}

	token := parts[1]
	user, err := s.authService.VerifyToken(token)
	if err != nil {
		return nil
	}

	return user
}

// renderTemplate renders an HTML template
func (s *Server) renderTemplate(w http.ResponseWriter, r *http.Request, templateName string, data map[string]interface{}) {
	theme := s.config.DefaultTheme
	if s.config.AllowThemeOverride {
		if requested := r.URL.Query().Get("theme"); requested != "" {
			theme = requested
		}
	}

	tmpl, err := s.loadSingleTemplate(theme, templateName)
	if err != nil {
		log.Printf("Template not loaded: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// loadSingleTemplate loads a specific template
func (s *Server) loadSingleTemplate(theme, templateName string) (*template.Template, error) {
	// Validate theme name
	if theme == "" {
		theme = s.config.DefaultTheme
	}
	
	themePath := fmt.Sprintf("frontend/themes/%s", theme)
	templatePath := fmt.Sprintf("%s/%s", themePath, templateName)
	
	// Check if this is a customer template
	customerTemplatePath := fmt.Sprintf("%s/customer/%s", themePath, templateName)
	
	// Try customer template first
	if tmpl, err := template.ParseFiles(customerTemplatePath); err == nil {
		return tmpl, nil
	}
	
	// Fall back to regular template
	return template.ParseFiles(templatePath)
}

// handleAdminUsers renders the admin users management page
func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil || user.Role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	translator := i18n.NewTranslator()
	
	data := map[string]interface{}{
		"Title": translator.T(lang, "nav.account"),
		"Lang":  lang,
		"User":  user,
		"T": func(key string) string {
			return translator.T(lang, key)
		},
	}

	s.renderTemplate(w, r, "admin/users.html", data)
}

// handleUsersAPI handles the API endpoint for listing users
func (s *Server) handleUsersAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil || user.Role != "admin" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	users := s.authService.ListUsers()
	writeJSON(w, http.StatusOK, users)
}
