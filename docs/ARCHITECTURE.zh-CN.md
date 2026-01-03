# OpenHost 架构设计

[English](ARCHITECTURE.md) | 简体中文

## 概述

OpenHost 遵循**整洁架构（Clean Architecture）**原则和**领域驱动设计（DDD）**模式，组织为**模块化单体（Modular Monolith）**架构，并通过进程隔离实现插件系统。

## 设计原则

### 1. 整洁架构

系统采用同心圆层次结构，依赖关系由外向内：

```
┌─────────────────────────────────────────────┐
│         基础设施层                           │
│  (HTTP, DB, Cache, 外部服务)                │
│  ┌───────────────────────────────────────┐  │
│  │      应用层                           │  │
│  │    (用例, 编排)                       │  │
│  │  ┌─────────────────────────────────┐  │  │
│  │  │     领域层                      │  │  │
│  │  │  (实体, 值对象,                 │  │  │
│  │  │   领域服务, 接口)               │  │  │
│  │  └─────────────────────────────────┘  │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

**核心规则：**
- 领域层无外部依赖
- 应用层仅依赖领域层
- 基础设施层实现领域/应用层定义的接口
- 依赖关系由外向内

### 2. 领域驱动设计

**核心概念：**
- **实体（Entities）**：具有唯一标识的对象（产品、订单、客户）
- **值对象（Value Objects）**：由属性定义的不可变对象（金额、邮箱、地址）
- **聚合（Aggregates）**：具有聚合根的相关实体集群（订单及订单项）
- **领域服务（Domain Services）**：不属于实体的业务逻辑
- **仓储（Repositories）**：数据持久化抽象
- **领域事件（Domain Events）**：跨边界通信状态变化

### 3. 模块化单体

应用程序结构为单一部署中的独立模块：

```
internal/
├── core/
│   ├── domain/          # 领域模型和接口
│   │   ├── catalog.go   # 产品目录领域
│   │   ├── order.go     # 订单管理领域
│   │   ├── ticket.go    # 工单支持领域
│   │   └── ipam.go      # IP地址管理领域
│   └── service/         # 领域服务
│       ├── billing/     # 计费服务
│       └── ipam/        # IPAM服务
└── infrastructure/
    ├── http/            # HTTP处理器
    ├── web/             # 模板渲染
    ├── plugin/          # 插件系统
    ├── tasks/           # 后台任务
    └── tickets/         # 工单处理
```

## 层次详解

### 领域层（`internal/core/domain/`）

**目的**：包含核心业务逻辑和规则

**组件：**
- **实体**：具有标识的业务对象
- **值对象**：不可变数据结构
- **领域服务**：复杂业务逻辑
- **仓储接口**：数据访问契约
- **领域事件**：状态变更通知

**示例：**
```go
// 产品实体
type Product struct {
    ID          uuid.UUID
    Name        string
    Price       decimal.Decimal
    ModuleName  string  // 插件引用
    Active      bool
}

// 领域服务
type BillingService interface {
    CalculateInvoice(orderID uuid.UUID) (*Invoice, error)
    ProcessPayment(invoiceID uuid.UUID, payment Payment) error
}
```

**规则：**
- 无外部依赖（仅Go标准库）
- 使用 `shopspring/decimal` 处理金额
- 线程安全操作
- 尽可能不可变

### 服务层（`internal/core/service/`）

**目的**：编排领域逻辑并协调操作

**组件：**
- 业务工作流
- 事务管理
- 领域事件发布
- 跨聚合操作

**示例：**
```go
type OrderService struct {
    orderRepo   domain.OrderRepository
    productRepo domain.ProductRepository
    billing     domain.BillingService
}

func (s *OrderService) CreateOrder(ctx context.Context, 
    customerID uuid.UUID, items []OrderItem) (*Order, error) {
    // 验证产品
    // 创建订单聚合
    // 生成发票
    // 发布领域事件
}
```

### 基础设施层（`internal/infrastructure/`）

**目的**：实现技术能力和外部集成

#### HTTP层（`http/`）
- RESTful API处理器
- 请求验证
- 响应格式化
- 认证/授权

```go
// 处理器示例
func (h *OrderHandler) CreateOrder(c *gin.Context) {
    var req CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    order, err := h.service.CreateOrder(c.Request.Context(), req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(201, order)
}
```

#### Web层（`web/`）
- 模板渲染
- 主题管理
- 扩展钩子系统

#### 插件层（`plugin/`）
- gRPC插件管理
- 进程隔离
- 插件发现和加载
- 健康检查

#### 任务层（`tasks/`）
- 后台任务处理
- 计划任务
- 使用Asynq的异步操作

## 插件架构

### 设计

OpenHost 使用 **HashiCorp go-plugin** 实现进程隔离的插件：

```
┌──────────────┐         gRPC           ┌──────────────┐
│              │◄──────────────────────►│              │
│   OpenHost   │                        │    插件      │
│   (宿主)     │  Protocol Buffers      │  (独立进程)  │
│              │                        │              │
└──────────────┘                        └──────────────┘
```

**优势：**
- **进程隔离**：插件崩溃不影响主程序
- **语言无关**：插件可用任何语言编写（通过gRPC）
- **版本控制**：多个插件版本可共存
- **安全性**：通过SHA-256验证的沙箱执行

### 插件生命周期

1. **发现**：宿主扫描 `./plugins/` 目录
2. **验证**：检查SHA-256校验和
3. **启动**：将插件作为子进程启动
4. **握手**：建立gRPC连接
5. **通信**：通过Protocol Buffers进行RPC调用
6. **清理**：宿主终止时优雅关闭

### 插件接口

定义在 `pkg/proto/provisioner/`：

```protobuf
service Provisioner {
    rpc Provision(ProvisionRequest) returns (ProvisionResponse);
    rpc Suspend(SuspendRequest) returns (SuspendResponse);
    rpc Unsuspend(UnsuspendRequest) returns (UnsuspendResponse);
    rpc Terminate(TerminateRequest) returns (TerminateResponse);
}
```

## 数据流

### 请求处理流程

```
┌─────────┐    ┌──────────┐    ┌─────────┐    ┌────────┐    ┌────────┐
│  客户端 │───►│  处理器  │───►│  服务   │───►│  领域  │───►│   DB   │
└─────────┘    └──────────┘    └─────────┘    └────────┘    └────────┘
    HTTP          验证请求      编排工作流    业务规则      持久化数据
```

### 插件集成流程

```
┌─────────┐    ┌─────────┐    ┌────────────┐    ┌────────┐
│  服务   │───►│ 插件    │───►│   gRPC     │───►│  插件  │
│         │    │ 管理器  │    │            │    │  进程  │
└─────────┘    └─────────┘    └────────────┘    └────────┘
   请求         查找插件       远程调用         执行逻辑
                             (Protobuf)
```

### 异步任务流程

```
┌─────────┐    ┌────────┐    ┌───────┐    ┌────────┐
│  服务   │───►│ Asynq  │───►│ Redis │───►│  工作  │
└─────────┘    └────────┘    └───────┘    └────────┘
   入队任务     调度任务       队列任务      处理任务
```

## 并发模型

### 线程安全

**原则：**
- **不可变领域对象**：值对象是不可变的
- **互斥锁保护**：共享状态使用 sync.RWMutex
- **上下文传播**：使用 context.Context 进行取消
- **通道通信**：优先使用通道而非共享内存

**示例：**
```go
type SafeCache struct {
    mu    sync.RWMutex
    items map[string]interface{}
}

func (c *SafeCache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.items[key]
    return val, ok
}

func (c *SafeCache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = value
}
```

### 数据库事务

使用GORM的事务支持实现原子操作：

```go
func (s *OrderService) CreateOrder(ctx context.Context, 
    req CreateOrderRequest) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 创建订单
        if err := tx.Create(&order).Error; err != nil {
            return err
        }
        
        // 创建发票
        if err := tx.Create(&invoice).Error; err != nil {
            return err
        }
        
        // 全部成功或全部失败
        return nil
    })
}
```

## 金额处理

**关键规则**：始终使用 `shopspring/decimal` 进行金额计算。

```go
import "github.com/shopspring/decimal"

// ✅ 正确
price := decimal.NewFromFloat(99.99)
quantity := decimal.NewFromInt(3)
total := price.Mul(quantity)

// ❌ 错误
price := 99.99
quantity := 3
total := price * float64(quantity) // 浮点数误差！
```

## 错误处理

### 错误类型

```go
// 领域错误
var (
    ErrProductNotFound = errors.New("product not found")
    ErrInsufficientFunds = errors.New("insufficient funds")
    ErrInvalidState = errors.New("invalid state transition")
)

// 包装错误以提供上下文
if err != nil {
    return fmt.Errorf("failed to create order: %w", err)
}
```

### 错误响应

```go
// HTTP错误响应
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code"`
    Details any    `json:"details,omitempty"`
}
```

## 配置管理

### 基于环境的配置

```go
type Config struct {
    Database DatabaseConfig
    Redis    RedisConfig
    Server   ServerConfig
    Plugins  PluginConfig
}

func LoadConfig() (*Config, error) {
    // 从环境变量或配置文件加载
}
```

## 安全考虑

### 1. 插件安全
- SHA-256校验和验证
- 进程隔离
- 资源限制
- 超时控制

### 2. 数据安全
- SQL注入防护（GORM参数化）
- XSS防护（模板转义）
- 表单CSRF令牌
- 输入验证

### 3. API安全
- 认证中间件
- 速率限制
- 请求大小限制
- CORS配置

## 测试策略

### 单元测试
- 领域逻辑（纯函数）
- 服务层（模拟依赖）
- 工具和助手

### 集成测试
- 使用测试数据库的HTTP处理器
- 插件集成
- 数据库操作

### 端到端测试
- 完整工作流场景
- API端点测试
- UI测试（如适用）

## 性能考虑

### 数据库
- 连接池
- 查询优化
- 适当索引
- 防止N+1查询

### 缓存
- Redis用于会话数据
- 内存缓存用于参考数据
- 缓存失效策略

### 后台任务
- Asynq异步处理
- 队列优先级
- 重试机制
- 死信队列

## 可扩展性

### 横向扩展
- 无状态应用设计
- 外部会话存储（Redis）
- 数据库连接池
- 负载均衡就绪

### 纵向扩展
- Go的高效协程
- 最小内存占用
- CPU密集型任务优化

## 监控和可观测性

### 日志
- 结构化日志
- 日志级别（DEBUG, INFO, WARN, ERROR）
- 上下文传播

### 指标
- 请求延迟
- 错误率
- 插件健康状况
- 队列深度

### 追踪
- 请求ID跟踪
- 分布式追踪支持

## 未来增强

- [ ] 事件溯源用于审计跟踪
- [ ] CQRS读写分离
- [ ] GraphQL API选项
- [ ] 多租户支持
- [ ] 高级插件市场

## 参考资料

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design by Eric Evans](https://www.domainlanguage.com/ddd/)
- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin)
- [GORM 文档](https://gorm.io/zh_CN/docs/)
- [Asynq 文档](https://github.com/hibiken/asynq)
