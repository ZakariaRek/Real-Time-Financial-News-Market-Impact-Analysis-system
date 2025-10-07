# Multi-stage build for News Ingestion Service
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies for building
RUN apk add --no-cache git ca-certificates tzdata

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY internal/ internal/
COPY proto/ proto/
COPY main.go .

# Copy config directory if it exists
COPY config* config/ 2>/dev/null || true

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o news-ingestion-service main.go

# Create minimal production image
FROM alpine:3.19

# Install ca-certificates and curl for health checks
RUN apk --no-cache add ca-certificates curl tzdata

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder --chown=appuser:appgroup /app/news-ingestion-service .

# Copy config directory if exists
COPY --from=builder --chown=appuser:appgroup /app/config* ./config/ 2>/dev/null || true

# Switch to non-root user
USER appuser:appgroup

# Expose HTTP and gRPC ports
EXPOSE 4001 4002

# Health check (HTTP endpoint)
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:4001/health || exit 1

# Run the application
ENTRYPOINT ["./news-ingestion-service"]