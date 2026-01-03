# Plugin Development Guide

[English](PLUGIN_DEVELOPMENT.md) | [简体中文](PLUGIN_DEVELOPMENT.zh-CN.md)

## Overview

OpenHost uses a plugin-based architecture for provisioning modules. Plugins are separate processes that communicate with the main application via gRPC using Protocol Buffers.

## Architecture

### Plugin System

```
┌─────────────────┐         gRPC/Protobuf        ┌─────────────────┐
│                 │◄────────────────────────────►│                 │
│   OpenHost      │                              │     Plugin      │
│   (Host)        │    Handshake & RPC Calls     │  (Provisioner)  │
│                 │                              │                 │
└─────────────────┘                              └─────────────────┘
      Process 1                                       Process 2
```

**Benefits:**
- **Crash Isolation**: Plugin crashes don't affect the host
- **Language Agnostic**: Write plugins in any language with gRPC support
- **Security**: SHA-256 checksum verification
- **Versioning**: Multiple plugin versions can coexist
- **Lifecycle Management**: Enable/disable plugins dynamically
- **Event System**: Subscribe to plugin events

## Plugin Manager API

The enhanced Plugin Manager provides comprehensive lifecycle management:

### Loading and Enabling Plugins

```go
import "github.com/openhost/openhost/internal/infrastructure/plugin"

// Create plugin manager
manager := plugin.NewPluginManager("./plugins", logger)

// Enable a plugin
err := manager.EnablePlugin("my_plugin")

// Get plugin state
state := manager.GetPluginState("my_plugin")
fmt.Printf("Status: %s\n", state.Status)  // active, inactive, error, loading
```

### Plugin Events

Subscribe to plugin lifecycle events:

```go
manager.OnEvent(func(event plugin.PluginEvent) {
    switch event.Type {
    case plugin.EventPluginLoaded:
        fmt.Printf("Plugin %s loaded at %s\n", event.Plugin, event.Timestamp)
    case plugin.EventPluginUnloaded:
        fmt.Printf("Plugin %s unloaded\n", event.Plugin)
    case plugin.EventPluginError:
        fmt.Printf("Plugin %s error: %s\n", event.Plugin, event.Error)
    case plugin.EventPluginConfigured:
        fmt.Printf("Plugin %s configured\n", event.Plugin)
    case plugin.EventProvisionStart:
        fmt.Printf("Provisioning started for %s\n", event.Plugin)
    case plugin.EventProvisionComplete:
        fmt.Printf("Provisioning completed for %s\n", event.Plugin)
    case plugin.EventProvisionError:
        fmt.Printf("Provisioning failed for %s: %s\n", event.Plugin, event.Error)
    }
})
```

### Plugin Configuration

```go
// Set plugin configuration
manager.SetPluginConfig("my_plugin", map[string]string{
    "api_url": "https://api.example.com",
    "api_key": "your-secret-key",
})

// Get plugin configuration
config := manager.GetPluginConfig("my_plugin")

// Save all configurations to disk
manager.SaveConfigs("./config/plugins.json")

// Load configurations from disk
manager.LoadConfigs("./config/plugins.json")
```

### Plugin State Tracking

```go
// Get all plugin states
states := manager.GetAllStates()
for name, state := range states {
    fmt.Printf("Plugin: %s\n", name)
    fmt.Printf("  Status: %s\n", state.Status)
    fmt.Printf("  Path: %s\n", state.Path)
    if state.LoadedAt != nil {
        fmt.Printf("  Loaded: %s\n", state.LoadedAt)
    }
    fmt.Printf("  Use Count: %d\n", state.UseCount)
}

// Scan with detailed info
plugins, _ := manager.ScanWithInfo()
for _, plugin := range plugins {
    fmt.Printf("%s: %s\n", plugin.Info.Name, plugin.Status)
}
```

### Disabling Plugins

```go
// Disable a plugin
err := manager.DisablePlugin("my_plugin")
if err != nil {
    log.Printf("Failed to disable: %v", err)
}
```

## Getting Started

### Prerequisites

- Go 1.22+ (for Go plugins)
- Protocol Buffer compiler (protoc)
- go-plugin knowledge (helpful)

### Quick Start

1. Clone the repository:
```bash
git clone https://github.com/lbyxiaolizi/xlpanel.git
cd xlpanel
```

2. Examine the example plugin:
```bash
cd cmd/mock_plugin
cat main.go
```

3. Build the example:
```bash
go build -o ../../plugins/mock-provisioner .
cd ../..
sha256sum plugins/mock-provisioner > plugins/mock-provisioner.sha256
```

## Plugin Interface

### Protocol Buffer Definition

Located in `pkg/proto/provisioner/provisioner.proto`:

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

// ... other messages
```

### Core Operations

Every plugin must implement these RPC methods:

1. **Provision**: Create/provision a new service
2. **Suspend**: Suspend an active service
3. **Unsuspend**: Reactivate a suspended service
4. **Terminate**: Permanently delete a service
5. **GetInfo**: Return plugin metadata and capabilities

## Implementing a Plugin in Go

### Step 1: Project Structure

```
my-plugin/
├── main.go
├── go.mod
└── README.md
```

### Step 2: Initialize Module

```bash
mkdir my-plugin
cd my-plugin
go mod init github.com/yourusername/my-plugin
```

### Step 3: Install Dependencies

```bash
go get github.com/hashicorp/go-plugin
go get github.com/hashicorp/go-hclog
go get google.golang.org/grpc
# Copy proto files or import the openhost proto package
```

### Step 4: Implement the Plugin

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

// Implement the Provisioner interface
type MyProvisioner struct {
    logger hclog.Logger
}

func (p *MyProvisioner) Provision(ctx context.Context, 
    req *pb.ProvisionRequest) (*pb.ProvisionResponse, error) {
    
    p.logger.Info("Provisioning service", "service_id", req.ServiceId)
    
    // Your provisioning logic here
    // e.g., create VM, configure DNS, setup database, etc.
    
    return &pb.ProvisionResponse{
        Success: true,
        Message: "Service provisioned successfully",
        Metadata: map[string]string{
            "ip_address": "192.168.1.100",
            "hostname": "server1.example.com",
        },
    }, nil
}

func (p *MyProvisioner) Suspend(ctx context.Context, 
    req *pb.SuspendRequest) (*pb.SuspendResponse, error) {
    
    p.logger.Info("Suspending service", "service_id", req.ServiceId)
    
    // Your suspend logic here
    
    return &pb.SuspendResponse{
        Success: true,
        Message: "Service suspended",
    }, nil
}

func (p *MyProvisioner) Unsuspend(ctx context.Context, 
    req *pb.UnsuspendRequest) (*pb.UnsuspendResponse, error) {
    
    p.logger.Info("Unsuspending service", "service_id", req.ServiceId)
    
    // Your unsuspend logic here
    
    return &pb.UnsuspendResponse{
        Success: true,
        Message: "Service unsuspended",
    }, nil
}

func (p *MyProvisioner) Terminate(ctx context.Context, 
    req *pb.TerminateRequest) (*pb.TerminateResponse, error) {
    
    p.logger.Info("Terminating service", "service_id", req.ServiceId)
    
    // Your termination logic here
    
    return &pb.TerminateResponse{
        Success: true,
        Message: "Service terminated",
    }, nil
}

func (p *MyProvisioner) GetInfo(ctx context.Context, 
    req *pb.GetInfoRequest) (*pb.GetInfoResponse, error) {
    
    return &pb.GetInfoResponse{
        Name: "my-provisioner",
        Version: "1.0.0",
        Description: "My custom provisioner plugin",
        Capabilities: []string{"provision", "suspend", "unsuspend", "terminate"},
    }, nil
}

// gRPC server implementation
type ProvisionerGRPCServer struct {
    pb.UnimplementedProvisionerServer
    Impl *MyProvisioner
}

func (s *ProvisionerGRPCServer) Provision(ctx context.Context, 
    req *pb.ProvisionRequest) (*pb.ProvisionResponse, error) {
    return s.Impl.Provision(ctx, req)
}

func (s *ProvisionerGRPCServer) Suspend(ctx context.Context, 
    req *pb.SuspendRequest) (*pb.SuspendResponse, error) {
    return s.Impl.Suspend(ctx, req)
}

func (s *ProvisionerGRPCServer) Unsuspend(ctx context.Context, 
    req *pb.UnsuspendRequest) (*pb.UnsuspendResponse, error) {
    return s.Impl.Unsuspend(ctx, req)
}

func (s *ProvisionerGRPCServer) Terminate(ctx context.Context, 
    req *pb.TerminateRequest) (*pb.TerminateResponse, error) {
    return s.Impl.Terminate(ctx, req)
}

func (s *ProvisionerGRPCServer) GetInfo(ctx context.Context, 
    req *pb.GetInfoRequest) (*pb.GetInfoResponse, error) {
    return s.Impl.GetInfo(ctx, req)
}

// Plugin implementation
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

### Step 5: Build and Deploy

```bash
# Build the plugin
go build -o my-provisioner .

# Generate checksum
sha256sum my-provisioner > my-provisioner.sha256

# Deploy to OpenHost plugins directory
cp my-provisioner /path/to/openhost/plugins/
cp my-provisioner.sha256 /path/to/openhost/plugins/
```

## Plugin Configuration

### Register Plugin in Database

```sql
INSERT INTO products (
    name, 
    slug, 
    description,
    module_name, 
    config_options,
    active
) VALUES (
    'My Custom Service',
    'my-custom-service',
    'Description of the service',
    'my-provisioner',  -- Must match plugin binary name
    '{"api_key": "", "region": "us-east-1"}',  -- JSON config schema
    true
);
```

### Configuration Options

Plugins can define configurable options that admins fill in:

```json
{
  "api_key": {
    "type": "string",
    "label": "API Key",
    "description": "Your provider API key",
    "required": true,
    "sensitive": true
  },
  "region": {
    "type": "select",
    "label": "Region",
    "options": ["us-east-1", "us-west-2", "eu-west-1"],
    "default": "us-east-1"
  },
  "instance_type": {
    "type": "string",
    "label": "Instance Type",
    "default": "t2.micro"
  }
}
```

## Advanced Features

### Custom Parameters

Pass custom parameters from the order to the plugin:

```go
func (p *MyProvisioner) Provision(ctx context.Context, 
    req *pb.ProvisionRequest) (*pb.ProvisionResponse, error) {
    
    // Access custom parameters
    hostname := req.Parameters["hostname"]
    diskSize := req.Parameters["disk_size"]
    region := req.Parameters["region"]
    
    // Use parameters in provisioning logic
    // ...
    
    return &pb.ProvisionResponse{
        Success: true,
        Message: fmt.Sprintf("Provisioned %s in %s", hostname, region),
        Metadata: map[string]string{
            "hostname": hostname,
            "ip_address": allocatedIP,
        },
    }, nil
}
```

### Error Handling

```go
func (p *MyProvisioner) Provision(ctx context.Context, 
    req *pb.ProvisionRequest) (*pb.ProvisionResponse, error) {
    
    // Validation
    if req.ServiceId == "" {
        return &pb.ProvisionResponse{
            Success: false,
            Message: "Service ID is required",
        }, nil
    }
    
    // API call with error handling
    result, err := providerAPI.CreateInstance(req.Parameters)
    if err != nil {
        p.logger.Error("Failed to provision", "error", err)
        return &pb.ProvisionResponse{
            Success: false,
            Message: fmt.Sprintf("Provisioning failed: %v", err),
        }, nil
    }
    
    return &pb.ProvisionResponse{
        Success: true,
        Message: "Provisioned successfully",
        Metadata: result.Metadata,
    }, nil
}
```

### Context and Timeouts

```go
func (p *MyProvisioner) Provision(ctx context.Context, 
    req *pb.ProvisionRequest) (*pb.ProvisionResponse, error) {
    
    // Create timeout context
    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()
    
    // Use context in API calls
    result, err := providerAPI.CreateInstanceWithContext(ctx, req.Parameters)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return &pb.ProvisionResponse{
                Success: false,
                Message: "Provisioning timed out",
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

### Logging Best Practices

```go
// Structured logging
p.logger.Info("operation started", 
    "operation", "provision",
    "service_id", req.ServiceId,
    "user_id", req.Parameters["user_id"])

// Error logging with context
p.logger.Error("operation failed",
    "operation", "provision",
    "service_id", req.ServiceId,
    "error", err)

// Debug logging
p.logger.Debug("api request",
    "endpoint", "/v1/instances",
    "method", "POST",
    "payload", payload)
```

## Testing Plugins

### Unit Tests

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
        t.Fatalf("Provision failed: %v", err)
    }
    
    if !resp.Success {
        t.Errorf("Expected success=true, got %v: %s", resp.Success, resp.Message)
    }
    
    if resp.Metadata["hostname"] != "test.example.com" {
        t.Errorf("Expected hostname in metadata")
    }
}
```

### Integration Tests

Test your plugin with the actual OpenHost application:

```bash
# Start OpenHost in test mode
./bin/server --test-mode

# Run plugin
./plugins/my-provisioner &

# Trigger provisioning
curl -X POST http://localhost:6421/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "123",
    "parameters": {
      "hostname": "test.example.com"
    }
  }'
```

## Plugin Examples

### 1. DigitalOcean Droplet Provisioner

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
        Message: "Droplet created",
        Metadata: map[string]string{
            "droplet_id": fmt.Sprintf("%d", droplet.ID),
            "ip_address": droplet.Networks.V4[0].IPAddress,
        },
    }, nil
}
```

### 2. cPanel/WHM Account Provisioner

```go
type CpanelProvisioner struct {
    apiURL string
    apiKey string
    logger hclog.Logger
}

func (p *CpanelProvisioner) Provision(ctx context.Context, 
    req *pb.ProvisionRequest) (*pb.ProvisionResponse, error) {
    
    // Create cPanel account via WHM API
    params := url.Values{}
    params.Set("username", req.Parameters["username"])
    params.Set("domain", req.Parameters["domain"])
    params.Set("plan", req.Parameters["package"])
    
    resp, err := p.callWHMAPI("createacct", params)
    if err != nil {
        return &pb.ProvisionResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }
    
    return &pb.ProvisionResponse{
        Success: true,
        Message: "cPanel account created",
        Metadata: map[string]string{
            "username": req.Parameters["username"],
            "domain": req.Parameters["domain"],
        },
    }, nil
}
```

## Security Best Practices

1. **Validate Input**: Always validate request parameters
2. **Secure Credentials**: Store API keys in environment variables or secrets manager
3. **Timeouts**: Set reasonable timeouts for all operations
4. **Logging**: Don't log sensitive data (passwords, API keys)
5. **Error Messages**: Don't expose internal details in error messages
6. **Resource Limits**: Implement rate limiting for API calls
7. **Checksums**: Always provide SHA-256 checksums

## Debugging

### Enable Debug Logging

```go
logger := hclog.New(&hclog.LoggerOptions{
    Level:      hclog.Debug,  // or hclog.Trace for verbose
    Output:     os.Stderr,
    JSONFormat: true,
})
```

### Test Plugin Standalone

```bash
# Run plugin directly (it will wait for host connection)
PLUGIN_DEBUG=1 ./plugins/my-provisioner
```

### Attach Debugger

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o my-provisioner .

# Use delve
dlv exec ./my-provisioner
```

## Plugin Lifecycle

1. **Discovery**: Host scans `./plugins/` directory
2. **Verification**: Checks SHA-256 checksum
3. **Launch**: Starts plugin process
4. **Handshake**: Establishes gRPC connection
5. **Ready**: Plugin ready for RPC calls
6. **Shutdown**: Graceful cleanup on host termination

## FAQs

**Q: Can I write plugins in languages other than Go?**
A: Yes! Any language with gRPC support can be used.

**Q: How do I update a plugin without downtime?**
A: Deploy the new version with a different name, update the product `module_name`, then remove the old version.

**Q: Can plugins access the OpenHost database?**
A: No, plugins should be stateless. The host manages all state.

**Q: What happens if a plugin crashes?**
A: The host detects the crash and marks the operation as failed. The service state remains consistent.

## Resources

- [HashiCorp go-plugin Documentation](https://github.com/hashicorp/go-plugin)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/)
- [Protocol Buffers Guide](https://developers.google.com/protocol-buffers)
- [OpenHost Architecture](ARCHITECTURE.md)

## Support

- Report plugin issues: [GitHub Issues](https://github.com/lbyxiaolizi/xlpanel/issues)
- Plugin development help: [Discussions](https://github.com/lbyxiaolizi/xlpanel/discussions)
