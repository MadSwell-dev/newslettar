# 📧 Newslettar

**Automated newsletter generator for Sonarr and Radarr**

Newslettar automatically generates beautiful, scheduled email newsletters summarizing new TV shows and movies from your Sonarr and Radarr installations. Keep your family, friends, or community informed about what's new on your media server!

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker)](https://hub.docker.com/r/agencefanfare/newslettar)

## ✨ Features

- **📺 Sonarr & Radarr Integration** - Automatically fetches new episodes and movies
- **🎬 Trakt.tv Integration** - Show trending series and movies in your newsletters
- **📅 Scheduled Newsletters** - Weekly automated emails at your preferred time
- **🎨 Beautiful HTML Templates** - Modern, responsive email design with poster images
- **⚙️ Web UI Configuration** - Easy setup and testing through browser interface
- **🚀 Lightweight** - Only ~12MB RAM usage, minimal CPU
- **🐳 Docker Ready** - Deploy with a single command
- **📦 Debian Package** - Native `.deb` installation available
- **🔒 Secure** - No data collection, runs entirely on your infrastructure
- **🌐 Timezone Aware** - Schedule in your local timezone

## 🚀 Quick Start

### Docker (Recommended)

The fastest way to get started:

```bash
# Clone the repository
git clone https://github.com/agencefanfare/newslettar.git
cd newslettar

# Run the setup script
bash docker-setup.sh

# Edit configuration
nano data/.env

# Start the container (or re-run setup script)
docker start newslettar
```

Access the web UI at `http://localhost:8080`

### Debian/Ubuntu Package

```bash
# Download the latest .deb package
wget https://github.com/agencefanfare/newslettar/releases/latest/download/newslettar_*_amd64.deb

# Install
sudo dpkg -i newslettar_*_amd64.deb

# Configure
newslettar-ctl edit

# Start
newslettar-ctl start
```

### One-Line Install (Debian/Ubuntu/Proxmox LXC)

```bash
curl -sSL https://raw.githubusercontent.com/agencefanfare/newslettar/main/install.sh | bash
```

## 📋 Requirements

### For Docker
- Docker installed
- 300MB disk space
- Port 8080 available (or configure different port)

### For Native Installation
- Debian 11+, Ubuntu 20.04+, or similar Linux distribution
- Go 1.23+ (automatically installed by install script)
- systemd (for service management)

### General Requirements
- Access to Sonarr and Radarr instances
- SMTP server for sending emails (Gmail, Mailgun, SendGrid, etc.)
- (Optional) Trakt.tv API key for trending content

## 📖 Installation Methods

### Method 1: Docker Compose

1. **Clone and prepare:**
   ```bash
   git clone https://github.com/agencefanfare/newslettar.git
   cd newslettar
   mkdir -p data
   cp .env.example data/.env
   ```

2. **Edit configuration:**
   ```bash
   nano data/.env
   ```

3. **Start:**
   ```bash
   docker-compose up -d
   ```

4. **Access:** `http://localhost:8080`

### Method 2: Debian Package (Advanced)

Build and install your own package:

```bash
# Build the package
make deb

# Install
sudo dpkg -i dist/newslettar_*.deb

# Configure
newslettar-ctl edit

# Start
newslettar-ctl start
```

### Method 3: Manual Build

```bash
# Install Go 1.23+
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone and build
git clone https://github.com/agencefanfare/newslettar.git
cd newslettar
make build

# Run
./build/newslettar -web
```

## ⚙️ Configuration

Configuration is done through a `.env` file. Copy `.env.example` to `.env` and customize:

```bash
# Sonarr Configuration
SONARR_URL=http://localhost:8989
SONARR_API_KEY=your_api_key_here

# Radarr Configuration
RADARR_URL=http://localhost:7878
RADARR_API_KEY=your_api_key_here

# Trakt Configuration (Optional - enables trending series and movies)
# Get your Client ID from https://trakt.tv/oauth/applications
TRAKT_CLIENT_ID=your_client_id_here

# Email Configuration (Works with any SMTP provider: Gmail, Mailgun, SendGrid, etc.)
SMTP_HOST=smtp.mailgun.org
SMTP_PORT=587
SMTP_USER=your_email@domain.com
SMTP_PASS=your_password
FROM_NAME=Newslettar
FROM_EMAIL=newsletter@yourdomain.com
TO_EMAILS=user1@example.com,user2@example.com

# Schedule Settings (Internal Cron - No systemd timer needed!)
TIMEZONE=America/New_York
SCHEDULE_DAY=Sun
SCHEDULE_TIME=09:00

# Template Settings
SHOW_POSTERS=true                    # Show poster images in emails
SHOW_DOWNLOADED=true                 # Include already downloaded content
SHOW_QUALITY_PROFILES=false          # Show quality profile information
SHOW_UNMONITORED=false               # Include unmonitored content
SHOW_SERIES_OVERVIEW=false           # Show series descriptions
SHOW_EPISODE_OVERVIEW=false          # Show episode summaries
DARK_MODE=true                       # Use dark mode theme

# Trakt Features (Requires TRAKT_CLIENT_ID)
SHOW_TRAKT_ANTICIPATED_SERIES=false  # Show anticipated TV series
SHOW_TRAKT_WATCHED_SERIES=false      # Show most watched TV series
SHOW_TRAKT_ANTICIPATED_MOVIES=false  # Show anticipated movies
SHOW_TRAKT_WATCHED_MOVIES=false      # Show most watched movies

# Web UI Port
WEBUI_PORT=8080

# Performance Tuning (Optional - defaults are fine for most users)
# API_PAGE_SIZE=1000                 # Items per API page
# MAX_RETRIES=3                      # API retry attempts
# PREVIEW_RETRIES=2                  # Preview generation retries
# API_TIMEOUT=30                     # API timeout in seconds
```

### Getting API Keys

**Sonarr/Radarr:**
1. Open Sonarr/Radarr web interface
2. Settings → General → Security → API Key
3. Copy the key to your `.env` file

**Trakt.tv (Optional):**
1. Create account at https://trakt.tv
2. Go to https://trakt.tv/oauth/applications
3. Create new application
4. Copy Client ID to your `.env` file

**SMTP (Email):**
- **Gmail:** Use App Passwords (requires 2FA enabled)
- **Mailgun:** Free tier available, get SMTP credentials from dashboard
- **SendGrid:** Free tier available, create API key for SMTP
- **Any SMTP server:** Just need host, port, username, password

## 🎯 Usage

### Web UI

Access the web interface at `http://localhost:8080` (or your configured port).

Features:
- **Dashboard** - View and test configuration
- **Test Buttons** - Test Sonarr, Radarr, and email connections
- **Preview** - See what your newsletter will look like
- **Send Test** - Send a test newsletter immediately
- **Configuration** - Edit settings through the UI

### Command Line (Native Install)

```bash
# Start/Stop/Restart service
newslettar-ctl start
newslettar-ctl stop
newslettar-ctl restart

# View logs (live)
newslettar-ctl logs

# Edit configuration
newslettar-ctl edit

# Send test newsletter now
newslettar-ctl test

# Check status
newslettar-ctl status

# Show Web UI URL
newslettar-ctl web

# Check memory usage
newslettar-ctl memory

# Show version
newslettar-ctl version
```

### Docker Commands

```bash
# View logs
docker-compose logs -f

# Restart
docker-compose restart

# Edit config
nano data/.env
docker-compose restart

# Send test newsletter
docker-compose exec newslettar ./newslettar

# Stop
docker-compose down

# Update to latest
git pull
docker-compose build --no-cache
docker-compose up -d
```

## 🔨 Development

### Prerequisites

- Go 1.23+
- Make (optional, for using Makefile)

### Building from Source

```bash
# Clone repository
git clone https://github.com/agencefanfare/newslettar.git
cd newslettar

# Build
make build

# Run
make run

# Build for all platforms
make build-all

# Run tests
make test

# Format code
make fmt
```

### Project Structure

```
newslettar/
├── main.go              # Entry point
├── types.go             # Data structures
├── config.go            # Configuration management
├── constants.go         # Application constants
├── api.go               # Sonarr/Radarr API client
├── trakt.go             # Trakt.tv API client
├── newsletter.go        # Newsletter generation logic
├── handlers.go          # HTTP handlers for Web UI
├── server.go            # HTTP server & scheduler
├── ui.go                # Web UI HTML templates
├── utils.go             # Utility functions
├── templates/
│   └── email.html       # Email HTML template
├── assets/
│   ├── newslettar_logo.svg      # Logo
│   ├── newslettar_black.svg     # Logo (black)
│   └── newslettar_white.svg     # Logo (white)
├── scripts/
│   └── build-deb.sh     # Debian package builder
├── Dockerfile           # Docker image definition
├── docker-compose.yml   # Docker Compose configuration
├── Makefile             # Build automation
└── README.md            # This file
```

### Development Workflow

1. **Make changes** to `.go` files
2. **Test locally:** `make run`
3. **Format code:** `make fmt`
4. **Run tests:** `make test`
5. **Build for production:** `make build`

### Adding Features

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Test thoroughly
5. Commit: `git commit -m 'Add amazing feature'`
6. Push: `git push origin feature/amazing-feature`
7. Open a Pull Request

## 🐳 Docker Details

### Building Docker Image

```bash
# Build locally
make docker-build

# Or manually
docker build -t newslettar:latest .
```

### Custom Docker Compose

Create a `docker-compose.override.yml`:

```yaml
version: '3.8'
services:
  newslettar:
    ports:
      - "8888:8080"  # Custom port
    environment:
      - WEBUI_PORT=8080
    mem_limit: 256m  # Custom memory limit
```

### Accessing Host Services from Docker

**On Linux:**
```bash
SONARR_URL=http://172.17.0.1:8989
RADARR_URL=http://172.17.0.1:7878
```

**On Mac/Windows:**
```bash
SONARR_URL=http://host.docker.internal:8989
RADARR_URL=http://host.docker.internal:7878
```

## 📦 Debian Package Details

### Building

```bash
make deb
```

This creates: `dist/newslettar_<version>_<arch>.deb`

### What Gets Installed

- Binary: `/opt/newslettar/newslettar`
- Templates: `/opt/newslettar/templates/`
- Assets: `/opt/newslettar/assets/`
- Config: `/opt/newslettar/.env`
- Service: `/etc/systemd/system/newslettar.service`
- Control: `/usr/local/bin/newslettar-ctl`

### Uninstalling

```bash
sudo dpkg -r newslettar
```

Configuration files remain in `/opt/newslettar/.env` for backup.

## 🔧 Troubleshooting

### Service Won't Start

```bash
# Check logs
journalctl -u newslettar.service -n 50

# Or for Docker
docker-compose logs newslettar
```

### Can't Connect to Sonarr/Radarr

1. Verify URLs are accessible from the server
2. Check API keys are correct
3. Ensure no firewall blocking
4. For Docker: use correct host address (see Docker Details above)

### Email Not Sending

1. Test SMTP credentials manually
2. Check spam folder
3. Verify FROM_EMAIL is allowed by your SMTP provider
4. Check logs for specific error messages

### Port 8080 Already in Use

Change the port:
```bash
# In .env
WEBUI_PORT=8888

# For Docker, also update docker-compose.yml
ports:
  - "8888:8080"
```

### Memory Issues

The application uses ~12MB normally. If you see high memory:
1. Check for configuration loops (schedule set to run too frequently)
2. Reduce `API_PAGE_SIZE` in `.env`
3. Disable poster images: `SHOW_POSTERS=false`

## 📊 Performance

- **Binary Size:** ~13 MB (stripped)
- **Memory Usage:** ~12 MB at runtime
- **Newsletter Generation:** 3-5 seconds with parallel API calls
- **Startup Time:** <1 second
- **CPU Usage:** <1% idle, <5% during newsletter generation

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Report Bugs** - Open an issue with details and reproduction steps
2. **Suggest Features** - Describe your idea in an issue
3. **Submit Pull Requests** - Fix bugs or add features
4. **Improve Documentation** - Help make docs clearer
5. **Share** - Star the repo and tell others!

### Development Guidelines

- Follow existing code style
- Add tests for new features
- Update documentation
- Keep commits atomic and well-described
- Ensure `make test` passes before submitting PR

## 📄 License

This project is licensed under the MIT License - see below for details:

```
MIT License

Copyright (c) 2025 Agency Fanfare

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## 🙏 Acknowledgments

- **Sonarr & Radarr** - For their excellent APIs
- **Trakt.tv** - For trending data
- **Go Community** - For amazing tools and libraries
- **Contributors** - Everyone who has helped improve this project

## 📞 Support

- **GitHub Issues:** https://github.com/agencefanfare/newslettar/issues
- **Discussions:** https://github.com/agencefanfare/newslettar/discussions
- **Email:** hello@agencefanfare.com

## 🗺️ Roadmap

- [ ] Multi-language support
- [ ] Custom email templates
- [ ] Plex integration
- [ ] Jellyfin integration
- [ ] Multiple newsletter recipients with different preferences
- [ ] Web-based template editor
- [ ] Statistics and analytics
- [ ] Discord/Slack notifications
- [ ] Docker Hub automated builds

---

**Made with ❤️ by [Agency Fanfare](https://agencefanfare.com)**

If you find this useful, please ⭐ star the repository!
