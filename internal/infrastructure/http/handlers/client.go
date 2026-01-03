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

func ClientAffiliates(c *gin.Context) {
	web.Render(c, "client/affiliates.html", gin.H{
		"Title":       "推广联盟",
		"Description": "管理您的推广联盟账户",
		"Year":        time.Now().Year(),
		"Section":     "affiliates",
	})
}

func ClientDomains(c *gin.Context) {
	web.Render(c, "client/domains.html", gin.H{
		"Title":       "域名管理",
		"Description": "管理您的域名",
		"Year":        time.Now().Year(),
		"Section":     "domains",
	})
}

func ClientSubUsers(c *gin.Context) {
	web.Render(c, "client/subusers.html", gin.H{
		"Title":       "子用户管理",
		"Description": "管理您的子用户",
		"Year":        time.Now().Year(),
		"Section":     "subusers",
	})
}

func ClientSecurity(c *gin.Context) {
	web.Render(c, "client/security.html", gin.H{
		"Title":       "账户安全",
		"Description": "管理您的账户安全设置",
		"Year":        time.Now().Year(),
		"Section":     "security",
	})
}
