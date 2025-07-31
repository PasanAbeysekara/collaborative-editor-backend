FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# -o /app/server builds the binary and places it in the /app directory
# CGO_ENABLED=0 is important for creating a static binary
# -ldflags="-w -s" strips debug information to make the binary smaller
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/server ./cmd/api

FROM alpine:latest

COPY --from=builder /app/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]