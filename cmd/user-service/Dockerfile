FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/server ./cmd/user-service

FROM alpine:latest
COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]