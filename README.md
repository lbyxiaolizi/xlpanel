# OpenHost

English | [简体中文](README.zh-CN.md)

A modern hosting and billing management system written in Go - an open-source alternative to WHMCS/Blesta.

## ✨ Features

- 🚀 **High Performance** - Built on Go 1.23+ with excellent concurrency
- 🔌 **Plugin Architecture** - Modular design using HashiCorp go-plugin
- 💰 **Precise Billing** - High-precision money calculations with shopspring/decimal
- 🏗️ **Clean Architecture** - Domain-Driven Design (DDD) and Clean Architecture principles
- 🔒 **Type Safety** - Strict typing and thread-safe guarantees
- 📊 **Modern Stack** - PostgreSQL, Redis, GORM, Gin, and other mature technologies

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.23+
- **Web Framework**: Gin
- **Database**: PostgreSQL + GORM
- **Cache/Queue**: Redis + Asynq
- **Plugin System**: HashiCorp go-plugin (gRPC)
- **API Docs**: Swagger/OpenAPI

### Frontend
- **Framework**: Vue.js (optional)
- **Template Engine**: Go HTML Templates
- **Theme System**: Customizable themes

## 📦 Quick Start

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 12+
- Redis 6+
- Make

### Installation

1. Clone the repository
```bash
git clone https://github.com/lbyxiaolizi/xlpanel.git
cd xlpanel
```

2. Install dependencies
```bash
go mod download
```

3. Build the project
```bash
make all
```

### Run the Server

Start the main server:
```bash
./bin/server
```

The server will start at `http://localhost:6421`.

API health check:
```bash
curl http://localhost:6421/api/v1/health
```

## 🔌 Plugin System

OpenHost uses a gRPC-based plugin system that supports dynamic loading of provisioning modules.

## Plugin Compilation & Registration

### Compile a plugin

Plugins are gRPC binaries launched by the host. Each plugin binary must include a matching SHA-256 checksum file.

```bash
go build -o plugins/provisioner-example ./cmd/mock_plugin
sha256sum plugins/provisioner-example > plugins/provisioner-example.sha256
```

### Register a plugin in the database

Register the plugin module name and metadata in your `products` table so services can reference the correct module name.

Example (SQL):

```sql
INSERT INTO products (name, slug, module_name, active)
VALUES ('Example VPS', 'example-vps', 'provisioner-example', true);
```

When provisioning, the system resolves `module_name` to a plugin binary under `./plugins/` and establishes a gRPC connection.

## 🏗️ Architecture

OpenHost follows a clean layered architecture:

```
openhost/
├── cmd/              # Application entry points
│   ├── server/      # Main API server
│   ├── emailpipe/   # Email processing service
│   └── mock_plugin/ # Example plugin
├── internal/        # Private application code
│   ├── core/       # Business logic layer
│   │   ├── domain/ # Domain models
│   │   └── service/# Domain services
│   └── infrastructure/ # Infrastructure layer
│       ├── http/   # HTTP handlers
│       ├── web/    # Template rendering
│       ├── plugin/ # Plugin management
│       └── tasks/  # Background tasks
├── pkg/            # Public libraries
│   └── proto/      # Protocol Buffer definitions
├── themes/         # Frontend themes
└── docs/           # Documentation
```

See [Architecture Documentation](docs/ARCHITECTURE.md) for details.

## 📚 Documentation

- [Architecture Design](docs/ARCHITECTURE.md) - System architecture and design principles
- [Deployment Guide](docs/DEPLOYMENT.md) - Production deployment instructions
- [Installation Guide](docs/INSTALLATION.md) - Web installer walkthrough
- [Plugin Development](docs/PLUGIN_DEVELOPMENT.md) - How to develop custom plugins
- [API Documentation](docs/API.md) - RESTful API reference
- [Contributing Guidelines](docs/CONTRIBUTING.md) - How to contribute to the project

## 🎨 Theme System

OpenHost supports a custom theme system allowing you to create personalized client and admin interfaces.

Themes are located in the `themes/` directory and written using Go HTML templates.

```
themes/
├── default/        # Default theme
│   ├── layouts/   # Layout templates
│   ├── pages/     # Page templates
│   └── assets/    # Static assets
└── custom/        # Custom themes
```

## 🔧 Build Targets

```bash
# Build all components
make all

# Build individually
make server       # API server
make emailpipe    # Email processing
make mock_plugin  # Example plugin
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./internal/core/service/...

# With coverage
go test -cover ./...
```

## 🤝 Contributing

We welcome all forms of contribution! Please read the [Contributing Guidelines](docs/CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin) - Plugin system
- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [GORM](https://gorm.io/) - ORM library
- [Asynq](https://github.com/hibiken/asynq) - Task queue
- [shopspring/decimal](https://github.com/shopspring/decimal) - Precise decimal arithmetic

## 📧 Contact

- Project Homepage: [https://github.com/lbyxiaolizi/xlpanel](https://github.com/lbyxiaolizi/xlpanel)
- Issue Tracker: [https://github.com/lbyxiaolizi/xlpanel/issues](https://github.com/lbyxiaolizi/xlpanel/issues)

---

**Note**: OpenHost is under active development. Contributions and feedback are welcome!
