# gRPC Implementation Guide

This collaborative editor backend has been updated to use gRPC for inter-service communication while maintaining backward compatibility with the REST API.

## Architecture

The system now supports both REST and gRPC:
- **HTTP/REST** - For external client communication (web/mobile apps)
- **gRPC** - For internal inter-service communication

## Services and Ports

| Service | HTTP Port | gRPC Port |
|---------|-----------|-----------|
| User Service | 8080 | 50051 |
| Document Service | 8080 | 50052 |
| Notification Service | - | 50053 |
| Realtime Service | 8080 | WebSocket |

## Protocol Buffer Setup

### 1. Install Protocol Buffer Compiler

**Windows:**
```powershell
# Using Chocolatey
choco install protoc

# Or download from https://github.com/protocolbuffers/protobuf/releases
```

**Linux/macOS:**
```bash
# Ubuntu/Debian
sudo apt install protobuf-compiler

# macOS
brew install protobuf
```

### 2. Install Go Plugins

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 3. Generate gRPC Code

**Windows:**
```powershell
.\generate_proto.ps1
```

**Linux/macOS:**
```bash
chmod +x generate_proto.sh
./generate_proto.sh
```

## Proto Files

The proto definitions are located in the `proto/` directory:
- `user.proto` - User authentication and management
- `document.proto` - Document CRUD and sharing operations
- `notification.proto` - Notification service

## Using gRPC Clients

### Example: Calling User Service from Another Service

```go
package main

import (
    "context"
    "log"
    
    "github.com/pasanAbeysekara/collaborative-editor/internal/grpcclient"
)

func main() {
    // Create a client for user-service
    userClient, err := grpcclient.NewUserServiceClient("user-service:50051")
    if err != nil {
        log.Fatal(err)
    }
    defer userClient.Close()
    
    // Get user by email
    user, err := userClient.GetUserByEmail(context.Background(), "user@example.com")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("User found: %s", user.Email)
}
```

### Example: Calling Document Service

```go
package main

import (
    "context"
    "log"
    
    "github.com/pasanAbeysekara/collaborative-editor/internal/grpcclient"
)

func main() {
    // Create a client for document-service
    docClient, err := grpcclient.NewDocumentServiceClient("document-service:50052")
    if err != nil {
        log.Fatal(err)
    }
    defer docClient.Close()
    
    // Check permission
    result, err := docClient.CheckPermission(context.Background(), "doc-123", "user-456")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Has permission: %v", result.HasPermission)
}
```

### Example: Sending Notification

```go
package main

import (
    "context"
    "log"
    
    "github.com/pasanAbeysekara/collaborative-editor/internal/grpcclient"
)

func main() {
    // Create a client for notification-service
    notifClient, err := grpcclient.NewNotificationServiceClient("notification-service:50053")
    if err != nil {
        log.Fatal(err)
    }
    defer notifClient.Close()
    
    // Send notification
    result, err := notifClient.SendNotification(
        context.Background(),
        "user-123",
        "You have been invited to collaborate",
        "document_share",
        map[string]string{"document_id": "doc-456"},
    )
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Notification sent: %s", result.Message)
}
```

## Kubernetes Deployment

The Kubernetes deployments have been updated to expose both HTTP and gRPC ports:

```yaml
# Example service configuration
apiVersion: v1
kind: Service
metadata:
  name: user-service
spec:
  selector:
    app: user-service
  ports:
  - name: http
    protocol: TCP
    port: 8080
    targetPort: 8080
  - name: grpc
    protocol: TCP
    port: 50051
    targetPort: 50051
```

## Development

### Running Services Locally

Each service now runs both HTTP and gRPC servers:

```bash
# User Service
cd cmd/user-service
go run main.go
# HTTP on :8080, gRPC on :50051

# Document Service
cd cmd/document-service
go run main.go
# HTTP on :8080, gRPC on :50052

# Notification Service
cd cmd/notification-service
go run main.go
# gRPC on :50053
```

### Testing gRPC Endpoints

You can use tools like:
- **grpcurl** - Command-line tool for gRPC
- **BloomRPC** - GUI client for gRPC
- **Postman** - Supports gRPC requests

Example with grpcurl:
```bash
# List services
grpcurl -plaintext localhost:50051 list

# Call a method
grpcurl -plaintext \
  -d '{"email": "test@example.com"}' \
  localhost:50051 proto.UserService/GetUserByEmail
```

## Benefits of gRPC

1. **Performance** - Binary protocol is faster than JSON/REST
2. **Type Safety** - Protocol buffers provide strong typing
3. **Streaming** - Support for bidirectional streaming (future enhancement)
4. **Code Generation** - Automatic client/server code generation
5. **Language Agnostic** - Can add services in other languages easily

## Migration Path

The system maintains backward compatibility:
1. External clients continue using REST APIs
2. Internal services can migrate to gRPC gradually
3. Both protocols run side-by-side

## Next Steps

1. Generate proto files: `./generate_proto.ps1` (Windows) or `./generate_proto.sh` (Linux/macOS)
2. Run `go mod tidy` to install dependencies
3. Build and deploy services
4. Update inter-service calls to use gRPC clients instead of HTTP

## Troubleshooting

### Proto generation fails
- Ensure `protoc` is installed and in PATH
- Ensure Go plugins are installed: `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`

### gRPC connection errors
- Check service is listening on correct port
- Verify firewall/network policies in Kubernetes
- Ensure service names resolve correctly (use full DNS names in K8s)

### Import errors
- Run `go mod tidy` after generating proto files
- Ensure proto package path matches go_package option in .proto files
