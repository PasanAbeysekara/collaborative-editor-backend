# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# -o /app/server builds the binary and places it in the /app directory
# CGO_ENABLED=0 is important for creating a static binary
# -ldflags="-w -s" strips debug information to make the binary smaller
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/server ./cmd/api

# Stage 2: Create the final, lightweight image
FROM alpine:latest

# Copy the built binary from the builder stage
COPY --from=builder /app/server /server

# Expose the port the app runs on
EXPOSE 8081

# Set the entrypoint to run the binary
ENTRYPOINT ["/server"]