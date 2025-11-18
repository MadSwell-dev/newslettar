# Multi-stage build for Newslettar
FROM golang:1.23.5-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy source files (new structure with cmd/)
COPY cmd/ cmd/
COPY go.mod go.sum version.json ./

# Build the application with optimizations
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o newslettar ./cmd/newslettar

# Final stage - minimal Alpine runtime image
# Alpine is more compatible with restricted environments (LXC, rootless Docker)
FROM alpine:3.20

# Install only essential runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /opt/newslettar

# Copy the compiled binary from builder (templates/assets are embedded in binary)
COPY --from=builder /build/newslettar .
COPY --from=builder /build/version.json ./

# Make binary executable
RUN chmod +x newslettar

# Expose web UI port
EXPOSE 8080

# Run the application
# Configuration can be provided via:
# 1. Environment variables (recommended for Docker)
# 2. Mounted .env file at /opt/newslettar/.env (recommended for docker-compose)
CMD ["./newslettar", "-web"]
