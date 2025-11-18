#!/bin/bash

# =============================================================================
# Newslettar Binary Installer v2.0.0
# =============================================================================
# Perfect for: Proxmox LXC containers, Debian/Ubuntu servers
# Downloads pre-built binary (~13MB) instead of compiling from source
#
# INSTALLATION:
#   curl -sSL https://raw.githubusercontent.com/MadSwell-dev/newslettar/main/install-binary.sh | sudo bash
#
# After install, configure via web UI at: http://your-server-ip:8080

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║    Newslettar Installer v2.0.0         ║${NC}"
echo -e "${GREEN}║    Pre-built Binary Installation       ║${NC}"
echo -e "${GREEN}║    For Debian/Ubuntu/Proxmox LXC       ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root (use sudo or run as root)${NC}"
    exit 1
fi

# Check if Debian-based
if [ ! -f /etc/debian_version ]; then
    echo -e "${RED}This script is designed for Debian-based systems${NC}"
    exit 1
fi

DEBIAN_VERSION=$(cat /etc/debian_version | cut -d'.' -f1)
echo -e "${BLUE}Detected Debian version: ${DEBIAN_VERSION}${NC}"
echo ""

INSTALL_DIR="/opt/newslettar"
GITHUB_REPO="MadSwell-dev/newslettar"

# Detect architecture
ARCH=$(dpkg --print-architecture)
case $ARCH in
    amd64) BINARY_ARCH="amd64" ;;
    arm64) BINARY_ARCH="arm64" ;;
    armhf) BINARY_ARCH="armv6" ;;
    *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

echo -e "${YELLOW}[1/5] Installing required dependencies...${NC}"
apt-get update -qq
apt-get install -y wget curl ca-certificates tar >/dev/null 2>&1
echo -e "${GREEN}✓ Dependencies installed${NC}"

echo -e "${YELLOW}[2/5] Downloading latest release...${NC}"

# Get latest release version
echo -e "${BLUE}  Fetching latest release info...${NC}"
LATEST_VERSION=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo -e "${RED}Could not fetch latest release version${NC}"
    echo -e "${YELLOW}No releases with binaries found yet.${NC}"
    echo ""
    echo -e "${YELLOW}To use this installer, create a release first:${NC}"
    echo -e "${BLUE}  git tag v0.8.0${NC}"
    echo -e "${BLUE}  git push origin v0.8.0${NC}"
    echo ""
    echo -e "${YELLOW}Or use the source installation method:${NC}"
    echo -e "${BLUE}  curl -sSL https://raw.githubusercontent.com/${GITHUB_REPO}/main/install.sh | sudo bash${NC}"
    exit 1
fi

echo -e "${BLUE}  Latest version: ${LATEST_VERSION}${NC}"

# Construct download URL
BINARY_NAME="newslettar_${LATEST_VERSION}_linux_${BINARY_ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${LATEST_VERSION}/${BINARY_NAME}"

echo -e "${BLUE}  Downloading ${BINARY_NAME}...${NC}"
wget -q --show-progress "$DOWNLOAD_URL" -O /tmp/newslettar.tar.gz || {
    echo -e "${RED}Failed to download binary${NC}"
    echo -e "${YELLOW}Tried: ${DOWNLOAD_URL}${NC}"
    echo -e "${YELLOW}Please check if this release exists or use source installation.${NC}"
    exit 1
}

echo -e "${GREEN}✓ Downloaded ${BINARY_NAME}${NC}"

echo -e "${YELLOW}[3/5] Installing Newslettar...${NC}"

# Remove old installation if exists
if [ -d "$INSTALL_DIR" ]; then
    echo -e "${BLUE}  Removing old installation...${NC}"
    systemctl stop newslettar.service 2>/dev/null || true
    rm -rf "$INSTALL_DIR"
fi

# Create installation directory
mkdir -p "$INSTALL_DIR"

# Extract binary and assets
cd /tmp
tar -xzf newslettar.tar.gz -C "$INSTALL_DIR"
rm newslettar.tar.gz

# Make binary executable
chmod +x "$INSTALL_DIR/newslettar"

BINARY_SIZE=$(du -h "$INSTALL_DIR/newslettar" | cut -f1)
echo -e "${GREEN}✓ Installed to $INSTALL_DIR (${BINARY_SIZE})${NC}"

echo -e "${YELLOW}[4/5] Creating configuration...${NC}"

# Only create .env if it doesn't exist (preserve existing config on updates)
if [ ! -f "$INSTALL_DIR/.env" ]; then
    cat > "$INSTALL_DIR/.env" << 'EOF'
# Sonarr Configuration
SONARR_URL=http://localhost:8989
SONARR_API_KEY=

# Radarr Configuration
RADARR_URL=http://localhost:7878
RADARR_API_KEY=

# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
FROM_NAME=Newslettar
FROM_EMAIL=newsletter@yourdomain.com
TO_EMAILS=user@example.com

# Schedule Settings
TIMEZONE=UTC
SCHEDULE_DAY=Sun
SCHEDULE_TIME=09:00

# Template Settings
SHOW_POSTERS=true
SHOW_DOWNLOADED=true

# Web UI Port
WEBUI_PORT=8080
EOF
    echo -e "${GREEN}✓ Configuration file created${NC}"
else
    echo -e "${GREEN}✓ Existing configuration preserved${NC}"
fi

echo -e "${YELLOW}[5/5] Setting up systemd service...${NC}"

cat > /etc/systemd/system/newslettar.service << 'SVCEOF'
[Unit]
Description=Newslettar Web UI with Internal Scheduler
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/newslettar
EnvironmentFile=/opt/newslettar/.env
ExecStart=/opt/newslettar/newslettar -web
Restart=always
RestartSec=10

# Optimize resource usage
MemoryMax=100M
CPUQuota=50%

[Install]
WantedBy=multi-user.target
SVCEOF

# Create management script
cat > /usr/local/bin/newslettar-ctl << 'CTLEOF'
#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

case "$1" in
    start)
        systemctl start newslettar.service
        echo -e "${GREEN}✓ Newslettar started${NC}"
        ;;
    stop)
        systemctl stop newslettar.service
        echo -e "${YELLOW}Newslettar stopped${NC}"
        ;;
    restart)
        systemctl restart newslettar.service
        echo -e "${GREEN}✓ Newslettar restarted${NC}"
        ;;
    status)
        systemctl status newslettar.service --no-pager
        ;;
    logs)
        journalctl -u newslettar.service -f
        ;;
    test)
        echo -e "${YELLOW}Sending test newsletter...${NC}"
        cd /opt/newslettar
        ./newslettar
        ;;
    edit)
        ${EDITOR:-nano} /opt/newslettar/.env
        echo -e "${YELLOW}Remember to restart: newslettar-ctl restart${NC}"
        ;;
    web)
        IP=$(hostname -I | awk '{print $1}')
        echo -e "${GREEN}Web UI:${NC} http://${IP}:8080"
        ;;
    update)
        echo -e "${YELLOW}Updating Newslettar...${NC}"
        curl -sSL https://raw.githubusercontent.com/MadSwell-dev/newslettar/main/install-binary.sh | bash
        ;;
    memory)
        echo -e "${BLUE}Memory Usage:${NC}"
        ps aux | grep newslettar | grep -v grep | awk '{print "  Process: "$11" - "$4"% ("$6/1024" MB)"}'
        ;;
    *)
        echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
        echo -e "${BLUE}║     Newslettar Control v2.0.0          ║${NC}"
        echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
        echo ""
        echo "Usage: newslettar-ctl {command}"
        echo ""
        echo "Commands:"
        echo "  start    - Start Newslettar service"
        echo "  stop     - Stop Newslettar service"
        echo "  restart  - Restart Newslettar service"
        echo "  status   - Show service status"
        echo "  logs     - View logs (live)"
        echo "  test     - Send test newsletter now"
        echo "  edit     - Edit configuration (.env file)"
        echo "  web      - Show Web UI URL"
        echo "  update   - Update to latest version"
        echo "  memory   - Show memory usage"
        exit 1
        ;;
esac
CTLEOF

chmod +x /usr/local/bin/newslettar-ctl

# Enable and start service
systemctl daemon-reload
systemctl enable --now newslettar.service

echo -e "${GREEN}✓ Service configured and started${NC}"

# Get IP address
IP=$(hostname -I | awk '{print $1}')

# Wait for service to fully start
sleep 2

echo ""
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║     Installation Complete! 🚀          ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}┌─ Web UI Access ──────────────────────┐${NC}"
echo -e "${BLUE}│${NC} ${GREEN}http://${IP}:8080${NC}"
echo -e "${BLUE}└──────────────────────────────────────┘${NC}"
echo ""
echo -e "${YELLOW}Quick Start:${NC}"
echo "  1. Open http://${IP}:8080 in your browser"
echo "  2. Configure Sonarr/Radarr in Configuration tab"
echo "  3. Configure email settings"
echo "  4. Test connections and send test newsletter"
echo ""
echo -e "${YELLOW}Command Line:${NC}"
echo "  newslettar-ctl web      - Show Web UI URL"
echo "  newslettar-ctl status   - Check service status"
echo "  newslettar-ctl test     - Send test newsletter"
echo "  newslettar-ctl logs     - View logs"
echo "  newslettar-ctl update   - Update to latest version"
echo ""
echo -e "${GREEN}Installed version: ${LATEST_VERSION}${NC}"
echo -e "${GREEN}Binary size: ${BINARY_SIZE}${NC}"
echo -e "${GREEN}Memory usage: ~12 MB${NC}"
echo ""
