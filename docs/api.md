# API 示例

以下示例展示核心功能的基础接口，服务默认监听 `http://127.0.0.1:8080`。

## 客户

```bash
curl -X POST http://127.0.0.1:8080/customers \
  -H 'Content-Type: application/json' \
  -d '{"name": "ACME", "email": "team@acme.test"}'
```

## 订单

```bash
curl -X POST http://127.0.0.1:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"customer_id": "<id>", "product_code": "vps-basic", "quantity": 1}'
```

## 账单

```bash
curl -X POST http://127.0.0.1:8080/billing/invoices \
  -H 'Content-Type: application/json' \
  -d '{"customer_id": "<id>", "amount": 99, "currency": "USD"}'
```

## 优惠券

```bash
curl -X POST http://127.0.0.1:8080/billing/coupons \
  -H 'Content-Type: application/json' \
  -d '{"code": "WELCOME10", "percent_off": 10, "recurring": true}'
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
