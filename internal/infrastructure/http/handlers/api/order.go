package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/openhost/openhost/internal/core/domain"
	"github.com/openhost/openhost/internal/core/service/order"
)

// OrderHandler handles order API endpoints
type OrderHandler struct {
	orderService *order.Service
	cartService  *order.CartService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderService *order.Service, cartService *order.CartService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		cartService:  cartService,
	}
}

// ListOrders godoc
// @Summary List orders
// @Description Returns the current user's orders
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} PaginatedResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/orders [get]
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := GetCurrentUserID(c)
	limit, offset := PaginationParams(c)

	orders, total, err := h.orderService.ListOrders(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch orders"})
		return
	}

	var response []OrderResponse
	for _, o := range orders {
		response = append(response, toOrderResponse(&o))
	}

	c.JSON(http.StatusOK, NewPaginatedResponse(response, total, limit, offset))
}

// GetOrder godoc
// @Summary Get order details
// @Description Returns details of a specific order
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} OrderDetailResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid order ID"})
		return
	}

	o, err := h.orderService.GetOrder(orderID)
	if err != nil {
		if err == order.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch order"})
		return
	}

	// Verify ownership (unless admin)
	user := GetCurrentUser(c)
	if o.CustomerID != user.ID && !user.IsAdmin() {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		return
	}

	c.JSON(http.StatusOK, toOrderDetailResponse(o))
}

// CreateOrder godoc
// @Summary Create order from cart
// @Description Creates a new order from the user's shopping cart
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Success 201 {object} OrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := GetCurrentUserID(c)
	ipAddress := c.ClientIP()

	// Get user's cart
	cart, err := h.cartService.GetOrCreateCart(&userID, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get cart"})
		return
	}

	o, err := h.orderService.CreateOrder(userID, cart.ID, ipAddress)
	if err != nil {
		if err == order.ErrCartEmpty {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Cart is empty"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, toOrderResponse(o))
}

// ListServices godoc
// @Summary List services
// @Description Returns the current user's services
// @Tags services
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (pending, active, suspended, terminated)"
// @Param limit query int false "Number of results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} PaginatedResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/services [get]
func (h *OrderHandler) ListServices(c *gin.Context) {
	userID := GetCurrentUserID(c)
	limit, offset := PaginationParams(c)
	status := domain.ServiceStatus(c.Query("status"))

	services, total, err := h.orderService.ListServices(userID, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch services"})
		return
	}

	var response []ServiceResponse
	for _, s := range services {
		response = append(response, toServiceResponse(&s))
	}

	c.JSON(http.StatusOK, NewPaginatedResponse(response, total, limit, offset))
}

// GetService godoc
// @Summary Get service details
// @Description Returns details of a specific service
// @Tags services
// @Produce json
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Success 200 {object} ServiceDetailResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/services/{id} [get]
func (h *OrderHandler) GetService(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid service ID"})
		return
	}

	s, err := h.orderService.GetService(serviceID)
	if err != nil {
		if err == order.ErrServiceNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Service not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch service"})
		return
	}

	// Verify ownership (unless admin)
	user := GetCurrentUser(c)
	if s.CustomerID != user.ID && !user.IsAdmin() {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Service not found"})
		return
	}

	c.JSON(http.StatusOK, toServiceDetailResponse(s))
}

// Cart endpoints

// GetCart godoc
// @Summary Get shopping cart
// @Description Returns the current user's shopping cart
// @Tags cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} CartSummaryResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/cart [get]
func (h *OrderHandler) GetCart(c *gin.Context) {
	var customerID *uint64
	sessionID := ""

	user := GetCurrentUser(c)
	if user != nil {
		customerID = &user.ID
	} else {
		sessionID = c.GetHeader("X-Session-ID")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Session ID required for guest cart"})
			return
		}
	}

	cart, err := h.cartService.GetOrCreateCart(customerID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get cart"})
		return
	}

	summary, err := h.cartService.GetCartSummary(cart.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get cart summary"})
		return
	}

	c.JSON(http.StatusOK, toCartSummaryResponse(summary))
}

// AddToCart godoc
// @Summary Add item to cart
// @Description Adds a product to the shopping cart
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddToCartRequest true "Cart item data"
// @Success 201 {object} CartItemResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/cart/items [post]
func (h *OrderHandler) AddToCart(c *gin.Context) {
	var customerID *uint64
	sessionID := ""

	user := GetCurrentUser(c)
	if user != nil {
		customerID = &user.ID
	} else {
		sessionID = c.GetHeader("X-Session-ID")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Session ID required for guest cart"})
			return
		}
	}

	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	cart, err := h.cartService.GetOrCreateCart(customerID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get cart"})
		return
	}

	item, err := h.cartService.AddItem(cart.ID, req.ProductID, req.Quantity, req.BillingCycle, req.Domain, req.Hostname, req.ConfigOptions)
	if err != nil {
		switch err {
		case order.ErrPricingNotFound:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Pricing not configured for this product"})
			return
		case order.ErrInvalidBillingCycle:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Billing cycle not available"})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, CartItemResponse{
		ID:           item.ID,
		ProductID:    item.ProductID,
		Quantity:     item.Quantity,
		BillingCycle: item.BillingCycle,
		SetupFee:     item.SetupFee.String(),
		RecurringFee: item.RecurringFee.String(),
		Discount:     item.Discount.String(),
		Total:        item.Total.String(),
	})
}

// UpdateCartItem godoc
// @Summary Update cart item
// @Description Updates the quantity of a cart item
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Cart item ID"
// @Param request body UpdateCartItemRequest true "Update data"
// @Success 200 {object} CartItemResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/cart/items/{id} [put]
func (h *OrderHandler) UpdateCartItem(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid item ID"})
		return
	}

	var req UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := h.cartService.UpdateItem(itemID, req.Quantity)
	if err != nil {
		if err == order.ErrCartItemNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Cart item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update cart item"})
		return
	}

	if item == nil {
		c.JSON(http.StatusOK, MessageResponse{Message: "Item removed from cart"})
		return
	}

	c.JSON(http.StatusOK, CartItemResponse{
		ID:           item.ID,
		ProductID:    item.ProductID,
		Quantity:     item.Quantity,
		BillingCycle: item.BillingCycle,
		SetupFee:     item.SetupFee.String(),
		RecurringFee: item.RecurringFee.String(),
		Discount:     item.Discount.String(),
		Total:        item.Total.String(),
	})
}

// RemoveCartItem godoc
// @Summary Remove cart item
// @Description Removes an item from the shopping cart
// @Tags cart
// @Produce json
// @Security BearerAuth
// @Param id path int true "Cart item ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/cart/items/{id} [delete]
func (h *OrderHandler) RemoveCartItem(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid item ID"})
		return
	}

	if err := h.cartService.RemoveItem(itemID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to remove item"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Item removed from cart"})
}

// ApplyCoupon godoc
// @Summary Apply coupon to cart
// @Description Applies a coupon code to the shopping cart
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ApplyCouponRequest true "Coupon code"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/cart/coupon [post]
func (h *OrderHandler) ApplyCoupon(c *gin.Context) {
	var customerID *uint64
	sessionID := ""

	user := GetCurrentUser(c)
	if user != nil {
		customerID = &user.ID
	} else {
		sessionID = c.GetHeader("X-Session-ID")
	}

	var req ApplyCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	cart, err := h.cartService.GetOrCreateCart(customerID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get cart"})
		return
	}

	if err := h.cartService.ApplyCoupon(cart.ID, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid or expired coupon"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Coupon applied successfully"})
}

// RemoveCoupon godoc
// @Summary Remove coupon from cart
// @Description Removes the applied coupon from the shopping cart
// @Tags cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Router /api/v1/cart/coupon [delete]
func (h *OrderHandler) RemoveCoupon(c *gin.Context) {
	var customerID *uint64
	sessionID := ""

	user := GetCurrentUser(c)
	if user != nil {
		customerID = &user.ID
	} else {
		sessionID = c.GetHeader("X-Session-ID")
	}

	cart, err := h.cartService.GetOrCreateCart(customerID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get cart"})
		return
	}

	_ = h.cartService.RemoveCoupon(cart.ID)
	c.JSON(http.StatusOK, MessageResponse{Message: "Coupon removed"})
}

// ClearCart godoc
// @Summary Clear cart
// @Description Removes all items from the shopping cart
// @Tags cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Router /api/v1/cart [delete]
func (h *OrderHandler) ClearCart(c *gin.Context) {
	var customerID *uint64
	sessionID := ""

	user := GetCurrentUser(c)
	if user != nil {
		customerID = &user.ID
	} else {
		sessionID = c.GetHeader("X-Session-ID")
	}

	cart, err := h.cartService.GetOrCreateCart(customerID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get cart"})
		return
	}

	_ = h.cartService.ClearCart(cart.ID)
	c.JSON(http.StatusOK, MessageResponse{Message: "Cart cleared"})
}

// Admin endpoints

// AdminListOrders godoc
// @Summary List all orders (Admin)
// @Description Returns all orders in the system
// @Tags admin/orders
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status"
// @Param customer_id query int false "Filter by customer"
// @Param limit query int false "Number of results per page" default(20)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} PaginatedResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/admin/orders [get]
func (h *OrderHandler) AdminListOrders(c *gin.Context) {
	limit, offset := PaginationParams(c)
	status := domain.OrderStatus(c.Query("status"))

	orders, total, err := h.orderService.ListAllOrders(status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch orders"})
		return
	}

	var response []OrderResponse
	for _, o := range orders {
		response = append(response, toOrderResponse(&o))
	}

	c.JSON(http.StatusOK, NewPaginatedResponse(response, total, limit, offset))
}

// AdminUpdateOrderStatus godoc
// @Summary Update order status (Admin)
// @Description Updates the status of an order
// @Tags admin/orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body UpdateOrderStatusRequest true "Status data"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/admin/orders/{id}/status [put]
func (h *OrderHandler) AdminUpdateOrderStatus(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid order ID"})
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.orderService.UpdateOrderStatus(orderID, domain.OrderStatus(req.Status)); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Order status updated"})
}

// AdminSuspendService godoc
// @Summary Suspend service (Admin)
// @Description Suspends a customer's service
// @Tags admin/services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Param request body SuspendServiceRequest true "Suspension reason"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/admin/services/{id}/suspend [post]
func (h *OrderHandler) AdminSuspendService(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid service ID"})
		return
	}

	var req SuspendServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.orderService.SuspendService(serviceID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to suspend service"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Service suspended"})
}

// AdminUnsuspendService godoc
// @Summary Unsuspend service (Admin)
// @Description Unsuspends a customer's service
// @Tags admin/services
// @Produce json
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/admin/services/{id}/unsuspend [post]
func (h *OrderHandler) AdminUnsuspendService(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid service ID"})
		return
	}

	if err := h.orderService.UnsuspendService(serviceID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to unsuspend service"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Service unsuspended"})
}

// AdminTerminateService godoc
// @Summary Terminate service (Admin)
// @Description Terminates a customer's service
// @Tags admin/services
// @Produce json
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/admin/services/{id}/terminate [post]
func (h *OrderHandler) AdminTerminateService(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid service ID"})
		return
	}

	if err := h.orderService.TerminateService(serviceID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to terminate service"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Service terminated"})
}

// Helper functions

func toOrderResponse(o *domain.Order) OrderResponse {
	return OrderResponse{
		ID:          o.ID,
		OrderNumber: o.OrderNumber,
		Status:      string(o.Status),
		Currency:    o.Currency,
		Total:       o.Total.String(),
		CreatedAt:   o.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toOrderDetailResponse(o *domain.Order) OrderDetailResponse {
	var items []OrderItemResponse
	for _, item := range o.Items {
		items = append(items, OrderItemResponse{
			ID:           item.ID,
			ProductID:    item.ProductID,
			Description:  item.Description,
			Quantity:     item.Quantity,
			BillingCycle: item.BillingCycle,
			SetupFee:     item.SetupFee.String(),
			RecurringFee: item.RecurringFee.String(),
			Discount:     item.Discount.String(),
			Total:        item.Total.String(),
		})
	}

	return OrderDetailResponse{
		ID:          o.ID,
		OrderNumber: o.OrderNumber,
		Status:      string(o.Status),
		Currency:    o.Currency,
		Subtotal:    o.Subtotal.String(),
		Discount:    o.Discount.String(),
		TaxAmount:   o.TaxAmount.String(),
		Total:       o.Total.String(),
		Items:       items,
		CreatedAt:   o.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toServiceResponse(s *domain.Service) ServiceResponse {
	return ServiceResponse{
		ID:              s.ID,
		ProductID:       s.ProductID,
		ProductName:     s.Product.Name,
		Status:          string(s.Status),
		Domain:          s.Domain,
		Hostname:        s.Hostname,
		BillingCycle:    s.BillingCycle,
		NextDueDate:     s.NextDueDate.Format("2006-01-02"),
		RecurringAmount: s.RecurringAmount.String(),
		Currency:        s.Currency,
	}
}

func toServiceDetailResponse(s *domain.Service) ServiceDetailResponse {
	resp := ServiceDetailResponse{
		ID:               s.ID,
		ProductID:        s.ProductID,
		ProductName:      s.Product.Name,
		Status:           string(s.Status),
		Domain:           s.Domain,
		Hostname:         s.Hostname,
		Username:         s.Username,
		BillingCycle:     s.BillingCycle,
		Currency:         s.Currency,
		RecurringAmount:  s.RecurringAmount.String(),
		NextDueDate:      s.NextDueDate.Format("2006-01-02"),
		RegistrationDate: s.RegistrationDate.Format("2006-01-02"),
		Notes:            s.Notes,
	}

	if s.IPAddress != nil {
		resp.IPAddress = s.IPAddress.IP
	}

	if s.SuspensionReason != "" {
		resp.SuspensionReason = s.SuspensionReason
	}

	return resp
}

func toCartSummaryResponse(summary *order.CartSummary) CartSummaryResponse {
	var items []CartItemSummaryResponse
	for _, item := range summary.Items {
		items = append(items, CartItemSummaryResponse{
			ID:           item.ID,
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			Quantity:     item.Quantity,
			BillingCycle: item.BillingCycle,
			SetupFee:     item.SetupFee.String(),
			RecurringFee: item.RecurringFee.String(),
			Discount:     item.Discount.String(),
			Total:        item.Total.String(),
		})
	}

	return CartSummaryResponse{
		CartID:        summary.CartID,
		Currency:      summary.Currency,
		Items:         items,
		Subtotal:      summary.Subtotal.String(),
		TotalDiscount: summary.TotalDiscount.String(),
		Tax:           summary.Tax.String(),
		Total:         summary.Total.String(),
		CouponCode:    summary.CouponCode,
	}
}

// Request/Response types

type OrderResponse struct {
	ID          uint64 `json:"id"`
	OrderNumber string `json:"order_number"`
	Status      string `json:"status"`
	Currency    string `json:"currency"`
	Total       string `json:"total"`
	CreatedAt   string `json:"created_at"`
}

type OrderDetailResponse struct {
	ID          uint64              `json:"id"`
	OrderNumber string              `json:"order_number"`
	Status      string              `json:"status"`
	Currency    string              `json:"currency"`
	Subtotal    string              `json:"subtotal"`
	Discount    string              `json:"discount"`
	TaxAmount   string              `json:"tax_amount"`
	Total       string              `json:"total"`
	Items       []OrderItemResponse `json:"items"`
	CreatedAt   string              `json:"created_at"`
}

type OrderItemResponse struct {
	ID           uint64 `json:"id"`
	ProductID    uint64 `json:"product_id"`
	Description  string `json:"description"`
	Quantity     int    `json:"quantity"`
	BillingCycle string `json:"billing_cycle"`
	SetupFee     string `json:"setup_fee"`
	RecurringFee string `json:"recurring_fee"`
	Discount     string `json:"discount"`
	Total        string `json:"total"`
}

type ServiceResponse struct {
	ID              uint64 `json:"id"`
	ProductID       uint64 `json:"product_id"`
	ProductName     string `json:"product_name"`
	Status          string `json:"status"`
	Domain          string `json:"domain,omitempty"`
	Hostname        string `json:"hostname,omitempty"`
	BillingCycle    string `json:"billing_cycle"`
	NextDueDate     string `json:"next_due_date"`
	RecurringAmount string `json:"recurring_amount"`
	Currency        string `json:"currency"`
}

type ServiceDetailResponse struct {
	ID               uint64 `json:"id"`
	ProductID        uint64 `json:"product_id"`
	ProductName      string `json:"product_name"`
	Status           string `json:"status"`
	Domain           string `json:"domain,omitempty"`
	Hostname         string `json:"hostname,omitempty"`
	Username         string `json:"username,omitempty"`
	IPAddress        string `json:"ip_address,omitempty"`
	BillingCycle     string `json:"billing_cycle"`
	Currency         string `json:"currency"`
	RecurringAmount  string `json:"recurring_amount"`
	NextDueDate      string `json:"next_due_date"`
	RegistrationDate string `json:"registration_date"`
	SuspensionReason string `json:"suspension_reason,omitempty"`
	Notes            string `json:"notes,omitempty"`
}

type CartSummaryResponse struct {
	CartID        uint64                    `json:"cart_id"`
	Currency      string                    `json:"currency"`
	Items         []CartItemSummaryResponse `json:"items"`
	Subtotal      string                    `json:"subtotal"`
	TotalDiscount string                    `json:"total_discount"`
	Tax           string                    `json:"tax"`
	Total         string                    `json:"total"`
	CouponCode    string                    `json:"coupon_code,omitempty"`
}

type CartItemSummaryResponse struct {
	ID           uint64 `json:"id"`
	ProductID    uint64 `json:"product_id"`
	ProductName  string `json:"product_name"`
	Quantity     int    `json:"quantity"`
	BillingCycle string `json:"billing_cycle"`
	SetupFee     string `json:"setup_fee"`
	RecurringFee string `json:"recurring_fee"`
	Discount     string `json:"discount"`
	Total        string `json:"total"`
}

type CartItemResponse struct {
	ID           uint64 `json:"id"`
	ProductID    uint64 `json:"product_id"`
	Quantity     int    `json:"quantity"`
	BillingCycle string `json:"billing_cycle"`
	SetupFee     string `json:"setup_fee"`
	RecurringFee string `json:"recurring_fee"`
	Discount     string `json:"discount"`
	Total        string `json:"total"`
}

type AddToCartRequest struct {
	ProductID     uint64         `json:"product_id" binding:"required"`
	Quantity      int            `json:"quantity"`
	BillingCycle  string         `json:"billing_cycle" binding:"required"`
	Domain        string         `json:"domain"`
	Hostname      string         `json:"hostname"`
	ConfigOptions domain.JSONMap `json:"config_options"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required"`
}

type ApplyCouponRequest struct {
	Code string `json:"code" binding:"required"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type SuspendServiceRequest struct {
	Reason string `json:"reason"`
}
