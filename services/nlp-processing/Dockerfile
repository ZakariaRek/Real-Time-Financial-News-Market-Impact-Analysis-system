# Multi-stage build for NLP Processing Service
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the generated proto files (already generated locally)
COPY proto/ proto/

# Copy source code
COPY internal/ internal/
COPY main.go .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o nlp-processing-service main.go

# Create minimal production image
FROM alpine:3.19

# Install runtime dependencies and health check tool
RUN apk --no-cache add ca-certificates curl tzdata && \
    wget -qO /usr/local/bin/grpc_health_probe \
    https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.24/grpc_health_probe-linux-amd64 && \
    chmod +x /usr/local/bin/grpc_health_probe

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Create directories and set ownership
RUN mkdir -p config models/sentiment && \
    chown -R appuser:appgroup /app

# Copy the binary from builder stage
COPY --from=builder --chown=appuser:appgroup /app/nlp-processing-service .

# Copy config and models
COPY --chown=appuser:appgroup config/ ./config/
COPY --chown=appuser:appgroup models/ ./models/

# Switch to non-root user
USER appuser:appgroup

# Expose gRPC port
EXPOSE 50052

# Health check using grpc-health-probe
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD /usr/local/bin/grpc_health_probe -addr=:50052 || exit 1

# Run the application
ENTRYPOINT ["./nlp-processing-service"]