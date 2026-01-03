package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/openhost/openhost/internal/core/domain"
)

var (
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserInactive          = errors.New("user account is inactive")
	ErrUserSuspended         = errors.New("user account is suspended")
	ErrEmailExists           = errors.New("email already exists")
	ErrInvalidToken          = errors.New("invalid or expired token")
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters")
	ErrSessionExpired        = errors.New("session has expired")
	ErrTooManyLoginAttempts  = errors.New("too many failed login attempts, please try again later")
)

const (
	SessionDuration       = 24 * time.Hour * 30 // 30 days
	PasswordResetDuration = 24 * time.Hour
	EmailVerifyDuration   = 7 * 24 * time.Hour
	MinPasswordLength     = 8
	BcryptCost            = 12
	MaxLoginAttempts      = 5
	LoginLockoutDuration  = 15 * time.Minute
)

// Service provides authentication operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new auth service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// Register creates a new user account
func (s *Service) Register(email, password, firstName, lastName string) (*domain.User, error) {
	if len(password) < MinPasswordLength {
		return nil, ErrPasswordTooShort
	}

	// Check if email already exists
	var existingUser domain.User
	if err := s.db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, ErrEmailExists
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		FirstName:    firstName,
		LastName:     lastName,
		Role:         domain.UserRoleCustomer,
		Status:       domain.UserStatusActive,
		Language:     "en",
		Currency:     "USD",
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and creates a session
func (s *Service) Login(email, password, ipAddress, userAgent string) (*domain.Session, error) {
	// Check for too many failed attempts
	if s.isLockedOut(email, ipAddress) {
		s.logLoginAttempt(email, ipAddress, userAgent, false, "locked_out")
		return nil, ErrTooManyLoginAttempts
	}

	var user domain.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logLoginAttempt(email, ipAddress, userAgent, false, "user_not_found")
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logLoginAttempt(email, ipAddress, userAgent, false, "invalid_password")
		return nil, ErrInvalidCredentials
	}

	// Check user status
	switch user.Status {
	case domain.UserStatusInactive:
		s.logLoginAttempt(email, ipAddress, userAgent, false, "inactive")
		return nil, ErrUserInactive
	case domain.UserStatusSuspended:
		s.logLoginAttempt(email, ipAddress, userAgent, false, "suspended")
		return nil, ErrUserSuspended
	}

	// Create session
	sessionID, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:        sessionID,
		UserID:    user.ID,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		ExpiresAt: time.Now().Add(SessionDuration),
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, err
	}

	// Update last login
	s.db.Model(&user).Updates(map[string]interface{}{
		"last_login_at": time.Now(),
		"last_login_ip": ipAddress,
	})

	// Log successful login
	s.logLoginAttempt(email, ipAddress, userAgent, true, "")

	return session, nil
}

// Logout invalidates a session
func (s *Service) Logout(sessionID string) error {
	return s.db.Delete(&domain.Session{}, "id = ?", sessionID).Error
}

// ValidateSession checks if a session is valid and returns the user
func (s *Service) ValidateSession(sessionID string) (*domain.User, error) {
	var session domain.Session
	if err := s.db.Preload("User").First(&session, "id = ?", sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	if session.IsExpired() {
		s.db.Delete(&session)
		return nil, ErrSessionExpired
	}

	if !session.User.IsActive() {
		return nil, ErrUserInactive
	}

	return &session.User, nil
}

// CreatePasswordResetToken creates a password reset token
func (s *Service) CreatePasswordResetToken(email string) (*domain.PasswordResetToken, error) {
	var user domain.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	token, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	resetToken := &domain.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(PasswordResetDuration),
	}

	if err := s.db.Create(resetToken).Error; err != nil {
		return nil, err
	}

	return resetToken, nil
}

// ResetPassword resets a user's password using a token
func (s *Service) ResetPassword(token, newPassword string) error {
	if len(newPassword) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	var resetToken domain.PasswordResetToken
	if err := s.db.Where("token = ?", token).First(&resetToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidToken
		}
		return err
	}

	if !resetToken.IsValid() {
		return ErrInvalidToken
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		return err
	}

	// Update password
	if err := s.db.Model(&domain.User{}).Where("id = ?", resetToken.UserID).
		Update("password_hash", string(passwordHash)).Error; err != nil {
		return err
	}

	// Mark token as used
	now := time.Now()
	s.db.Model(&resetToken).Update("used_at", &now)

	// Invalidate all sessions for this user
	s.db.Delete(&domain.Session{}, "user_id = ?", resetToken.UserID)

	return nil
}

// ChangePassword changes a user's password
func (s *Service) ChangePassword(userID uint64, currentPassword, newPassword string) error {
	if len(newPassword) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	var user domain.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return ErrUserNotFound
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		return err
	}

	return s.db.Model(&user).Update("password_hash", string(passwordHash)).Error
}

// CreateEmailVerificationToken creates an email verification token
func (s *Service) CreateEmailVerificationToken(userID uint64) (*domain.EmailVerificationToken, error) {
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	verifyToken := &domain.EmailVerificationToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(EmailVerifyDuration),
	}

	if err := s.db.Create(verifyToken).Error; err != nil {
		return nil, err
	}

	return verifyToken, nil
}

// VerifyEmail verifies a user's email address
func (s *Service) VerifyEmail(token string) error {
	var verifyToken domain.EmailVerificationToken
	if err := s.db.Where("token = ?", token).First(&verifyToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidToken
		}
		return err
	}

	if !verifyToken.IsValid() {
		return ErrInvalidToken
	}

	// Update user
	if err := s.db.Model(&domain.User{}).Where("id = ?", verifyToken.UserID).
		Update("email_verified", true).Error; err != nil {
		return err
	}

	// Mark token as used
	now := time.Now()
	s.db.Model(&verifyToken).Update("used_at", &now)

	return nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(id uint64) (*domain.User, error) {
	var user domain.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (s *Service) GetUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// UpdateProfile updates a user's profile
func (s *Service) UpdateProfile(userID uint64, firstName, lastName, company, phone, address1, address2, city, state, postalCode, country string) error {
	updates := map[string]interface{}{
		"first_name":  firstName,
		"last_name":   lastName,
		"company":     company,
		"phone":       phone,
		"address1":    address1,
		"address2":    address2,
		"city":        city,
		"state":       state,
		"postal_code": postalCode,
		"country":     country,
	}

	return s.db.Model(&domain.User{}).Where("id = ?", userID).Updates(updates).Error
}

// CleanupExpiredSessions removes expired sessions
func (s *Service) CleanupExpiredSessions() error {
	return s.db.Delete(&domain.Session{}, "expires_at < ?", time.Now()).Error
}

// logLoginAttempt records a login attempt
func (s *Service) logLoginAttempt(email, ipAddress, userAgent string, success bool, failReason string) {
	attempt := &domain.LoginAttempt{
		Email:      email,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Success:    success,
		FailReason: failReason,
	}
	s.db.Create(attempt)
}

// isLockedOut checks if an IP/email is locked out due to failed attempts
func (s *Service) isLockedOut(email, ipAddress string) bool {
	cutoff := time.Now().Add(-LoginLockoutDuration)
	var count int64
	s.db.Model(&domain.LoginAttempt{}).
		Where("(email = ? OR ip_address = ?) AND success = false AND created_at > ?", email, ipAddress, cutoff).
		Count(&count)
	return count >= MaxLoginAttempts
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
