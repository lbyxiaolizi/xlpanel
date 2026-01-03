# 插件开发指南

[English](PLUGIN_DEVELOPMENT.md) | 简体中文

## 概述

OpenHost 使用基于插件的架构来实现供应模块。插件是独立的进程，通过使用 Protocol Buffers 的 gRPC 与主应用程序通信。

## 架构

### 插件系统

```
┌─────────────────┐      gRPC/Protobuf       ┌─────────────────┐
│                 │◄────────────────────────►│                 │
│   OpenHost      │                          │     插件        │
│   (宿主)        │  握手和 RPC 调用          │  (供应程序)     │
│                 │                          │                 │
└─────────────────┘                          └─────────────────┘
      进程 1                                       进程 2
```

**优势：**
- **崩溃隔离**：插件崩溃不影响宿主
- **语言无关**：使用任何支持 gRPC 的语言编写插件
- **安全性**：SHA-256 校验和验证
- **版本控制**：多个插件版本可以共存

## 入门

### 前置要求

- Go 1.22+（用于 Go 插件）
- Protocol Buffer 编译器 (protoc)
- go-plugin 知识（有帮助）

### 快速开始

1. 克隆仓库：
```bash
git clone https://github.com/lbyxiaolizi/xlpanel.git
cd xlpanel
```

2. 查看示例插件：
```bash
cd cmd/mock_plugin
cat main.go
```

3. 构建示例：
```bash
go build -o ../../plugins/mock-provisioner .
cd ../..
sha256sum plugins/mock-provisioner > plugins/mock-provisioner.sha256
```

## 插件接口

### Protocol Buffer 定义

位于 `pkg/proto/provisioner/provisioner.proto`：

```protobuf
syntax = "proto3";

package provisioner;
option go_package = "github.com/openhost/openhost/pkg/proto/provisioner";

service Provisioner {
    rpc Provision(ProvisionRequest) returns (ProvisionResponse);
    rpc Suspend(SuspendRequest) returns (SuspendResponse);
    rpc Unsuspend(UnsuspendRequest) returns (UnsuspendResponse);
    rpc Terminate(TerminateRequest) returns (TerminateResponse);
    rpc GetInfo(GetInfoRequest) returns (GetInfoResponse);
}

message ProvisionRequest {
    string service_id = 1;
    map<string, string> parameters = 2;
}

message ProvisionResponse {
    bool success = 1;
    string message = 2;
    map<string, string> metadata = 3;
}

// ... 其他消息
```

### 核心操作

每个插件必须实现这些 RPC 方法：

1. **Provision**: 创建/配置新服务
2. **Suspend**: 暂停活动服务
3. **Unsuspend**: 重新激活暂停的服务
4. **Terminate**: 永久删除服务
5. **GetInfo**: 返回插件元数据和功能

## 在 Go 中实现插件

### 步骤 1: 项目结构

```
my-plugin/
├── main.go
├── go.mod
└── README.md
```

### 步骤 2: 初始化模块

```bash
mkdir my-plugin
cd my-plugin
go mod init github.com/yourusername/my-plugin
```

### 步骤 3: 安装依赖

```bash
go get github.com/hashicorp/go-plugin
go get github.com/hashicorp/go-hclog
go get google.golang.org/grpc
# 复制 proto 文件或导入 openhost proto 包
```

### 步骤 4: 实现插件

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/hashicorp/go-hclog"
    "github.com/hashicorp/go-plugin"
    "google.golang.org/grpc"
    
    pb "github.com/openhost/openhost/pkg/proto/provisioner"
)

// 实现 Provisioner 接口
type MyProvisioner struct {
    logger hclog.Logger
}

func (p *MyProvisioner) Provision(ctx context.Context, 
    req *pb.ProvisionRequest) (*pb.ProvisionResponse, error) {
    
    p.logger.Info("配置服务", "service_id", req.ServiceId)
    
    // 您的配置逻辑在这里
    // 例如，创建虚拟机、配置 DNS、设置数据库等
    
    return &pb.ProvisionResponse{
        Success: true,
        Message: "服务配置成功",
        Metadata: map[string]string{
            "ip_address": "192.168.1.100",
            "hostname": "server1.example.com",
        },
    }, nil
}

func (p *MyProvisioner) Suspend(ctx context.Context, 
    req *pb.SuspendRequest) (*pb.SuspendResponse, error) {
    
    p.logger.Info("暂停服务", "service_id", req.ServiceId)
    
    // 您的暂停逻辑在这里
    
    return &pb.SuspendResponse{
        Success: true,
        Message: "服务已暂停",
    }, nil
}

func (p *MyProvisioner) Unsuspend(ctx context.Context, 
    req *pb.UnsuspendRequest) (*pb.UnsuspendResponse, error) {
    
    p.logger.Info("取消暂停服务", "service_id", req.ServiceId)
    
    // 您的取消暂停逻辑在这里
    
    return &pb.UnsuspendResponse{
        Success: true,
        Message: "服务已取消暂停",
    }, nil
}

func (p *MyProvisioner) Terminate(ctx context.Context, 
    req *pb.TerminateRequest) (*pb.TerminateResponse, error) {
    
    p.logger.Info("终止服务", "service_id", req.ServiceId)
    
    // 您的终止逻辑在这里
    
    return &pb.TerminateResponse{
        Success: true,
        Message: "服务已终止",
    }, nil
}

func (p *MyProvisioner) GetInfo(ctx context.Context, 
    req *pb.GetInfoRequest) (*pb.GetInfoResponse, error) {
    
    return &pb.GetInfoResponse{
        Name: "my-provisioner",
        Version: "1.0.0",
        Description: "我的自定义供应程序插件",
        Capabilities: []string{"provision", "suspend", "unsuspend", "terminate"},
    }, nil
}

// gRPC 服务器实现
type ProvisionerGRPCServer struct {
    pb.UnimplementedProvisionerServer
    Impl *MyProvisioner
}

func (s *ProvisionerGRPCServer) Provision(ctx context.Context, 
    req *pb.ProvisionRequest) (*pb.ProvisionResponse, error) {
    return s.Impl.Provision(ctx, req)
}

// ... 其他方法实现

// 插件实现
type ProvisionerPlugin struct {
    Impl *MyProvisioner
}

func (p *ProvisionerPlugin) GRPCServer(broker *plugin.GRPCBroker, 
    s *grpc.Server) error {
    pb.RegisterProvisionerServer(s, &ProvisionerGRPCServer{Impl: p.Impl})
    return nil
}

func (p *ProvisionerPlugin) GRPCClient(ctx context.Context, 
    broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
    return pb.NewProvisionerClient(c), nil
}

func main() {
    logger := hclog.New(&hclog.LoggerOptions{
        Level:      hclog.Debug,
        Output:     os.Stderr,
        JSONFormat: true,
    })

    provisioner := &MyProvisioner{
        logger: logger,
    }

    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: plugin.HandshakeConfig{
            ProtocolVersion:  1,
            MagicCookieKey:   "OPENHOST_PLUGIN",
            MagicCookieValue: "provisioner",
        },
        Plugins: map[string]plugin.Plugin{
            "provisioner": &ProvisionerPlugin{Impl: provisioner},
        },
        GRPCServer: plugin.DefaultGRPCServer,
        Logger:     logger,
    })
}
```

### 步骤 5: 构建和部署

```bash
# 构建插件
go build -o my-provisioner .

# 生成校验和
sha256sum my-provisioner > my-provisioner.sha256

# 部署到 OpenHost 插件目录
cp my-provisioner /path/to/openhost/plugins/
cp my-provisioner.sha256 /path/to/openhost/plugins/
```

## 插件配置

### 在数据库中注册插件

```sql
INSERT INTO products (
    name, 
    slug, 
    description,
    module_name, 
    config_options,
    active
) VALUES (
    '我的自定义服务',
    'my-custom-service',
    '服务描述',
    'my-provisioner',  -- 必须匹配插件二进制名称
    '{"api_key": "", "region": "us-east-1"}',  -- JSON 配置架构
    true
);
```

## 高级功能

### 自定义参数

从订单传递自定义参数到插件：

```go
func (p *MyProvisioner) Provision(ctx context.Context, 
    req *pb.ProvisionRequest) (*pb.ProvisionResponse, error) {
    
    // 访问自定义参数
    hostname := req.Parameters["hostname"]
    diskSize := req.Parameters["disk_size"]
    region := req.Parameters["region"]
    
    // 在配置逻辑中使用参数
    // ...
    
    return &pb.ProvisionResponse{
        Success: true,
        Message: fmt.Sprintf("在 %s 配置了 %s", region, hostname),
        Metadata: map[string]string{
            "hostname": hostname,
            "ip_address": allocatedIP,
        },
    }, nil
}
```

### 错误处理

```go
func (p *MyProvisioner) Provision(ctx context.Context, 
    req *pb.ProvisionRequest) (*pb.ProvisionResponse, error) {
    
    // 验证
    if req.ServiceId == "" {
        return &pb.ProvisionResponse{
            Success: false,
            Message: "需要服务 ID",
        }, nil
    }
    
    // 带错误处理的 API 调用
    result, err := providerAPI.CreateInstance(req.Parameters)
    if err != nil {
        p.logger.Error("配置失败", "error", err)
        return &pb.ProvisionResponse{
            Success: false,
            Message: fmt.Sprintf("配置失败: %v", err),
        }, nil
    }
    
    return &pb.ProvisionResponse{
        Success: true,
        Message: "配置成功",
        Metadata: result.Metadata,
    }, nil
}
```

### 上下文和超时

```go
func (p *MyProvisioner) Provision(ctx context.Context, 
    req *pb.ProvisionRequest) (*pb.ProvisionResponse, error) {
    
    // 创建超时上下文
    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()
    
    // 在 API 调用中使用上下文
    result, err := providerAPI.CreateInstanceWithContext(ctx, req.Parameters)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return &pb.ProvisionResponse{
                Success: false,
                Message: "配置超时",
            }, nil
        }
        return &pb.ProvisionResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }
    
    return &pb.ProvisionResponse{Success: true}, nil
}
```

## 测试插件

### 单元测试

```go
package main

import (
    "context"
    "testing"
    
    "github.com/hashicorp/go-hclog"
    pb "github.com/openhost/openhost/pkg/proto/provisioner"
)

func TestProvision(t *testing.T) {
    logger := hclog.NewNullLogger()
    provisioner := &MyProvisioner{logger: logger}
    
    req := &pb.ProvisionRequest{
        ServiceId: "test-service-123",
        Parameters: map[string]string{
            "hostname": "test.example.com",
            "region": "us-east-1",
        },
    }
    
    resp, err := provisioner.Provision(context.Background(), req)
    
    if err != nil {
        t.Fatalf("配置失败: %v", err)
    }
    
    if !resp.Success {
        t.Errorf("期望 success=true，得到 %v: %s", resp.Success, resp.Message)
    }
    
    if resp.Metadata["hostname"] != "test.example.com" {
        t.Errorf("期望元数据中有 hostname")
    }
}
```

## 插件示例

### 1. DigitalOcean Droplet 供应程序

```go
import "github.com/digitalocean/godo"

type DOProvisioner struct {
    client *godo.Client
    logger hclog.Logger
}

func (p *DOProvisioner) Provision(ctx context.Context, 
    req *pb.ProvisionRequest) (*pb.ProvisionResponse, error) {
    
    createReq := &godo.DropletCreateRequest{
        Name:   req.Parameters["hostname"],
        Region: req.Parameters["region"],
        Size:   req.Parameters["size"],
        Image:  godo.DropletCreateImage{Slug: "ubuntu-22-04-x64"},
    }
    
    droplet, _, err := p.client.Droplets.Create(ctx, createReq)
    if err != nil {
        return &pb.ProvisionResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }
    
    return &pb.ProvisionResponse{
        Success: true,
        Message: "Droplet 已创建",
        Metadata: map[string]string{
            "droplet_id": fmt.Sprintf("%d", droplet.ID),
            "ip_address": droplet.Networks.V4[0].IPAddress,
        },
    }, nil
}
```

## 安全最佳实践

1. **验证输入**：始终验证请求参数
2. **保护凭证**：将 API 密钥存储在环境变量或密钥管理器中
3. **超时**：为所有操作设置合理的超时
4. **日志记录**：不要记录敏感数据（密码、API 密钥）
5. **错误消息**：不要在错误消息中暴露内部详细信息
6. **资源限制**：为 API 调用实现速率限制
7. **校验和**：始终提供 SHA-256 校验和

## 调试

### 启用调试日志

```go
logger := hclog.New(&hclog.LoggerOptions{
    Level:      hclog.Debug,  // 或 hclog.Trace 用于详细
    Output:     os.Stderr,
    JSONFormat: true,
})
```

### 独立测试插件

```bash
# 直接运行插件（它将等待宿主连接）
PLUGIN_DEBUG=1 ./plugins/my-provisioner
```

## 插件生命周期

1. **发现**：宿主扫描 `./plugins/` 目录
2. **验证**：检查 SHA-256 校验和
3. **启动**：启动插件进程
4. **握手**：建立 gRPC 连接
5. **就绪**：插件准备好接受 RPC 调用
6. **关闭**：宿主终止时优雅清理

## 常见问题

**问：我可以用 Go 以外的语言编写插件吗？**
答：可以！任何支持 gRPC 的语言都可以使用。

**问：如何在不停机的情况下更新插件？**
答：使用不同名称部署新版本，更新产品 `module_name`，然后删除旧版本。

**问：插件可以访问 OpenHost 数据库吗？**
答：不可以，插件应该是无状态的。宿主管理所有状态。

**问：如果插件崩溃会发生什么？**
答：宿主检测到崩溃并将操作标记为失败。服务状态保持一致。

## 资源

- [HashiCorp go-plugin 文档](https://github.com/hashicorp/go-plugin)
- [gRPC Go 教程](https://grpc.io/docs/languages/go/)
- [Protocol Buffers 指南](https://developers.google.com/protocol-buffers)
- [OpenHost 架构](ARCHITECTURE.zh-CN.md)

## 支持

- 报告插件问题: [GitHub Issues](https://github.com/lbyxiaolizi/xlpanel/issues)
- 插件开发帮助: [讨论](https://github.com/lbyxiaolizi/xlpanel/discussions)
