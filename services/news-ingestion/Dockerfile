# Multi-stage build for News Ingestion Service
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies including protoc
RUN apk add --no-cache git ca-certificates tzdata

# Install Go protoc plugins
COPY go.mod go.sum ./

# Download dependencies

RUN go mod download

# Copy proto files first
COPY proto/ proto/



# Copy source code
COPY internal/ internal/
COPY main.go .

# Copy config if exist
# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o news-ingestion-service main.go

# Create minimal production image
FROM alpine:3.19

# Install ca-certificates and curl
#RUN #apk --no-cache add ca-certificates curl tzdata

# Install runtime dependencies and health check tool
RUN apk --no-cache add ca-certificates curl tzdata && \
    wget -qO /usr/local/bin/grpc_health_probe \
    https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.24/grpc_health_probe-linux-amd64 && \
    chmod +x /usr/local/bin/grpc_health_probe

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary
COPY --from=builder --chown=appuser:appgroup /app/news-ingestion-service .

# Copy config if exists
COPY --chown=appuser:appgroup config/ ./config/

# Switch to non-root user
USER appuser:appgroup

# Expose HTTP and gRPC ports
EXPOSE 4001 4002

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:4001/health || exit 1

# Run the application
ENTRYPOINT ["./news-ingestion-service"]