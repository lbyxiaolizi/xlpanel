# OpenHost Architecture

[English](ARCHITECTURE.md) | [简体中文](ARCHITECTURE.zh-CN.md)

## Overview

OpenHost follows **Clean Architecture** principles with **Domain-Driven Design (DDD)** patterns, organized as a **Modular Monolith** with process isolation for plugins.

## Design Principles

### 1. Clean Architecture

The system is organized in concentric layers with dependencies pointing inward:

```
┌─────────────────────────────────────────────┐
│         Infrastructure Layer                │
│  (HTTP, DB, Cache, External Services)       │
│  ┌───────────────────────────────────────┐  │
│  │      Application Layer                │  │
│  │    (Use Cases, Orchestration)         │  │
│  │  ┌─────────────────────────────────┐  │  │
│  │  │     Domain Layer                │  │  │
│  │  │  (Entities, Value Objects,      │  │  │
│  │  │   Domain Services, Interfaces)  │  │  │
│  │  └─────────────────────────────────┘  │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

**Key Rules:**
- Domain layer has no dependencies
- Application layer depends only on domain
- Infrastructure implements interfaces defined in domain/application
- Dependencies point inward

### 2. Domain-Driven Design

**Core Concepts:**
- **Entities**: Objects with unique identity (Product, Order, Customer)
- **Value Objects**: Immutable objects defined by attributes (Money, Email, Address)
- **Aggregates**: Clusters of related entities with a root (Order with OrderItems)
- **Domain Services**: Business logic that doesn't belong to entities
- **Repositories**: Abstraction for data persistence
- **Domain Events**: Communicate state changes across boundaries

### 3. Modular Monolith

The application is structured as independent modules within a single deployment:

```
internal/
├── core/
│   ├── domain/          # Domain models and interfaces
│   │   ├── catalog.go   # Product catalog domain
│   │   ├── order.go     # Order management domain
│   │   ├── ticket.go    # Support ticket domain
│   │   └── ipam.go      # IP address management domain
│   └── service/         # Domain services
│       ├── billing/     # Billing service
│       └── ipam/        # IPAM service
└── infrastructure/
    ├── http/            # HTTP handlers
    ├── web/             # Template rendering
    ├── plugin/          # Plugin system
    ├── tasks/           # Background tasks
    └── tickets/         # Ticket processing
```

## Layers in Detail

### Domain Layer (`internal/core/domain/`)

**Purpose**: Contains core business logic and rules

**Components:**
- **Entities**: Business objects with identity
- **Value Objects**: Immutable data structures
- **Domain Services**: Complex business logic
- **Repository Interfaces**: Data access contracts
- **Domain Events**: State change notifications

**Example:**
```go
// Product entity
type Product struct {
    ID          uuid.UUID
    Name        string
    Price       decimal.Decimal
    ModuleName  string  // Plugin reference
    Active      bool
}

// Domain service
type BillingService interface {
    CalculateInvoice(orderID uuid.UUID) (*Invoice, error)
    ProcessPayment(invoiceID uuid.UUID, payment Payment) error
}
```

**Rules:**
- No external dependencies (only Go stdlib)
- Use `shopspring/decimal` for money
- Thread-safe operations
- Immutable where possible

### Service Layer (`internal/core/service/`)

**Purpose**: Orchestrates domain logic and coordinates operations

**Components:**
- Business workflows
- Transaction management
- Domain event publishing
- Cross-aggregate operations

**Example:**
```go
type OrderService struct {
    orderRepo   domain.OrderRepository
    productRepo domain.ProductRepository
    billing     domain.BillingService
}

func (s *OrderService) CreateOrder(ctx context.Context, 
    customerID uuid.UUID, items []OrderItem) (*Order, error) {
    // Validate products
    // Create order aggregate
    // Generate invoice
    // Publish domain event
}
```

### Infrastructure Layer (`internal/infrastructure/`)

**Purpose**: Implements technical capabilities and external integrations

#### HTTP Layer (`http/`)
- RESTful API handlers
- Request validation
- Response formatting
- Authentication/Authorization

```go
// Handler example
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

#### Web Layer (`web/`)
- Template rendering
- Theme management
- Hook system for extensions

#### Plugin Layer (`plugin/`)
- gRPC plugin management
- Process isolation
- Plugin discovery and loading
- Health checking

#### Tasks Layer (`tasks/`)
- Background job processing
- Scheduled tasks
- Async operations using Asynq

## Plugin Architecture

### Design

OpenHost uses **HashiCorp go-plugin** for process-isolated plugins:

```
┌──────────────┐         gRPC           ┌──────────────┐
│              │◄──────────────────────►│              │
│   OpenHost   │                        │   Plugin     │
│   (Host)     │  Protocol Buffers      │  (Separate   │
│              │                        │   Process)   │
└──────────────┘                        └──────────────┘
```

**Benefits:**
- **Process Isolation**: Plugin crashes don't affect host
- **Language Agnostic**: Plugins can be written in any language (with gRPC)
- **Versioning**: Multiple plugin versions can coexist
- **Security**: Sandboxed execution with SHA-256 verification

### Plugin Lifecycle

1. **Discovery**: Host scans `./plugins/` directory
2. **Verification**: Check SHA-256 checksums
3. **Launch**: Start plugin as subprocess
4. **Handshake**: Establish gRPC connection
5. **Communication**: RPC calls via Protocol Buffers
6. **Cleanup**: Graceful shutdown on host termination

### Plugin Interface

Defined in `pkg/proto/provisioner/`:

```protobuf
service Provisioner {
    rpc Provision(ProvisionRequest) returns (ProvisionResponse);
    rpc Suspend(SuspendRequest) returns (SuspendResponse);
    rpc Unsuspend(UnsuspendRequest) returns (UnsuspendResponse);
    rpc Terminate(TerminateRequest) returns (TerminateResponse);
}
```

## Data Flow

### Request Processing Flow

```
┌─────────┐    ┌──────────┐    ┌─────────┐    ┌────────┐    ┌────────┐
│ Client  │───►│ Handler  │───►│ Service │───►│ Domain │───►│   DB   │
└─────────┘    └──────────┘    └─────────┘    └────────┘    └────────┘
    HTTP          Validate      Orchestrate    Business      Persist
                  Request       Workflow       Rules         Data
```

### Plugin Integration Flow

```
┌─────────┐    ┌─────────┐    ┌────────────┐    ┌────────┐
│ Service │───►│ Plugin  │───►│   gRPC     │───►│ Plugin │
│         │    │ Manager │    │            │    │Process │
└─────────┘    └─────────┘    └────────────┘    └────────┘
   Request      Lookup         Remote Call       Execute
                Plugin         (Protobuf)        Logic
```

### Async Task Flow

```
┌─────────┐    ┌────────┐    ┌───────┐    ┌────────┐
│ Service │───►│ Asynq  │───►│ Redis │───►│ Worker │
└─────────┘    └────────┘    └───────┘    └────────┘
   Enqueue     Schedule       Queue        Process
   Task        Job           Job           Task
```

## Concurrency Model

### Thread Safety

**Principles:**
- **Immutable Domain Objects**: Value objects are immutable
- **Mutex Protection**: Shared state uses sync.RWMutex
- **Context Propagation**: Use context.Context for cancellation
- **Channel Communication**: Prefer channels over shared memory

**Example:**
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

### Database Transactions

Use GORM's transaction support for atomic operations:

```go
func (s *OrderService) CreateOrder(ctx context.Context, 
    req CreateOrderRequest) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // Create order
        if err := tx.Create(&order).Error; err != nil {
            return err
        }
        
        // Create invoice
        if err := tx.Create(&invoice).Error; err != nil {
            return err
        }
        
        // All or nothing
        return nil
    })
}
```

## Money Handling

**Critical Rule**: Always use `shopspring/decimal` for money calculations.

```go
import "github.com/shopspring/decimal"

// ✅ Correct
price := decimal.NewFromFloat(99.99)
quantity := decimal.NewFromInt(3)
total := price.Mul(quantity)

// ❌ Wrong
price := 99.99
quantity := 3
total := price * float64(quantity) // Floating point errors!
```

## Error Handling

### Error Types

```go
// Domain errors
var (
    ErrProductNotFound = errors.New("product not found")
    ErrInsufficientFunds = errors.New("insufficient funds")
    ErrInvalidState = errors.New("invalid state transition")
)

// Wrapped errors for context
if err != nil {
    return fmt.Errorf("failed to create order: %w", err)
}
```

### Error Responses

```go
// HTTP error response
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code"`
    Details any    `json:"details,omitempty"`
}
```

## Configuration

### Environment-Based Config

```go
type Config struct {
    Database DatabaseConfig
    Redis    RedisConfig
    Server   ServerConfig
    Plugins  PluginConfig
}

func LoadConfig() (*Config, error) {
    // Load from env vars or config file
}
```

## Security Considerations

### 1. Plugin Security
- SHA-256 checksum verification
- Process isolation
- Resource limits
- Timeout controls

### 2. Data Security
- SQL injection prevention (GORM parameterization)
- XSS protection (template escaping)
- CSRF tokens for forms
- Input validation

### 3. API Security
- Authentication middleware
- Rate limiting
- Request size limits
- CORS configuration

## Testing Strategy

### Unit Tests
- Domain logic (pure functions)
- Service layer (mocked dependencies)
- Utilities and helpers

### Integration Tests
- HTTP handlers with test database
- Plugin integration
- Database operations

### E2E Tests
- Full workflow scenarios
- API endpoint testing
- UI testing (if applicable)

## Performance Considerations

### Database
- Connection pooling
- Query optimization
- Proper indexing
- N+1 query prevention

### Caching
- Redis for session data
- In-memory cache for reference data
- Cache invalidation strategy

### Background Jobs
- Asynq for async processing
- Queue prioritization
- Retry mechanisms
- Dead letter queues

## Scalability

### Horizontal Scaling
- Stateless application design
- External session storage (Redis)
- Database connection pooling
- Load balancer ready

### Vertical Scaling
- Go's efficient goroutines
- Minimal memory footprint
- CPU-bound task optimization

## Monitoring and Observability

### Logging
- Structured logging
- Log levels (DEBUG, INFO, WARN, ERROR)
- Context propagation

### Metrics
- Request latency
- Error rates
- Plugin health
- Queue depth

### Tracing
- Request ID tracking
- Distributed tracing support

## Future Enhancements

- [ ] Event sourcing for audit trail
- [ ] CQRS for read/write separation
- [ ] GraphQL API option
- [ ] Multi-tenancy support
- [ ] Advanced plugin marketplace

## References

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design by Eric Evans](https://www.domainlanguage.com/ddd/)
- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin)
- [GORM Documentation](https://gorm.io/docs/)
- [Asynq Documentation](https://github.com/hibiken/asynq)
