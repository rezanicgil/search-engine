# Dockerfile for Go backend API
# Multi-stage build for optimized production image

# Stage 1: Build stage
# This stage compiles the Go application
FROM golang:1.24-alpine AS builder

# Update package index and upgrade all packages to latest versions with security patches
# Install build dependencies (git for go modules, ca-certificates for HTTPS)
RUN apk update && \
    apk upgrade --no-cache && \
    apk add --no-cache git ca-certificates && \
    rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /build

# Copy go mod files first (for better layer caching)
COPY go.mod go.sum ./

# Download dependencies (cached layer if go.mod doesn't change)
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a statically linked binary (no C dependencies)
# -ldflags="-s -w" strips debug info to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /build/bin/api \
    ./cmd/api

# Stage 2: Runtime stage
# This stage creates a minimal production image
FROM alpine:3.20

# Update package index, upgrade packages, and install ca-certificates
# This ensures we have the latest security patches
RUN apk update && \
    apk upgrade --no-cache && \
    apk add --no-cache ca-certificates wget && \
    rm -rf /var/cache/apk/*

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/bin/api /app/api

# Copy migrations directory (needed at runtime)
COPY --from=builder /build/migrations /app/migrations

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port (default 8080, can be overridden via env)
EXPOSE 8080

# Health check endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${SERVER_PORT:-8080}/health || exit 1

# Run the application
CMD ["./api"]
