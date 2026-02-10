# PowerShell script to generate Go code from Protocol Buffer definitions

# Install protoc plugins if not already installed
Write-Host "Installing protoc plugins..." -ForegroundColor Green
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Create output directories
Write-Host "Creating output directories..." -ForegroundColor Green
New-Item -ItemType Directory -Force -Path "proto\user" | Out-Null
New-Item -ItemType Directory -Force -Path "proto\document" | Out-Null
New-Item -ItemType Directory -Force -Path "proto\notification" | Out-Null

# Generate Go code from proto files
Write-Host "Generating gRPC code from proto files..." -ForegroundColor Green

protoc --go_out=. --go_opt=paths=source_relative `
    --go-grpc_out=. --go-grpc_opt=paths=source_relative `
    proto/user.proto

protoc --go_out=. --go_opt=paths=source_relative `
    --go-grpc_out=. --go-grpc_opt=paths=source_relative `
    proto/document.proto

protoc --go_out=. --go_opt=paths=source_relative `
    --go-grpc_out=. --go-grpc_opt=paths=source_relative `
    proto/notification.proto

Write-Host "gRPC code generation complete!" -ForegroundColor Green
