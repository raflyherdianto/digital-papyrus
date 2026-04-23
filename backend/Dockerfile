# ============================================================
# Digital Papyrus API — Multi-Stage Production Dockerfile
# ============================================================

# Stage 1: Build
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -trimpath \
    -o /build/server \
    ./cmd/server

# Stage 2: Production
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata su-exec \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /build/server /app/server
COPY entrypoint.sh /app/entrypoint.sh

# Create data directory for SQLite with correct permissions
RUN mkdir -p /app/data && \
    chmod +x /app/entrypoint.sh && \
    chown -R appuser:appgroup /app

# Removed USER appuser so the container can start as root
# and drop privileges in entrypoint.sh

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --spider http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/server"]
