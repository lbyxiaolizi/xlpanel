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

	customerRepo := infra.NewRepository(func(c domain.Customer) string { return c.ID })
	orderRepo := infra.NewRepository(func(o domain.Order) string { return o.ID })
	invoiceRepo := infra.NewRepository(func(i domain.Invoice) string { return i.ID })
	vpsRepo := infra.NewRepository(func(v domain.VPSInstance) string { return v.ID })
	ipRepo := infra.NewRepository(func(ip domain.IPAllocation) string { return ip.ID })
	ticketRepo := infra.NewRepository(func(t domain.Ticket) string { return t.ID })

	billingService := service.NewBillingService(invoiceRepo, infra.NewCouponRepository(), metrics)
	orderService := service.NewOrderService(orderRepo, metrics)
	hostingService := service.NewHostingService(vpsRepo, ipRepo, metrics)
	supportService := service.NewSupportService(ticketRepo, metrics)
	paymentsService := service.NewPaymentsService(metrics)
	automationService := service.NewAutomationService(metrics)

	server := api.NewServer(
		core.DefaultConfig,
		metrics,
		customerRepo,
		orderService,
		billingService,
		hostingService,
		supportService,
		paymentsService,
		automationService,
	)

	log.Println("XLPanel running on :8080")
	if err := http.ListenAndServe(":8080", server.Routes()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
