# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev protobuf-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o file-upload-server .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/file-upload-server .

# Copy static files
COPY --from=builder /build/static ./static

# Create storage directory with proper permissions
RUN mkdir -p /app/storage && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose ports
# HTTP server port
EXPOSE 8080
# TCP upload port (HTTP port + 1)
EXPOSE 8081

# Set environment variables with defaults
ENV STORAGE_DIRECTORY=/app/storage \
    SERVER_PORT=8080 \
    BASE_URL=http://localhost:8080 \
    MAX_UPLOAD_SIZE_MB=1024 \
    MAX_FILES_PER_UPLOAD=10 \
    RSA_KEY_PATH=/app/storage/private_key.pem \
    CLEANUP_INTERVAL_SECONDS=60 \
    MIN_TTL_SECONDS=60 \
    MAX_TTL_SECONDS=2592000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./file-upload-server"]
