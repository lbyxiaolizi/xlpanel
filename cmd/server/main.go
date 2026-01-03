package main

import (
	"log"
	"net/http"

	"xlpanel/internal/api"
	"xlpanel/internal/core"
	"xlpanel/internal/domain"
	"xlpanel/internal/infra"
	"xlpanel/internal/service"
)

func main() {
	metrics := infra.NewMetricsRegistry()

	tenantRepo := infra.NewRepository(func(t domain.Tenant) string { return t.ID })
	customerRepo := infra.NewRepository(func(c domain.Customer) string { return c.ID })
	orderRepo := infra.NewRepository(func(o domain.Order) string { return o.ID })
	invoiceRepo := infra.NewRepository(func(i domain.Invoice) string { return i.ID })
	paymentRepo := infra.NewRepository(func(p domain.Payment) string { return p.ID })
	productRepo := infra.NewRepository(func(p domain.Product) string { return p.Code })
	subscriptionRepo := infra.NewRepository(func(s domain.Subscription) string { return s.ID })
	vpsRepo := infra.NewRepository(func(v domain.VPSInstance) string { return v.ID })
	ipRepo := infra.NewRepository(func(ip domain.IPAllocation) string { return ip.ID })
	ticketRepo := infra.NewRepository(func(t domain.Ticket) string { return t.ID })

	billingService := service.NewBillingService(invoiceRepo, infra.NewCouponRepository(), paymentRepo, metrics)
	orderService := service.NewOrderService(orderRepo, metrics)
	hostingService := service.NewHostingService(vpsRepo, ipRepo, metrics)
	supportService := service.NewSupportService(ticketRepo, metrics)
	catalogService := service.NewCatalogService(productRepo)
	subService := service.NewSubscriptionService(subscriptionRepo, catalogService, billingService, metrics)
	paymentsService := service.NewPaymentsService(metrics)
	automationService := service.NewAutomationService(metrics)

	server := api.NewServer(
		core.DefaultConfig,
		metrics,
		tenantRepo,
		customerRepo,
		orderService,
		billingService,
		hostingService,
		supportService,
		catalogService,
		subService,
		paymentsService,
		automationService,
	)

	log.Println("XLPanel running on :8080")
	if err := http.ListenAndServe(":8080", server.Routes()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
