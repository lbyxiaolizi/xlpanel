# Protocol Buffers

## Generate Go code

From the repository root:

```bash
protoc \
  --go_out=. \
  --go-grpc_out=. \
  pkg/proto/provisioner.proto \
  pkg/proto/payment.proto
```
