#!/bin/bash

# Script to generate Go code from Protocol Buffer definitions

# Install protoc plugins if not already installed
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Create output directories
mkdir -p proto/user
mkdir -p proto/document
mkdir -p proto/notification

# Generate Go code from proto files
echo "Generating gRPC code from proto files..."

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/user.proto

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/document.proto

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/notification.proto

echo "gRPC code generation complete!"
