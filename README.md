# OpenHost

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

## Build Targets

```bash
make server
make emailpipe
make mock_plugin
```
