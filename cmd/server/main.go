package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/openhost/openhost/docs"
	"github.com/openhost/openhost/internal/infrastructure/http/handlers"
)

// @title OpenHost API
// @version 0.1
// @description OpenHost API for provisioning and billing.
// @BasePath /api/v1
func main() {
	router := gin.New()
	router.Use(gin.Recovery())

	// Static files
	router.Static("/static", "./themes")

	// Frontend routes
	router.GET("/", handlers.Root)
	router.GET("/home", handlers.Home)
	router.GET("/products", handlers.Products)
	router.GET("/dashboard", handlers.Dashboard)
	router.GET("/install", handlers.InstallForm)
	router.POST("/install", handlers.InstallSubmit)

	// Additional frontend routes
	router.GET("/pricing", handlers.Pricing)
	router.GET("/login", handlers.Login)
	router.GET("/register", handlers.Register)
	router.GET("/features", handlers.Features)
	router.GET("/about", handlers.About)
	router.GET("/contact", handlers.Contact)
	router.GET("/docs", handlers.Docs)
	router.GET("/support", handlers.Support)

	// User panel routes
	router.GET("/client", handlers.ClientDashboard)
	router.GET("/client/services", handlers.ClientServices)
	router.GET("/client/services/:id", handlers.ClientServiceDetail)
	router.GET("/client/invoices", handlers.ClientInvoices)
	router.GET("/client/invoices/:id", handlers.ClientInvoiceDetail)
	router.GET("/client/tickets", handlers.ClientTickets)
	router.GET("/client/tickets/:id", handlers.ClientTicketDetail)
	router.GET("/client/tickets/new", handlers.ClientNewTicket)
	router.GET("/client/profile", handlers.ClientProfile)
	router.GET("/client/affiliates", handlers.ClientAffiliates)
	router.GET("/client/domains", handlers.ClientDomains)
	router.GET("/client/subusers", handlers.ClientSubUsers)
	router.GET("/client/security", handlers.ClientSecurity)

	// Admin panel routes
	router.GET("/admin", handlers.AdminDashboard)
	router.GET("/admin/customers", handlers.AdminCustomers)
	router.GET("/admin/orders", handlers.AdminOrders)
	router.GET("/admin/services", handlers.AdminServices)
	router.GET("/admin/invoices", handlers.AdminInvoices)
	router.GET("/admin/tickets", handlers.AdminTickets)
	router.GET("/admin/products", handlers.AdminProducts)
	router.GET("/admin/servers", handlers.AdminServers)
	router.GET("/admin/settings", handlers.AdminSettings)

	// Shopping cart routes
	router.GET("/cart", handlers.Cart)
	router.GET("/checkout", handlers.Checkout)
	router.GET("/order/configure/:slug", handlers.ConfigureProduct)

	// API routes
	api := router.Group("/api/v1")
	api.GET("/health", handlers.Health)

	// TODO: Initialize services and wire up API handlers when database is ready
	// For now, just register the health endpoint

	_ = http.ListenAndServe(":6421", router)
}
