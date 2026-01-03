package handlers

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/openhost/openhost/internal/infrastructure/web"
)

// Admin Panel Handlers

func AdminDashboard(c *gin.Context) {
	web.Render(c, "admin/dashboard.html", gin.H{
		"Title":       "管理后台",
		"Description": "OpenHost 管理后台",
		"Year":        time.Now().Year(),
		"Section":     "dashboard",
	})
}

func AdminCustomers(c *gin.Context) {
	web.Render(c, "admin/customers.html", gin.H{
		"Title":       "客户管理",
		"Description": "管理客户",
		"Year":        time.Now().Year(),
		"Section":     "customers",
	})
}

func AdminOrders(c *gin.Context) {
	web.Render(c, "admin/orders.html", gin.H{
		"Title":       "订单管理",
		"Description": "管理订单",
		"Year":        time.Now().Year(),
		"Section":     "orders",
	})
}

func AdminServices(c *gin.Context) {
	web.Render(c, "admin/services.html", gin.H{
		"Title":       "服务管理",
		"Description": "管理服务",
		"Year":        time.Now().Year(),
		"Section":     "services",
	})
}

func AdminInvoices(c *gin.Context) {
	web.Render(c, "admin/invoices.html", gin.H{
		"Title":       "账单管理",
		"Description": "管理账单",
		"Year":        time.Now().Year(),
		"Section":     "invoices",
	})
}

func AdminTickets(c *gin.Context) {
	web.Render(c, "admin/tickets.html", gin.H{
		"Title":       "工单管理",
		"Description": "管理工单",
		"Year":        time.Now().Year(),
		"Section":     "tickets",
	})
}

func AdminProducts(c *gin.Context) {
	web.Render(c, "admin/products.html", gin.H{
		"Title":       "产品管理",
		"Description": "管理产品",
		"Year":        time.Now().Year(),
		"Section":     "products",
	})
}

func AdminServers(c *gin.Context) {
	web.Render(c, "admin/servers.html", gin.H{
		"Title":       "服务器管理",
		"Description": "管理服务器",
		"Year":        time.Now().Year(),
		"Section":     "servers",
	})
}

func AdminSettings(c *gin.Context) {
	web.Render(c, "admin/settings.html", gin.H{
		"Title":       "系统设置",
		"Description": "系统设置",
		"Year":        time.Now().Year(),
		"Section":     "settings",
	})
}
