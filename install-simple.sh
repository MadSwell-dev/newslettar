#!/bin/bash
# Newslettar Simple Installer - Step by step
# This script is more transparent and easier to debug

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  Newslettar Simple Installer           ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}ERROR: Please run as root${NC}"
    exit 1
fi

INSTALL_DIR="/opt/newslettar"

# Step 1: Install dependencies
echo -e "${YELLOW}Step 1: Installing system dependencies...${NC}"
apt-get update -qq
apt-get install -y wget curl git ca-certificates build-essential >/dev/null 2>&1
echo -e "${GREEN}✓ Dependencies installed${NC}"
echo ""

# Step 2: Install Go
echo -e "${YELLOW}Step 2: Installing Go 1.23.5...${NC}"
if ! command -v go &> /dev/null; then
    ARCH=$(dpkg --print-architecture)
    case $ARCH in
        amd64) GO_ARCH="amd64" ;;
        arm64) GO_ARCH="arm64" ;;
        armhf) GO_ARCH="armv6l" ;;
        *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
    esac
    
    cd /tmp
    wget -q https://go.dev/dl/go1.23.5.linux-${GO_ARCH}.tar.gz
    tar -C /usr/local -xzf go1.23.5.linux-${GO_ARCH}.tar.gz
    rm go1.23.5.linux-${GO_ARCH}.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
fi
echo -e "${GREEN}✓ Go installed: $(go version | awk '{print $3}')${NC}"
export PATH=$PATH:/usr/local/go/bin
echo ""

# Step 3: Create installation directory
echo -e "${YELLOW}Step 3: Creating installation directory...${NC}"
rm -rf "$INSTALL_DIR" 2>/dev/null || true
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"
echo -e "${GREEN}✓ Directory: $INSTALL_DIR${NC}"
echo ""

# Step 4: Download or clone repository
echo -e "${YELLOW}Step 4: Getting source code...${NC}"
echo -e "${BLUE}Method: Git clone${NC}"

# Create temp directory for clone
TEMP_CLONE=$(mktemp -d)
echo -e "${BLUE}  Temp directory: $TEMP_CLONE${NC}"
git clone --depth 1 --branch main "https://github.com/agencefanfare/newslettar.git" "$TEMP_CLONE" 2>&1

if [ $? -eq 0 ]; then
    echo -e "${BLUE}  Clone successful, checking files...${NC}"
    ls -la "$TEMP_CLONE" | head -20
    
    if [ -f "$TEMP_CLONE/main.go" ]; then
        # Copy all files explicitly to ensure nothing is missed
        echo -e "${BLUE}  Copying files to $INSTALL_DIR...${NC}"
        cp "$TEMP_CLONE"/api.go "$INSTALL_DIR/" || echo "Failed to copy api.go"
        cp "$TEMP_CLONE"/config.go "$INSTALL_DIR/" || echo "Failed to copy config.go"
        cp "$TEMP_CLONE"/handlers.go "$INSTALL_DIR/" || echo "Failed to copy handlers.go"
        cp "$TEMP_CLONE"/main.go "$INSTALL_DIR/" || echo "Failed to copy main.go"
        cp "$TEMP_CLONE"/newsletter.go "$INSTALL_DIR/" || echo "Failed to copy newsletter.go"
        cp "$TEMP_CLONE"/server.go "$INSTALL_DIR/" || echo "Failed to copy server.go"
        cp "$TEMP_CLONE"/types.go "$INSTALL_DIR/" || echo "Failed to copy types.go"
        cp "$TEMP_CLONE"/ui.go "$INSTALL_DIR/" || echo "Failed to copy ui.go"
        cp "$TEMP_CLONE"/utils.go "$INSTALL_DIR/" || echo "Failed to copy utils.go"
        cp "$TEMP_CLONE"/go.mod "$INSTALL_DIR/" || echo "Failed to copy go.mod"
        cp "$TEMP_CLONE"/go.sum "$INSTALL_DIR/" || echo "Failed to copy go.sum"
        cp "$TEMP_CLONE"/version.json "$INSTALL_DIR/" || echo "Failed to copy version.json"
        mkdir -p "$INSTALL_DIR/templates"
        cp "$TEMP_CLONE"/templates/email.html "$INSTALL_DIR/templates/" || echo "Failed to copy email.html"
        cp -r "$TEMP_CLONE"/.git "$INSTALL_DIR/" 2>/dev/null || true
        cp "$TEMP_CLONE"/.gitignore "$INSTALL_DIR/" 2>/dev/null || true
        
        echo -e "${BLUE}  Verifying copy...${NC}"
        ls -la "$INSTALL_DIR" | head -20
        
        # Final check: ensure all critical files exist
        MISSING=""
        for file in main.go types.go config.go api.go newsletter.go handlers.go server.go utils.go ui.go go.mod; do
            if [ ! -f "$INSTALL_DIR/$file" ]; then
                MISSING="$MISSING $file"
            fi
        done
        
        if [ -n "$MISSING" ]; then
            echo -e "${RED}ERROR: Missing files:$MISSING${NC}"
            exit 1
        fi
        
        rm -rf "$TEMP_CLONE"
        echo -e "${GREEN}✓ Source code downloaded${NC}"
    else
        echo -e "${RED}ERROR: main.go not found in $TEMP_CLONE${NC}"
        ls -la "$TEMP_CLONE"
        exit 1
    fi
else
    echo -e "${YELLOW}Git clone failed, trying wget fallback...${NC}"
    rm -rf "$TEMP_CLONE"
    mkdir -p "$INSTALL_DIR/templates"
    for file in main.go types.go config.go api.go newsletter.go handlers.go server.go utils.go ui.go go.mod go.sum version.json; do
        echo -e "${BLUE}  Downloading ${file}...${NC}"
        wget -q -O "$INSTALL_DIR/$file" "https://raw.githubusercontent.com/agencefanfare/newslettar/main/${file}" || echo -e "${RED}Failed: $file${NC}"
    done
    wget -q -O "$INSTALL_DIR/templates/email.html" "https://raw.githubusercontent.com/agencefanfare/newslettar/main/templates/email.html" || echo -e "${RED}Failed: email.html${NC}"
    echo -e "${GREEN}✓ Source code downloaded (fallback)${NC}"
fi
echo ""

# Step 5: Build
echo -e "${YELLOW}Step 5: Building Newslettar...${NC}"
go mod tidy
go build -ldflags="-s -w" -trimpath -o newslettar main.go
chmod +x newslettar
BINARY_SIZE=$(du -h newslettar | cut -f1)
echo -e "${GREEN}✓ Built successfully (${BINARY_SIZE})${NC}"
echo ""

# Step 6: Create configuration
echo -e "${YELLOW}Step 6: Creating configuration file...${NC}"
cat > .env << 'EOF'
# Sonarr Configuration
SONARR_URL=http://localhost:8989
SONARR_API_KEY=

# Radarr Configuration
RADARR_URL=http://localhost:7878
RADARR_API_KEY=

# Email Configuration (SMTP)
MAILGUN_SMTP=smtp.mailgun.org
MAILGUN_PORT=587
MAILGUN_USER=your-email@domain.com
MAILGUN_PASS=your-app-password

# Email Settings
FROM_EMAIL=newsletter@yourdomain.com
FROM_NAME=Newslettar
TO_EMAILS=recipient1@example.com,recipient2@example.com

# Schedule Settings
TIMEZONE=America/New_York
SCHEDULE_DAY=Sun
SCHEDULE_TIME=09:00

# Web UI
WEBUI_PORT=8080

# Display Options
SHOW_POSTERS=true
SHOW_DOWNLOADED=true
EOF
echo -e "${GREEN}✓ Configuration file created (.env)${NC}"
echo ""

# Step 7: Create systemd service
echo -e "${YELLOW}Step 7: Setting up systemd service...${NC}"
cat > /etc/systemd/system/newslettar.service << 'SVCEOF'
[Unit]
Description=Newslettar - Sonarr/Radarr Newsletter Generator
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/newslettar
ExecStart=/opt/newslettar/newslettar -web
Restart=on-failure
RestartSec=5
StandardOutput=append:/opt/newslettar/logs.txt
StandardError=append:/opt/newslettar/logs.txt
MemoryMax=100M
CPUQuota=50%

[Install]
WantedBy=multi-user.target
SVCEOF

systemctl daemon-reload
systemctl enable newslettar.service
echo -e "${GREEN}✓ Service configured${NC}"
echo ""

# Step 8: Create management CLI
echo -e "${YELLOW}Step 8: Creating management command...${NC}"
cat > /usr/local/bin/newslettar-ctl << 'CTLEOF'
#!/bin/bash
case "$1" in
    start) systemctl start newslettar.service ;;
    stop) systemctl stop newslettar.service ;;
    restart) systemctl restart newslettar.service ;;
    status) systemctl status newslettar.service ;;
    logs) journalctl -u newslettar.service -f ;;
    edit) nano /opt/newslettar/.env; systemctl restart newslettar.service ;;
    web) echo "Web UI: http://$(hostname -I | awk '{print $1}'):8080" ;;
    test) cd /opt/newslettar && ./newslettar ;;
    update) cd /opt/newslettar && git fetch origin main -q && git reset --hard origin/main -q && go build -ldflags="-s -w" -trimpath -o newslettar main.go && systemctl restart newslettar.service && echo "✓ Updated" ;;
    memory) ps aux | grep newslettar | grep -v grep | awk '{print "Memory: "$6/1024" MB"}' ;;
    *) echo "Usage: newslettar-ctl {start|stop|restart|status|logs|edit|web|test|update|memory}" ;;
esac
CTLEOF

chmod +x /usr/local/bin/newslettar-ctl
echo -e "${GREEN}✓ Management command installed${NC}"
echo ""

# Step 9: Start service
echo -e "${YELLOW}Step 9: Starting service...${NC}"
systemctl start newslettar.service
sleep 2

if systemctl is-active --quiet newslettar.service; then
    echo -e "${GREEN}✓ Service started successfully${NC}"
else
    echo -e "${RED}⚠ Service may not have started. Check logs:${NC}"
    echo -e "${BLUE}journalctl -u newslettar.service -n 20${NC}"
fi
echo ""

# Summary
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║    Installation Complete!             ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Edit configuration:"
echo -e "   ${BLUE}newslettar-ctl edit${NC}"
echo ""
echo "2. View logs:"
echo -e "   ${BLUE}newslettar-ctl logs${NC}"
echo ""
echo "3. Access Web UI:"
IP=$(hostname -I | awk '{print $1}')
echo -e "   ${BLUE}http://${IP}:8080${NC}"
echo ""
echo -e "${YELLOW}Management Commands:${NC}"
echo "   newslettar-ctl start     - Start service"
echo "   newslettar-ctl stop      - Stop service"
echo "   newslettar-ctl restart   - Restart service"
echo "   newslettar-ctl status    - Check status"
echo "   newslettar-ctl logs      - View logs"
echo "   newslettar-ctl edit      - Edit configuration"
echo "   newslettar-ctl test      - Send test newsletter"
echo "   newslettar-ctl update    - Update to latest version"
echo "   newslettar-ctl web       - Show Web UI URL"
echo "   newslettar-ctl memory    - Show memory usage"
echo ""
