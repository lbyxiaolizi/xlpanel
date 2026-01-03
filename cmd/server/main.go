package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"xlpanel/internal/api"
	"xlpanel/internal/core"
	"xlpanel/internal/domain"
	"xlpanel/internal/infra"
	"xlpanel/internal/plugins"
	"xlpanel/internal/plugins/providers"
	"xlpanel/internal/service"
)

func main() {
	config := core.LoadConfigFromEnv()
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

	registry := plugins.NewRegistry()
	registry.Register(providers.NewAlipayFaceToFace())
	registry.Register(providers.NewUnionPay())
	registry.Register(providers.NewStripe())
	registry.Register(providers.NewCrypto())

	billingService := service.NewBillingService(invoiceRepo, infra.NewCouponRepository(), paymentRepo, metrics)
	orderService := service.NewOrderService(orderRepo, metrics)
	hostingService := service.NewHostingService(vpsRepo, ipRepo, metrics)
	supportService := service.NewSupportService(ticketRepo, metrics)
	catalogService := service.NewCatalogService(productRepo)
	subService := service.NewSubscriptionService(subscriptionRepo, catalogService, billingService, metrics)
	paymentsService := service.NewPaymentsService(metrics, registry)
	automationService := service.NewAutomationService(metrics)

	server := api.NewServer(
		config,
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

	httpServer := &http.Server{
		Addr:              config.Address,
		Handler:           server.Routes(),
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("XLPanel running on %s", config.Address)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-shutdownCtx.Done()
	log.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		_ = httpServer.Close()
	}
}
