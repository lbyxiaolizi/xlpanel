# 贡献指南

[English](CONTRIBUTING.md) | 简体中文

感谢您对 OpenHost 项目的关注！本文档提供贡献指南和说明。

## 行为准则

### 我们的承诺

我们承诺为所有贡献者提供友好和包容的环境。

### 我们的标准

- 使用友好和包容的语言
- 尊重不同的观点和经验
- 优雅地接受建设性批评
- 专注于对社区最有利的事情
- 对其他社区成员表现出同理心

## 如何贡献

### 报告错误

在创建错误报告之前：
1. 检查[现有问题](https://github.com/lbyxiaolizi/xlpanel/issues)
2. 尝试最新版本看错误是否仍然存在

创建错误报告时，请包括：
- 清晰描述性的标题
- 重现行为的步骤
- 预期行为
- 实际行为
- 截图（如适用）
- 环境详细信息（操作系统、Go 版本等）

**示例：**
```markdown
### 错误：无效插件校验和时服务器崩溃

**重现步骤：**
1. 创建没有 SHA256 文件的插件
2. 启动服务器
3. 触发插件加载

**预期：** 记录错误消息，服务器继续运行
**实际：** 服务器崩溃并出现 panic

**环境：**
- 操作系统: Ubuntu 22.04
- Go: 1.23.0
- OpenHost: v0.1.0
```

### 建议增强功能

增强建议作为 GitHub issues 跟踪。包括：
- 清晰描述性的标题
- 建议功能的详细描述
- 用例和好处
- 可能的实现方法
- 类似项目的示例（如适用）

### 拉取请求

1. **Fork 仓库**
```bash
git clone https://github.com/yourusername/xlpanel.git
cd xlpanel
```

2. **创建功能分支**
```bash
git checkout -b feature/your-feature-name
# 或
git checkout -b fix/your-bug-fix
```

3. **进行更改**
- 遵循编码标准（见下文）
- 编写或更新测试
- 更新文档

4. **提交更改**
```bash
git add .
git commit -m "feat: 添加新功能"
# 或
git commit -m "fix: 解决配置中的错误"
```

遵循[约定式提交](https://www.conventionalcommits.org/zh-hans/)：
- `feat:` 新功能
- `fix:` 错误修复
- `docs:` 文档更改
- `style:` 代码样式更改（格式化等）
- `refactor:` 代码重构
- `test:` 添加或更新测试
- `chore:` 维护任务

5. **推送到您的 fork**
```bash
git push origin feature/your-feature-name
```

6. **创建拉取请求**
- 使用清晰的标题和描述
- 引用相关问题
- 为 UI 更改包含截图
- 确保所有测试通过
- 请求维护者审查

## 开发设置

### 前置要求

- Go 1.23 或更高版本
- PostgreSQL 12+
- Redis 6+
- Make
- Git

### 本地开发

1. **克隆和设置**
```bash
git clone https://github.com/lbyxiaolizi/xlpanel.git
cd xlpanel
go mod download
```

2. **设置数据库**
```bash
createdb openhost
psql -d openhost -f schema.sql
```

3. **配置环境**
```bash
cp .env.example .env
# 编辑 .env 配置您的设置
```

4. **构建**
```bash
make all
```

5. **运行测试**
```bash
go test ./...
```

6. **启动服务器**
```bash
./bin/server
```

## 编码标准

### Go 风格指南

遵循 [Uber Go 风格指南](https://github.com/uber-go/guide/blob/master/style.md) 和 [Effective Go](https://golang.org/doc/effective_go.html)。

### 关键原则

1. **整洁架构**
   - 保持领域层纯净（无依赖）
   - 使用接口进行依赖
   - 依赖注入

2. **类型安全**
   - 使用 `shopspring/decimal` 处理金额
   - 尽可能避免 `interface{}`
   - 显式错误处理

3. **线程安全**
   - 使用互斥锁保护共享状态
   - 使用通道进行通信
   - 使用上下文进行取消

### 代码示例

**良好：**
```go
// 使用 decimal 处理金额
import "github.com/shopspring/decimal"

type Product struct {
    ID    uuid.UUID
    Name  string
    Price decimal.Decimal  // ✅
}

func (p *Product) CalculateTotal(quantity int) decimal.Decimal {
    return p.Price.Mul(decimal.NewFromInt(int64(quantity)))
}
```

**不良：**
```go
type Product struct {
    ID    string
    Name  string
    Price float64  // ❌ 不要使用 float 处理金额
}

func (p *Product) CalculateTotal(quantity int) float64 {
    return p.Price * float64(quantity)  // ❌ 浮点数误差
}
```

**良好：**
```go
// 适当的错误处理
func (s *OrderService) CreateOrder(ctx context.Context, 
    req CreateOrderRequest) (*Order, error) {
    
    product, err := s.productRepo.FindByID(ctx, req.ProductID)
    if err != nil {
        return nil, fmt.Errorf("failed to find product: %w", err)
    }
    
    if !product.Active {
        return nil, ErrProductInactive
    }
    
    // ... 其余逻辑
}
```

### 命名约定

```go
// 接口：名词或形容词
type Repository interface {}
type Provisioner interface {}

// 结构体：清晰的名词
type Order struct {}
type Product struct {}

// 函数：动词-名词或清晰目的
func CreateOrder() {}
func ValidateInput() {}
func (s *Service) ProcessPayment() {}

// 变量：描述性
orderID := uuid.New()
totalPrice := calculateTotal(items)

// 常量：SCREAMING_SNAKE_CASE 或 CamelCase
const MaxRetries = 3
const defaultTimeout = 30 * time.Second
```

## 测试

### 测试结构

```go
package service_test

import (
    "context"
    "testing"
    
    "github.com/openhost/openhost/internal/core/service"
    "github.com/openhost/openhost/internal/core/domain"
)

func TestOrderService_CreateOrder(t *testing.T) {
    // 准备
    mockRepo := &MockOrderRepository{}
    svc := service.NewOrderService(mockRepo)
    
    // 执行
    order, err := svc.CreateOrder(context.Background(), createRequest)
    
    // 断言
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if order.ID == uuid.Nil {
        t.Error("expected order ID to be set")
    }
}
```

### 测试覆盖率

- 业务逻辑目标 80%+ 覆盖率
- 关键路径 100% 覆盖率（支付、配置）
- 不测试琐碎的 getter/setter
- 关注行为，而非实现

### 运行测试

```bash
# 所有测试
go test ./...

# 带覆盖率
go test -cover ./...

# 详细
go test -v ./...

# 特定包
go test ./internal/core/service/...

# 带竞态检测器
go test -race ./...
```

## 文档

### 代码文档

```go
// Package billing 提供计费和发票管理服务。
package billing

// Order 表示客户购买订单。
type Order struct {
    // ID 是订单的唯一标识符。
    ID uuid.UUID
    
    // CustomerID 引用购买客户。
    CustomerID uuid.UUID
}

// CreateOrder 为给定客户创建新订单。
// 它验证产品可用性并计算定价。
//
// 如果产品不活动或缺货，返回错误。
func (s *OrderService) CreateOrder(ctx context.Context, 
    req CreateOrderRequest) (*Order, error) {
    // 实现
}
```

### API 文档

使用 Swagger 注释：

```go
// CreateOrder godoc
// @Summary 创建新订单
// @Description 为客户创建具有指定产品的新订单
// @Tags orders
// @Accept json
// @Produce json
// @Param order body CreateOrderRequest true "订单详情"
// @Success 201 {object} Order
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
    // 实现
}
```

## 性能指南

### 数据库

```go
// ✅ 使用预加载避免 N+1
orders, err := db.Preload("Items").Find(&orders).Error

// ❌ 避免 N+1 查询
for _, order := range orders {
    db.Find(&order.Items)  // N 次查询！
}

// ✅ 使用批量操作
db.CreateInBatches(orders, 100)

// ✅ 使用索引
type Order struct {
    ID         uuid.UUID `gorm:"primaryKey"`
    CustomerID uuid.UUID `gorm:"index"`  // 已索引
    CreatedAt  time.Time `gorm:"index"`  // 已索引
}
```

### 并发

```go
// ✅ 使用上下文进行取消
func (s *Service) Process(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case result := <-s.work():
        return s.handle(result)
    }
}

// ✅ 对读取密集的工作负载使用 sync.RWMutex
type Cache struct {
    mu    sync.RWMutex
    items map[string]interface{}
}

func (c *Cache) Get(key string) interface{} {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.items[key]
}
```

## 安全

### 输入验证

```go
// ✅ 验证所有输入
func (h *Handler) CreateOrder(c *gin.Context) {
    var req CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }
    
    if err := req.Validate(); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // 处理请求
}
```

### SQL 注入防护

```go
// ✅ 使用 GORM 参数化
db.Where("email = ?", email).First(&user)

// ❌ 永远不要拼接 SQL
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)  // 易受攻击！
```

### 密钥管理

```go
// ✅ 使用环境变量
dbPassword := os.Getenv("DB_PASSWORD")

// ❌ 永远不要硬编码密钥
const dbPassword = "secret123"  // 永远不要！
```

## 审查流程

### 拉取请求审查检查表

- [ ] 代码遵循风格指南
- [ ] 包含测试并通过
- [ ] 文档已更新
- [ ] 没有引入安全漏洞
- [ ] 记录了破坏性更改
- [ ] 提交消息遵循约定
- [ ] 分支与主分支保持最新

### 对于审查者

- 具有建设性和尊重性
- 解释为什么需要更改
- 建议具体改进
- 准备好后批准或请求更改
- 如果可能，测试更改

## 获取帮助

- **问题**: [GitHub 讨论](https://github.com/lbyxiaolizi/xlpanel/discussions)
- **错误**: [GitHub Issues](https://github.com/lbyxiaolizi/xlpanel/issues)
- **文档**: [docs/](docs/)
- **架构**: [ARCHITECTURE.zh-CN.md](ARCHITECTURE.zh-CN.md)

## 许可证

通过贡献，您同意您的贡献将在 MIT 许可证下授权。

## 致谢

贡献者将得到认可：
- GitHub 贡献者页面
- 发布说明（对于重要贡献）
- 项目文档

感谢您为 OpenHost 做出贡献！🎉
