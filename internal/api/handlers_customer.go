package api

import (
	"fmt"
	"net/http"
	"time"

	"xlpanel/internal/core"
	"xlpanel/internal/domain"
	"xlpanel/internal/i18n"
)

// handleCustomerDashboard renders the customer dashboard
func (s *Server) handleCustomerDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	translator := i18n.NewTranslator()
	
	data := map[string]interface{}{
		"Title": translator.T(lang, "dashboard.title"),
		"Lang":  lang,
		"User":  user,
		"T": func(key string) string {
			return translator.T(lang, key)
		},
	}

	s.renderTemplate(w, r, "customer/dashboard.html", data)
}

// handleCustomerProducts renders the customer product catalog
func (s *Server) handleCustomerProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	translator := i18n.NewTranslator()
	
	data := map[string]interface{}{
		"Title": translator.T(lang, "products.title"),
		"Lang":  lang,
		"User":  user,
		"T": func(key string) string {
			return translator.T(lang, key)
		},
	}

	s.renderTemplate(w, r, "customer/products.html", data)
}

// handleCustomerCart renders the customer cart
func (s *Server) handleCustomerCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	translator := i18n.NewTranslator()
	
	data := map[string]interface{}{
		"Title": translator.T(lang, "cart.title"),
		"Lang":  lang,
		"User":  user,
		"T": func(key string) string {
			return translator.T(lang, key)
		},
	}

	s.renderTemplate(w, r, "customer/cart.html", data)
}

// handleCustomerOrders renders the customer orders page
func (s *Server) handleCustomerOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	translator := i18n.NewTranslator()
	
	data := map[string]interface{}{
		"Title": translator.T(lang, "orders.title"),
		"Lang":  lang,
		"User":  user,
		"T": func(key string) string {
			return translator.T(lang, key)
		},
	}

	s.renderTemplate(w, r, "customer/orders.html", data)
}

// handleCustomerInvoices renders the customer invoices page
func (s *Server) handleCustomerInvoices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}

	translator := i18n.NewTranslator()
	
	data := map[string]interface{}{
		"Title": translator.T(lang, "invoices.title"),
		"Lang":  lang,
		"User":  user,
		"T": func(key string) string {
			return translator.T(lang, key)
		},
	}

	s.renderTemplate(w, r, "customer/invoices.html", data)
}

// Cart API handlers

// handleCartGet retrieves the current user's cart
func (s *Server) handleCartGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cart := s.cartService.GetOrCreateCart(user.ID, user.TenantID)
	writeJSON(w, http.StatusOK, cart)
}

// handleCartAdd adds an item to the cart
func (s *Server) handleCartAdd(w http.ResponseWriter, r *http.Request) {
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
		ProductCode string `json:"product_code"`
		Quantity    int    `json:"quantity"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	cart := s.cartService.GetOrCreateCart(user.ID, user.TenantID)
	updatedCart, err := s.cartService.AddItem(cart.ID, req.ProductCode, req.Quantity)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updatedCart)
}

// handleCartUpdate updates the quantity of an item in the cart
func (s *Server) handleCartUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ProductCode string `json:"product_code"`
		Quantity    int    `json:"quantity"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cart := s.cartService.GetOrCreateCart(user.ID, user.TenantID)
	updatedCart, err := s.cartService.UpdateItem(cart.ID, req.ProductCode, req.Quantity)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updatedCart)
}

// handleCartRemove removes an item from the cart
func (s *Server) handleCartRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ProductCode string `json:"product_code"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cart := s.cartService.GetOrCreateCart(user.ID, user.TenantID)
	updatedCart, err := s.cartService.RemoveItem(cart.ID, req.ProductCode)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updatedCart)
}

// handleCartClear clears all items from the cart
func (s *Server) handleCartClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cart := s.cartService.GetOrCreateCart(user.ID, user.TenantID)
	if err := s.cartService.ClearCart(cart.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "cart cleared"})
}

// handleCartCheckout processes the cart checkout
func (s *Server) handleCartCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	user := s.getUserFromRequest(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cart := s.cartService.GetOrCreateCart(user.ID, user.TenantID)
	if len(cart.Items) == 0 {
		writeError(w, http.StatusBadRequest, "cart is empty")
		return
	}

	// Create orders for each item
	orders := []domain.Order{}
	for _, item := range cart.Items {
		order := domain.Order{
			ID:          core.NewID(),
			TenantID:    user.TenantID,
			CustomerID:  user.ID,
			ProductCode: item.ProductCode,
			Quantity:    item.Quantity,
			Status:      "pending",
			CreatedAt:   time.Now().UTC(),
		}
		createdOrder := s.orderService.CreateOrder(order)
		orders = append(orders, createdOrder)
	}

	// Create invoice
	invoiceLines := []domain.InvoiceLine{}
	for _, item := range cart.Items {
		invoiceLines = append(invoiceLines, domain.InvoiceLine{
			Description: item.ProductCode,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Total:       item.UnitPrice * float64(item.Quantity),
		})
	}

	invoice := s.billingService.CreateInvoiceWithLines(
		user.TenantID,
		user.ID,
		invoiceLines,
		"USD",
	)

	// Send invoice email
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}
	if err := s.emailService.SendInvoiceEmail(
		user.Email,
		user.Name,
		invoice.ID,
		fmt.Sprintf("%.2f", invoice.Amount),
		invoice.Currency,
		lang,
	); err != nil {
		// Log but don't fail the request
		fmt.Printf("Failed to send invoice email: %v\n", err)
	}

	// Clear the cart
	s.cartService.ClearCart(cart.ID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"orders":  orders,
		"invoice": invoice,
	})
}
