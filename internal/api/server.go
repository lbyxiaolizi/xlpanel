package api

import (
	"encoding/json"
	"net/http"
	"time"

	"xlpanel/internal/core"
	"xlpanel/internal/domain"
	"xlpanel/internal/infra"
	"xlpanel/internal/service"
)

type Server struct {
	config          core.AppConfig
	metrics         *infra.MetricsRegistry
	customers       *infra.Repository[domain.Customer]
	orderService    *service.OrderService
	billingService  *service.BillingService
	hostingService  *service.HostingService
	supportService  *service.SupportService
	paymentsService *service.PaymentsService
	autoService     *service.AutomationService
}

func NewServer(
	config core.AppConfig,
	metrics *infra.MetricsRegistry,
	customers *infra.Repository[domain.Customer],
	orderService *service.OrderService,
	billingService *service.BillingService,
	hostingService *service.HostingService,
	supportService *service.SupportService,
	paymentsService *service.PaymentsService,
	autoService *service.AutomationService,
) *Server {
	return &Server{
		config:          config,
		metrics:         metrics,
		customers:       customers,
		orderService:    orderService,
		billingService:  billingService,
		hostingService:  hostingService,
		supportService:  supportService,
		paymentsService: paymentsService,
		autoService:     autoService,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/customers", s.handleCustomers)
	mux.HandleFunc("/orders", s.handleOrders)
	mux.HandleFunc("/billing/invoices", s.handleInvoices)
	mux.HandleFunc("/billing/coupons", s.handleCoupons)
	mux.HandleFunc("/hosting/vps", s.handleVPS)
	mux.HandleFunc("/hosting/ips", s.handleIPs)
	mux.HandleFunc("/support/tickets", s.handleTickets)
	mux.HandleFunc("/payments/gateways", s.handleGateways)
	mux.HandleFunc("/automation/jobs", s.handleJobs)
	return mux
}

type customerCreate struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type orderCreate struct {
	CustomerID  string `json:"customer_id"`
	ProductCode string `json:"product_code"`
	Quantity    int    `json:"quantity"`
}

type invoiceCreate struct {
	CustomerID string  `json:"customer_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

type couponCreate struct {
	Code           string `json:"code"`
	PercentOff     int    `json:"percent_off"`
	Recurring      bool   `json:"recurring"`
	MaxRedemptions *int   `json:"max_redemptions"`
}

type ticketCreate struct {
	CustomerID string `json:"customer_id"`
	Subject    string `json:"subject"`
}

type vpsCreate struct {
	CustomerID string `json:"customer_id"`
	Hostname   string `json:"hostname"`
	PlanCode   string `json:"plan_code"`
}

type ipCreate struct {
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
		customer := domain.Customer{
			ID:        core.NewID(),
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
		order := domain.Order{
			ID:          core.NewID(),
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
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}
	invoice := s.billingService.CreateInvoice(req.CustomerID, req.Amount, req.Currency)
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
