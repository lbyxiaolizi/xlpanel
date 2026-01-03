# XLPanel - 完整功能清单 / Complete Feature List

XLPanel 是一个面向托管服务商的财务与自动化平台，支持中文/英文双语界面，包含完整的产品选购、购物车、用户认证、管理端功能和邮件通知系统。

XLPanel is a financial and automation platform for hosting service providers, supporting Chinese/English bilingual interfaces, with complete product selection, shopping cart, user authentication, admin panel, and email notification features.

## 新增功能 / New Features

### 1. 国际化 (i18n) / Internationalization

- ✅ 支持中文/英文双语切换 / Support Chinese/English language switching
- ✅ 语言选择器集成在所有页面 / Language selector integrated on all pages
- ✅ 超过100个翻译键值对 / Over 100 translation key-value pairs
- ✅ 可扩展的翻译系统 / Extensible translation system

**位置 / Location**: `internal/i18n/translator.go`

### 2. 用户认证系统 / Authentication System

#### 登录/注册 / Login/Registration
- ✅ JWT 令牌认证 / JWT token authentication
- ✅ 密码加密 (bcrypt) / Password hashing (bcrypt)
- ✅ 会话管理 / Session management
- ✅ 注册欢迎邮件 / Welcome email on registration

**前端页面 / Frontend Pages**:
- `/login` - 登录页面 / Login page
- `/register` - 注册页面 / Registration page

**API 端点 / API Endpoints**:
- `POST /api/login` - 用户登录 / User login
- `POST /api/register` - 用户注册 / User registration
- `POST /api/logout` - 用户登出 / User logout

#### 双因素认证 (2FA) / Two-Factor Authentication
- ✅ TOTP (Time-based One-Time Password) 支持 / TOTP support
- ✅ 2FA 设置邮件通知 / 2FA setup email notification
- ✅ 启用/禁用 2FA / Enable/disable 2FA

**API 端点 / API Endpoints**:
- `POST /api/2fa/enable` - 启用 2FA / Enable 2FA
- `POST /api/2fa/verify` - 验证 2FA 代码 / Verify 2FA code
- `POST /api/2fa/disable` - 禁用 2FA / Disable 2FA

### 3. 客户端功能 / Customer Portal

#### 产品目录 / Product Catalog
- ✅ 产品浏览界面 / Product browsing interface
- ✅ 产品详情展示 / Product details display
- ✅ 周期性/一次性产品标识 / Recurring/one-time product labels
- ✅ 一键加入购物车 / One-click add to cart

**页面 / Page**: `/customer/products`

#### 购物车 / Shopping Cart
- ✅ 添加/移除商品 / Add/remove items
- ✅ 更新数量 / Update quantities
- ✅ 实时价格计算 / Real-time price calculation
- ✅ 清空购物车 / Clear cart
- ✅ 结账功能 / Checkout functionality

**页面 / Page**: `/customer/cart`

**API 端点 / API Endpoints**:
- `GET /api/cart` - 获取购物车 / Get cart
- `POST /api/cart/add` - 添加商品 / Add item
- `PUT /api/cart/update` - 更新数量 / Update quantity
- `DELETE /api/cart/remove` - 移除商品 / Remove item
- `DELETE /api/cart/clear` - 清空购物车 / Clear cart
- `POST /api/cart/checkout` - 结账 / Checkout

#### 客户仪表板 / Customer Dashboard
- ✅ 活跃服务统计 / Active services statistics
- ✅ 购物车商品数量 / Cart items count
- ✅ 未付账单统计 / Unpaid invoices count
- ✅ 快捷操作链接 / Quick action links
- ✅ 最近订单展示 / Recent orders display

**页面 / Page**: `/customer`

#### 订单管理 / Order Management
- ✅ 订单列表展示 / Order list display
- ✅ 订单状态跟踪 / Order status tracking
- ✅ 订单详情查看 / Order details view

**页面 / Page**: `/customer/orders`

#### 账单管理 / Invoice Management
- ✅ 账单列表展示 / Invoice list display
- ✅ 账单状态查看 / Invoice status view
- ✅ 支付功能接口 / Payment interface (ready for integration)
- ✅ 账单下载功能 / Invoice download feature (ready for integration)

**页面 / Page**: `/customer/invoices`

### 4. 管理端功能 / Admin Panel

#### 用户管理 / User Management
- ✅ 用户列表展示 / User list display
- ✅ 用户角色管理 / User role management
- ✅ 2FA 状态查看 / 2FA status view
- ✅ 添加新用户 / Add new user
- ✅ 用户创建时间 / User creation time

**页面 / Page**: `/admin/users`

**API 端点 / API Endpoints**:
- `GET /api/users` - 获取用户列表 / Get user list

### 5. 邮件功能 / Email Features

#### SMTP 配置 / SMTP Configuration
- ✅ 可配置的 SMTP 服务器 / Configurable SMTP server
- ✅ 支持认证 / Authentication support
- ✅ 自定义发件人 / Custom sender
- ✅ HTML 邮件模板 / HTML email templates

**环境变量 / Environment Variables**:
```bash
XLPANEL_SMTP_HOST=smtp.example.com
XLPANEL_SMTP_PORT=587
XLPANEL_SMTP_USERNAME=user@example.com
XLPANEL_SMTP_PASSWORD=password
XLPANEL_SMTP_FROM=noreply@example.com
XLPANEL_SMTP_FROM_NAME=XLPanel
```

#### 邮件模板 / Email Templates
- ✅ 欢迎邮件 / Welcome email
- ✅ 2FA 设置邮件 / 2FA setup email
- ✅ 账单通知邮件 / Invoice notification email
- ✅ 密码重置邮件 / Password reset email
- ✅ 中英文双语模板 / Chinese/English bilingual templates

**服务位置 / Service Location**: `internal/service/email.go`

### 6. 安全功能 / Security Features

- ✅ JWT 令牌认证 / JWT token authentication
- ✅ 密码哈希 (bcrypt) / Password hashing (bcrypt)
- ✅ 2FA 支持 / 2FA support
- ✅ TOTP 密钥生成 / TOTP secret generation
- ✅ 会话管理 / Session management
- ✅ 安全请求头 / Security headers
- ✅ CSRF 保护 (准备就绪) / CSRF protection (ready)

## 配置 / Configuration

### JWT 配置 / JWT Configuration
```bash
XLPANEL_JWT_SECRET=your-secret-key-change-in-production
XLPANEL_JWT_EXPIRATION=24h
```

### 主题配置 / Theme Configuration
```bash
XLPANEL_DEFAULT_THEME=classic
XLPANEL_ALLOW_THEME_OVERRIDE=true
```

## 数据模型 / Data Models

### 新增实体 / New Entities

#### User (用户)
```go
type User struct {
    ID               string
    TenantID         string
    Email            string
    PasswordHash     string
    Name             string
    Role             string  // admin, customer
    TwoFactorSecret  string
    TwoFactorEnabled bool
    CreatedAt        time.Time
    LastLoginAt      *time.Time
}
```

#### Session (会话)
```go
type Session struct {
    ID        string
    UserID    string
    Token     string
    ExpiresAt time.Time
    CreatedAt time.Time
}
```

#### Cart (购物车)
```go
type Cart struct {
    ID         string
    CustomerID string
    TenantID   string
    Items      []CartItem
    UpdatedAt  time.Time
}
```

#### CartItem (购物车项)
```go
type CartItem struct {
    ProductCode string
    Quantity    int
    UnitPrice   float64
}
```

## 依赖项 / Dependencies

```go
require (
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/pquerna/otp v1.4.0
    golang.org/x/crypto v0.18.0
)
```

## 目录结构 / Directory Structure

```
frontend/themes/classic/
├── base.html                  # 管理端基础模板 / Admin base template
├── home.html                  # 管理端首页 / Admin home
├── login.html                 # 登录页面 / Login page
├── register.html              # 注册页面 / Registration page
├── customer/
│   ├── base.html             # 客户端基础模板 / Customer base template
│   ├── dashboard.html        # 客户仪表板 / Customer dashboard
│   ├── products.html         # 产品目录 / Product catalog
│   ├── cart.html             # 购物车 / Shopping cart
│   ├── orders.html           # 订单列表 / Order list
│   └── invoices.html         # 账单列表 / Invoice list
└── admin/
    └── users.html            # 用户管理 / User management

internal/
├── api/
│   ├── server.go             # 主服务器 / Main server
│   ├── handlers_auth.go      # 认证处理器 / Auth handlers
│   └── handlers_customer.go  # 客户处理器 / Customer handlers
├── service/
│   ├── auth.go               # 认证服务 / Auth service
│   ├── cart.go               # 购物车服务 / Cart service
│   └── email.go              # 邮件服务 / Email service
├── i18n/
│   └── translator.go         # 翻译服务 / Translation service
├── domain/
│   └── entities.go           # 实体定义 / Entity definitions
└── core/
    └── config.go             # 配置管理 / Configuration management
```

## 使用示例 / Usage Examples

### 启动服务器 / Start Server
```bash
# 设置环境变量 / Set environment variables
export XLPANEL_JWT_SECRET=your-secret-key
export XLPANEL_SMTP_HOST=smtp.gmail.com
export XLPANEL_SMTP_PORT=587
export XLPANEL_SMTP_USERNAME=your-email@gmail.com
export XLPANEL_SMTP_PASSWORD=your-app-password

# 运行服务器 / Run server
go run ./cmd/server
```

### 访问系统 / Access System
- 管理端首页 / Admin Home: http://localhost:8080/
- 客户端首页 / Customer Home: http://localhost:8080/customer
- 产品目录 / Product Catalog: http://localhost:8080/customer/products
- 登录页面 / Login: http://localhost:8080/login
- 注册页面 / Register: http://localhost:8080/register

### 语言切换 / Language Switching
在任何页面添加 `?lang=zh` 或 `?lang=en` 参数即可切换语言。
Add `?lang=zh` or `?lang=en` parameter to any page to switch language.

例如 / Example:
- http://localhost:8080/customer?lang=zh
- http://localhost:8080/login?lang=en

## 后续开发 / Future Development

- [ ] 支付网关集成 / Payment gateway integration
- [ ] 账单PDF生成 / Invoice PDF generation
- [ ] 工单系统增强 / Enhanced ticket system
- [ ] 实时通知 / Real-time notifications
- [ ] 移动端响应式优化 / Mobile responsive optimization
- [ ] API 文档 / API documentation
- [ ] 单元测试 / Unit tests
- [ ] 集成测试 / Integration tests

## 许可证 / License

请参考主 README.md 文件。
Please refer to the main README.md file.
