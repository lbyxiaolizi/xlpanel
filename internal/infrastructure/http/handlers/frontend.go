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

func Pricing(c *gin.Context) {
	web.Render(c, "pricing.html", gin.H{
		"Title":       "价格",
		"Description": "OpenHost 价格方案",
		"Year":        time.Now().Year(),
	})
}

func Login(c *gin.Context) {
	web.Render(c, "login.html", gin.H{
		"Title":       "登录",
		"Description": "登录到 OpenHost",
		"Year":        time.Now().Year(),
	})
}

func Register(c *gin.Context) {
	web.Render(c, "register.html", gin.H{
		"Title":       "注册",
		"Description": "注册 OpenHost 账户",
		"Year":        time.Now().Year(),
	})
}

func Features(c *gin.Context) {
	web.Render(c, "features.html", gin.H{
		"Title":       "功能特性",
		"Description": "OpenHost 功能特性",
		"Year":        time.Now().Year(),
	})
}

func About(c *gin.Context) {
	web.Render(c, "about.html", gin.H{
		"Title":       "关于我们",
		"Description": "关于 OpenHost",
		"Year":        time.Now().Year(),
	})
}

func Contact(c *gin.Context) {
	web.Render(c, "contact.html", gin.H{
		"Title":       "联系我们",
		"Description": "联系 OpenHost",
		"Year":        time.Now().Year(),
	})
}

func Docs(c *gin.Context) {
	web.Render(c, "docs.html", gin.H{
		"Title":       "文档",
		"Description": "OpenHost 文档",
		"Year":        time.Now().Year(),
	})
}

func Support(c *gin.Context) {
	web.Render(c, "support.html", gin.H{
		"Title":       "支持",
		"Description": "OpenHost 支持中心",
		"Year":        time.Now().Year(),
	})
}

func Cart(c *gin.Context) {
	web.Render(c, "cart.html", gin.H{
		"Title":       "购物车",
		"Description": "您的购物车",
		"Year":        time.Now().Year(),
	})
}

func Checkout(c *gin.Context) {
	web.Render(c, "checkout.html", gin.H{
		"Title":       "结账",
		"Description": "完成订单",
		"Year":        time.Now().Year(),
	})
}

func ConfigureProduct(c *gin.Context) {
	slug := c.Param("slug")
	web.Render(c, "configure.html", gin.H{
		"Title":       "配置产品",
		"Description": "配置您的产品",
		"Year":        time.Now().Year(),
		"ProductSlug": slug,
	})
}
