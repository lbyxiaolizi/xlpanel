# API Documentation

[English](API.md) | [简体中文](API.zh-CN.md)

## Overview

OpenHost provides a RESTful API for managing hosting services, billing, and customer accounts. All API endpoints return JSON and follow standard HTTP conventions.

## Base URL

```
Production: https://api.yourdomain.com/api/v1
Development: http://localhost:6421/api/v1
```

## Authentication

### API Key Authentication

Include your API key in the `Authorization` header:

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.yourdomain.com/api/v1/orders
```

### Session Authentication

For web applications, use session-based authentication:

```bash
# Login
curl -X POST https://api.yourdomain.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "secret"}'

# Response includes session cookie
```

## Response Format

### Success Response

```json
{
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "VPS Basic",
    "price": "9.99"
  }
}
```

### Error Response

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Product not found",
    "details": {
      "product_id": "invalid-id"
    }
  }
}
```

### Pagination

List endpoints support pagination:

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

## HTTP Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created
- `400 Bad Request` - Invalid request
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `422 Unprocessable Entity` - Validation error
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

## Rate Limiting

API requests are rate limited:
- **Authenticated**: 1000 requests per hour
- **Unauthenticated**: 100 requests per hour

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640000000
```

## Endpoints

### Health Check

Check API availability.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

---

### Products

#### List Products

List all available products.

**Endpoint:** `GET /products`

**Parameters:**
- `page` (integer): Page number (default: 1)
- `per_page` (integer): Items per page (default: 20, max: 100)
- `active` (boolean): Filter by active status

**Example:**
```bash
curl https://api.yourdomain.com/api/v1/products?active=true
```

**Response:**
```json
{
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "VPS Basic",
      "slug": "vps-basic",
      "description": "Basic VPS with 1 CPU and 1GB RAM",
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

#### Get Product

Get a specific product by ID.

**Endpoint:** `GET /products/:id`

**Example:**
```bash
curl https://api.yourdomain.com/api/v1/products/123e4567-e89b-12d3-a456-426614174000
```

**Response:**
```json
{
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "VPS Basic",
    "slug": "vps-basic",
    "description": "Basic VPS with 1 CPU and 1GB RAM",
    "price": "9.99",
    "billing_cycle": "monthly",
    "module_name": "provisioner-vps",
    "active": true,
    "config_options": {
      "cpu": "1",
      "ram": "1GB",
      "disk": "25GB"
    }
  }
}
```

#### Create Product (Admin)

Create a new product.

**Endpoint:** `POST /products`

**Request Body:**
```json
{
  "name": "VPS Premium",
  "slug": "vps-premium",
  "description": "Premium VPS with 4 CPU and 8GB RAM",
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

**Response:**
```json
{
  "data": {
    "id": "new-uuid-here",
    "name": "VPS Premium",
    // ... other fields
  }
}
```

---

### Orders

#### List Orders

List orders for the authenticated customer.

**Endpoint:** `GET /orders`

**Parameters:**
- `status` (string): Filter by status (`pending`, `active`, `suspended`, `cancelled`)
- `page` (integer): Page number
- `per_page` (integer): Items per page

**Example:**
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.yourdomain.com/api/v1/orders?status=active
```

**Response:**
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

#### Create Order

Create a new order.

**Endpoint:** `POST /orders`

**Request Body:**
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

**Response:**
```json
{
  "data": {
    "id": "new-order-uuid",
    "status": "pending",
    "total": "9.99",
    "invoice_id": "invoice-uuid"
  }
}
```

#### Get Order

Get order details.

**Endpoint:** `GET /orders/:id`

**Response:**
```json
{
  "data": {
    "id": "order-uuid",
    "customer_id": "customer-uuid",
    "product_id": "product-uuid",
    "product_name": "VPS Basic",
    "status": "active",
    "total": "9.99",
    "service_metadata": {
      "ip_address": "192.168.1.100",
      "hostname": "server1.example.com"
    }
  }
}
```

---

### Invoices

#### List Invoices

List invoices for the authenticated customer.

**Endpoint:** `GET /invoices`

**Parameters:**
- `status` (string): Filter by status (`unpaid`, `paid`, `cancelled`)
- `page` (integer): Page number
- `per_page` (integer): Items per page

**Response:**
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
          "description": "VPS Basic - Monthly",
          "quantity": 1,
          "unit_price": "9.99",
          "total": "9.99"
        }
      ]
    }
  ]
}
```

#### Get Invoice

Get invoice details.

**Endpoint:** `GET /invoices/:id`

#### Pay Invoice

Pay an invoice.

**Endpoint:** `POST /invoices/:id/pay`

**Request Body:**
```json
{
  "payment_method": "credit_card",
  "payment_gateway": "stripe",
  "payment_details": {
    "token": "tok_visa"
  }
}
```

**Response:**
```json
{
  "data": {
    "invoice_id": "invoice-uuid",
    "status": "paid",
    "payment_id": "payment-uuid",
    "paid_at": "2024-01-01T12:00:00Z"
  }
}
```

---

### Customers

#### Get Customer Profile

Get the authenticated customer's profile.

**Endpoint:** `GET /customers/me`

**Response:**
```json
{
  "data": {
    "id": "customer-uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "company": "Acme Inc",
    "phone": "+1234567890",
    "address": {
      "line1": "123 Main St",
      "city": "New York",
      "state": "NY",
      "postal_code": "10001",
      "country": "US"
    },
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

#### Update Customer Profile

Update profile information.

**Endpoint:** `PATCH /customers/me`

**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890"
}
```

---

### Support Tickets

#### List Tickets

List support tickets.

**Endpoint:** `GET /tickets`

**Parameters:**
- `status` (string): Filter by status (`open`, `in_progress`, `closed`)

**Response:**
```json
{
  "data": [
    {
      "id": "ticket-uuid",
      "subject": "VPS not responding",
      "status": "open",
      "priority": "high",
      "department": "technical",
      "created_at": "2024-01-01T12:00:00Z",
      "last_reply_at": "2024-01-01T13:00:00Z"
    }
  ]
}
```

#### Create Ticket

Create a new support ticket.

**Endpoint:** `POST /tickets`

**Request Body:**
```json
{
  "subject": "VPS not responding",
  "department": "technical",
  "priority": "high",
  "message": "My VPS at 192.168.1.100 is not responding to SSH connections."
}
```

#### Reply to Ticket

Add a reply to a ticket.

**Endpoint:** `POST /tickets/:id/replies`

**Request Body:**
```json
{
  "message": "I've checked and the issue persists."
}
```

---

## Webhooks

OpenHost can send webhooks for various events.

### Webhook Events

- `order.created` - New order created
- `order.activated` - Order activated
- `order.suspended` - Order suspended
- `order.terminated` - Order terminated
- `invoice.created` - New invoice created
- `invoice.paid` - Invoice paid
- `invoice.overdue` - Invoice overdue
- `ticket.created` - New ticket created
- `ticket.replied` - Ticket reply added

### Webhook Payload

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

### Webhook Security

Webhooks include an `X-OpenHost-Signature` header with HMAC-SHA256 signature:

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

## SDKs and Libraries

### Official SDKs

- **Go**: `github.com/openhost/openhost-go`
- **Python**: `pip install openhost`
- **PHP**: `composer require openhost/openhost-php`
- **Node.js**: `npm install @openhost/openhost-js`

### Example (Node.js)

```javascript
const OpenHost = require('@openhost/openhost-js');

const client = new OpenHost({
  apiKey: 'your-api-key',
  baseURL: 'https://api.yourdomain.com'
});

// List products
const products = await client.products.list();

// Create order
const order = await client.orders.create({
  product_id: 'product-uuid',
  billing_cycle: 'monthly'
});
```

## Postman Collection

Download our Postman collection: [OpenHost API Collection](https://api.yourdomain.com/postman.json)

## GraphQL API (Coming Soon)

GraphQL endpoint will be available at `/graphql` in a future release.

## OpenAPI Specification

View the full OpenAPI specification at:
- Swagger UI: `https://api.yourdomain.com/docs`
- OpenAPI JSON: `https://api.yourdomain.com/openapi.json`

## Support

- API Documentation Issues: [GitHub Issues](https://github.com/lbyxiaolizi/xlpanel/issues)
- General API Help: [Discussions](https://github.com/lbyxiaolizi/xlpanel/discussions)
- Email: api-support@yourdomain.com
