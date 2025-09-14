# Multi-stage build for solana-etl
# Stage 1: Build the application
FROM golang:1.19-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
#RUN go mod download

# Copy source code
COPY . .

# Build the solana-etl binary
RUN CGO_ENABLED=0 GOOS=linux go build --mod=vendor -a -installsuffix cgo -o solana-etl ./cmd/solana-etl/main.go

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S solana && \
    adduser -u 1001 -S solana -G solana

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/solana-etl .

# Copy example script for reference
COPY --from=builder /app/docker/example.sh .

# Change ownership to non-root user
RUN chown -R solana:solana /app

# Switch to non-root user
USER solana

# Expose Prometheus metrics port (default :20000)
EXPOSE 20000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:20000/metrics || exit 1

# Default command - show help
CMD ["./solana-etl", "--help"]
