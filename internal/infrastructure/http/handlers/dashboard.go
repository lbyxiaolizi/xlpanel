package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/openhost/openhost/internal/infrastructure/config"
	"github.com/openhost/openhost/internal/infrastructure/web"
)

type dashboardStats struct {
	ActiveServices  int
	PendingInvoices int
	OpenTickets     int
	TotalSpent      string
}

type dashboardService struct {
	ID           int
	ProductName  string
	Hostname     string
	Currency     string
	Price        string
	BillingCycle string
	Status       string
}

type dashboardInvoice struct {
	ID            int
	InvoiceNumber string
	CreatedAt     time.Time
	Currency      string
	Total         string
	Status        string
}

type dashboardUser struct {
	FirstName string
}

func Dashboard(c *gin.Context) {
	installed, err := config.Exists(config.DefaultPath)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if !installed {
		c.Redirect(http.StatusSeeOther, "/install")
		return
	}
	web.Render(c, "dashboard.html", gin.H{
		"Title":       "控制面板",
		"Description": "OpenHost 控制面板概览",
		"Year":        time.Now().Year(),
		"Currency":    "¥",
		"User":        dashboardUser{FirstName: "Admin"},
		"Stats": dashboardStats{
			ActiveServices:  2,
			PendingInvoices: 1,
			OpenTickets:     0,
			TotalSpent:      "128.00",
		},
		"Services": []dashboardService{
			{
				ID:           1,
				ProductName:  "云服务器 Pro",
				Hostname:     "vps-01.openhost.local",
				Currency:     "¥",
				Price:        "64.00",
				BillingCycle: "月",
				Status:       "active",
			},
			{
				ID:           2,
				ProductName:  "企业独立服务器",
				Hostname:     "dedicated-02.openhost.local",
				Currency:     "¥",
				Price:        "64.00",
				BillingCycle: "月",
				Status:       "active",
			},
		},
		"Invoices": []dashboardInvoice{
			{
				ID:            1001,
				InvoiceNumber: "INV-2024-1001",
				CreatedAt:     time.Now().AddDate(0, 0, -2),
				Currency:      "¥",
				Total:         "64.00",
				Status:        "unpaid",
			},
			{
				ID:            1000,
				InvoiceNumber: "INV-2024-1000",
				CreatedAt:     time.Now().AddDate(0, 0, -32),
				Currency:      "¥",
				Total:         "64.00",
				Status:        "paid",
			},
		},
	})
}
