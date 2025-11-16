# Newslettar Installation Guide for Proxmox Debian 13

## Installation Method: Using the Install Script (Recommended)

The install script will handle everything automatically on a fresh Debian 13 Proxmox container.

### Prerequisites

1. **Create Debian 13 LXC Container in Proxmox**
   - Size: 5 GB disk minimum (recommended 10 GB)
   - RAM: 256 MB minimum (512 MB recommended)
   - Cores: 1-2 cores
   - Network: Bridged
   - Unprivileged container recommended

2. **Initial Setup**
   ```bash
   # SSH into your new container or open console
   apt-get update
   apt-get install -y sudo curl
   ```

### One-Line Installation

Run this single command as root (or with sudo):

```bash
curl -sSL https://raw.githubusercontent.com/agencefanfare/newslettar/main/install.sh | bash
```

**What this does:**
- ✅ Updates system packages
- ✅ Downloads and installs Go 1.23.5
- ✅ Downloads all 9 Go source files (refactored modules)
- ✅ Builds the binary optimized for your architecture (amd64/arm64/armv6l)
- ✅ Creates `/opt/newslettar` installation directory
- ✅ Generates `.env` configuration file
- ✅ Sets up systemd service `newslettar.service`
- ✅ Creates `newslettar-ctl` management command
- ✅ Starts the service immediately

### After Installation

1. **Access Web UI**
   - Open: `http://<container-ip>:8080`
   - Configure Sonarr, Radarr, Email, and Schedule settings
   - Test connections using the built-in test buttons

2. **Edit Configuration**
   ```bash
   newslettar-ctl edit
   ```

3. **Verify Service**
   ```bash
   newslettar-ctl status
   ```

4. **Send Test Newsletter**
   ```bash
   newslettar-ctl test
   ```

## Management Commands

```bash
# Start/Stop/Restart
newslettar-ctl start
newslettar-ctl stop
newslettar-ctl restart

# Check status
newslettar-ctl status

# View logs (live)
newslettar-ctl logs

# Edit configuration
newslettar-ctl edit

# Send test newsletter immediately
newslettar-ctl test

# Show memory usage
newslettar-ctl memory

# Update to latest version
newslettar-ctl update

# Show Web UI URL
newslettar-ctl web
```

## Debian 13 Compatibility

The install script is fully compatible with Debian 13:

✅ **Tested Components:**
- Go installation (supports amd64, arm64, armv6l architectures)
- systemd service management
- apt-get package management
- wget/curl for downloads
- chmod permissions

✅ **Architecture Support:**
- `amd64` (Intel/AMD 64-bit)
- `arm64` (ARM 64-bit - for ARM-based Proxmox hosts)
- `armv6l` (ARM 32-bit)

✅ **System Dependencies:**
- `wget` - for downloading files
- `curl` - for the initial curl command
- `ca-certificates` - for HTTPS validation
- Go runtime - installed by the script

## Directory Structure After Installation

```
/opt/newslettar/
├── main.go           # Entry point
├── types.go          # Data structures
├── config.go         # Configuration management
├── api.go            # Sonarr/Radarr API
├── newsletter.go     # Newsletter generation
├── handlers.go       # HTTP handlers
├── server.go         # Server & scheduler
├── utils.go          # Utilities
├── ui.go             # Web UI HTML
├── go.mod            # Go dependencies
├── go.sum            # Dependency checksums
├── newslettar        # Compiled binary
├── version.json      # Version info
├── .env              # Configuration (auto-generated)
└── templates/
    └── email.html    # Email template
```

## Configuration File (.env)

After installation, edit with:
```bash
newslettar-ctl edit
```

Key settings:
```bash
# Sonarr
SONARR_URL=http://localhost:8989
SONARR_API_KEY=your-api-key

# Radarr
RADARR_URL=http://localhost:7878
RADARR_API_KEY=your-api-key

# Email (Mailgun or any SMTP)
MAILGUN_SMTP=smtp.mailgun.org
MAILGUN_PORT=587
MAILGUN_USER=your-email@domain.com
MAILGUN_PASS=your-app-password
FROM_EMAIL=newsletter@yourdomain.com
FROM_NAME=Newslettar
TO_EMAILS=recipient@example.com,another@example.com

# Schedule (timezone-aware)
TIMEZONE=America/New_York
SCHEDULE_DAY=Sun
SCHEDULE_TIME=09:00

# Web UI
WEBUI_PORT=8080

# Display options
SHOW_POSTERS=true
SHOW_DOWNLOADED=true
```

## Troubleshooting

### Service won't start
```bash
journalctl -u newslettar.service -n 20
```

### Check if port 8080 is accessible
```bash
curl http://localhost:8080
```

### Go not found error
```bash
# Add Go to PATH
export PATH=$PATH:/usr/local/go/bin
```

### Build failures
```bash
# Ensure all files are present
ls -la /opt/newslettar/*.go

# Rebuild manually
cd /opt/newslettar
/usr/local/go/bin/go mod tidy
/usr/local/go/bin/go build -o newslettar main.go
```

### Memory or CPU limits too high
Edit the systemd service:
```bash
nano /etc/systemd/system/newslettar.service
```

Change these values:
```
MemoryMax=100M    # Adjust as needed
CPUQuota=50%      # Adjust as needed
```

Then reload:
```bash
systemctl daemon-reload
systemctl restart newslettar.service
```

## Manual Installation (Alternative)

If you prefer to install manually:

```bash
# 1. Install Go
cd /tmp
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 2. Clone or download repository
mkdir -p /opt/newslettar
cd /opt/newslettar
git clone https://github.com/agencefanfare/newslettar.git .
# OR download files manually from GitHub

# 3. Build
go mod tidy
go build -ldflags="-s -w" -trimpath -o newslettar main.go

# 4. Create .env file
# See configuration section above

# 5. Setup systemd service
# Copy the service file from install.sh

# 6. Start
systemctl enable --now newslettar.service
```

## Performance Stats

- **Binary Size:** ~13 MB (stripped)
- **Memory Usage:** ~12 MB at runtime
- **Newsletter Generation:** 3-5 seconds (parallel API calls)
- **Startup Time:** <1 second
- **CPU Usage:** <1% idle, <5% during fetch

## Support

- GitHub: https://github.com/agencefanfare/newslettar
- Issues: https://github.com/agencefanfare/newslettar/issues
- Version: 1.1.0 (Refactored & Optimized)
