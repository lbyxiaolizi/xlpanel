# XLPanel

一个面向托管服务商的财务与自动化平台骨架，参考 WHMCS / Blesta 的功能域，使用 Go 构建，强调清晰的分层结构与可扩展性。

## 功能覆盖

- 多租户 / 多客户管理（Tenant / Customer）
- 产品目录与订阅
- 计费与账单（Invoice / InvoiceLine）
- 订单管理（Order）
- 服务器托管 / VPS 处理
- IP 分配
- 优惠券与循环优惠券
- 支付通道对接与收款记录
- 支付插件机制（支付宝当面付 / 银联 / Stripe / 虚拟币）
- 客户支持（工单）
- 自动化任务（可扩展）
- 可替换前端主题（配置驱动）

## 结构总览

```
xlpanel/
├─ cmd/server/          # 入口
├─ internal/
│  ├─ api/              # HTTP 路由与处理器
│  ├─ core/             # 全局配置与基础能力
│  ├─ domain/           # 领域模型
│  ├─ infra/            # 基础设施（仓储 / 观测）
│  ├─ plugins/          # 支付插件接口与实现
│  └─ service/          # 业务服务
└─ docs/                # 架构与使用文档
```

## 快速开始

```bash
go run ./cmd/server
```

服务默认监听 `http://127.0.0.1:8080`。

## 文档

- [架构设计](docs/architecture.md)
- [模块说明](docs/overview.md)
- [API 示例](docs/api.md)
- [支付插件](docs/payment_plugins.md)
- [主题与前端](docs/theming.md)
- [自动化与集成](docs/automation.md)
