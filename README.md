# Newslettar

**Automated newsletter generator for Sonarr and Radarr**

Generate beautiful, scheduled email newsletters summarizing new TV shows and movies from your Sonarr and Radarr installations.

[![License: Unlicense](https://img.shields.io/badge/license-Unlicense-blue.svg)](https://unlicense.org)
[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker)](https://hub.docker.com/r/madswell/newslettar)

## Features

- Sonarr & Radarr Integration - Automatically fetches new episodes and movies
- Trakt.tv Integration - Show trending series and movies in newsletters
- Scheduled Newsletters - Weekly automated emails at your preferred time
- Beautiful HTML Templates - Modern, responsive email design with poster images
- Web UI Configuration - Easy setup and testing through browser interface
- Lightweight - Only ~12MB RAM usage, minimal CPU
- Secure - No data collection, runs entirely on your infrastructure

## Architecture Support

| Platform | Support |
|----------|---------|
| amd64    | ✅      |
| arm64    | ✅      |
| armv6    | ✅      |

## Quick Installation (Docker Compose)

```yaml
services:
  newslettar:
    image: madswell/newslettar:latest
    container_name: newslettar
    ports:
      - 8080:8080
    environment:
      - SONARR_URL=http://YOUR-SERVER-IP:8989
      - SONARR_API_KEY=your-api-key
      - RADARR_URL=http://YOUR-SERVER-IP:7878
      - RADARR_API_KEY=your-api-key
      - SMTP_HOST=smtp.gmail.com
      - SMTP_PORT=587
      - SMTP_USER=your-email@gmail.com
      - SMTP_PASS=your-app-password
      - FROM_EMAIL=newsletter@yourdomain.com
      - TO_EMAILS=user@example.com
      - TIMEZONE=America/New_York
      - SCHEDULE_DAY=Sun
      - SCHEDULE_TIME=09:00
    restart: unless-stopped
```

After deployment, access the web UI at http://localhost:8080 to configure and test.

**Alternative:** Download standalone compose file with inline configuration:
```bash
wget https://raw.githubusercontent.com/MadSwell-dev/newslettar/main/docker-compose.simple.yml
nano docker-compose.simple.yml  # Edit your settings
docker compose -f docker-compose.simple.yml up -d
```

## Docker CLI

Recommended for Proxmox LXC and restricted Docker environments:

```bash
docker run -d \
  --name newslettar \
  --network host \
  -e SONARR_URL=http://YOUR-SERVER-IP:8989 \
  -e SONARR_API_KEY=your-api-key \
  -e RADARR_URL=http://YOUR-SERVER-IP:7878 \
  -e RADARR_API_KEY=your-api-key \
  -e SMTP_HOST=smtp.gmail.com \
  -e SMTP_PORT=587 \
  -e SMTP_USER=your-email@gmail.com \
  -e SMTP_PASS=your-app-password \
  -e FROM_EMAIL=newsletter@yourdomain.com \
  -e TO_EMAILS=user@example.com \
  -e TIMEZONE=America/New_York \
  -e SCHEDULE_DAY=Sun \
  -e SCHEDULE_TIME=09:00 \
  --restart unless-stopped \
  madswell/newslettar:latest
```

With `--network host`, access the UI at `http://<your-host-ip>:8080`.

**Alternative (standard Docker):** Replace `--network host` with `-p 8080:8080` for port mapping.

## Native Installation (Linux)

One-command installation for Debian/Ubuntu servers and Proxmox LXC containers:

```bash
curl -sSL https://raw.githubusercontent.com/MadSwell-dev/newslettar/main/install-binary.sh | sudo bash
```

The installer downloads a pre-built binary (~13MB), installs it to `/opt/newslettar`, creates a systemd service, and starts it automatically.

**Management commands:**
```bash
newslettar-ctl start      # Start service
newslettar-ctl stop       # Stop service
newslettar-ctl status     # Check status
newslettar-ctl logs       # View logs
newslettar-ctl web        # Show Web UI URL
newslettar-ctl update     # Update to latest version
```

## Security

This is a self-hosted application designed to run on your own infrastructure.

**Important considerations:**

- **Web UI Authentication**: The web UI does not include built-in authentication. For secure access:
  - Run locally only (recommended for home servers)
  - Place behind a reverse proxy with authentication (e.g., Nginx with basic auth, Authelia, Authentik)
  - Use firewall rules to restrict access to trusted networks

- **Secrets Management**: All sensitive data (API keys, SMTP passwords) are stored in the `.env` file and never committed to version control

- **Network Exposure**: If exposing to the internet, always use HTTPS and proper authentication

## Configuration

All configuration can be done through the web UI at http://localhost:8080, or via environment variables for Docker deployments.

Required settings:
- Sonarr or Radarr URL and API key
- SMTP email credentials
- Email recipients
- Schedule (day and time)

Optional:
- Trakt.tv Client ID (for trending content)
- Template customization (posters, overviews, dark mode)

Get API keys:
- Sonarr/Radarr: Settings → General → Security → API Key
- Trakt.tv: https://trakt.tv/oauth/applications
- Gmail: Use App Passwords (requires 2FA)

---

_This project was mostly vibe-coded with a lot of help from Claude. Feel free to do whatever you want with it._
