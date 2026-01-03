package service

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"

	"xlpanel/internal/core"
	"xlpanel/internal/domain"
	"xlpanel/internal/infra"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalid2FACode     = errors.New("invalid 2FA code")
)

type AuthService struct {
	users     *infra.Repository[domain.User]
	sessions  *infra.Repository[domain.Session]
	jwtSecret string
	jwtExpiry time.Duration
}

func NewAuthService(users *infra.Repository[domain.User], sessions *infra.Repository[domain.Session], jwtSecret string, jwtExpiry time.Duration) *AuthService {
	return &AuthService{
		users:     users,
		sessions:  sessions,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// Register creates a new user account
func (s *AuthService) Register(tenantID, email, password, name, role string) (*domain.User, error) {
	// Check if user already exists
	for _, u := range s.users.List() {
		if u.Email == email && u.TenantID == tenantID {
			return nil, ErrUserExists
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := domain.User{
		ID:           core.NewID(),
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	}

	s.users.Add(user)
	return &user, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(tenantID, email, password string) (string, *domain.User, error) {
	var foundUser *domain.User
	for _, u := range s.users.List() {
		if u.Email == email && u.TenantID == tenantID {
			userCopy := u // Create a copy to avoid pointer issues
			foundUser = &userCopy
			break
		}
	}

	if foundUser == nil {
		return "", nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateJWT(foundUser)
	if err != nil {
		return "", nil, err
	}

	// Update last login time
	now := time.Now().UTC()
	foundUser.LastLoginAt = &now
	s.users.Add(*foundUser)

	// Create session
	session := domain.Session{
		ID:        core.NewID(),
		UserID:    foundUser.ID,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(s.jwtExpiry),
		CreatedAt: time.Now().UTC(),
	}
	s.sessions.Add(session)

	return token, foundUser, nil
}

// VerifyToken validates a JWT token and returns the user
func (s *AuthService) VerifyToken(tokenString string) (*domain.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	user, ok := s.users.Get(userID)
	if !ok {
		return nil, ErrUserNotFound
	}

	return &user, nil
}

// Enable2FA enables two-factor authentication for a user
func (s *AuthService) Enable2FA(userID string) (string, error) {
	user, ok := s.users.Get(userID)
	if !ok {
		return "", ErrUserNotFound
	}

	// Generate a new TOTP secret
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	secretBase32 := base32.StdEncoding.EncodeToString(secret)

	user.TwoFactorSecret = secretBase32
	user.TwoFactorEnabled = true
	s.users.Add(user)

	return secretBase32, nil
}

// Verify2FA validates a 2FA code
func (s *AuthService) Verify2FA(userID, code string) error {
	user, ok := s.users.Get(userID)
	if !ok {
		return ErrUserNotFound
	}

	if !user.TwoFactorEnabled {
		return errors.New("2FA not enabled for this user")
	}

	valid := totp.Validate(code, user.TwoFactorSecret)
	if !valid {
		return ErrInvalid2FACode
	}

	return nil
}

// Disable2FA disables two-factor authentication for a user
func (s *AuthService) Disable2FA(userID string) error {
	user, ok := s.users.Get(userID)
	if !ok {
		return ErrUserNotFound
	}

	user.TwoFactorSecret = ""
	user.TwoFactorEnabled = false
	s.users.Add(user)

	return nil
}

func (s *AuthService) generateJWT(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"email":     user.Email,
		"role":      user.Role,
		"tenant_id": user.TenantID,
		"exp":       time.Now().Add(s.jwtExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// GetUserByEmail retrieves a user by email
func (s *AuthService) GetUserByEmail(tenantID, email string) (*domain.User, error) {
	for _, u := range s.users.List() {
		if u.Email == email && u.TenantID == tenantID {
			userCopy := u // Create a copy to avoid pointer issues
			return &userCopy, nil
		}
	}
	return nil, ErrUserNotFound
}

// ListUsers returns all users
func (s *AuthService) ListUsers() []domain.User {
	return s.users.List()
}
