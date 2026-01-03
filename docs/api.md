# API 示例

以下示例展示核心功能的基础接口，服务默认监听 `http://127.0.0.1:8080`。

## 租户

```bash
curl -X POST http://127.0.0.1:8080/tenants \
  -H 'Content-Type: application/json' \
  -d '{"name": "ACME Hosting", "contact": "ops@acme.test"}'
```

## 客户

```bash
curl -X POST http://127.0.0.1:8080/customers \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id": "<tenant>", "name": "ACME", "email": "team@acme.test"}'
```

## 产品目录

```bash
curl -X POST http://127.0.0.1:8080/catalog/products \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id": "<tenant>", "code": "vps-basic", "name": "VPS Basic", "unit_price": 9.9, "currency": "USD", "recurring": true, "billing_days": 30}'
```

## 订单

```bash
curl -X POST http://127.0.0.1:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id": "<tenant>", "customer_id": "<id>", "product_code": "vps-basic", "quantity": 1}'
```

## 订阅

```bash
curl -X POST http://127.0.0.1:8080/subscriptions \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id": "<tenant>", "customer_id": "<id>", "product_code": "vps-basic"}'
```

## 订阅开票

```bash
curl -X POST http://127.0.0.1:8080/subscriptions/invoices \
  -H 'Content-Type: application/json' \
  -d '{"subscription_id": "<subscription>"}'
```

## 账单

```bash
curl -X POST http://127.0.0.1:8080/billing/invoices \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id": "<tenant>", "customer_id": "<id>", "currency": "USD", "lines": [{"description": "Setup", "quantity": 1, "unit_price": 99}]}'
```

## 优惠券

```bash
curl -X POST http://127.0.0.1:8080/billing/coupons \
  -H 'Content-Type: application/json' \
  -d '{"code": "WELCOME10", "percent_off": 10, "recurring": true}'
```

## 收款

```bash
curl -X POST http://127.0.0.1:8080/billing/payments \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id": "<tenant>", "customer_id": "<id>", "invoice_id": "<invoice>", "amount": 99, "currency": "USD", "method": "card"}'
```

## 支付插件

```bash
curl -X POST http://127.0.0.1:8080/payments/plugins/charge \
  -H 'Content-Type: application/json' \
  -d '{"provider": "stripe", "tenant_id": "<tenant>", "customer_id": "<customer>", "invoice_id": "<invoice>", "amount": 99, "currency": "USD"}'
```

## 支付通道

```bash
curl -X POST http://127.0.0.1:8080/payments/gateways \
  -H 'Content-Type: application/json' \
  -d '{"name": "Stripe", "provider": "stripe", "enabled": true}'
```

## 自动化任务

```bash
curl -X POST http://127.0.0.1:8080/automation/jobs \
  -H 'Content-Type: application/json' \
  -d '{"name": "invoice-dunning", "schedule": "0 */6 * * *"}'
```
