# Theme Development Guide / 主题开发指南

[English](#english) | [简体中文](#chinese)

---

<a name="english"></a>
# English

## Overview

OpenHost uses a powerful, decoupled theme system that allows you to completely customize the frontend appearance while keeping the backend logic intact. Themes are built using Go HTML templates with full i18n (internationalization) support.

## Theme Structure

```
themes/
├── default/                    # Default theme
│   ├── theme.json             # Theme manifest (optional)
│   ├── README.md              # Theme documentation
│   ├── layouts/
│   │   ├── base.html          # Base layout (required)
│   │   ├── client.html        # Client area layout
│   │   └── admin.html         # Admin area layout
│   ├── pages/
│   │   ├── home.html          # Homepage
│   │   ├── login.html         # Login page
│   │   ├── register.html      # Registration page
│   │   ├── products.html      # Products listing
│   │   ├── pricing.html       # Pricing page
│   │   ├── cart.html          # Shopping cart
│   │   ├── client/            # Client panel pages
│   │   │   ├── dashboard.html
│   │   │   ├── services.html
│   │   │   ├── invoices.html
│   │   │   └── ...
│   │   └── admin/             # Admin panel pages
│   │       ├── dashboard.html
│   │       ├── customers.html
│   │       └── ...
│   ├── partials/              # Reusable components
│   │   ├── header.html
│   │   ├── footer.html
│   │   └── sidebar.html
│   └── assets/
│       ├── css/
│       │   └── main.css       # Main stylesheet
│       ├── js/
│       │   └── main.js        # Main JavaScript
│       └── images/
│           └── logo.svg
└── custom/                    # Custom theme example
    └── ...
```

## Creating a New Theme

### Step 1: Create Theme Directory

```bash
cp -r themes/default themes/mytheme
```

### Step 2: Create Theme Manifest (Optional)

Create `themes/mytheme/theme.json`:

```json
{
  "name": "My Custom Theme",
  "slug": "mytheme",
  "version": "1.0.0",
  "author": "Your Name",
  "author_url": "https://example.com",
  "description": "A custom theme for OpenHost",
  "screenshot": "assets/images/screenshot.png",
  "type": "both",
  "supports": {
    "dark_mode": true,
    "rtl": false,
    "custom_css": true,
    "custom_js": true
  }
}
```

### Step 3: Customize the Base Layout

Edit `themes/mytheme/layouts/base.html`:

```html
<!DOCTYPE html>
<html lang="{{ .Lang }}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ if .Title }}{{ .Title }} - {{ end }}{{ t "common.app_name" }}</title>
    <link rel="stylesheet" href="/static/css/main.css">
</head>
<body>
    <!-- Navigation -->
    <nav class="navbar">
        <!-- Your navigation here -->
    </nav>

    <!-- Main Content -->
    <main>
        {{ template "content" . }}
    </main>

    <!-- Footer -->
    <footer>
        <!-- Your footer here -->
    </footer>
    
    <script src="/static/js/main.js"></script>
</body>
</html>
```

### Step 4: Activate Your Theme

In your application configuration or code:

```go
import "github.com/openhost/openhost/internal/infrastructure/web"

// Set the active theme
web.SetRenderer(web.NewRenderer("mytheme"))
```

## Template Functions

### Translation Function (`t`)

The `t` function translates keys to the current language:

```html
<!-- Simple translation -->
<h1>{{ t "nav.home" }}</h1>

<!-- Translation with arguments -->
<p>{{ t "common.copyright" .Year }}</p>

<!-- Nested translation keys -->
<span>{{ t "auth.login.title" }}</span>
```

### Available Template Functions

| Function | Description | Example |
|----------|-------------|---------|
| `t` | Translate a key | `{{ t "nav.home" }}` |
| `T` | Alias for `t` | `{{ T "nav.home" }}` |
| `hook` | Render plugin hooks | `{{ hook "header_scripts" . }}` |
| `dict` | Create a dictionary | `{{ dict "key" "value" }}` |
| `safe` | Mark HTML as safe | `{{ safe .HTML }}` |
| `eq` | Equality check | `{{ if eq .Status "active" }}` |
| `ne` | Not equal check | `{{ if ne .Status "inactive" }}` |
| `lt` | Less than | `{{ if lt .Count 10 }}` |
| `gt` | Greater than | `{{ if gt .Count 0 }}` |
| `add` | Add integers | `{{ add .Page 1 }}` |
| `sub` | Subtract integers | `{{ sub .Page 1 }}` |

## Available Template Variables

Every template has access to these variables:

| Variable | Type | Description |
|----------|------|-------------|
| `.User` | *User | Current authenticated user (nil if not logged in) |
| `.CSRFToken` | string | CSRF protection token |
| `.Currency` | string | Current currency code |
| `.Year` | int | Current year |
| `.Lang` | string | Current language code (e.g., "en", "zh") |
| `.Languages` | []*Language | List of available languages |
| `.Theme` | string | Current theme name |
| `.T` | func | Translation function |
| `.Title` | string | Page title |
| `.Description` | string | Page description |
| `.Flash` | *Flash | Flash message (if any) |

## Internationalization (i18n)

### Using Translations in Templates

```html
<!-- Navigation -->
<nav>
    <a href="/">{{ t "nav.home" }}</a>
    <a href="/products">{{ t "nav.products" }}</a>
    <a href="/pricing">{{ t "nav.pricing" }}</a>
</nav>

<!-- Conditional by language -->
{{ if eq .Lang "zh" }}
    <p>中文内容</p>
{{ else }}
    <p>English content</p>
{{ end }}

<!-- Language switcher -->
<div class="lang-switcher">
    {{ if eq .Lang "zh" }}
    <a href="?lang=en">English</a>
    {{ else }}
    <a href="?lang=zh">中文</a>
    {{ end }}
</div>
```

### Translation Key Structure

Translations are organized hierarchically:

```
common.app_name           -> "OpenHost"
common.copyright          -> "© %d OpenHost. All rights reserved."
nav.home                  -> "Home"
nav.products              -> "Products"
auth.login.title          -> "Welcome Back"
auth.login.email          -> "Email Address"
client.dashboard.title    -> "Dashboard"
admin.dashboard.title     -> "Admin Dashboard"
```

## Styling Guidelines

### CSS Variables

Use CSS variables for consistent styling:

```css
:root {
    /* Colors */
    --primary: #667eea;
    --primary-dark: #5568d3;
    --secondary: #764ba2;
    
    /* Text */
    --text-primary: #2d3748;
    --text-secondary: #4a5568;
    
    /* Backgrounds */
    --bg-primary: #ffffff;
    --bg-secondary: #f7fafc;
    
    /* Borders */
    --border: #e2e8f0;
    
    /* Spacing */
    --spacing-sm: 0.5rem;
    --spacing-md: 1rem;
    --spacing-lg: 1.5rem;
    
    /* Border Radius */
    --radius-sm: 0.25rem;
    --radius-md: 0.5rem;
    --radius-lg: 0.75rem;
}
```

### Dark Mode Support

```css
@media (prefers-color-scheme: dark) {
    :root {
        --bg-primary: #1a202c;
        --bg-secondary: #2d3748;
        --text-primary: #f7fafc;
        --text-secondary: #e2e8f0;
        --border: #4a5568;
    }
}
```

### Responsive Design

```css
/* Mobile */
@media (max-width: 768px) {
    .navbar-menu {
        display: none;
    }
    
    .navbar-menu.active {
        display: flex;
    }
}

/* Tablet */
@media (max-width: 1024px) {
    .grid-4 {
        grid-template-columns: repeat(2, 1fr);
    }
}
```

## Component Examples

### Card Component

```html
<div class="card">
    <div class="card-header">
        <h3 class="card-title">{{ t "client.services.title" }}</h3>
    </div>
    <div class="card-body">
        <!-- Content -->
    </div>
    <div class="card-footer">
        <a href="#" class="btn btn-primary">{{ t "common.view" }}</a>
    </div>
</div>
```

### Alert Component

```html
{{ if .Flash }}
<div class="alert alert-{{ .Flash.Type }}">
    {{ .Flash.Message }}
</div>
{{ end }}
```

### Form Components

```html
<form method="post" action="/login">
    {{ if .CSRFToken }}
    <input type="hidden" name="csrf_token" value="{{ .CSRFToken }}">
    {{ end }}
    
    <div class="form-group">
        <label class="form-label" for="email">{{ t "auth.login.email" }}</label>
        <input type="email" 
               id="email" 
               name="email" 
               class="form-control" 
               placeholder="{{ t "auth.login.email_placeholder" }}"
               required>
    </div>
    
    <button type="submit" class="btn btn-primary">{{ t "auth.login.submit" }}</button>
</form>
```

## Plugin Hooks

Themes can include plugin hook points:

```html
<!-- In base.html -->
<head>
    {{ hook "head_start" . }}
    <!-- ... -->
    {{ hook "head_end" . }}
</head>
<body>
    {{ hook "body_start" . }}
    <!-- ... -->
    {{ hook "body_end" . }}
</body>
```

Common hook points:
- `head_start` / `head_end` - Inside `<head>` tag
- `body_start` / `body_end` - Inside `<body>` tag
- `navbar_start` / `navbar_end` - Navigation area
- `footer_start` / `footer_end` - Footer area
- `sidebar_start` / `sidebar_end` - Sidebar area
- `dashboard_widgets` - Dashboard widget area

## Best Practices

1. **Always use i18n** - Never hardcode text, always use translation keys
2. **Use semantic HTML** - Use proper HTML5 elements
3. **Accessibility** - Include ARIA labels and keyboard navigation
4. **Mobile First** - Design for mobile, then scale up
5. **CSS Variables** - Use variables for theming flexibility
6. **Performance** - Minimize CSS/JS, optimize images
7. **Hook Points** - Include plugin hooks for extensibility

---

<a name="chinese"></a>
# 简体中文

## 概述

OpenHost 使用强大的解耦主题系统，允许您完全自定义前端外观，同时保持后端逻辑不变。主题使用 Go HTML 模板构建，完全支持 i18n（国际化）。

## 主题结构

```
themes/
├── default/                    # 默认主题
│   ├── theme.json             # 主题清单（可选）
│   ├── README.md              # 主题文档
│   ├── layouts/
│   │   ├── base.html          # 基础布局（必需）
│   │   ├── client.html        # 客户区域布局
│   │   └── admin.html         # 管理区域布局
│   ├── pages/
│   │   ├── home.html          # 首页
│   │   ├── login.html         # 登录页面
│   │   ├── register.html      # 注册页面
│   │   ├── products.html      # 产品列表
│   │   ├── pricing.html       # 价格页面
│   │   ├── cart.html          # 购物车
│   │   ├── client/            # 客户面板页面
│   │   │   ├── dashboard.html
│   │   │   ├── services.html
│   │   │   ├── invoices.html
│   │   │   └── ...
│   │   └── admin/             # 管理面板页面
│   │       ├── dashboard.html
│   │       ├── customers.html
│   │       └── ...
│   ├── partials/              # 可重用组件
│   │   ├── header.html
│   │   ├── footer.html
│   │   └── sidebar.html
│   └── assets/
│       ├── css/
│       │   └── main.css       # 主样式表
│       ├── js/
│       │   └── main.js        # 主 JavaScript
│       └── images/
│           └── logo.svg
└── custom/                    # 自定义主题示例
    └── ...
```

## 创建新主题

### 第一步：创建主题目录

```bash
cp -r themes/default themes/mytheme
```

### 第二步：创建主题清单（可选）

创建 `themes/mytheme/theme.json`：

```json
{
  "name": "我的自定义主题",
  "slug": "mytheme",
  "version": "1.0.0",
  "author": "您的名字",
  "author_url": "https://example.com",
  "description": "OpenHost 的自定义主题",
  "screenshot": "assets/images/screenshot.png",
  "type": "both",
  "supports": {
    "dark_mode": true,
    "rtl": false,
    "custom_css": true,
    "custom_js": true
  }
}
```

### 第三步：自定义基础布局

编辑 `themes/mytheme/layouts/base.html`：

```html
<!DOCTYPE html>
<html lang="{{ .Lang }}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ if .Title }}{{ .Title }} - {{ end }}{{ t "common.app_name" }}</title>
    <link rel="stylesheet" href="/static/css/main.css">
</head>
<body>
    <!-- 导航 -->
    <nav class="navbar">
        <!-- 您的导航内容 -->
    </nav>

    <!-- 主要内容 -->
    <main>
        {{ template "content" . }}
    </main>

    <!-- 页脚 -->
    <footer>
        <!-- 您的页脚内容 -->
    </footer>
    
    <script src="/static/js/main.js"></script>
</body>
</html>
```

### 第四步：激活您的主题

在应用程序配置或代码中：

```go
import "github.com/openhost/openhost/internal/infrastructure/web"

// 设置活动主题
web.SetRenderer(web.NewRenderer("mytheme"))
```

## 模板函数

### 翻译函数 (`t`)

`t` 函数将键翻译为当前语言：

```html
<!-- 简单翻译 -->
<h1>{{ t "nav.home" }}</h1>

<!-- 带参数的翻译 -->
<p>{{ t "common.copyright" .Year }}</p>

<!-- 嵌套翻译键 -->
<span>{{ t "auth.login.title" }}</span>
```

### 可用的模板函数

| 函数 | 描述 | 示例 |
|------|------|------|
| `t` | 翻译键 | `{{ t "nav.home" }}` |
| `T` | `t` 的别名 | `{{ T "nav.home" }}` |
| `hook` | 渲染插件钩子 | `{{ hook "header_scripts" . }}` |
| `dict` | 创建字典 | `{{ dict "key" "value" }}` |
| `safe` | 标记 HTML 为安全 | `{{ safe .HTML }}` |
| `eq` | 相等检查 | `{{ if eq .Status "active" }}` |
| `ne` | 不相等检查 | `{{ if ne .Status "inactive" }}` |
| `lt` | 小于 | `{{ if lt .Count 10 }}` |
| `gt` | 大于 | `{{ if gt .Count 0 }}` |
| `add` | 整数加法 | `{{ add .Page 1 }}` |
| `sub` | 整数减法 | `{{ sub .Page 1 }}` |

## 可用的模板变量

每个模板都可以访问这些变量：

| 变量 | 类型 | 描述 |
|------|------|------|
| `.User` | *User | 当前已认证用户（未登录则为 nil） |
| `.CSRFToken` | string | CSRF 保护令牌 |
| `.Currency` | string | 当前货币代码 |
| `.Year` | int | 当前年份 |
| `.Lang` | string | 当前语言代码（例如 "en"、"zh"） |
| `.Languages` | []*Language | 可用语言列表 |
| `.Theme` | string | 当前主题名称 |
| `.T` | func | 翻译函数 |
| `.Title` | string | 页面标题 |
| `.Description` | string | 页面描述 |
| `.Flash` | *Flash | Flash 消息（如果有） |

## 国际化 (i18n)

### 在模板中使用翻译

```html
<!-- 导航 -->
<nav>
    <a href="/">{{ t "nav.home" }}</a>
    <a href="/products">{{ t "nav.products" }}</a>
    <a href="/pricing">{{ t "nav.pricing" }}</a>
</nav>

<!-- 按语言条件渲染 -->
{{ if eq .Lang "zh" }}
    <p>中文内容</p>
{{ else }}
    <p>English content</p>
{{ end }}

<!-- 语言切换器 -->
<div class="lang-switcher">
    {{ if eq .Lang "zh" }}
    <a href="?lang=en">English</a>
    {{ else }}
    <a href="?lang=zh">中文</a>
    {{ end }}
</div>
```

### 翻译键结构

翻译按层次结构组织：

```
common.app_name           -> "OpenHost"
common.copyright          -> "© %d OpenHost. 保留所有权利。"
nav.home                  -> "首页"
nav.products              -> "产品"
auth.login.title          -> "欢迎回来"
auth.login.email          -> "邮箱地址"
client.dashboard.title    -> "控制面板"
admin.dashboard.title     -> "管理面板"
```

## 样式指南

### CSS 变量

使用 CSS 变量保持样式一致性：

```css
:root {
    /* 颜色 */
    --primary: #667eea;
    --primary-dark: #5568d3;
    --secondary: #764ba2;
    
    /* 文字 */
    --text-primary: #2d3748;
    --text-secondary: #4a5568;
    
    /* 背景 */
    --bg-primary: #ffffff;
    --bg-secondary: #f7fafc;
    
    /* 边框 */
    --border: #e2e8f0;
    
    /* 间距 */
    --spacing-sm: 0.5rem;
    --spacing-md: 1rem;
    --spacing-lg: 1.5rem;
    
    /* 边框圆角 */
    --radius-sm: 0.25rem;
    --radius-md: 0.5rem;
    --radius-lg: 0.75rem;
}
```

### 深色模式支持

```css
@media (prefers-color-scheme: dark) {
    :root {
        --bg-primary: #1a202c;
        --bg-secondary: #2d3748;
        --text-primary: #f7fafc;
        --text-secondary: #e2e8f0;
        --border: #4a5568;
    }
}
```

## 最佳实践

1. **始终使用 i18n** - 永远不要硬编码文本，始终使用翻译键
2. **使用语义化 HTML** - 使用正确的 HTML5 元素
3. **无障碍性** - 包含 ARIA 标签和键盘导航
4. **移动优先** - 先设计移动端，然后扩展
5. **CSS 变量** - 使用变量提高主题灵活性
6. **性能** - 最小化 CSS/JS，优化图片
7. **钩子点** - 包含插件钩子以实现可扩展性

## 许可证

主题遵循 OpenHost 的 MIT 许可证。
