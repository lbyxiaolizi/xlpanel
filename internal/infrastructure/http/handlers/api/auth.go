package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/openhost/openhost/internal/core/domain"
	"github.com/openhost/openhost/internal/core/service/auth"
)

// AuthHandler handles authentication API endpoints
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration data"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.authService.Register(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		if err == auth.ErrEmailExists {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Email already registered"})
			return
		}
		if err == auth.ErrPasswordTooShort {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Password must be at least 8 characters"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Registration failed"})
		return
	}

	c.JSON(http.StatusCreated, UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      string(user.Role),
		Status:    string(user.Status),
	})
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// Login godoc
// @Summary User login
// @Description Authenticates a user and returns a session token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	session, err := h.authService.Login(req.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		switch err {
		case auth.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
		case auth.ErrUserInactive:
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Account is inactive"})
		case auth.ErrUserSuspended:
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Account is suspended"})
		case auth.ErrTooManyLoginAttempts:
			c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: "Too many login attempts"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Login failed"})
		}
		return
	}

	// Get user details
	user, _ := h.authService.GetUserByID(session.UserID)

	c.JSON(http.StatusOK, LoginResponse{
		Token: session.ID,
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      string(user.Role),
			Status:    string(user.Status),
		},
	})
}

// Logout godoc
// @Summary User logout
// @Description Invalidates the current session
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		c.JSON(http.StatusOK, MessageResponse{Message: "Logged out"})
		return
	}

	_ = h.authService.Logout(token)
	c.JSON(http.StatusOK, MessageResponse{Message: "Logged out successfully"})
}

// GetCurrentUser godoc
// @Summary Get current user
// @Description Returns the currently authenticated user's profile
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserDetailResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, UserDetailResponse{
		ID:            user.ID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Company:       user.Company,
		Phone:         user.Phone,
		Address1:      user.Address1,
		Address2:      user.Address2,
		City:          user.City,
		State:         user.State,
		PostalCode:    user.PostalCode,
		Country:       user.Country,
		Role:          string(user.Role),
		Status:        string(user.Status),
		Language:      user.Language,
		Currency:      user.Currency,
		EmailVerified: user.EmailVerified,
		Credit:        user.Credit.String(),
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Company    string `json:"company"`
	Phone      string `json:"phone"`
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Updates the current user's profile
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Profile data"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Not authenticated"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := h.authService.UpdateProfile(
		user.ID,
		req.FirstName, req.LastName, req.Company, req.Phone,
		req.Address1, req.Address2, req.City, req.State, req.PostalCode, req.Country,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Profile updated successfully"})
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword godoc
// @Summary Change password
// @Description Changes the current user's password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "Password data"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Not authenticated"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := h.authService.ChangePassword(user.ID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Current password is incorrect"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to change password"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Password changed successfully"})
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Sends a password reset email to the user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Email address"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Create token (don't reveal if user exists for security)
	// Log the attempt internally for monitoring purposes
	_, err := h.authService.CreatePasswordResetToken(req.Email)
	if err != nil {
		// Log failed attempt for monitoring (user not found, etc.)
		// but don't reveal this to the client
		_ = err // Intentionally ignoring - logged in service layer
	}

	// TODO: Send email with reset link

	c.JSON(http.StatusOK, MessageResponse{Message: "If an account exists with that email, a password reset link has been sent"})
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResetPassword godoc
// @Summary Reset password
// @Description Resets the user's password using a reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset data"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := h.authService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		if err == auth.ErrInvalidToken {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid or expired reset token"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Password has been reset successfully"})
}

// AuthMiddleware validates authentication for protected routes
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "Authentication required"})
			return
		}

		user, err := h.authService.ValidateSession(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired session"})
			return
		}

		SetCurrentUser(c, user)
		c.Next()
	}
}

// AdminMiddleware restricts access to admin users
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil || !user.IsAdmin() {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{Error: "Admin access required"})
			return
		}
		c.Next()
	}
}

// StaffMiddleware restricts access to staff and admin users
func StaffMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil || !user.IsStaff() {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{Error: "Staff access required"})
			return
		}
		c.Next()
	}
}

// extractToken extracts the bearer token from the request
func extractToken(c *gin.Context) string {
	// Check Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Check query parameter
	if token := c.Query("token"); token != "" {
		return token
	}

	// Check cookie
	if cookie, err := c.Cookie("session"); err == nil {
		return cookie
	}

	return ""
}

// Context keys
const (
	userContextKey = "authenticated_user"
)

// SetCurrentUser sets the current user in the context
func SetCurrentUser(c *gin.Context, user *domain.User) {
	c.Set(userContextKey, user)
}

// GetCurrentUser retrieves the current user from the context
func GetCurrentUser(c *gin.Context) *domain.User {
	if user, exists := c.Get(userContextKey); exists {
		if u, ok := user.(*domain.User); ok {
			return u
		}
	}
	return nil
}

// GetCurrentUserID retrieves the current user's ID from the context
func GetCurrentUserID(c *gin.Context) uint64 {
	user := GetCurrentUser(c)
	if user != nil {
		return user.ID
	}
	return 0
}

// Response types
type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type UserResponse struct {
	ID        uint64 `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	Status    string `json:"status"`
}

type UserDetailResponse struct {
	ID            uint64 `json:"id"`
	Email         string `json:"email"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Company       string `json:"company,omitempty"`
	Phone         string `json:"phone,omitempty"`
	Address1      string `json:"address1,omitempty"`
	Address2      string `json:"address2,omitempty"`
	City          string `json:"city,omitempty"`
	State         string `json:"state,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	Language      string `json:"language"`
	Currency      string `json:"currency"`
	EmailVerified bool   `json:"email_verified"`
	Credit        string `json:"credit"`
	CreatedAt     string `json:"created_at"`
}

// PaginationParams extracts pagination parameters from the request
func PaginationParams(c *gin.Context) (limit, offset int) {
	limit = 20
	offset = 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if page := c.Query("page"); page != "" {
		if parsed, err := strconv.Atoi(page); err == nil && parsed > 0 {
			offset = (parsed - 1) * limit
		}
	} else if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(data interface{}, total int64, limit, offset int) PaginatedResponse {
	page := (offset / limit) + 1
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PerPage:    limit,
		TotalPages: totalPages,
	}
}
