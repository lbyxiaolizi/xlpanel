# Contributing to OpenHost

[English](CONTRIBUTING.md) | [简体中文](CONTRIBUTING.zh-CN.md)

Thank you for your interest in contributing to OpenHost! This document provides guidelines and instructions for contributing.

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inclusive environment for all contributors.

### Our Standards

- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

## How to Contribute

### Reporting Bugs

Before creating a bug report:
1. Check the [existing issues](https://github.com/lbyxiaolizi/xlpanel/issues)
2. Try the latest version to see if the bug still exists

When creating a bug report, include:
- A clear and descriptive title
- Steps to reproduce the behavior
- Expected behavior
- Actual behavior
- Screenshots (if applicable)
- Environment details (OS, Go version, etc.)

**Example:**
```markdown
### Bug: Server crashes on invalid plugin checksum

**Steps to Reproduce:**
1. Create plugin without SHA256 file
2. Start server
3. Trigger plugin load

**Expected:** Error message logged, server continues
**Actual:** Server crashes with panic

**Environment:**
- OS: Ubuntu 22.04
- Go: 1.23.0
- OpenHost: v0.1.0
```

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. Include:
- Clear and descriptive title
- Detailed description of the proposed functionality
- Use cases and benefits
- Possible implementation approach
- Examples from similar projects (if applicable)

### Pull Requests

1. **Fork the repository**
```bash
git clone https://github.com/yourusername/xlpanel.git
cd xlpanel
```

2. **Create a feature branch**
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

3. **Make your changes**
- Follow the coding standards (see below)
- Write or update tests
- Update documentation

4. **Commit your changes**
```bash
git add .
git commit -m "feat: add new feature"
# or
git commit -m "fix: resolve bug in provisioning"
```

Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Maintenance tasks

5. **Push to your fork**
```bash
git push origin feature/your-feature-name
```

6. **Create Pull Request**
- Use a clear title and description
- Reference related issues
- Include screenshots for UI changes
- Ensure all tests pass
- Request review from maintainers

## Development Setup

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 12+
- Redis 6+
- Make
- Git

### Local Development

1. **Clone and setup**
```bash
git clone https://github.com/lbyxiaolizi/xlpanel.git
cd xlpanel
go mod download
```

2. **Setup database**
```bash
createdb openhost
psql -d openhost -f schema.sql
```

3. **Configure environment**
```bash
cp .env.example .env
# Edit .env with your settings
```

4. **Build**
```bash
make all
```

5. **Run tests**
```bash
go test ./...
```

6. **Start server**
```bash
./bin/server
```

## Coding Standards

### Go Style Guide

Follow the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) and [Effective Go](https://golang.org/doc/effective_go.html).

### Key Principles

1. **Clean Architecture**
   - Keep domain layer pure (no dependencies)
   - Use interfaces for dependencies
   - Dependency injection

2. **Type Safety**
   - Use `shopspring/decimal` for money
   - Avoid `interface{}` when possible
   - Explicit error handling

3. **Thread Safety**
   - Protect shared state with mutexes
   - Use channels for communication
   - Context for cancellation

### Code Examples

**Good:**
```go
// Use decimal for money
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

**Bad:**
```go
type Product struct {
    ID    string
    Name  string
    Price float64  // ❌ Don't use float for money
}

func (p *Product) CalculateTotal(quantity int) float64 {
    return p.Price * float64(quantity)  // ❌ Floating point errors
}
```

**Good:**
```go
// Proper error handling
func (s *OrderService) CreateOrder(ctx context.Context, 
    req CreateOrderRequest) (*Order, error) {
    
    product, err := s.productRepo.FindByID(ctx, req.ProductID)
    if err != nil {
        return nil, fmt.Errorf("failed to find product: %w", err)
    }
    
    if !product.Active {
        return nil, ErrProductInactive
    }
    
    // ... rest of logic
}
```

**Bad:**
```go
// Poor error handling
func (s *OrderService) CreateOrder(ctx context.Context, 
    req CreateOrderRequest) (*Order, error) {
    
    product, _ := s.productRepo.FindByID(ctx, req.ProductID)  // ❌ Ignoring errors
    
    // No validation ❌
    
    // ... rest of logic
}
```

### Naming Conventions

```go
// Interfaces: noun or adjective
type Repository interface {}
type Provisioner interface {}

// Structs: clear nouns
type Order struct {}
type Product struct {}

// Functions: verb-noun or clear purpose
func CreateOrder() {}
func ValidateInput() {}
func (s *Service) ProcessPayment() {}

// Variables: descriptive
orderID := uuid.New()
totalPrice := calculateTotal(items)

// Constants: SCREAMING_SNAKE_CASE or CamelCase
const MaxRetries = 3
const defaultTimeout = 30 * time.Second
```

### Package Organization

```go
// domain package - pure business logic
package domain

type Order struct {
    ID         uuid.UUID
    CustomerID uuid.UUID
    Items      []OrderItem
    Total      decimal.Decimal
}

// No external dependencies in domain!

// service package - orchestration
package service

import "github.com/openhost/openhost/internal/core/domain"

type OrderService struct {
    orderRepo   domain.OrderRepository
    productRepo domain.ProductRepository
}

// infrastructure package - implementations
package postgres

import (
    "github.com/openhost/openhost/internal/core/domain"
    "gorm.io/gorm"
)

type OrderRepository struct {
    db *gorm.DB
}

func (r *OrderRepository) Save(order *domain.Order) error {
    return r.db.Create(order).Error
}
```

## Testing

### Test Structure

```go
package service_test

import (
    "context"
    "testing"
    
    "github.com/openhost/openhost/internal/core/service"
    "github.com/openhost/openhost/internal/core/domain"
)

func TestOrderService_CreateOrder(t *testing.T) {
    // Arrange
    mockRepo := &MockOrderRepository{}
    svc := service.NewOrderService(mockRepo)
    
    // Act
    order, err := svc.CreateOrder(context.Background(), createRequest)
    
    // Assert
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if order.ID == uuid.Nil {
        t.Error("expected order ID to be set")
    }
}
```

### Test Coverage

- Aim for 80%+ coverage on business logic
- 100% coverage on critical paths (payment, provisioning)
- Don't test trivial getters/setters
- Focus on behavior, not implementation

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Verbose
go test -v ./...

# Specific package
go test ./internal/core/service/...

# With race detector
go test -race ./...
```

## Documentation

### Code Documentation

```go
// Package billing provides billing and invoice management services.
package billing

// Order represents a customer purchase order.
type Order struct {
    // ID is the unique identifier for the order.
    ID uuid.UUID
    
    // CustomerID references the purchasing customer.
    CustomerID uuid.UUID
}

// CreateOrder creates a new order for the given customer.
// It validates the product availability and calculates pricing.
//
// Returns an error if the product is inactive or out of stock.
func (s *OrderService) CreateOrder(ctx context.Context, 
    req CreateOrderRequest) (*Order, error) {
    // implementation
}
```

### API Documentation

Use Swagger annotations:

```go
// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order for a customer with specified products
// @Tags orders
// @Accept json
// @Produce json
// @Param order body CreateOrderRequest true "Order details"
// @Success 201 {object} Order
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
    // implementation
}
```

### README and Guides

- Keep documentation up to date
- Include code examples
- Add screenshots for UI features
- Translate important docs to Chinese

## Performance Guidelines

### Database

```go
// ✅ Use preloading to avoid N+1
orders, err := db.Preload("Items").Find(&orders).Error

// ❌ Avoid N+1 queries
for _, order := range orders {
    db.Find(&order.Items)  // N queries!
}

// ✅ Use batch operations
db.CreateInBatches(orders, 100)

// ✅ Use indexes
type Order struct {
    ID         uuid.UUID `gorm:"primaryKey"`
    CustomerID uuid.UUID `gorm:"index"`  // Indexed
    CreatedAt  time.Time `gorm:"index"`  // Indexed
}
```

### Concurrency

```go
// ✅ Use context for cancellation
func (s *Service) Process(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case result := <-s.work():
        return s.handle(result)
    }
}

// ✅ Use sync.RWMutex for read-heavy workloads
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

## Security

### Input Validation

```go
// ✅ Validate all inputs
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
    
    // Process request
}
```

### SQL Injection Prevention

```go
// ✅ Use GORM parameterization
db.Where("email = ?", email).First(&user)

// ❌ Never concatenate SQL
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)  // Vulnerable!
```

### Secrets Management

```go
// ✅ Use environment variables
dbPassword := os.Getenv("DB_PASSWORD")

// ❌ Never hardcode secrets
const dbPassword = "secret123"  // Never!
```

## Review Process

### Pull Request Review Checklist

- [ ] Code follows style guidelines
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] No security vulnerabilities introduced
- [ ] Breaking changes are documented
- [ ] Commit messages follow conventions
- [ ] Branch is up to date with main

### For Reviewers

- Be constructive and respectful
- Explain why changes are needed
- Suggest specific improvements
- Approve when ready or request changes
- Test the changes if possible

## Getting Help

- **Questions**: [GitHub Discussions](https://github.com/lbyxiaolizi/xlpanel/discussions)
- **Bugs**: [GitHub Issues](https://github.com/lbyxiaolizi/xlpanel/issues)
- **Documentation**: [docs/](docs/)
- **Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be recognized in:
- GitHub contributors page
- Release notes (for significant contributions)
- Project documentation

Thank you for contributing to OpenHost! 🎉
