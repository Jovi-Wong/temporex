# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go files and vendor directory
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .

# Build the application using vendor
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o kelutral .

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/kelutral .

# Expose the default port
EXPOSE 8080

# Run the application
CMD ["./kelutral"]
