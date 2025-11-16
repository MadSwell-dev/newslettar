# Docker Installation Guide

Docker provides the easiest and most reliable way to deploy Newslettar. It handles all dependencies automatically and ensures consistency across different systems.

## Quick Start (2 methods)

### Method 1: Using Setup Script (Recommended)

```bash
git clone https://github.com/agencefanfare/newslettar.git
cd newslettar
bash docker-setup.sh
nano data/.env  # Edit with your settings
```

### Method 2: Using docker-compose

```bash
git clone https://github.com/agencefanfare/newslettar.git
cd newslettar
mkdir -p data
cp .env.example data/.env
nano data/.env  # Edit with your settings
docker-compose up -d
```

Then open `http://localhost:8080` in your browser.

## Requirements

- Docker: [Install Docker](https://docs.docker.com/get-docker/)
- For docker-compose method: Docker Compose 1.29+ (usually included with Docker)
- ~300MB disk space for the image
- Port 8080 available (or change it in docker-compose.yml)

## Configuration

### Using Setup Script (Easiest - Works everywhere)

1. **Clone and run setup:**
   ```bash
   git clone https://github.com/agencefanfare/newslettar.git
   cd newslettar
   bash docker-setup.sh
   ```

2. **Script will prompt you to edit configuration:**
   ```bash
   nano data/.env
   ```

3. **Key settings to configure:**
   - `SONARR_URL` and `SONARR_API_KEY` - Your Sonarr instance
   - `RADARR_URL` and `RADARR_API_KEY` - Your Radarr instance
   - `MAILGUN_SMTP`, `MAILGUN_USER`, `MAILGUN_PASS` - Email configuration
   - `FROM_EMAIL` and `TO_EMAILS` - Email addresses
   - `TIMEZONE` - Your timezone (e.g., `America/New_York`, `Europe/London`)
   - `SCHEDULE_DAY` and `SCHEDULE_TIME` - When to send newsletters

4. **Start container:**
   ```bash
   docker start newslettar
   # or run setup script again
   ```

5. **Access the web UI:**
   ```bash
   # Open: http://localhost:8080
   ```

### Using docker-compose (If installed)

1. **Clone the repository:**
   ```bash
   git clone https://github.com/agencefanfare/newslettar.git
   cd newslettar
   ```

2. **Create data directory and configuration:**
   ```bash
   mkdir -p data
   cp .env.example data/.env
   nano data/.env
   ```

3. **Start the containers:**
   ```bash
   docker-compose up -d
   ```

4. **Access the web UI:**
   ```bash
   # Open: http://localhost:8080
   ```

### Manual Docker commands (if not using scripts)

```bash
# Build the image
docker build -t newslettar:latest .

# Run the container
docker run -d \
  --name newslettar \
  -p 8080:8080 \
  -v $(pwd)/data/.env:/opt/newslettar/.env \
  --restart unless-stopped \
  newslettar:latest
```

## Common Tasks

### View logs
```bash
docker-compose logs -f newslettar
```

### Restart the service
```bash
docker-compose restart
```

### Stop the service
```bash
docker-compose down
```

### Update to latest version
```bash
# Pull latest changes
git pull

# Rebuild the image (will use latest source code)
docker-compose build --no-cache

# Restart with new image
docker-compose up -d
```

### Edit configuration while running
```bash
nano data/.env
docker-compose restart
```

### Access container shell
```bash
docker-compose exec newslettar /bin/bash
```

## Environment Variables

All configuration is done through the `.env` file. Here are the available settings:

```env
# Sonarr - Your TV show management
SONARR_URL=http://sonarr:8989
SONARR_API_KEY=your_api_key_here

# Radarr - Your movie management
RADARR_URL=http://radarr:7878
RADARR_API_KEY=your_api_key_here

# Email - How to send newsletters
MAILGUN_SMTP=smtp.mailgun.org
MAILGUN_PORT=587
MAILGUN_USER=your_mailgun_username
MAILGUN_PASS=your_mailgun_password
FROM_NAME=Newslettar
FROM_EMAIL=newsletter@yourdomain.com
TO_EMAILS=user@example.com

# Schedule - When to send newsletters
TIMEZONE=UTC
SCHEDULE_DAY=Sun
SCHEDULE_TIME=09:00

# Display - What to show in newsletters
SHOW_POSTERS=true
SHOW_DOWNLOADED=true

# Web UI
WEBUI_PORT=8080
```

## Networking

### Accessing local services (Sonarr, Radarr) from Docker

If Sonarr/Radarr are running on your host machine:

**On Linux:**
```env
SONARR_URL=http://172.17.0.1:8989
RADARR_URL=http://172.17.0.1:7878
```

**On Mac/Windows (Docker Desktop):**
```env
SONARR_URL=http://host.docker.internal:8989
RADARR_URL=http://host.docker.internal:7878
```

### Using docker-compose with other services

Create a `docker-compose.override.yml`:
```yaml
version: '3.8'
services:
  newslettar:
    environment:
      - SONARR_URL=http://sonarr:8989
      - RADARR_URL=http://radarr:7878
```

If Sonarr and Radarr are also in Docker Compose, they'll be on the same network automatically.

## Troubleshooting

### Container exits immediately
```bash
docker-compose logs newslettar
```
Check for configuration errors in the logs.

### Can't connect to Sonarr/Radarr
- Verify URLs are correct (use container names if in same docker-compose)
- Check API keys are valid
- Ensure services are accessible from the container

### Permission errors
The container runs as root. If you have permission issues:
```bash
sudo chown -R $(id -u):$(id -g) data/
```

### Out of memory
The container has a default 512MB memory limit. To increase:
```yaml
# In docker-compose.yml
services:
  newslettar:
    mem_limit: 1g
```

### Port already in use
Change the port mapping in `docker-compose.yml`:
```yaml
ports:
  - "8888:8080"  # Access on http://localhost:8888
```

## Proxmox LXC Container

To run Dockerized Newslettar in a Proxmox LXC container:

1. **Create a Debian 13 container** in Proxmox with at least:
   - 2 CPU cores
   - 1GB RAM
   - 5GB storage

2. **Install Docker inside the container:**
   ```bash
   # Inside the container
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh
   ```

3. **Follow the Quick Start above**

4. **Access from host:**
   ```bash
   # Find container IP
   hostname -I
   
   # Open http://<container-ip>:8080 in browser
   ```

## Performance

The Docker image is optimized for:
- **Size**: ~13MB stripped binary, ~30MB total image
- **Memory**: ~12-15MB base usage
- **Startup**: <2 seconds
- **Building**: Cached layers for fast rebuilds

## Security Notes

- The container runs as root (for simplicity)
- Use a firewall to limit access to port 8080
- Keep your Sonarr/Radarr API keys secure
- Consider using a reverse proxy (nginx, caddy) for HTTPS

## Support

For issues or questions:
- Check logs: `docker-compose logs -f`
- Review `.env` file for correct configuration
- Verify Sonarr/Radarr are accessible and API keys are correct
