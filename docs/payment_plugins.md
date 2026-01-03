# 支付插件开发

XLPanel 支持通过插件扩展支付通道，插件实现统一接口后即可注册到系统。

## 插件接口

接口位于 `internal/plugins/payments.go`：

```go
type PaymentPlugin interface {
    Name() string
    Provider() string
    Initialize(config map[string]string) error
    Charge(ctx context.Context, request PaymentRequest) (PaymentResponse, error)
}
```

## 注册插件

在启动时注册插件：

```go
registry := plugins.NewRegistry()
registry.Register(providers.NewAlipayFaceToFace())
registry.Register(providers.NewUnionPay())
registry.Register(providers.NewStripe())
registry.Register(providers.NewCrypto())
```

## 触发支付

HTTP API 示例：

```bash
curl -X POST http://127.0.0.1:8080/payments/plugins/charge \
  -H 'Content-Type: application/json' \
  -d '{"provider": "stripe", "tenant_id": "<tenant>", "customer_id": "<customer>", "invoice_id": "<invoice>", "amount": 99, "currency": "USD"}'
```

## 内置示例插件

- `providers.AlipayFaceToFace`（支付宝当面付）
- `providers.UnionPay`（银联）
- `providers.Stripe`
- `providers.Crypto`（虚拟币）

这些插件是演示实现，实际接入时请在 `Charge` 中调用第三方 SDK。
