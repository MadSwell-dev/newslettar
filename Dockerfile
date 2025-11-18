# Multi-stage build for Newslettar
FROM golang:1.23.5-bookworm AS builder

WORKDIR /build

# Copy source files (new structure with cmd/)
COPY cmd/ cmd/
COPY go.mod go.sum version.json ./

# Build the application with optimizations
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o newslettar ./cmd/newslettar

# Final stage - minimal runtime image
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/newslettar

# Copy the compiled binary from builder (templates/assets are embedded in binary)
COPY --from=builder /build/newslettar .
COPY --from=builder /build/version.json ./

# Make binary executable
RUN chmod +x newslettar

# Expose web UI port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
# Configuration can be provided via:
# 1. Environment variables (recommended for Docker)
# 2. Mounted .env file at /opt/newslettar/.env (recommended for docker-compose)
CMD ["./newslettar", "-web"]
