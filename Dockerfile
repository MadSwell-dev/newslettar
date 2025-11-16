# Multi-stage build for Newslettar
FROM golang:1.23.5-bookworm AS builder

WORKDIR /build

# Copy all source files explicitly (glob expansion doesn't work reliably)
COPY main.go types.go config.go api.go newsletter.go handlers.go server.go utils.go ui.go ./
COPY go.mod go.sum ./
COPY templates/ templates/

# Build the application with optimizations
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o newslettar .

# Final stage - minimal runtime image
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/newslettar

# Copy the compiled binary from builder
COPY --from=builder /build/newslettar .
COPY --from=builder /build/templates/ templates/
COPY --from=builder /build/go.mod go.sum version.json ./

# Create default .env file
RUN echo "# Sonarr Configuration\n\
SONARR_URL=http://localhost:8989\n\
SONARR_API_KEY=\n\
\n\
# Radarr Configuration\n\
RADARR_URL=http://localhost:7878\n\
RADARR_API_KEY=\n\
\n\
# Email Configuration\n\
MAILGUN_SMTP=smtp.mailgun.org\n\
MAILGUN_PORT=587\n\
MAILGUN_USER=\n\
MAILGUN_PASS=\n\
FROM_NAME=Newslettar\n\
FROM_EMAIL=newsletter@yourdomain.com\n\
TO_EMAILS=user@example.com\n\
\n\
# Schedule Settings\n\
TIMEZONE=UTC\n\
SCHEDULE_DAY=Sun\n\
SCHEDULE_TIME=09:00\n\
\n\
# Template Settings\n\
SHOW_POSTERS=true\n\
SHOW_DOWNLOADED=true\n\
\n\
# Web UI Port\n\
WEBUI_PORT=8080" > .env.example

# Make binary executable
RUN chmod +x newslettar

# Expose web UI port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["./newslettar", "-web"]
