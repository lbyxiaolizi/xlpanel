package handlers

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/openhost/openhost/internal/infrastructure/web"
)

// Client Panel Handlers

func ClientDashboard(c *gin.Context) {
	web.Render(c, "client/dashboard.html", gin.H{
		"Title":       "控制面板",
		"Description": "您的控制面板",
		"Year":        time.Now().Year(),
		"Section":     "dashboard",
	})
}

func ClientServices(c *gin.Context) {
	web.Render(c, "client/services.html", gin.H{
		"Title":       "我的服务",
		"Description": "管理您的服务",
		"Year":        time.Now().Year(),
		"Section":     "services",
	})
}

func ClientServiceDetail(c *gin.Context) {
	serviceID := c.Param("id")
	web.Render(c, "client/service_detail.html", gin.H{
		"Title":       "服务详情",
		"Description": "服务详情",
		"Year":        time.Now().Year(),
		"Section":     "services",
		"ServiceID":   serviceID,
	})
}

func ClientInvoices(c *gin.Context) {
	web.Render(c, "client/invoices.html", gin.H{
		"Title":       "我的账单",
		"Description": "查看您的账单",
		"Year":        time.Now().Year(),
		"Section":     "invoices",
	})
}

func ClientInvoiceDetail(c *gin.Context) {
	invoiceID := c.Param("id")
	web.Render(c, "client/invoice_detail.html", gin.H{
		"Title":       "账单详情",
		"Description": "账单详情",
		"Year":        time.Now().Year(),
		"Section":     "invoices",
		"InvoiceID":   invoiceID,
	})
}

func ClientTickets(c *gin.Context) {
	web.Render(c, "client/tickets.html", gin.H{
		"Title":       "工单支持",
		"Description": "查看您的工单",
		"Year":        time.Now().Year(),
		"Section":     "tickets",
	})
}

func ClientTicketDetail(c *gin.Context) {
	ticketID := c.Param("id")
	web.Render(c, "client/ticket_detail.html", gin.H{
		"Title":       "工单详情",
		"Description": "工单详情",
		"Year":        time.Now().Year(),
		"Section":     "tickets",
		"TicketID":    ticketID,
	})
}

func ClientNewTicket(c *gin.Context) {
	web.Render(c, "client/new_ticket.html", gin.H{
		"Title":       "提交工单",
		"Description": "提交新工单",
		"Year":        time.Now().Year(),
		"Section":     "tickets",
	})
}

func ClientProfile(c *gin.Context) {
	web.Render(c, "client/profile.html", gin.H{
		"Title":       "个人资料",
		"Description": "管理您的个人资料",
		"Year":        time.Now().Year(),
		"Section":     "profile",
	})
}
