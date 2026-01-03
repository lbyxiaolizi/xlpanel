package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/openhost/openhost/internal/core/domain"
	"github.com/openhost/openhost/internal/core/service/subuser"
)

// SubUserHandler handles sub-user API endpoints
type SubUserHandler struct {
	service *subuser.Service
}

// NewSubUserHandler creates a new sub-user handler
func NewSubUserHandler(service *subuser.Service) *SubUserHandler {
	return &SubUserHandler{service: service}
}

// ListSubUsers lists sub-users for the current customer
// @Summary List sub-users
// @Description Get a list of sub-users for the current customer
// @Tags SubUsers
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/subusers [get]
func (h *SubUserHandler) ListSubUsers(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	subUsers, err := h.service.ListSubUsers(customerID.(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sub_users": subUsers})
}

// CreateInvite creates an invitation for a new sub-user
// @Summary Invite sub-user
// @Description Create an invitation for a new sub-user
// @Tags SubUsers
// @Accept json
// @Produce json
// @Param request body CreateInviteRequest true "Invite request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/subusers/invite [post]
func (h *SubUserHandler) CreateInvite(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, _ := c.Get("user_id") // Inviter

	var req CreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permissions := subuser.DefaultPermissionsForRole(req.Role)
	if req.Permissions != nil {
		permissions = *req.Permissions
	}

	invite, err := h.service.CreateInvite(customerID.(uint64), userID.(uint64), req.Email, req.Role, permissions)
	if err != nil {
		if err == subuser.ErrSubUserExists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation sent",
		"invite":  invite,
	})
}

// AcceptInvite accepts a sub-user invitation
// @Summary Accept invitation
// @Description Accept an invitation to become a sub-user
// @Tags SubUsers
// @Accept json
// @Produce json
// @Param token path string true "Invitation token"
// @Param request body AcceptInviteRequest true "Accept request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/subusers/invite/{token}/accept [post]
func (h *SubUserHandler) AcceptInvite(c *gin.Context) {
	token := c.Param("token")

	var req AcceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subUser, err := h.service.AcceptInvite(token, req.Password, req.FirstName, req.LastName, req.Phone)
	if err != nil {
		switch err {
		case subuser.ErrInviteNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
		case subuser.ErrInviteExpired:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invitation has expired"})
		case subuser.ErrInviteAlreadyUsed:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invitation has already been used"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Account created successfully",
		"sub_user": subUser,
	})
}

// UpdateSubUser updates a sub-user
// @Summary Update sub-user
// @Description Update a sub-user's information and permissions
// @Tags SubUsers
// @Accept json
// @Produce json
// @Param id path int true "Sub-user ID"
// @Param request body UpdateSubUserRequest true "Update request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/subusers/{id} [put]
func (h *SubUserHandler) UpdateSubUser(c *gin.Context) {
	subUserID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sub-user ID"})
		return
	}

	var req UpdateSubUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateSubUser(subUserID, req.FirstName, req.LastName, req.Phone, req.Role, req.Permissions, req.Active); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sub-user updated"})
}

// DeleteSubUser deletes a sub-user
// @Summary Delete sub-user
// @Description Delete a sub-user
// @Tags SubUsers
// @Produce json
// @Param id path int true "Sub-user ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/subusers/{id} [delete]
func (h *SubUserHandler) DeleteSubUser(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	subUserID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sub-user ID"})
		return
	}

	if err := h.service.DeleteSubUser(subUserID, customerID.(uint64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sub-user deleted"})
}

// SubUserLogin authenticates a sub-user
// @Summary Sub-user login
// @Description Authenticate a sub-user
// @Tags SubUsers
// @Accept json
// @Produce json
// @Param request body SubUserLoginRequest true "Login request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/subusers/login [post]
func (h *SubUserHandler) SubUserLogin(c *gin.Context) {
	var req SubUserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.service.Login(req.Email, req.Password, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		if err == subuser.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		if err == subuser.ErrSubUserInactive {
			c.JSON(http.StatusForbidden, gin.H{"error": "Account is inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Login successful",
		"session_id": session.ID,
		"expires_at": session.ExpiresAt,
	})
}

// SubUserLogout logs out a sub-user
// @Summary Sub-user logout
// @Description Log out a sub-user
// @Tags SubUsers
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/subusers/logout [post]
func (h *SubUserHandler) SubUserLogout(c *gin.Context) {
	sessionID, exists := c.Get("sub_user_session_id")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
		return
	}

	h.service.Logout(sessionID.(string))
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

// ChangePassword changes a sub-user's password
// @Summary Change password
// @Description Change the sub-user's password
// @Tags SubUsers
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Password change request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/subusers/password [put]
func (h *SubUserHandler) ChangePassword(c *gin.Context) {
	subUserID, exists := c.Get("sub_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req SubUserChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ChangePassword(subUserID.(uint64), req.CurrentPassword, req.NewPassword); err != nil {
		if err == subuser.ErrInvalidCredentials {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed"})
}

// GetPendingInvites lists pending invitations
// @Summary List pending invites
// @Description Get a list of pending sub-user invitations
// @Tags SubUsers
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/subusers/invites [get]
func (h *SubUserHandler) GetPendingInvites(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	invites, err := h.service.GetPendingInvites(customerID.(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invites": invites})
}

// CancelInvite cancels a pending invitation
// @Summary Cancel invite
// @Description Cancel a pending sub-user invitation
// @Tags SubUsers
// @Produce json
// @Param id path int true "Invite ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/subusers/invites/{id} [delete]
func (h *SubUserHandler) CancelInvite(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	inviteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite ID"})
		return
	}

	if err := h.service.CancelInvite(inviteID, customerID.(uint64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invite cancelled"})
}

// Request/Response types
type CreateInviteRequest struct {
	Email       string                      `json:"email" binding:"required,email"`
	Role        domain.SubUserRole          `json:"role" binding:"required"`
	Permissions *domain.SubUserPermissions  `json:"permissions"`
}

type AcceptInviteRequest struct {
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
}

type UpdateSubUserRequest struct {
	FirstName   string                     `json:"first_name" binding:"required"`
	LastName    string                     `json:"last_name" binding:"required"`
	Phone       string                     `json:"phone"`
	Role        domain.SubUserRole         `json:"role" binding:"required"`
	Permissions domain.SubUserPermissions  `json:"permissions" binding:"required"`
	Active      bool                       `json:"active"`
}

type SubUserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SubUserChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}
