package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	_ "github.com/openhost/openhost/docs"
	"github.com/openhost/openhost/internal/core/domain"
	"github.com/openhost/openhost/internal/core/service/affiliate"
	"github.com/openhost/openhost/internal/core/service/auth"
	"github.com/openhost/openhost/internal/core/service/invoice"
	"github.com/openhost/openhost/internal/core/service/knowledgebase"
	"github.com/openhost/openhost/internal/core/service/notification"
	"github.com/openhost/openhost/internal/core/service/order"
	"github.com/openhost/openhost/internal/core/service/payment"
	"github.com/openhost/openhost/internal/core/service/product"
	"github.com/openhost/openhost/internal/core/service/subuser"
	"github.com/openhost/openhost/internal/core/service/ticket"
	"github.com/openhost/openhost/internal/infrastructure/config"
	"github.com/openhost/openhost/internal/infrastructure/database"
	"github.com/openhost/openhost/internal/infrastructure/http/handlers"
	apiHandlers "github.com/openhost/openhost/internal/infrastructure/http/handlers/api"
	"github.com/openhost/openhost/internal/infrastructure/web"
)

// @title OpenHost API
// @version 0.1
// @description OpenHost API for provisioning and billing.
// @BasePath /api/v1
func main() {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(web.RequestIDMiddleware())
	router.Use(web.SecurityHeaders())
	router.Use(web.RecoveryMiddleware())
	router.Use(web.LanguageMiddleware())
	router.Use(web.ThemeMiddleware("default"))
	router.Use(web.CurrencyMiddleware("USD"))

	router.NoRoute(web.NotFoundHandler)
	router.NoMethod(web.MethodNotAllowedHandler)

	// Static files
	router.Static("/static", "./themes")

	// Frontend routes
	router.GET("/", handlers.Root)
	router.GET("/home", handlers.Home)
	router.GET("/dashboard", handlers.Dashboard)
	router.GET("/install", handlers.InstallForm)
	router.POST("/install", handlers.InstallSubmit)

	// Additional frontend routes
	router.GET("/pricing", handlers.Pricing)
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
	// Shopping cart routes are registered after installation.

	// API routes
	api := router.Group("/api/v1")

	installed, err := config.Exists(config.DefaultPath)
	if err != nil {
		log.Fatalf("failed to check install status: %v", err)
	}

	if installed {
		cfg, err := config.Load(config.DefaultPath)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		db, err := database.Open(cfg.Database)
		if err != nil {
			log.Fatalf("failed to open database: %v", err)
		}
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}
		if err := ensureAdminUser(db, cfg.Admin); err != nil {
			log.Fatalf("failed to ensure admin user: %v", err)
		}
		if err := ensureDefaultCatalog(db); err != nil {
			log.Fatalf("failed to ensure default catalog: %v", err)
		}
		api.GET("/health", handlers.Health)
		registerAPIRoutes(api, db)
		registerFrontendRoutes(router, db)
	} else {
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, apiHandlers.ErrorResponse{Error: "Service not installed"})
		})
	}

	_ = http.ListenAndServe(":6421", router)
}

func registerFrontendRoutes(router *gin.Engine, db *gorm.DB) {
	authService := auth.NewService(db)
	productService := product.NewService(db)
	orderService := order.NewService(db)
	cartService := order.NewCartService(db)
	invoiceService := invoice.NewService(db)

	frontendHandler := handlers.NewFrontendHandler(authService, productService, cartService, orderService, invoiceService)
	frontend := router.Group("/", frontendHandler.SessionMiddleware())

	frontend.GET("/login", frontendHandler.LoginForm)
	frontend.POST("/login", frontendHandler.LoginSubmit)
	frontend.GET("/register", frontendHandler.RegisterForm)
	frontend.POST("/register", frontendHandler.RegisterSubmit)
	frontend.GET("/logout", frontendHandler.Logout)

	frontend.GET("/products", frontendHandler.Products)
	frontend.GET("/order/configure/:slug", frontendHandler.ConfigureProduct)
	frontend.POST("/order/configure/:slug", frontendHandler.AddToCartFromProduct)

	frontend.GET("/cart", frontendHandler.Cart)
	frontend.POST("/cart/coupon", frontendHandler.ApplyCoupon)
	frontend.GET("/checkout", frontendHandler.Checkout)
	frontend.POST("/checkout", frontendHandler.PlaceOrder)
}

func registerAPIRoutes(api *gin.RouterGroup, db *gorm.DB) {
	authService := auth.NewService(db)
	productService := product.NewService(db)
	orderService := order.NewService(db)
	cartService := order.NewCartService(db)
	invoiceService := invoice.NewService(db)
	ticketService := ticket.NewService(db)
	paymentService := payment.NewService(db)
	affiliateService := affiliate.NewService(db)
	notificationService := notification.NewService(db)
	knowledgebaseService := knowledgebase.NewService(db)
	subUserService := subuser.NewService(db)

	authHandler := apiHandlers.NewAuthHandler(authService)
	productHandler := apiHandlers.NewProductHandler(productService)
	orderHandler := apiHandlers.NewOrderHandler(orderService, cartService)
	invoiceHandler := apiHandlers.NewInvoiceHandler(invoiceService)
	ticketHandler := apiHandlers.NewTicketHandler(ticketService)
	paymentHandler := apiHandlers.NewPaymentHandler(paymentService)
	affiliateHandler := apiHandlers.NewAffiliateHandler(affiliateService)
	notificationHandler := apiHandlers.NewNotificationHandler(notificationService)
	knowledgeBaseHandler := apiHandlers.NewKnowledgeBaseHandler(knowledgebaseService)
	subUserHandler := apiHandlers.NewSubUserHandler(subUserService)

	// Public endpoints
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/forgot-password", authHandler.ForgotPassword)
	api.POST("/auth/reset-password", authHandler.ResetPassword)

	api.GET("/products/groups", productHandler.ListProductGroups)
	api.GET("/products/groups/:slug", productHandler.GetProductGroup)
	api.GET("/products", productHandler.ListProducts)
	api.GET("/products/:slug", productHandler.GetProduct)
	api.POST("/products/:id/pricing", productHandler.GetProductPricing)

	api.GET("/cart", orderHandler.GetCart)
	api.POST("/cart/items", orderHandler.AddToCart)
	api.PUT("/cart/items/:id", orderHandler.UpdateCartItem)
	api.DELETE("/cart/items/:id", orderHandler.RemoveCartItem)
	api.POST("/cart/coupon", orderHandler.ApplyCoupon)
	api.DELETE("/cart/coupon", orderHandler.RemoveCoupon)
	api.DELETE("/cart", orderHandler.ClearCart)

	api.GET("/kb/categories", knowledgeBaseHandler.ListCategories)
	api.GET("/kb/categories/:slug", knowledgeBaseHandler.GetCategory)
	api.GET("/kb/articles/:slug", knowledgeBaseHandler.GetArticle)
	api.GET("/kb/search", knowledgeBaseHandler.SearchArticles)
	api.POST("/kb/articles/:slug/rate", knowledgeBaseHandler.RateArticle)
	api.GET("/kb/popular", knowledgeBaseHandler.GetPopularArticles)

	api.GET("/payments/gateways", paymentHandler.ListGateways)
	api.POST("/payments/callback/:gateway", paymentHandler.ProcessCallback)

	api.POST("/subusers/invite/:token/accept", subUserHandler.AcceptInvite)
	api.POST("/subusers/login", subUserHandler.SubUserLogin)

	api.GET("/ref/:code", affiliateHandler.TrackClick)

	// Authenticated endpoints
	authGroup := api.Group("", authHandler.AuthMiddleware())
	authGroup.POST("/auth/logout", authHandler.Logout)
	authGroup.GET("/auth/me", authHandler.GetCurrentUser)
	authGroup.PUT("/auth/profile", authHandler.UpdateProfile)
	authGroup.PUT("/auth/password", authHandler.ChangePassword)

	authGroup.GET("/orders", orderHandler.ListOrders)
	authGroup.GET("/orders/:id", orderHandler.GetOrder)
	authGroup.POST("/orders", orderHandler.CreateOrder)
	authGroup.GET("/services", orderHandler.ListServices)
	authGroup.GET("/services/:id", orderHandler.GetService)

	authGroup.GET("/invoices", invoiceHandler.ListInvoices)
	authGroup.GET("/invoices/:id", invoiceHandler.GetInvoice)
	authGroup.GET("/invoices/unpaid", invoiceHandler.GetUnpaidInvoices)

	authGroup.GET("/tickets", ticketHandler.ListTickets)
	authGroup.GET("/tickets/:id", ticketHandler.GetTicket)
	authGroup.POST("/tickets", ticketHandler.CreateTicket)
	authGroup.POST("/tickets/:id/reply", ticketHandler.ReplyToTicket)
	authGroup.POST("/tickets/:id/close", ticketHandler.CloseTicket)
	authGroup.GET("/tickets/stats", ticketHandler.GetTicketStats)

	authGroup.GET("/affiliate", affiliateHandler.GetAffiliate)
	authGroup.POST("/affiliate", affiliateHandler.Apply)
	authGroup.GET("/affiliate/commissions", affiliateHandler.GetCommissions)
	authGroup.GET("/affiliate/withdrawals", affiliateHandler.GetWithdrawals)
	authGroup.POST("/affiliate/withdraw", affiliateHandler.RequestWithdrawal)
	authGroup.PUT("/affiliate/settings", affiliateHandler.UpdateSettings)
	authGroup.GET("/affiliate/banners", affiliateHandler.GetBanners)

	authGroup.GET("/notifications", notificationHandler.GetUnreadNotifications)
	authGroup.POST("/notifications/:id/read", notificationHandler.MarkAsRead)
	authGroup.POST("/notifications/read-all", notificationHandler.MarkAllAsRead)

	authGroup.POST("/payments", paymentHandler.CreatePaymentRequest)
	authGroup.POST("/payments/:id/process", paymentHandler.ProcessPayment)
	authGroup.POST("/payments/credit", paymentHandler.PayWithCredit)
	authGroup.POST("/payments/methods", paymentHandler.SavePaymentMethod)
	authGroup.POST("/payments/methods/:id/default", paymentHandler.SetDefaultPaymentMethod)
	authGroup.DELETE("/payments/methods/:id", paymentHandler.DeletePaymentMethod)
	authGroup.POST("/payments/auto", paymentHandler.SetupAutoPayment)
	authGroup.GET("/payments/auto", paymentHandler.GetAutoPaymentConfig)

	authGroup.GET("/subusers", subUserHandler.ListSubUsers)
	authGroup.POST("/subusers/invite", subUserHandler.CreateInvite)
	authGroup.PUT("/subusers/:id", subUserHandler.UpdateSubUser)
	authGroup.DELETE("/subusers/:id", subUserHandler.DeleteSubUser)
	authGroup.POST("/subusers/logout", subUserHandler.SubUserLogout)
	authGroup.PUT("/subusers/password", subUserHandler.ChangePassword)
	authGroup.GET("/subusers/invites", subUserHandler.GetPendingInvites)
	authGroup.DELETE("/subusers/invites/:id", subUserHandler.CancelInvite)

	// Admin endpoints
	adminGroup := api.Group("/admin", authHandler.AuthMiddleware(), apiHandlers.AdminMiddleware())
	adminGroup.GET("/orders", orderHandler.AdminListOrders)
	adminGroup.PUT("/orders/:id/status", orderHandler.AdminUpdateOrderStatus)
	adminGroup.POST("/services/:id/suspend", orderHandler.AdminSuspendService)
	adminGroup.POST("/services/:id/unsuspend", orderHandler.AdminUnsuspendService)
	adminGroup.POST("/services/:id/terminate", orderHandler.AdminTerminateService)

	adminGroup.GET("/invoices", invoiceHandler.AdminListInvoices)
	adminGroup.POST("/invoices/:id/cancel", invoiceHandler.AdminCancelInvoice)

	adminGroup.GET("/tickets", ticketHandler.AdminListTickets)
	adminGroup.GET("/tickets/stats", ticketHandler.AdminGetTicketStats)
	adminGroup.PUT("/tickets/:id/status", ticketHandler.AdminUpdateTicketStatus)
	adminGroup.PUT("/tickets/:id/priority", ticketHandler.AdminUpdateTicketPriority)
	adminGroup.DELETE("/tickets/:id", ticketHandler.AdminDeleteTicket)

	adminGroup.POST("/products/groups", productHandler.CreateProductGroup)
	adminGroup.POST("/products", productHandler.CreateProduct)
	adminGroup.PUT("/products/:id", productHandler.UpdateProduct)
	adminGroup.DELETE("/products/:id", productHandler.DeleteProduct)

	adminGroup.GET("/kb/categories", knowledgeBaseHandler.AdminListCategories)
	adminGroup.POST("/kb/categories", knowledgeBaseHandler.AdminCreateCategory)
	adminGroup.PUT("/kb/categories/:id", knowledgeBaseHandler.AdminUpdateCategory)
	adminGroup.DELETE("/kb/categories/:id", knowledgeBaseHandler.AdminDeleteCategory)
	adminGroup.GET("/kb/articles", knowledgeBaseHandler.AdminListArticles)
	adminGroup.POST("/kb/articles", knowledgeBaseHandler.AdminCreateArticle)
	adminGroup.PUT("/kb/articles/:id", knowledgeBaseHandler.AdminUpdateArticle)
	adminGroup.POST("/kb/articles/:id/publish", knowledgeBaseHandler.AdminPublishArticle)
	adminGroup.POST("/kb/articles/:id/unpublish", knowledgeBaseHandler.AdminUnpublishArticle)
	adminGroup.DELETE("/kb/articles/:id", knowledgeBaseHandler.AdminDeleteArticle)
	adminGroup.GET("/kb/search-stats", knowledgeBaseHandler.AdminGetSearchStats)

	adminGroup.POST("/notifications/send", notificationHandler.AdminSendNotification)
	adminGroup.GET("/email-templates", notificationHandler.AdminListEmailTemplates)
	adminGroup.POST("/email-templates", notificationHandler.AdminCreateEmailTemplate)
	adminGroup.PUT("/email-templates/:id", notificationHandler.AdminUpdateEmailTemplate)
	adminGroup.POST("/email-templates/test", notificationHandler.AdminTestEmail)
	adminGroup.POST("/webhooks", notificationHandler.AdminCreateWebhook)

	adminGroup.POST("/payments/credit", paymentHandler.AdminAddCredit)
	adminGroup.POST("/payments/:id/refund", paymentHandler.AdminRefundPayment)

	adminGroup.GET("/affiliates", affiliateHandler.AdminListAffiliates)
	adminGroup.POST("/affiliates/:id/approve", affiliateHandler.AdminApproveAffiliate)
	adminGroup.POST("/affiliates/:id/suspend", affiliateHandler.AdminSuspendAffiliate)
	adminGroup.POST("/affiliates/withdrawals/:id/process", affiliateHandler.AdminProcessWithdrawal)
}

func ensureAdminUser(db *gorm.DB, admin config.AdminConfig) error {
	if admin.Email == "" || admin.PasswordHash == "" {
		return nil
	}

	var user domain.User
	if err := db.Where("email = ?", admin.Email).First(&user).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		user = domain.User{
			Email:        admin.Email,
			PasswordHash: admin.PasswordHash,
			FirstName:    "Admin",
			LastName:     "User",
			Role:         domain.UserRoleAdmin,
			Status:       domain.UserStatusActive,
			Language:     "en",
			Currency:     "USD",
		}
		return db.Create(&user).Error
	}

	updates := map[string]any{
		"role":          domain.UserRoleAdmin,
		"status":        domain.UserStatusActive,
		"password_hash": admin.PasswordHash,
	}
	return db.Model(&user).Updates(updates).Error
}

func ensureDefaultCatalog(db *gorm.DB) error {
	var count int64
	if err := db.Model(&domain.Product{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	group := domain.ProductGroup{
		Name:        "Default",
		Slug:        "default",
		Description: "默认产品分组",
		Active:      true,
	}
	if err := db.Create(&group).Error; err != nil {
		return err
	}

	productItem := domain.Product{
		ProductGroupID: group.ID,
		Name:           "Starter VPS",
		Slug:           "starter-vps",
		Description:    "入门级 VPS 产品，适合小型站点与开发环境。",
		ModuleName:     "provisioner-example",
		Active:         true,
	}
	if err := db.Create(&productItem).Error; err != nil {
		return err
	}

	pricing := domain.ProductPricing{
		ProductID: productItem.ID,
		Currency:  "USD",
		SetupFee:  decimal.Zero,
		Monthly:   decimal.NewFromFloat(19.99),
		Quarterly: decimal.NewFromFloat(54.99),
		Annually:  decimal.NewFromFloat(199.99),
	}
	return db.Create(&pricing).Error
}
