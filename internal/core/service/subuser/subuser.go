package subuser

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
	ErrSubUserNotFound       = errors.New("sub-user not found")
	ErrSubUserExists         = errors.New("email already registered")
	ErrSubUserInactive       = errors.New("sub-user account is inactive")
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrInviteNotFound        = errors.New("invitation not found")
	ErrInviteExpired         = errors.New("invitation has expired")
	ErrInviteAlreadyUsed     = errors.New("invitation has already been used")
	ErrPermissionDenied      = errors.New("permission denied")
)

const (
	SessionDuration   = 24 * time.Hour * 30
	InviteDuration    = 7 * 24 * time.Hour
	MinPasswordLength = 8
	BcryptCost        = 12
)

// Service provides sub-user management operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new sub-user service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateInvite creates an invitation for a new sub-user
func (s *Service) CreateInvite(customerID, invitedBy uint64, email string, role domain.SubUserRole, permissions domain.SubUserPermissions) (*domain.SubUserInvite, error) {
	// Check if email already exists
	var existingSubUser domain.SubUser
	if err := s.db.Where("email = ?", email).First(&existingSubUser).Error; err == nil {
		return nil, ErrSubUserExists
	}

	// Generate invite token
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	invite := &domain.SubUserInvite{
		CustomerID:  customerID,
		Email:       email,
		Token:       token,
		Role:        role,
		Permissions: permissions,
		InvitedBy:   invitedBy,
		ExpiresAt:   time.Now().Add(InviteDuration),
	}

	if err := s.db.Create(invite).Error; err != nil {
		return nil, err
	}

	return invite, nil
}

// AcceptInvite accepts an invitation and creates the sub-user
func (s *Service) AcceptInvite(token, password, firstName, lastName, phone string) (*domain.SubUser, error) {
	var invite domain.SubUserInvite
	if err := s.db.Where("token = ?", token).First(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInviteNotFound
		}
		return nil, err
	}

	if invite.UsedAt != nil {
		return nil, ErrInviteAlreadyUsed
	}

	if time.Now().After(invite.ExpiresAt) {
		return nil, ErrInviteExpired
	}

	if len(password) < MinPasswordLength {
		return nil, errors.New("password must be at least 8 characters")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return nil, err
	}

	subUser := &domain.SubUser{
		CustomerID:   invite.CustomerID,
		Email:        invite.Email,
		PasswordHash: string(passwordHash),
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        phone,
		Role:         invite.Role,
		Permissions:  invite.Permissions,
		Active:       true,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(subUser).Error; err != nil {
			return err
		}

		// Mark invite as used
		now := time.Now()
		return tx.Model(&invite).Update("used_at", &now).Error
	})

	if err != nil {
		return nil, err
	}

	return subUser, nil
}

// Login authenticates a sub-user
func (s *Service) Login(email, password, ipAddress, userAgent string) (*domain.SubUserSession, error) {
	var subUser domain.SubUser
	if err := s.db.Where("email = ?", email).First(&subUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !subUser.Active {
		return nil, ErrSubUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(subUser.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Create session
	sessionID, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	session := &domain.SubUserSession{
		ID:        sessionID,
		SubUserID: subUser.ID,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		ExpiresAt: time.Now().Add(SessionDuration),
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, err
	}

	// Update last login
	now := time.Now()
	s.db.Model(&subUser).Updates(map[string]interface{}{
		"last_login_at": &now,
		"last_login_ip": ipAddress,
	})

	return session, nil
}

// ValidateSession validates a sub-user session
func (s *Service) ValidateSession(sessionID string) (*domain.SubUser, error) {
	var session domain.SubUserSession
	if err := s.db.Preload("SubUser").Where("id = ?", sessionID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid session")
		}
		return nil, err
	}

	if session.IsExpired() {
		s.db.Delete(&session)
		return nil, errors.New("session expired")
	}

	if !session.SubUser.Active {
		return nil, ErrSubUserInactive
	}

	return &session.SubUser, nil
}

// Logout invalidates a sub-user session
func (s *Service) Logout(sessionID string) error {
	return s.db.Delete(&domain.SubUserSession{}, "id = ?", sessionID).Error
}

// GetSubUser retrieves a sub-user by ID
func (s *Service) GetSubUser(id uint64) (*domain.SubUser, error) {
	var subUser domain.SubUser
	if err := s.db.Preload("Customer").First(&subUser, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubUserNotFound
		}
		return nil, err
	}
	return &subUser, nil
}

// ListSubUsers lists sub-users for a customer
func (s *Service) ListSubUsers(customerID uint64) ([]domain.SubUser, error) {
	var subUsers []domain.SubUser
	if err := s.db.Where("customer_id = ?", customerID).Find(&subUsers).Error; err != nil {
		return nil, err
	}
	return subUsers, nil
}

// UpdateSubUser updates a sub-user
func (s *Service) UpdateSubUser(id uint64, firstName, lastName, phone string, role domain.SubUserRole, permissions domain.SubUserPermissions, active bool) error {
	return s.db.Model(&domain.SubUser{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"first_name":  firstName,
			"last_name":   lastName,
			"phone":       phone,
			"role":        role,
			"permissions": permissions,
			"active":      active,
		}).Error
}

// DeleteSubUser deletes a sub-user
func (s *Service) DeleteSubUser(id, customerID uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete sessions
		if err := tx.Delete(&domain.SubUserSession{}, "sub_user_id = ?", id).Error; err != nil {
			return err
		}
		// Delete sub-user
		return tx.Where("id = ? AND customer_id = ?", id, customerID).Delete(&domain.SubUser{}).Error
	})
}

// ChangePassword changes a sub-user's password
func (s *Service) ChangePassword(id uint64, currentPassword, newPassword string) error {
	var subUser domain.SubUser
	if err := s.db.First(&subUser, id).Error; err != nil {
		return ErrSubUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(subUser.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	if len(newPassword) < MinPasswordLength {
		return errors.New("password must be at least 8 characters")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		return err
	}

	return s.db.Model(&subUser).Update("password_hash", string(passwordHash)).Error
}

// HasPermission checks if a sub-user has a specific permission
func (s *Service) HasPermission(subUser *domain.SubUser, permission string) bool {
	switch permission {
	case "view_services":
		return subUser.Permissions.ViewServices
	case "manage_services":
		return subUser.Permissions.ManageServices
	case "view_invoices":
		return subUser.Permissions.ViewInvoices
	case "pay_invoices":
		return subUser.Permissions.PayInvoices
	case "view_tickets":
		return subUser.Permissions.ViewTickets
	case "create_tickets":
		return subUser.Permissions.CreateTickets
	case "reply_tickets":
		return subUser.Permissions.ReplyTickets
	case "view_profile":
		return subUser.Permissions.ViewProfile
	case "edit_profile":
		return subUser.Permissions.EditProfile
	case "manage_payments":
		return subUser.Permissions.ManagePayments
	case "place_orders":
		return subUser.Permissions.PlaceOrders
	case "manage_sub_users":
		return subUser.Permissions.ManageSubUsers
	case "view_affiliates":
		return subUser.Permissions.ViewAffiliates
	case "manage_domains":
		return subUser.Permissions.ManageDomains
	case "access_api":
		return subUser.Permissions.AccessAPI
	default:
		return false
	}
}

// CheckPermission checks permission and returns an error if denied
func (s *Service) CheckPermission(subUser *domain.SubUser, permission string) error {
	if !s.HasPermission(subUser, permission) {
		return ErrPermissionDenied
	}
	return nil
}

// GetPendingInvites gets pending invites for a customer
func (s *Service) GetPendingInvites(customerID uint64) ([]domain.SubUserInvite, error) {
	var invites []domain.SubUserInvite
	if err := s.db.Where("customer_id = ? AND used_at IS NULL AND expires_at > ?", customerID, time.Now()).
		Find(&invites).Error; err != nil {
		return nil, err
	}
	return invites, nil
}

// CancelInvite cancels a pending invite
func (s *Service) CancelInvite(inviteID, customerID uint64) error {
	return s.db.Where("id = ? AND customer_id = ? AND used_at IS NULL", inviteID, customerID).
		Delete(&domain.SubUserInvite{}).Error
}

// ResendInvite resends an invite by extending its expiration
func (s *Service) ResendInvite(inviteID, customerID uint64) error {
	return s.db.Model(&domain.SubUserInvite{}).
		Where("id = ? AND customer_id = ? AND used_at IS NULL", inviteID, customerID).
		Update("expires_at", time.Now().Add(InviteDuration)).Error
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// DefaultPermissionsForRole returns default permissions for a role
func DefaultPermissionsForRole(role domain.SubUserRole) domain.SubUserPermissions {
	switch role {
	case domain.SubUserRoleBilling:
		return domain.SubUserPermissions{
			ViewServices:   true,
			ViewInvoices:   true,
			PayInvoices:    true,
			ViewTickets:    false,
			CreateTickets:  false,
			ReplyTickets:   false,
			ViewProfile:    true,
			EditProfile:    false,
			ManagePayments: true,
			PlaceOrders:    true,
			ManageSubUsers: false,
			ViewAffiliates: false,
			ManageDomains:  false,
			AccessAPI:      false,
		}
	case domain.SubUserRoleTechnical:
		return domain.SubUserPermissions{
			ViewServices:   true,
			ManageServices: true,
			ViewInvoices:   false,
			PayInvoices:    false,
			ViewTickets:    true,
			CreateTickets:  true,
			ReplyTickets:   true,
			ViewProfile:    true,
			EditProfile:    false,
			ManagePayments: false,
			PlaceOrders:    false,
			ManageSubUsers: false,
			ViewAffiliates: false,
			ManageDomains:  true,
			AccessAPI:      true,
		}
	case domain.SubUserRoleAdmin:
		return domain.SubUserPermissions{
			ViewServices:   true,
			ManageServices: true,
			ViewInvoices:   true,
			PayInvoices:    true,
			ViewTickets:    true,
			CreateTickets:  true,
			ReplyTickets:   true,
			ViewProfile:    true,
			EditProfile:    true,
			ManagePayments: true,
			PlaceOrders:    true,
			ManageSubUsers: true,
			ViewAffiliates: true,
			ManageDomains:  true,
			AccessAPI:      true,
		}
	default:
		return domain.SubUserPermissions{}
	}
}
