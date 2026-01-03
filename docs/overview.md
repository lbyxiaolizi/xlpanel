# 模块说明

## api
- `internal/api.Server` 负责注册对外 HTTP API。
- 处理器聚合多个服务并提供 JSON 接口。

## domain
- `internal/domain` 包含 Tenant、Customer、Order、Invoice、Coupon、Ticket、VPS、IP 等核心实体。

## service
- `CatalogService` 管理产品目录。
- `SubscriptionService` 维护订阅与周期性出账。
- `BillingService` 处理账单、优惠券与收款。
- `OrderService` 处理订单生命周期。
- `HostingService` 管理 VPS 与 IP 分配。
- `SupportService` 管理工单。
- `PaymentsService` 管理支付通道与插件调用。
- `AutomationService` 维护自动化作业。

## plugins
- `internal/plugins` 提供支付插件接口与注册表。
- `internal/plugins/providers` 包含支付宝当面付、银联、Stripe、虚拟币的示例实现。

## infra
- `Repository` 展示可替换的数据访问模式。
- `MetricsRegistry` 用于统计领域事件。

## core
- `AppConfig` 统一配置入口，例如默认主题与主题覆盖策略。
- `NewID` 提供服务端 ID 生成策略。
