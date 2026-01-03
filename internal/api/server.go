package api

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"xlpanel/internal/core"
	"xlpanel/internal/domain"
	"xlpanel/internal/infra"
	"xlpanel/internal/service"
)

type Server struct {
	config          core.AppConfig
	metrics         *infra.MetricsRegistry
	tenants         *infra.Repository[domain.Tenant]
	customers       *infra.Repository[domain.Customer]
	orderService    *service.OrderService
	billingService  *service.BillingService
	hostingService  *service.HostingService
	supportService  *service.SupportService
	catalogService  *service.CatalogService
	subService      *service.SubscriptionService
	paymentsService *service.PaymentsService
	autoService     *service.AutomationService
	authService     *service.AuthService
	cartService     *service.CartService
	emailService    *service.EmailService
	templates       map[string]*template.Template
	templateMu      sync.RWMutex
}

func NewServer(
	config core.AppConfig,
	metrics *infra.MetricsRegistry,
	tenants *infra.Repository[domain.Tenant],
	customers *infra.Repository[domain.Customer],
	orderService *service.OrderService,
	billingService *service.BillingService,
	hostingService *service.HostingService,
	supportService *service.SupportService,
	catalogService *service.CatalogService,
	subService *service.SubscriptionService,
	paymentsService *service.PaymentsService,
	autoService *service.AutomationService,
	authService *service.AuthService,
	cartService *service.CartService,
	emailService *service.EmailService,
) *Server {
	s := &Server{
		config:          config,
		metrics:         metrics,
		tenants:         tenants,
		customers:       customers,
		orderService:    orderService,
		billingService:  billingService,
		hostingService:  hostingService,
		supportService:  supportService,
		catalogService:  catalogService,
		subService:      subService,
		paymentsService: paymentsService,
		autoService:     autoService,
		authService:     authService,
		cartService:     cartService,
		emailService:    emailService,
		templates:       make(map[string]*template.Template),
	}

	// Parse and cache templates at startup
	if _, err := s.loadTemplate(config.DefaultTheme); err != nil {
		log.Printf("Warning: failed to load templates: %v", err)
	}

	return s
}

func (s *Server) loadTemplate(theme string) (*template.Template, error) {
	// Validate theme name to prevent path traversal
	if theme == "" {
		theme = s.config.DefaultTheme
	}
	// Check for empty, path separators, or relative path components
	if theme == "" ||
		filepath.Base(theme) != theme ||
		theme == "." ||
		theme == ".." ||
		strings.Contains(theme, "/") ||
		strings.Contains(theme, "\\") {
		return nil, filepath.ErrBadPattern
	}

	s.templateMu.RLock()
	if tmpl, ok := s.templates[theme]; ok {
		s.templateMu.RUnlock()
		return tmpl, nil
	}
	s.templateMu.RUnlock()

	// Build theme path
	themePath := filepath.Join("frontend", "themes", theme)
	basePath := filepath.Join(themePath, "base.html")
	homePath := filepath.Join(themePath, "home.html")

	// Parse templates
	tmpl, err := template.ParseFiles(basePath, homePath)
	if err != nil {
		return nil, err
	}

	s.templateMu.Lock()
	s.templates[theme] = tmpl
	s.templateMu.Unlock()
	return tmpl, nil
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	// Admin routes
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/tenants", s.handleTenants)
	mux.HandleFunc("/customers", s.handleCustomers)
	mux.HandleFunc("/orders", s.handleOrders)
	mux.HandleFunc("/catalog/products", s.handleProducts)
	mux.HandleFunc("/subscriptions", s.handleSubscriptions)
	mux.HandleFunc("/subscriptions/invoices", s.handleSubscriptionInvoices)
	mux.HandleFunc("/billing/invoices", s.handleInvoices)
	mux.HandleFunc("/billing/coupons", s.handleCoupons)
	mux.HandleFunc("/billing/payments", s.handlePayments)
	mux.HandleFunc("/hosting/vps", s.handleVPS)
	mux.HandleFunc("/hosting/ips", s.handleIPs)
	mux.HandleFunc("/support/tickets", s.handleTickets)
	mux.HandleFunc("/payments/gateways", s.handleGateways)
	mux.HandleFunc("/payments/plugins/charge", s.handlePluginCharge)
	mux.HandleFunc("/automation/jobs", s.handleJobs)
	
	// Auth routes
	mux.HandleFunc("/login", s.handleLoginPage)
	mux.HandleFunc("/register", s.handleRegisterPage)
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/register", s.handleRegister)
	mux.HandleFunc("/api/logout", s.handleLogout)
	mux.HandleFunc("/api/2fa/enable", s.handle2FAEnable)
	mux.HandleFunc("/api/2fa/verify", s.handle2FAVerify)
	mux.HandleFunc("/api/2fa/disable", s.handle2FADisable)
	
	// Customer routes
	mux.HandleFunc("/customer", s.handleCustomerDashboard)
	mux.HandleFunc("/customer/products", s.handleCustomerProducts)
	mux.HandleFunc("/customer/cart", s.handleCustomerCart)
	mux.HandleFunc("/customer/orders", s.handleCustomerOrders)
	mux.HandleFunc("/customer/invoices", s.handleCustomerInvoices)
	
	// Admin routes
	mux.HandleFunc("/admin/users", s.handleAdminUsers)
	
	// Cart API routes
	mux.HandleFunc("/api/cart", s.handleCartGet)
	mux.HandleFunc("/api/cart/add", s.handleCartAdd)
	mux.HandleFunc("/api/cart/update", s.handleCartUpdate)
	mux.HandleFunc("/api/cart/remove", s.handleCartRemove)
	mux.HandleFunc("/api/cart/clear", s.handleCartClear)
	mux.HandleFunc("/api/cart/checkout", s.handleCartCheckout)
	
	// User API routes
	mux.HandleFunc("/api/users", s.handleUsersAPI)
	
	return s.applyMiddlewares(mux)
}

type customerCreate struct {
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

type orderCreate struct {
	TenantID    string `json:"tenant_id"`
	CustomerID  string `json:"customer_id"`
	ProductCode string `json:"product_code"`
	Quantity    int    `json:"quantity"`
}

type invoiceCreate struct {
	TenantID   string               `json:"tenant_id"`
	CustomerID string               `json:"customer_id"`
	Currency   string               `json:"currency"`
	CouponCode string               `json:"coupon_code"`
	Lines      []domain.InvoiceLine `json:"lines"`
}

type couponCreate struct {
	Code           string `json:"code"`
	PercentOff     int    `json:"percent_off"`
	Recurring      bool   `json:"recurring"`
	MaxRedemptions *int   `json:"max_redemptions"`
}

type ticketCreate struct {
	TenantID   string `json:"tenant_id"`
	CustomerID string `json:"customer_id"`
	Subject    string `json:"subject"`
}

type vpsCreate struct {
	TenantID   string `json:"tenant_id"`
	CustomerID string `json:"customer_id"`
	Hostname   string `json:"hostname"`
	PlanCode   string `json:"plan_code"`
}

type ipCreate struct {
	TenantID   string `json:"tenant_id"`
	CustomerID string `json:"customer_id"`
	IPAddress  string `json:"ip_address"`
}

type gatewayCreate struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Enabled  bool   `json:"enabled"`
}

type jobCreate struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
}

type tenantCreate struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
}

type productCreate struct {
	TenantID    string  `json:"tenant_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	UnitPrice   float64 `json:"unit_price"`
	Currency    string  `json:"currency"`
	Recurring   bool    `json:"recurring"`
	BillingDays int     `json:"billing_days"`
}

type subscriptionCreate struct {
	TenantID    string `json:"tenant_id"`
	CustomerID  string `json:"customer_id"`
	ProductCode string `json:"product_code"`
}

type paymentCreate struct {
	TenantID   string  `json:"tenant_id"`
	CustomerID string  `json:"customer_id"`
	InvoiceID  string  `json:"invoice_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Method     string  `json:"method"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	payload := map[string]any{
		"status":  "ok",
		"metrics": s.metrics.Snapshot(),
		"theme":   s.config.DefaultTheme,
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	theme := s.config.DefaultTheme
	if s.config.AllowThemeOverride {
		if requested := r.URL.Query().Get("theme"); requested != "" {
			theme = requested
		}
	}

	tmpl, err := s.loadTemplate(theme)
	if err != nil {
		log.Printf("Templates not loaded: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Prepare data
	data := map[string]string{
		"Title": "Dashboard",
	}

	// Execute template
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleTenants(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.tenants.List())
	case http.MethodPost:
		var req tenantCreate
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		tenant := domain.Tenant{
			ID:        core.NewID(),
			Name:      req.Name,
			Contact:   req.Contact,
			CreatedAt: time.Now().UTC(),
		}
		s.tenants.Add(tenant)
		writeJSON(w, http.StatusCreated, tenant)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleCustomers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.customers.List())
	case http.MethodPost:
		var req customerCreate
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.TenantID == "" {
			writeError(w, http.StatusBadRequest, "tenant_id is required")
			return
		}
		customer := domain.Customer{
			ID:        core.NewID(),
			TenantID:  req.TenantID,
			Name:      req.Name,
			Email:     req.Email,
			CreatedAt: time.Now().UTC(),
		}
		s.customers.Add(customer)
		writeJSON(w, http.StatusCreated, customer)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.orderService.ListOrders())
	case http.MethodPost:
		var req orderCreate
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Quantity <= 0 {
			writeError(w, http.StatusBadRequest, "quantity must be positive")
			return
		}
		if req.TenantID == "" {
			writeError(w, http.StatusBadRequest, "tenant_id is required")
			return
		}
		order := domain.Order{
			ID:          core.NewID(),
			TenantID:    req.TenantID,
			CustomerID:  req.CustomerID,
			ProductCode: req.ProductCode,
			Quantity:    req.Quantity,
			Status:      "pending",
			CreatedAt:   time.Now().UTC(),
		}
		writeJSON(w, http.StatusCreated, s.orderService.CreateOrder(order))
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.catalogService.ListProducts())
	case http.MethodPost:
		var req productCreate
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Code == "" || req.Name == "" {
			writeError(w, http.StatusBadRequest, "code and name are required")
			return
		}
		if req.Currency == "" {
			req.Currency = "USD"
		}
		product := domain.Product{
			Code:        req.Code,
			TenantID:    req.TenantID,
			Name:        req.Name,
			Description: req.Description,
			UnitPrice:   req.UnitPrice,
			Currency:    req.Currency,
			Recurring:   req.Recurring,
			BillingDays: req.BillingDays,
		}
		writeJSON(w, http.StatusCreated, s.catalogService.AddProduct(product))
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleSubscriptions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.subService.ListSubscriptions())
	case http.MethodPost:
		var req subscriptionCreate
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		subscription, err := s.subService.CreateSubscription(domain.Subscription{
			TenantID:    req.TenantID,
			CustomerID:  req.CustomerID,
			ProductCode: req.ProductCode,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, subscription)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleSubscriptionInvoices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var payload struct {
		SubscriptionID string `json:"subscription_id"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	invoice, err := s.subService.GenerateInvoice(payload.SubscriptionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, invoice)
}

func (s *Server) handleInvoices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req invoiceCreate
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}
	if req.TenantID == "" {
		writeError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}
	if len(req.Lines) == 0 {
		writeError(w, http.StatusBadRequest, "lines are required")
		return
	}
	invoice := s.billingService.CreateInvoiceWithLines(req.TenantID, req.CustomerID, req.Lines, req.Currency)
	if req.CouponCode != "" {
		adjusted, ok := s.billingService.ApplyCoupon(req.CouponCode, invoice.Amount)
		if ok {
			invoice.Amount = adjusted
		}
	}
	writeJSON(w, http.StatusCreated, invoice)
}

func (s *Server) handleCoupons(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.billingService.ListCoupons())
	case http.MethodPost:
		var req couponCreate
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.PercentOff < 0 || req.PercentOff > 100 {
			writeError(w, http.StatusBadRequest, "percent_off must be 0-100")
			return
		}
		coupon := domain.Coupon{
			Code:           req.Code,
			PercentOff:     req.PercentOff,
			Recurring:      req.Recurring,
			MaxRedemptions: req.MaxRedemptions,
		}
		writeJSON(w, http.StatusCreated, s.billingService.AddCoupon(coupon))
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handlePayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req paymentCreate
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}
	payment := domain.Payment{
		TenantID:   req.TenantID,
		CustomerID: req.CustomerID,
		InvoiceID:  req.InvoiceID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Method:     req.Method,
	}
	writeJSON(w, http.StatusCreated, s.billingService.RecordPayment(payment))
}

func (s *Server) handleVPS(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.hostingService.ListVPS())
	case http.MethodPost:
		var req vpsCreate
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		instance := domain.VPSInstance{
			ID:         core.NewID(),
			TenantID:   req.TenantID,
			CustomerID: req.CustomerID,
			Hostname:   req.Hostname,
			PlanCode:   req.PlanCode,
			Status:     "provisioning",
			CreatedAt:  time.Now().UTC(),
		}
		writeJSON(w, http.StatusCreated, s.hostingService.ProvisionVPS(instance))
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleIPs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.hostingService.ListIPs())
	case http.MethodPost:
		var req ipCreate
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		allocation := domain.IPAllocation{
			ID:         core.NewID(),
			TenantID:   req.TenantID,
			CustomerID: req.CustomerID,
			IPAddress:  req.IPAddress,
			AssignedAt: time.Now().UTC(),
		}
		writeJSON(w, http.StatusCreated, s.hostingService.AllocateIP(allocation))
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleTickets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.supportService.ListTickets())
	case http.MethodPost:
		var req ticketCreate
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		ticket := domain.Ticket{
			ID:         core.NewID(),
			TenantID:   req.TenantID,
			CustomerID: req.CustomerID,
			Subject:    req.Subject,
			Status:     "open",
			CreatedAt:  time.Now().UTC(),
		}
		writeJSON(w, http.StatusCreated, s.supportService.OpenTicket(ticket))
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleGateways(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.paymentsService.ListGateways())
	case http.MethodPost:
		var req gatewayCreate
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		gateway := service.PaymentGateway{
			Name:     req.Name,
			Provider: req.Provider,
			Enabled:  req.Enabled,
		}
		writeJSON(w, http.StatusCreated, s.paymentsService.RegisterGateway(gateway))
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.autoService.ListJobs())
	case http.MethodPost:
		var req jobCreate
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		job := service.AutomationJob{
			Name:     req.Name,
			Schedule: req.Schedule,
		}
		writeJSON(w, http.StatusCreated, s.autoService.RegisterJob(job))
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(data)
	r.size += n
	return n, err
}

func (s *Server) applyMiddlewares(next http.Handler) http.Handler {
	return s.withRecovery(s.withSecurityHeaders(s.withRequestID(s.withBodyLimit(s.withAuth(s.withLogging(next))))))
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		duration := time.Since(start)
		requestID := r.Header.Get("X-Request-ID")
		log.Printf("request_id=%s method=%s path=%s status=%d duration=%s bytes=%d remote=%s",
			requestID,
			r.Method,
			r.URL.Path,
			rec.status,
			duration.String(),
			rec.size,
			clientIP(r),
		)
	})
}

func (s *Server) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				requestID := r.Header.Get("X-Request-ID")
				log.Printf("panic request_id=%s err=%v", requestID, recovered)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Request-ID") == "" {
			r.Header.Set("X-Request-ID", core.NewID())
		}
		w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withBodyLimit(next http.Handler) http.Handler {
	if s.config.MaxBodyBytes <= 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			r.Body = http.MaxBytesReader(w, r.Body, s.config.MaxBodyBytes)
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withAuth(next http.Handler) http.Handler {
	if s.config.APIKey == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicRoute(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		if authorized(r, s.config.APIKey) {
			next.ServeHTTP(w, r)
			return
		}
		writeError(w, http.StatusUnauthorized, "unauthorized")
	})
}

func isPublicRoute(path string) bool {
	publicRoutes := []string{
		"/",
		"/health",
		"/login",
		"/register",
		"/api/login",
		"/api/register",
		"/customer/products",
		"/catalog/products",
	}
	
	for _, route := range publicRoutes {
		if path == route {
			return true
		}
	}
	
	return false
}

func authorized(r *http.Request, apiKey string) bool {
	if apiKey == "" {
		return true
	}
	if key := strings.TrimSpace(r.Header.Get("X-API-Key")); key != "" {
		return key == apiKey
	}
	if auth := strings.TrimSpace(r.Header.Get("Authorization")); auth != "" {
		const prefix = "Bearer "
		if strings.HasPrefix(auth, prefix) {
			return strings.TrimPrefix(auth, prefix) == apiKey
		}
	}
	return false
}

func clientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host := r.RemoteAddr
	if host == "" {
		return ""
	}
	if strings.Contains(host, ":") {
		return strings.Split(host, ":")[0]
	}
	return host
}
