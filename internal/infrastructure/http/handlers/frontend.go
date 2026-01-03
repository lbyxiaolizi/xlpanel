package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/openhost/openhost/internal/infrastructure/config"
	"github.com/openhost/openhost/internal/infrastructure/web"
)

type productCard struct {
	ID           int
	Slug         string
	Name         string
	Description  string
	Category     string
	Currency     string
	Price        string
	BillingCycle string
	IsPopular    bool
	Features     []string
}

func Root(c *gin.Context) {
	installed, err := config.Exists(config.DefaultPath)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if !installed {
		c.Redirect(http.StatusSeeOther, "/install")
		return
	}
	Home(c)
}

func Home(c *gin.Context) {
	web.Render(c, "home.html", gin.H{
		"Title":       "首页",
		"Description": "OpenHost 现代化主机与账单管理平台",
		"Year":        time.Now().Year(),
	})
}

func Products(c *gin.Context) {
	web.Render(c, "products.html", gin.H{
		"Title":       "产品",
		"Description": "OpenHost 产品与价格方案",
		"Year":        time.Now().Year(),
		"Products": []productCard{
			{
				ID:           1,
				Slug:         "vps-starter",
				Name:         "VPS Starter",
				Description:  "入门级 VPS，适合轻量应用与个人站点。",
				Category:     "vps",
				Currency:     "¥",
				Price:        "29.00",
				BillingCycle: "月",
				IsPopular:    true,
				Features: []string{
					"2 vCPU",
					"2 GB 内存",
					"50 GB SSD",
					"2 TB 流量",
				},
			},
			{
				ID:           2,
				Slug:         "vps-pro",
				Name:         "VPS Pro",
				Description:  "面向中型业务的稳定 VPS 方案。",
				Category:     "vps",
				Currency:     "¥",
				Price:        "64.00",
				BillingCycle: "月",
				IsPopular:    false,
				Features: []string{
					"4 vCPU",
					"4 GB 内存",
					"80 GB SSD",
					"4 TB 流量",
				},
			},
			{
				ID:           3,
				Slug:         "dedicated-enterprise",
				Name:         "Dedicated Enterprise",
				Description:  "企业级独服，适合高负载应用。",
				Category:     "dedicated",
				Currency:     "¥",
				Price:        "499.00",
				BillingCycle: "月",
				IsPopular:    false,
				Features: []string{
					"8 核 CPU",
					"32 GB 内存",
					"2x 1 TB NVMe",
					"10 TB 流量",
				},
			},
		},
	})
}
