package i18n

import (
	"encoding/json"
	"fmt"
)

// Translator handles translations
type Translator struct {
	translations map[string]map[string]string
}

// NewTranslator creates a new translator with default translations
func NewTranslator() *Translator {
	t := &Translator{
		translations: make(map[string]map[string]string),
	}
	t.loadDefaultTranslations()
	return t
}

// T translates a key to the specified language
func (t *Translator) T(lang, key string) string {
	if langMap, ok := t.translations[lang]; ok {
		if translation, ok := langMap[key]; ok {
			return translation
		}
	}
	// Fallback to English
	if langMap, ok := t.translations["en"]; ok {
		if translation, ok := langMap[key]; ok {
			return translation
		}
	}
	return key
}

func (t *Translator) loadDefaultTranslations() {
	// English translations
	t.translations["en"] = map[string]string{
		// Navigation
		"nav.dashboard":       "Dashboard",
		"nav.products":        "Products",
		"nav.cart":            "Cart",
		"nav.orders":          "Orders",
		"nav.invoices":        "Invoices",
		"nav.support":         "Support",
		"nav.account":         "Account",
		"nav.logout":          "Logout",
		"nav.login":           "Login",
		"nav.register":        "Register",
		
		// Dashboard
		"dashboard.title":           "Dashboard",
		"dashboard.welcome":         "Welcome",
		"dashboard.active_services": "Active Services",
		"dashboard.unpaid_invoices": "Unpaid Invoices",
		"dashboard.total_customers": "Total Customers",
		"dashboard.monthly_revenue": "Monthly Revenue",
		
		// Products
		"products.title":       "Products",
		"products.name":        "Product Name",
		"products.description": "Description",
		"products.price":       "Price",
		"products.add_to_cart": "Add to Cart",
		"products.select":      "Select Product",
		"products.recurring":   "Recurring",
		"products.one_time":    "One-time",
		
		// Cart
		"cart.title":         "Shopping Cart",
		"cart.empty":         "Your cart is empty",
		"cart.item":          "Item",
		"cart.quantity":      "Quantity",
		"cart.price":         "Price",
		"cart.total":         "Total",
		"cart.checkout":      "Checkout",
		"cart.clear":         "Clear Cart",
		"cart.remove":        "Remove",
		"cart.update":        "Update",
		
		// Auth
		"auth.login":              "Login",
		"auth.register":           "Register",
		"auth.email":              "Email",
		"auth.password":           "Password",
		"auth.name":               "Name",
		"auth.confirm_password":   "Confirm Password",
		"auth.forgot_password":    "Forgot Password?",
		"auth.remember_me":        "Remember Me",
		"auth.2fa_code":           "2FA Code",
		"auth.enable_2fa":         "Enable 2FA",
		"auth.disable_2fa":        "Disable 2FA",
		"auth.2fa_enabled":        "Two-Factor Authentication Enabled",
		"auth.2fa_setup":          "Setup Two-Factor Authentication",
		
		// Orders
		"orders.title":        "Orders",
		"orders.order_id":     "Order ID",
		"orders.date":         "Date",
		"orders.status":       "Status",
		"orders.total":        "Total",
		"orders.view_details": "View Details",
		
		// Invoices
		"invoices.title":       "Invoices",
		"invoices.invoice_id":  "Invoice ID",
		"invoices.date":        "Date",
		"invoices.due_date":    "Due Date",
		"invoices.amount":      "Amount",
		"invoices.status":      "Status",
		"invoices.pay":         "Pay",
		"invoices.download":    "Download",
		
		// Common
		"common.save":        "Save",
		"common.cancel":      "Cancel",
		"common.delete":      "Delete",
		"common.edit":        "Edit",
		"common.add":         "Add",
		"common.search":      "Search",
		"common.filter":      "Filter",
		"common.actions":     "Actions",
		"common.submit":      "Submit",
		"common.back":        "Back",
		"common.next":        "Next",
		"common.previous":    "Previous",
		"common.language":    "Language",
		
		// Messages
		"msg.success":           "Success",
		"msg.error":             "Error",
		"msg.loading":           "Loading...",
		"msg.no_data":           "No data available",
		"msg.confirm_delete":    "Are you sure you want to delete this item?",
		"msg.item_added":        "Item added to cart",
		"msg.item_removed":      "Item removed from cart",
		"msg.login_success":     "Login successful",
		"msg.register_success":  "Registration successful",
		"msg.logout_success":    "Logout successful",
	}

	// Chinese translations
	t.translations["zh"] = map[string]string{
		// Navigation
		"nav.dashboard":       "仪表板",
		"nav.products":        "产品",
		"nav.cart":            "购物车",
		"nav.orders":          "订单",
		"nav.invoices":        "账单",
		"nav.support":         "支持",
		"nav.account":         "账户",
		"nav.logout":          "登出",
		"nav.login":           "登录",
		"nav.register":        "注册",
		
		// Dashboard
		"dashboard.title":           "仪表板",
		"dashboard.welcome":         "欢迎",
		"dashboard.active_services": "活跃服务",
		"dashboard.unpaid_invoices": "未付账单",
		"dashboard.total_customers": "客户总数",
		"dashboard.monthly_revenue": "月度收入",
		
		// Products
		"products.title":       "产品",
		"products.name":        "产品名称",
		"products.description": "描述",
		"products.price":       "价格",
		"products.add_to_cart": "加入购物车",
		"products.select":      "选择产品",
		"products.recurring":   "周期性",
		"products.one_time":    "一次性",
		
		// Cart
		"cart.title":         "购物车",
		"cart.empty":         "您的购物车是空的",
		"cart.item":          "项目",
		"cart.quantity":      "数量",
		"cart.price":         "价格",
		"cart.total":         "总计",
		"cart.checkout":      "结账",
		"cart.clear":         "清空购物车",
		"cart.remove":        "移除",
		"cart.update":        "更新",
		
		// Auth
		"auth.login":              "登录",
		"auth.register":           "注册",
		"auth.email":              "邮箱",
		"auth.password":           "密码",
		"auth.name":               "姓名",
		"auth.confirm_password":   "确认密码",
		"auth.forgot_password":    "忘记密码？",
		"auth.remember_me":        "记住我",
		"auth.2fa_code":           "双因素认证码",
		"auth.enable_2fa":         "启用双因素认证",
		"auth.disable_2fa":        "禁用双因素认证",
		"auth.2fa_enabled":        "双因素认证已启用",
		"auth.2fa_setup":          "设置双因素认证",
		
		// Orders
		"orders.title":        "订单",
		"orders.order_id":     "订单编号",
		"orders.date":         "日期",
		"orders.status":       "状态",
		"orders.total":        "总计",
		"orders.view_details": "查看详情",
		
		// Invoices
		"invoices.title":       "账单",
		"invoices.invoice_id":  "账单编号",
		"invoices.date":        "日期",
		"invoices.due_date":    "到期日期",
		"invoices.amount":      "金额",
		"invoices.status":      "状态",
		"invoices.pay":         "支付",
		"invoices.download":    "下载",
		
		// Common
		"common.save":        "保存",
		"common.cancel":      "取消",
		"common.delete":      "删除",
		"common.edit":        "编辑",
		"common.add":         "添加",
		"common.search":      "搜索",
		"common.filter":      "筛选",
		"common.actions":     "操作",
		"common.submit":      "提交",
		"common.back":        "返回",
		"common.next":        "下一步",
		"common.previous":    "上一步",
		"common.language":    "语言",
		
		// Messages
		"msg.success":           "成功",
		"msg.error":             "错误",
		"msg.loading":           "加载中...",
		"msg.no_data":           "无可用数据",
		"msg.confirm_delete":    "您确定要删除此项吗？",
		"msg.item_added":        "项目已添加到购物车",
		"msg.item_removed":      "项目已从购物车中移除",
		"msg.login_success":     "登录成功",
		"msg.register_success":  "注册成功",
		"msg.logout_success":    "登出成功",
	}
}

// GetAvailableLanguages returns a list of available languages
func (t *Translator) GetAvailableLanguages() []string {
	langs := make([]string, 0, len(t.translations))
	for lang := range t.translations {
		langs = append(langs, lang)
	}
	return langs
}

// Export exports translations as JSON
func (t *Translator) Export(lang string) (string, error) {
	if langMap, ok := t.translations[lang]; ok {
		data, err := json.MarshalIndent(langMap, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", fmt.Errorf("language not found: %s", lang)
}
