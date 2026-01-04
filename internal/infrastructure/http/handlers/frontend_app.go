package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/openhost/openhost/internal/core/domain"
	"github.com/openhost/openhost/internal/core/service/auth"
	"github.com/openhost/openhost/internal/core/service/invoice"
	"github.com/openhost/openhost/internal/core/service/order"
	"github.com/openhost/openhost/internal/core/service/product"
	"github.com/openhost/openhost/internal/infrastructure/web"
)

const (
	cartSessionCookie     = "cart_session"
	frontendSessionCookie = "session"
)

type FrontendHandler struct {
	authService    *auth.Service
	productService *product.Service
	cartService    *order.CartService
	orderService   *order.Service
	invoiceService *invoice.Service
}

func NewFrontendHandler(
	authService *auth.Service,
	productService *product.Service,
	cartService *order.CartService,
	orderService *order.Service,
	invoiceService *invoice.Service,
) *FrontendHandler {
	return &FrontendHandler{
		authService:    authService,
		productService: productService,
		cartService:    cartService,
		orderService:   orderService,
		invoiceService: invoiceService,
	}
}

func (h *FrontendHandler) SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(frontendSessionCookie)
		if err == nil && token != "" {
			user, err := h.authService.ValidateSession(token)
			if err == nil {
				c.Set(web.ContextUserKey, user)
			}
		}
		c.Next()
	}
}

func (h *FrontendHandler) LoginForm(c *gin.Context) {
	renderAuthPage(c, "login.html", "")
}

func (h *FrontendHandler) LoginSubmit(c *gin.Context) {
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")

	if email == "" || password == "" {
		renderAuthPage(c, "login.html", "请输入邮箱和密码。")
		return
	}

	session, err := h.authService.Login(email, password, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		renderAuthPage(c, "login.html", "账号或密码错误，请重试。")
		return
	}

	setSessionCookie(c, session.ID)
	c.Redirect(http.StatusSeeOther, "/client")
}

func (h *FrontendHandler) RegisterForm(c *gin.Context) {
	renderAuthPage(c, "register.html", "")
}

func (h *FrontendHandler) RegisterSubmit(c *gin.Context) {
	firstName := strings.TrimSpace(c.PostForm("first_name"))
	lastName := strings.TrimSpace(c.PostForm("last_name"))
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")
	confirm := c.PostForm("confirm_password")
	acceptTerms := c.PostForm("accept_terms")

	if firstName == "" || lastName == "" || email == "" || password == "" {
		renderAuthPage(c, "register.html", "请完整填写注册信息。")
		return
	}
	if password != confirm {
		renderAuthPage(c, "register.html", "两次输入的密码不一致。")
		return
	}
	if acceptTerms == "" {
		renderAuthPage(c, "register.html", "请先同意服务条款。")
		return
	}

	_, err := h.authService.Register(email, password, firstName, lastName)
	if err != nil {
		switch err {
		case auth.ErrEmailExists:
			renderAuthPage(c, "register.html", "该邮箱已注册。")
		case auth.ErrPasswordTooShort:
			renderAuthPage(c, "register.html", "密码长度至少 8 位。")
		default:
			renderAuthPage(c, "register.html", "注册失败，请稍后再试。")
		}
		return
	}

	session, err := h.authService.Login(email, password, c.ClientIP(), c.GetHeader("User-Agent"))
	if err == nil {
		setSessionCookie(c, session.ID)
	}

	c.Redirect(http.StatusSeeOther, "/client")
}

func (h *FrontendHandler) Logout(c *gin.Context) {
	clearSessionCookie(c)
	c.Redirect(http.StatusSeeOther, "/")
}

func (h *FrontendHandler) Products(c *gin.Context) {
	products, _, err := h.productService.ListProducts(nil, true, 100, 0)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	currencyCode := "USD"
	if value, ok := c.Get(web.ContextCurrencyKey); ok {
		if code, ok := value.(string); ok && code != "" {
			currencyCode = code
		}
	}

	cards := make([]productCard, 0, len(products))
	for _, item := range products {
		priceLabel := ""
		billingCycle := ""
		if pricing, err := h.productService.GetPricing(item.ID, currencyCode); err == nil {
			cycle, amount := pickPreferredCycle(pricing)
			billingCycle = cycle
			priceLabel = amount.StringFixed(2)
		}
		cards = append(cards, productCard{
			ID:           int(item.ID),
			Slug:         item.Slug,
			Name:         item.Name,
			Description:  item.Description,
			Currency:     currencyCode,
			Price:        priceLabel,
			BillingCycle: billingCycle,
		})
	}

	web.Render(c, "products.html", gin.H{
		"Title":       "产品",
		"Description": "OpenHost 产品与价格方案",
		"Year":        time.Now().Year(),
		"Products":    cards,
	})
}

func (h *FrontendHandler) ConfigureProduct(c *gin.Context) {
	slug := c.Param("slug")
	productItem, err := h.productService.GetProductBySlug(slug)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	currencyCode := "USD"
	if value, ok := c.Get(web.ContextCurrencyKey); ok {
		if code, ok := value.(string); ok && code != "" {
			currencyCode = code
		}
	}
	pricing, _ := h.productService.GetPricing(productItem.ID, currencyCode)
	billingCycles := availableCycles(pricing)
	configGroups := buildConfigGroups(productItem.ConfigGroups)
	cycle, amount := pickPreferredCycle(pricing)

	h.renderConfigure(c, productItem, billingCycles, configGroups, cycle, amount.StringFixed(2), currencyCode, "")
}

func (h *FrontendHandler) AddToCartFromProduct(c *gin.Context) {
	slug := c.Param("slug")
	productItem, err := h.productService.GetProductBySlug(slug)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	quantity, _ := strconv.Atoi(c.PostForm("quantity"))
	billingCycle := c.PostForm("billing_cycle")
	configOptions := parseConfigOptions(c)

	cart, err := h.getOrCreateCart(c)
	if err != nil {
		h.renderConfigureFromProduct(c, productItem, "无法创建购物车，请稍后再试。")
		return
	}

	_, err = h.cartService.AddItem(cart.ID, productItem.ID, quantity, billingCycle, "", "", configOptions)
	if err != nil {
		h.renderConfigureFromProduct(c, productItem, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/cart")
}

func (h *FrontendHandler) Cart(c *gin.Context) {
	h.renderCart(c, "")
}

func (h *FrontendHandler) ApplyCoupon(c *gin.Context) {
	cart, err := h.getOrCreateCart(c)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	coupon := strings.TrimSpace(c.PostForm("coupon"))
	if coupon == "" {
		h.renderCart(c, "请输入优惠码。")
		return
	}

	if err := h.cartService.ApplyCoupon(cart.ID, coupon); err != nil {
		h.renderCart(c, "优惠码无效或已过期。")
		return
	}

	c.Redirect(http.StatusSeeOther, "/cart")
}

func (h *FrontendHandler) Checkout(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	cart, err := h.cartService.GetOrCreateCart(&user.ID, "")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	summary, err := h.cartService.GetCartSummary(cart.ID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	view := cartSummaryViewFrom(summary)

	web.Render(c, "checkout.html", gin.H{
		"Title":       "结账",
		"Description": "完成订单",
		"Year":        time.Now().Year(),
		"Cart":        view,
		"User":        user,
	})
}

func (h *FrontendHandler) PlaceOrder(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	cart, err := h.cartService.GetOrCreateCart(&user.ID, "")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	orderRecord, err := h.orderService.CreateOrder(user.ID, cart.ID, c.ClientIP())
	if err != nil {
		h.renderCart(c, "订单创建失败，请稍后再试。")
		return
	}

	invoiceRecord, err := h.invoiceService.CreateInvoiceFromOrder(orderRecord, time.Now().Add(7*24*time.Hour))
	if err != nil {
		h.renderCart(c, "账单生成失败，请联系支持。")
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/client/invoices/%d", invoiceRecord.ID))
}

func (h *FrontendHandler) renderCart(c *gin.Context, errorMessage string) {
	cart, err := h.getOrCreateCart(c)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	summary, err := h.cartService.GetCartSummary(cart.ID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	view := cartSummaryViewFrom(summary)

	data := gin.H{
		"Title":       "购物车",
		"Description": "您的购物车",
		"Year":        time.Now().Year(),
		"Cart":        view,
	}
	if errorMessage != "" {
		data["Flash"] = &web.Flash{Type: "error", Message: errorMessage}
	}

	web.Render(c, "cart.html", data)
}

func (h *FrontendHandler) renderConfigureFromProduct(c *gin.Context, productItem *domain.Product, message string) {
	currencyCode := "USD"
	if value, ok := c.Get(web.ContextCurrencyKey); ok {
		if code, ok := value.(string); ok && code != "" {
			currencyCode = code
		}
	}
	pricing, _ := h.productService.GetPricing(productItem.ID, currencyCode)
	billingCycles := availableCycles(pricing)
	configGroups := buildConfigGroups(productItem.ConfigGroups)
	cycle, amount := pickPreferredCycle(pricing)
	h.renderConfigure(c, productItem, billingCycles, configGroups, cycle, amount.StringFixed(2), currencyCode, message)
}

func (h *FrontendHandler) renderConfigure(
	c *gin.Context,
	productItem *domain.Product,
	billingCycles []string,
	configGroups []configGroupView,
	defaultCycle string,
	price string,
	currency string,
	message string,
) {
	data := gin.H{
		"Title":        "配置产品",
		"Description":  "配置您的产品",
		"Year":         time.Now().Year(),
		"Product":      productItem,
		"Billing":      billingCycles,
		"DefaultCycle": defaultCycle,
		"Price":        price,
		"Currency":     currency,
		"ConfigGroups": configGroups,
	}
	if message != "" {
		data["Flash"] = &web.Flash{Type: "error", Message: message}
	}
	web.Render(c, "configure.html", data)
}

func (h *FrontendHandler) getOrCreateCart(c *gin.Context) (*domain.Cart, error) {
	user := currentUser(c)
	if user != nil {
		return h.cartService.GetOrCreateCart(&user.ID, "")
	}

	sessionID, err := c.Cookie(cartSessionCookie)
	if err != nil || sessionID == "" {
		newSession, err := generateSessionToken()
		if err != nil {
			return nil, err
		}
		sessionID = newSession
		c.SetCookie(cartSessionCookie, sessionID, int((30 * 24 * time.Hour).Seconds()), "/", "", c.Request.TLS != nil, true)
	}
	return h.cartService.GetOrCreateCart(nil, sessionID)
}

func currentUser(c *gin.Context) *domain.User {
	value, ok := c.Get(web.ContextUserKey)
	if !ok {
		return nil
	}
	user, ok := value.(*domain.User)
	if !ok {
		return nil
	}
	return user
}

func setSessionCookie(c *gin.Context, token string) {
	maxAge := int(auth.SessionDuration.Seconds())
	c.SetCookie(frontendSessionCookie, token, maxAge, "/", "", c.Request.TLS != nil, true)
}

func clearSessionCookie(c *gin.Context) {
	c.SetCookie(frontendSessionCookie, "", -1, "/", "", c.Request.TLS != nil, true)
}

func generateSessionToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func parseConfigOptions(c *gin.Context) domain.JSONMap {
	if err := c.Request.ParseForm(); err != nil {
		return domain.JSONMap{}
	}
	config := domain.JSONMap{}
	for key, values := range c.Request.PostForm {
		if !strings.HasPrefix(key, "option_") || len(values) == 0 {
			continue
		}
		optionID := strings.TrimPrefix(key, "option_")
		config[optionID] = values[0]
	}
	return config
}

type configGroupView struct {
	Name    string
	Options []configOptionView
}

type configOptionView struct {
	ID         uint64
	Name       string
	Required   bool
	SubOptions []configSubOptionView
}

type configSubOptionView struct {
	ID    uint64
	Name  string
	Price string
}

func buildConfigGroups(groups []domain.ConfigGroup) []configGroupView {
	views := make([]configGroupView, 0, len(groups))
	for _, group := range groups {
		view := configGroupView{Name: group.Name}
		for _, option := range group.Options {
			optView := configOptionView{
				ID:       option.ID,
				Name:     option.Name,
				Required: option.Required,
			}
			for _, subOption := range option.SubOptions {
				optView.SubOptions = append(optView.SubOptions, configSubOptionView{
					ID:    subOption.ID,
					Name:  subOption.Name,
					Price: subOption.Pricing.Monthly.StringFixed(2),
				})
			}
			view.Options = append(view.Options, optView)
		}
		views = append(views, view)
	}
	return views
}

func availableCycles(pricing *domain.ProductPricing) []string {
	if pricing == nil {
		return []string{"monthly"}
	}
	cycles := []string{}
	if pricing.IsCycleEnabled("monthly") {
		cycles = append(cycles, "monthly")
	}
	if pricing.IsCycleEnabled("quarterly") {
		cycles = append(cycles, "quarterly")
	}
	if pricing.IsCycleEnabled("annually") {
		cycles = append(cycles, "annually")
	}
	if pricing.IsCycleEnabled("biennially") {
		cycles = append(cycles, "biennially")
	}
	if pricing.IsCycleEnabled("triennially") {
		cycles = append(cycles, "triennially")
	}
	if len(cycles) == 0 {
		cycles = append(cycles, "monthly")
	}
	return cycles
}

func pickPreferredCycle(pricing *domain.ProductPricing) (string, decimal.Decimal) {
	if pricing == nil {
		return "monthly", decimal.Zero
	}
	if pricing.IsCycleEnabled("monthly") {
		return "monthly", pricing.Monthly
	}
	if pricing.IsCycleEnabled("quarterly") {
		return "quarterly", pricing.Quarterly
	}
	if pricing.IsCycleEnabled("annually") {
		return "annually", pricing.Annually
	}
	if pricing.IsCycleEnabled("biennially") {
		return "biennially", pricing.Biennially
	}
	if pricing.IsCycleEnabled("triennially") {
		return "triennially", pricing.Triennially
	}
	return "monthly", pricing.Monthly
}

type cartItemView struct {
	ID           uint64
	ProductName  string
	BillingCycle string
	Quantity     int
	Total        string
}

type cartSummaryView struct {
	Items      []cartItemView
	Subtotal   string
	Discount   string
	Tax        string
	Total      string
	Currency   string
	HasItems   bool
	CouponCode string
}

func cartSummaryViewFrom(summary *order.CartSummary) cartSummaryView {
	view := cartSummaryView{
		Currency:   summary.Currency,
		CouponCode: summary.CouponCode,
	}
	for _, item := range summary.Items {
		view.Items = append(view.Items, cartItemView{
			ID:           item.ID,
			ProductName:  item.ProductName,
			BillingCycle: item.BillingCycle,
			Quantity:     item.Quantity,
			Total:        item.Total.StringFixed(2),
		})
	}
	view.Subtotal = summary.Subtotal.StringFixed(2)
	view.Discount = summary.TotalDiscount.StringFixed(2)
	view.Tax = summary.Tax.StringFixed(2)
	view.Total = summary.Total.StringFixed(2)
	view.HasItems = len(view.Items) > 0
	return view
}

func renderAuthPage(c *gin.Context, templateName string, message string) {
	data := gin.H{
		"Title":       "认证",
		"Description": "登录或注册 OpenHost",
		"Year":        time.Now().Year(),
	}
	if message != "" {
		data["Flash"] = &web.Flash{Type: "error", Message: message}
	}
	web.Render(c, templateName, data)
}
