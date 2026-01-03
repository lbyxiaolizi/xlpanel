package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sort"
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
	customers := s.customers.List()
	orders := s.orderService.ListOrders()
	subscriptions := s.subService.ListSubscriptions()
	vpsInstances := s.hostingService.ListVPS()
	invoices := s.billingService.ListInvoices()
	payments := s.billingService.ListPayments()
	tickets := s.supportService.ListTickets()

	activeServices := len(subscriptions) + len(vpsInstances)
	unpaidCount, unpaidAmounts := summarizeUnpaidInvoices(invoices)
	unpaidCurrency, unpaidTotal := formatCurrencyTotals(unpaidAmounts)
	monthlyCurrency, monthlyRevenue := formatCurrencyTotals(sumPaymentsLast30Days(payments))

	data := dashboardData{
		Title: "Dashboard",
		Stats: dashboardStats{
			ActiveServices:     activeServices,
			UnpaidInvoices:     unpaidCount,
			UnpaidAmount:       unpaidTotal,
			UnpaidCurrency:     unpaidCurrency,
			TotalCustomers:     len(customers),
			MonthlyRevenue:     monthlyRevenue,
			MonthlyRevenueCode: monthlyCurrency,
		},
		RecentActivities: buildRecentActivities(customers, orders, invoices, payments, tickets),
	}

	// Execute template
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

type dashboardData struct {
	Title            string
	Stats            dashboardStats
	RecentActivities []dashboardActivity
}

type dashboardStats struct {
	ActiveServices     int
	UnpaidInvoices     int
	UnpaidAmount       string
	UnpaidCurrency     string
	TotalCustomers     int
	MonthlyRevenue     string
	MonthlyRevenueCode string
}

type dashboardActivity struct {
	Icon      string
	IconBg    string
	IconColor string
	Title     string
	Subtitle  string
	Time      string
}

type activityEvent struct {
	occurred time.Time
	item     dashboardActivity
}

func summarizeUnpaidInvoices(invoices []domain.Invoice) (int, map[string]float64) {
	count := 0
	amounts := make(map[string]float64)
	for _, invoice := range invoices {
		if strings.EqualFold(invoice.Status, "unpaid") {
			count++
			currency := invoice.Currency
			if currency == "" {
				currency = "USD"
			}
			amounts[currency] += invoice.Amount
		}
	}
	return count, amounts
}

func sumPaymentsLast30Days(payments []domain.Payment) map[string]float64 {
	totals := make(map[string]float64)
	cutoff := time.Now().AddDate(0, 0, -30)
	for _, payment := range payments {
		if payment.ReceivedAt.IsZero() || payment.ReceivedAt.Before(cutoff) {
			continue
		}
		currency := payment.Currency
		if currency == "" {
			currency = "USD"
		}
		totals[currency] += payment.Amount
	}
	return totals
}

func formatCurrencyTotals(amounts map[string]float64) (string, string) {
	if len(amounts) == 0 {
		return "USD", "0.00"
	}
	if len(amounts) == 1 {
		for currency, total := range amounts {
			return currency, fmt.Sprintf("%.2f", total)
		}
	}
	total := 0.0
	for _, amount := range amounts {
		total += amount
	}
	return "Multiple", fmt.Sprintf("%.2f", total)
}

func buildRecentActivities(
	customers []domain.Customer,
	orders []domain.Order,
	invoices []domain.Invoice,
	payments []domain.Payment,
	tickets []domain.Ticket,
) []dashboardActivity {
	events := make([]activityEvent, 0, len(customers)+len(orders)+len(payments)+len(tickets))

	for _, customer := range customers {
		if customer.CreatedAt.IsZero() {
			continue
		}
		events = append(events, activityEvent{
			occurred: customer.CreatedAt,
			item: dashboardActivity{
				Icon:      "fa-user-plus",
				IconBg:    "bg-purple-100",
				IconColor: "text-purple-600",
				Title:     "New customer registered",
				Subtitle:  fmt.Sprintf("Name: %s • Email: %s", customer.Name, customer.Email),
				Time:      formatRelativeTime(customer.CreatedAt),
			},
		})
	}

	for _, order := range orders {
		if order.CreatedAt.IsZero() {
			continue
		}
		title := "New order created"
		if order.ID != "" {
			title = fmt.Sprintf("New order %s created", order.ID)
		}
		events = append(events, activityEvent{
			occurred: order.CreatedAt,
			item: dashboardActivity{
				Icon:      "fa-shopping-cart",
				IconBg:    "bg-blue-100",
				IconColor: "text-blue-600",
				Title:     title,
				Subtitle:  fmt.Sprintf("Product: %s • Qty: %d", order.ProductCode, order.Quantity),
				Time:      formatRelativeTime(order.CreatedAt),
			},
		})
	}

	for _, payment := range payments {
		if payment.ReceivedAt.IsZero() {
			continue
		}
		method := payment.Method
		if method == "" {
			method = "unknown"
		}
		currency := payment.Currency
		if currency == "" {
			currency = "USD"
		}
		events = append(events, activityEvent{
			occurred: payment.ReceivedAt,
			item: dashboardActivity{
				Icon:      "fa-credit-card",
				IconBg:    "bg-green-100",
				IconColor: "text-green-600",
				Title:     "Payment received",
				Subtitle:  fmt.Sprintf("Amount: %s %.2f • Method: %s", currency, payment.Amount, method),
				Time:      formatRelativeTime(payment.ReceivedAt),
			},
		})
	}

	for _, ticket := range tickets {
		if ticket.CreatedAt.IsZero() {
			continue
		}
		events = append(events, activityEvent{
			occurred: ticket.CreatedAt,
			item: dashboardActivity{
				Icon:      "fa-ticket-alt",
				IconBg:    "bg-yellow-100",
				IconColor: "text-yellow-600",
				Title:     "New support ticket opened",
				Subtitle:  fmt.Sprintf("Subject: %s", ticket.Subject),
				Time:      formatRelativeTime(ticket.CreatedAt),
			},
		})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].occurred.After(events[j].occurred)
	})

	limit := 5
	if len(events) < limit {
		limit = len(events)
	}

	activities := make([]dashboardActivity, 0, limit)
	for i := 0; i < limit; i++ {
		activities = append(activities, events[i].item)
	}
	return activities
}

func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	elapsed := time.Since(t)
	switch {
	case elapsed < time.Minute:
		return "just now"
	case elapsed < time.Hour:
		return fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
	case elapsed < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(elapsed.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(elapsed.Hours()/24))
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
		if wantsHTML(r) {
			customers := s.customers.List()
			rows := make([][]string, 0, len(customers))
			for _, customer := range customers {
				rows = append(rows, []string{
					customer.ID,
					customer.Name,
					customer.Email,
					formatDateTime(customer.CreatedAt),
				})
			}
			s.renderAdminList(w, r, adminListData{
				Title:        "Customers",
				Description:  "Review and manage all customer accounts.",
				Headers:      []string{"ID", "Name", "Email", "Created"},
				Rows:         rows,
				EmptyMessage: "No customers found yet.",
				Count:        len(rows),
			})
			return
		}
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
		if wantsHTML(r) {
			orders := s.orderService.ListOrders()
			rows := make([][]string, 0, len(orders))
			for _, order := range orders {
				rows = append(rows, []string{
					order.ID,
					order.CustomerID,
					order.ProductCode,
					fmt.Sprintf("%d", order.Quantity),
					order.Status,
				})
			}
			s.renderAdminList(w, r, adminListData{
				Title:        "Orders",
				Description:  "Track recent orders across all customers.",
				Headers:      []string{"Order ID", "Customer ID", "Product", "Quantity", "Status"},
				Rows:         rows,
				EmptyMessage: "No orders have been created yet.",
				Count:        len(rows),
			})
			return
		}
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
		if wantsHTML(r) {
			products := s.catalogService.ListProducts()
			rows := make([][]string, 0, len(products))
			for _, product := range products {
				rows = append(rows, []string{
					product.Code,
					product.Name,
					formatCurrencyAmount(product.Currency, product.UnitPrice),
					formatRecurring(product.Recurring),
				})
			}
			s.renderAdminList(w, r, adminListData{
				Title:        "Products",
				Description:  "Manage the product catalog and pricing.",
				Headers:      []string{"Code", "Name", "Price", "Recurring"},
				Rows:         rows,
				EmptyMessage: "No products are available yet.",
				Count:        len(rows),
			})
			return
		}
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
		if wantsHTML(r) {
			subscriptions := s.subService.ListSubscriptions()
			rows := make([][]string, 0, len(subscriptions))
			for _, subscription := range subscriptions {
				rows = append(rows, []string{
					subscription.ID,
					subscription.CustomerID,
					subscription.ProductCode,
					subscription.Status,
					formatDateTime(subscription.NextBillAt),
				})
			}
			s.renderAdminList(w, r, adminListData{
				Title:        "Subscriptions",
				Description:  "Monitor recurring customer subscriptions.",
				Headers:      []string{"Subscription ID", "Customer ID", "Product", "Status", "Next Bill"},
				Rows:         rows,
				EmptyMessage: "No subscriptions are active yet.",
				Count:        len(rows),
			})
			return
		}
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
	switch r.Method {
	case http.MethodGet:
		if wantsHTML(r) {
			invoices := s.billingService.ListInvoices()
			rows := make([][]string, 0, len(invoices))
			for _, invoice := range invoices {
				rows = append(rows, []string{
					invoice.ID,
					invoice.CustomerID,
					formatCurrencyAmount(invoice.Currency, invoice.Amount),
					invoice.Status,
					formatDateTime(invoice.DueAt),
				})
			}
			s.renderAdminList(w, r, adminListData{
				Title:        "Invoices",
				Description:  "Review billing invoices and payment status.",
				Headers:      []string{"Invoice ID", "Customer ID", "Amount", "Status", "Due"},
				Rows:         rows,
				EmptyMessage: "No invoices have been issued yet.",
				Count:        len(rows),
			})
			return
		}
		writeJSON(w, http.StatusOK, s.billingService.ListInvoices())
	case http.MethodPost:
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
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
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
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.billingService.ListPayments())
	case http.MethodPost:
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
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleVPS(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if wantsHTML(r) {
			vpsInstances := s.hostingService.ListVPS()
			rows := make([][]string, 0, len(vpsInstances))
			for _, instance := range vpsInstances {
				rows = append(rows, []string{
					instance.ID,
					instance.Hostname,
					instance.PlanCode,
					instance.Status,
					formatDateTime(instance.CreatedAt),
				})
			}
			s.renderAdminList(w, r, adminListData{
				Title:        "VPS Hosting",
				Description:  "Provisioned VPS instances and their status.",
				Headers:      []string{"Instance ID", "Hostname", "Plan", "Status", "Created"},
				Rows:         rows,
				EmptyMessage: "No VPS instances have been provisioned yet.",
				Count:        len(rows),
			})
			return
		}
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
		if wantsHTML(r) {
			tickets := s.supportService.ListTickets()
			rows := make([][]string, 0, len(tickets))
			for _, ticket := range tickets {
				rows = append(rows, []string{
					ticket.ID,
					ticket.Subject,
					ticket.Status,
					formatDateTime(ticket.CreatedAt),
				})
			}
			s.renderAdminList(w, r, adminListData{
				Title:        "Support Tickets",
				Description:  "Open and resolved customer support requests.",
				Headers:      []string{"Ticket ID", "Subject", "Status", "Created"},
				Rows:         rows,
				EmptyMessage: "No support tickets have been opened yet.",
				Count:        len(rows),
			})
			return
		}
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
		if wantsHTML(r) {
			gateways := s.paymentsService.ListGateways()
			rows := make([][]string, 0, len(gateways))
			for _, gateway := range gateways {
				status := "Disabled"
				if gateway.Enabled {
					status = "Enabled"
				}
				rows = append(rows, []string{
					gateway.Name,
					gateway.Provider,
					status,
				})
			}
			s.renderAdminList(w, r, adminListData{
				Title:        "Payment Gateways",
				Description:  "Manage payment providers and enablement.",
				Headers:      []string{"Gateway", "Provider", "Status"},
				Rows:         rows,
				EmptyMessage: "No payment gateways are configured yet.",
				Count:        len(rows),
			})
			return
		}
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
		if wantsHTML(r) {
			jobs := s.autoService.ListJobs()
			rows := make([][]string, 0, len(jobs))
			for _, job := range jobs {
				rows = append(rows, []string{
					job.Name,
					job.Schedule,
					formatDateTime(job.LastRun),
				})
			}
			s.renderAdminList(w, r, adminListData{
				Title:        "Automation Jobs",
				Description:  "Scheduled jobs and their last run times.",
				Headers:      []string{"Job", "Schedule", "Last Run"},
				Rows:         rows,
				EmptyMessage: "No automation jobs have been registered yet.",
				Count:        len(rows),
			})
			return
		}
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

type adminListData struct {
	Title        string
	Description  string
	Headers      []string
	Rows         [][]string
	EmptyMessage string
	Count        int
}

func wantsHTML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

func formatDateTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("2006-01-02 15:04")
}

func formatCurrencyAmount(currency string, amount float64) string {
	if currency == "" {
		currency = "USD"
	}
	return fmt.Sprintf("%s %.2f", currency, amount)
}

func formatRecurring(recurring bool) string {
	if recurring {
		return "Yes"
	}
	return "No"
}

func (s *Server) renderAdminList(w http.ResponseWriter, r *http.Request, data adminListData) {
	theme := s.config.DefaultTheme
	if s.config.AllowThemeOverride {
		if requested := r.URL.Query().Get("theme"); requested != "" {
			theme = requested
		}
	}
	basePath := filepath.Join("frontend", "themes", theme, "base.html")
	listPath := filepath.Join("frontend", "themes", theme, "admin", "list.html")
	tmpl, err := template.ParseFiles(basePath, listPath)
	if err != nil {
		log.Printf("Admin template not loaded: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Error executing admin template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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
