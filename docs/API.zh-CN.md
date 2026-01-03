# API 文档

[English](API.md) | 简体中文

## 概述

OpenHost 提供 RESTful API 用于管理托管服务、计费和客户账户。所有 API 端点返回 JSON 格式，遵循标准 HTTP 约定。

## 基础 URL

```
生产环境: https://api.yourdomain.com/api/v1
开发环境: http://localhost:6421/api/v1
```

## 认证

### API 密钥认证

在 `Authorization` 头中包含您的 API 密钥：

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.yourdomain.com/api/v1/orders
```

### 会话认证

对于 Web 应用程序，使用基于会话的认证：

```bash
# 登录
curl -X POST https://api.yourdomain.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "secret"}'

# 响应包含会话 cookie
```

## 响应格式

### 成功响应

```json
{
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "VPS Basic",
    "price": "9.99"
  }
}
```

### 错误响应

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "产品未找到",
    "details": {
      "product_id": "invalid-id"
    }
  }
}
```

### 分页

列表端点支持分页：

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

## HTTP 状态码

- `200 OK` - 请求成功
- `201 Created` - 资源已创建
- `400 Bad Request` - 无效请求
- `401 Unauthorized` - 需要认证
- `403 Forbidden` - 权限不足
- `404 Not Found` - 资源未找到
- `422 Unprocessable Entity` - 验证错误
- `429 Too Many Requests` - 超过速率限制
- `500 Internal Server Error` - 服务器错误

## 速率限制

API 请求受速率限制：
- **已认证**: 每小时 1000 次请求
- **未认证**: 每小时 100 次请求

速率限制头：
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640000000
```

## 端点

### 健康检查

检查 API 可用性。

**端点:** `GET /health`

**响应:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

---

### 产品

#### 列出产品

列出所有可用产品。

**端点:** `GET /products`

**参数:**
- `page` (整数): 页码 (默认: 1)
- `per_page` (整数): 每页项数 (默认: 20, 最大: 100)
- `active` (布尔值): 按活动状态过滤

**示例:**
```bash
curl https://api.yourdomain.com/api/v1/products?active=true
```

**响应:**
```json
{
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "VPS Basic",
      "slug": "vps-basic",
      "description": "基础 VPS，1 CPU 和 1GB RAM",
      "price": "9.99",
      "billing_cycle": "monthly",
      "module_name": "provisioner-vps",
      "active": true,
      "config_options": {
        "cpu": "1",
        "ram": "1GB",
        "disk": "25GB"
      },
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 5,
    "total_pages": 1
  }
}
```

#### 获取产品

通过 ID 获取特定产品。

**端点:** `GET /products/:id`

**示例:**
```bash
curl https://api.yourdomain.com/api/v1/products/123e4567-e89b-12d3-a456-426614174000
```

#### 创建产品 (管理员)

创建新产品。

**端点:** `POST /products`

**请求体:**
```json
{
  "name": "VPS Premium",
  "slug": "vps-premium",
  "description": "高级 VPS，4 CPU 和 8GB RAM",
  "price": "49.99",
  "billing_cycle": "monthly",
  "module_name": "provisioner-vps",
  "active": true,
  "config_options": {
    "cpu": "4",
    "ram": "8GB",
    "disk": "100GB"
  }
}
```

---

### 订单

#### 列出订单

列出已认证客户的订单。

**端点:** `GET /orders`

**参数:**
- `status` (字符串): 按状态过滤 (`pending`, `active`, `suspended`, `cancelled`)
- `page` (整数): 页码
- `per_page` (整数): 每页项数

**示例:**
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.yourdomain.com/api/v1/orders?status=active
```

**响应:**
```json
{
  "data": [
    {
      "id": "order-uuid",
      "customer_id": "customer-uuid",
      "product_id": "product-uuid",
      "product_name": "VPS Basic",
      "status": "active",
      "total": "9.99",
      "currency": "USD",
      "billing_cycle": "monthly",
      "next_due_date": "2024-02-01",
      "created_at": "2024-01-01T12:00:00Z",
      "service_metadata": {
        "ip_address": "192.168.1.100",
        "hostname": "server1.example.com"
      }
    }
  ]
}
```

#### 创建订单

创建新订单。

**端点:** `POST /orders`

**请求体:**
```json
{
  "product_id": "product-uuid",
  "billing_cycle": "monthly",
  "parameters": {
    "hostname": "myserver",
    "region": "us-east-1"
  }
}
```

---

### 发票

#### 列出发票

列出已认证客户的发票。

**端点:** `GET /invoices`

**参数:**
- `status` (字符串): 按状态过滤 (`unpaid`, `paid`, `cancelled`)
- `page` (整数): 页码
- `per_page` (整数): 每页项数

**响应:**
```json
{
  "data": [
    {
      "id": "invoice-uuid",
      "customer_id": "customer-uuid",
      "invoice_number": "INV-2024-0001",
      "status": "unpaid",
      "subtotal": "9.99",
      "tax": "0.80",
      "total": "10.79",
      "currency": "USD",
      "due_date": "2024-01-15",
      "created_at": "2024-01-01T12:00:00Z",
      "items": [
        {
          "description": "VPS Basic - 月付",
          "quantity": 1,
          "unit_price": "9.99",
          "total": "9.99"
        }
      ]
    }
  ]
}
```

#### 支付发票

支付发票。

**端点:** `POST /invoices/:id/pay`

**请求体:**
```json
{
  "payment_method": "credit_card",
  "payment_gateway": "stripe",
  "payment_details": {
    "token": "tok_visa"
  }
}
```

---

### 客户

#### 获取客户资料

获取已认证客户的资料。

**端点:** `GET /customers/me`

**响应:**
```json
{
  "data": {
    "id": "customer-uuid",
    "email": "user@example.com",
    "first_name": "张",
    "last_name": "三",
    "company": "示例公司",
    "phone": "+86138000000000",
    "address": {
      "line1": "主街 123 号",
      "city": "北京",
      "state": "北京",
      "postal_code": "100000",
      "country": "CN"
    },
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

#### 更新客户资料

更新资料信息。

**端点:** `PATCH /customers/me`

**请求体:**
```json
{
  "first_name": "张",
  "last_name": "三",
  "phone": "+86138000000000"
}
```

---

### 支持工单

#### 列出工单

列出支持工单。

**端点:** `GET /tickets`

**参数:**
- `status` (字符串): 按状态过滤 (`open`, `in_progress`, `closed`)

**响应:**
```json
{
  "data": [
    {
      "id": "ticket-uuid",
      "subject": "VPS 无响应",
      "status": "open",
      "priority": "high",
      "department": "technical",
      "created_at": "2024-01-01T12:00:00Z",
      "last_reply_at": "2024-01-01T13:00:00Z"
    }
  ]
}
```

#### 创建工单

创建新支持工单。

**端点:** `POST /tickets`

**请求体:**
```json
{
  "subject": "VPS 无响应",
  "department": "technical",
  "priority": "high",
  "message": "我的 VPS (192.168.1.100) 无法通过 SSH 连接。"
}
```

---

## Webhooks

OpenHost 可以为各种事件发送 webhooks。

### Webhook 事件

- `order.created` - 新订单创建
- `order.activated` - 订单激活
- `order.suspended` - 订单暂停
- `order.terminated` - 订单终止
- `invoice.created` - 新发票创建
- `invoice.paid` - 发票已支付
- `invoice.overdue` - 发票逾期
- `ticket.created` - 新工单创建
- `ticket.replied` - 工单回复添加

### Webhook 载荷

```json
{
  "event": "order.activated",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "id": "order-uuid",
    "customer_id": "customer-uuid",
    "product_id": "product-uuid",
    "status": "active"
  }
}
```

### Webhook 安全

Webhooks 包含带有 HMAC-SHA256 签名的 `X-OpenHost-Signature` 头：

```python
import hmac
import hashlib

def verify_webhook(payload, signature, secret):
    expected = hmac.new(
        secret.encode(),
        payload.encode(),
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(expected, signature)
```

## SDK 和库

### 官方 SDK

- **Go**: `github.com/openhost/openhost-go`
- **Python**: `pip install openhost`
- **PHP**: `composer require openhost/openhost-php`
- **Node.js**: `npm install @openhost/openhost-js`

### 示例 (Node.js)

```javascript
const OpenHost = require('@openhost/openhost-js');

const client = new OpenHost({
  apiKey: 'your-api-key',
  baseURL: 'https://api.yourdomain.com'
});

// 列出产品
const products = await client.products.list();

// 创建订单
const order = await client.orders.create({
  product_id: 'product-uuid',
  billing_cycle: 'monthly'
});
```

## OpenAPI 规范

查看完整的 OpenAPI 规范：
- Swagger UI: `https://api.yourdomain.com/docs`
- OpenAPI JSON: `https://api.yourdomain.com/openapi.json`

## 支持

- API 文档问题: [GitHub Issues](https://github.com/lbyxiaolizi/xlpanel/issues)
- 通用 API 帮助: [讨论](https://github.com/lbyxiaolizi/xlpanel/discussions)
- 邮箱: api-support@yourdomain.com
