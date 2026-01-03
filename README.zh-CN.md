# OpenHost

[English](README.md) | 简体中文

OpenHost 是一个现代化的主机托管和计费管理系统，使用 Go 语言编写，是 WHMCS/Blesta 的开源替代方案。

## ✨ 特性

- 🚀 **高性能** - 基于 Go 1.23+ 构建，具有出色的并发性能
- 🔌 **插件化架构** - 使用 HashiCorp go-plugin 实现模块化设计
- 💰 **精确计费** - 使用 shopspring/decimal 进行高精度货币计算
- 🏗️ **清晰架构** - 采用领域驱动设计（DDD）和整洁架构原则
- 🔒 **类型安全** - 严格的类型系统和线程安全保证
- 📊 **现代技术栈** - PostgreSQL、Redis、GORM、Gin 等成熟技术

## 🛠️ 技术栈

### 后端
- **语言**: Go 1.23+
- **Web 框架**: Gin
- **数据库**: PostgreSQL + GORM
- **缓存/队列**: Redis + Asynq
- **插件系统**: HashiCorp go-plugin (gRPC)
- **API 文档**: Swagger/OpenAPI

### 前端
- **框架**: Vue.js (可选)
- **模板引擎**: Go HTML Templates
- **主题系统**: 可自定义主题

## 📦 快速开始

### 前置要求

- Go 1.23 或更高版本
- PostgreSQL 12+
- Redis 6+
- Make

### 安装

1. 克隆仓库
```bash
git clone https://github.com/lbyxiaolizi/xlpanel.git
cd xlpanel
```

2. 安装依赖
```bash
go mod download
```

3. 构建项目
```bash
make all
```

### 运行服务

启动主服务器：
```bash
./bin/server
```

服务将在 `http://localhost:6421` 启动。

API 健康检查：
```bash
curl http://localhost:6421/api/v1/health
```

## 🔌 插件系统

OpenHost 使用基于 gRPC 的插件系统，支持动态加载供应商模块。

### 编译插件

插件是由主程序启动的 gRPC 二进制文件。每个插件二进制文件必须包含匹配的 SHA-256 校验和文件。

```bash
# 构建插件
go build -o plugins/provisioner-example ./cmd/mock_plugin

# 生成校验和
sha256sum plugins/provisioner-example > plugins/provisioner-example.sha256
```

### 注册插件

在数据库中注册插件模块名称和元数据，以便服务可以引用正确的模块名称。

示例（SQL）：
```sql
INSERT INTO products (name, slug, module_name, active)
VALUES ('示例 VPS', 'example-vps', 'provisioner-example', true);
```

在配置时，系统会将 `module_name` 解析为 `./plugins/` 下的插件二进制文件，并建立 gRPC 连接。

## 🏗️ 架构

OpenHost 采用清晰的分层架构：

```
openhost/
├── cmd/              # 应用程序入口点
│   ├── server/      # 主 API 服务器
│   ├── emailpipe/   # 邮件处理服务
│   └── mock_plugin/ # 示例插件
├── internal/        # 私有应用程序代码
│   ├── core/       # 业务逻辑层
│   │   ├── domain/ # 领域模型
│   │   └── service/# 领域服务
│   └── infrastructure/ # 基础设施层
│       ├── http/   # HTTP 处理器
│       ├── web/    # 模板渲染
│       ├── plugin/ # 插件管理
│       └── tasks/  # 后台任务
├── pkg/            # 公共库
│   └── proto/      # Protocol Buffer 定义
├── themes/         # 前端主题
└── docs/           # 文档
```

详细架构说明请参阅 [架构文档](docs/ARCHITECTURE.zh-CN.md)。

## 📚 文档

- [架构设计](docs/ARCHITECTURE.zh-CN.md) - 系统架构和设计原则
- [部署指南](docs/DEPLOYMENT.zh-CN.md) - 生产环境部署说明
- [安装指南](docs/INSTALLATION.zh-CN.md) - Web 安装向导说明
- [插件开发](docs/PLUGIN_DEVELOPMENT.zh-CN.md) - 如何开发自定义插件
- [API 文档](docs/API.zh-CN.md) - RESTful API 参考
- [贡献指南](docs/CONTRIBUTING.zh-CN.md) - 如何参与项目开发

## 🎨 主题系统

OpenHost 支持自定义主题系统，允许您创建个性化的客户端和管理界面。

主题位于 `themes/` 目录下，使用 Go HTML 模板语言编写。

```
themes/
├── default/        # 默认主题
│   ├── layouts/   # 布局模板
│   ├── pages/     # 页面模板
│   └── assets/    # 静态资源
└── custom/        # 自定义主题
```

## 🔧 构建目标

```bash
# 构建所有组件
make all

# 单独构建
make server       # API 服务器
make emailpipe    # 邮件处理
make mock_plugin  # 示例插件
```

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/core/service/...

# 带覆盖率
go test -cover ./...
```

## 🤝 贡献

我们欢迎所有形式的贡献！请阅读 [贡献指南](docs/CONTRIBUTING.zh-CN.md) 了解详情。

### 开发流程

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个 Pull Request

## 📝 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin) - 插件系统
- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [GORM](https://gorm.io/) - ORM 库
- [Asynq](https://github.com/hibiken/asynq) - 任务队列
- [shopspring/decimal](https://github.com/shopspring/decimal) - 精确数值计算

## 📧 联系方式

- 项目主页: [https://github.com/lbyxiaolizi/xlpanel](https://github.com/lbyxiaolizi/xlpanel)
- Issue 追踪: [https://github.com/lbyxiaolizi/xlpanel/issues](https://github.com/lbyxiaolizi/xlpanel/issues)

---

**注意**: OpenHost 目前处于积极开发阶段。欢迎贡献和反馈！
