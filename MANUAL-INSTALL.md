# Manual Installation Guide - Copy & Paste Friendly

This guide walks through each installation step so you can see exactly what's happening.

## Prerequisites

You need:
- Root access (or sudo)
- Debian 13 LXC container
- Internet connection

## Installation Steps

### Step 1: Update System
```bash
apt-get update
apt-get install -y wget curl git ca-certificates build-essential
```

### Step 2: Install Go 1.23.5

First, detect your architecture:
```bash
ARCH=$(dpkg --print-architecture)
echo "Architecture: $ARCH"
```

Then install Go:
```bash
# For amd64 (Intel/AMD 64-bit)
cd /tmp
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz
rm go1.23.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile

# Verify installation
go version
```

**For ARM64 or ARM32v6l:**
Replace `amd64` with `arm64` or `armv6l` in the wget URL above.

### Step 3: Create Installation Directory
```bash
mkdir -p /opt/newslettar
cd /opt/newslettar
```

### Step 4: Clone Repository
```bash
git clone --depth 1 --branch main "https://github.com/agencefanfare/newslettar.git" .
```

**If git clone fails**, download files manually:
```bash
mkdir -p templates
wget https://raw.githubusercontent.com/agencefanfare/newslettar/main/main.go
wget https://raw.githubusercontent.com/agencefanfare/newslettar/main/types.go
wget https://raw.githubusercontent.com/agencefanfare/newslettar/main/config.go
wget https://raw.githubusercontent.com/agencefanfare/newslettar/main/api.go
wget https://raw.githubusercontent.com/agencefanfare/newslettar/main/newsletter.go
wget https://raw.githubusercontent.com/agencefanfare/newslettar/main/handlers.go
wget https://raw.githubusercontent.com/agencefanfare/newslettar/main/server.go
wget https://raw.githubusercontent.com/agencefanfare/newslettar/main/utils.go
wget https://raw.githubusercontent.com/agencefanfare/newslettar/main/ui.go
wget https://raw.githubusercontent.com/agencefanfare/newslettar/main/go.mod
wget https://raw.githubusercontent.com/agencefanfare/newslettar/main/go.sum
wget https://raw.githubusercontent.com/agencefanfare/newslettar/main/version.json
wget https://raw.githubusercontent.com/agencefanfare/newslettar/main/templates/email.html -O templates/email.html
```

### Step 5: Build the Application
```bash
cd /opt/newslettar
go mod tidy
go build -ldflags="-s -w" -trimpath -o newslettar main.go
chmod +x newslettar
ls -lh newslettar  # Should be ~13 MB
```

### Step 6: Create Configuration File
```bash
cd /opt/newslettar
cat > .env << 'EOF'
# Sonarr Configuration
SONARR_URL=http://localhost:8989
SONARR_API_KEY=your-api-key-here

# Radarr Configuration
RADARR_URL=http://localhost:7878
RADARR_API_KEY=your-api-key-here

# Email Configuration
MAILGUN_SMTP=smtp.mailgun.org
MAILGUN_PORT=587
MAILGUN_USER=your-email@domain.com
MAILGUN_PASS=your-password

# Email Settings
FROM_EMAIL=newsletter@yourdomain.com
FROM_NAME=Newslettar
TO_EMAILS=recipient1@example.com,recipient2@example.com

# Schedule (Timezone-aware)
TIMEZONE=America/New_York
SCHEDULE_DAY=Sun
SCHEDULE_TIME=09:00

# Web UI
WEBUI_PORT=8080

# Display Options
SHOW_POSTERS=true
SHOW_DOWNLOADED=true
EOF
```

### Step 7: Create Systemd Service
```bash
cat > /etc/systemd/system/newslettar.service << 'EOF'
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
EOF

systemctl daemon-reload
systemctl enable newslettar.service
```

### Step 8: Create Management Command
```bash
cat > /usr/local/bin/newslettar-ctl << 'EOF'
#!/bin/bash
case "$1" in
    start)   systemctl start newslettar.service ;;
    stop)    systemctl stop newslettar.service ;;
    restart) systemctl restart newslettar.service ;;
    status)  systemctl status newslettar.service ;;
    logs)    journalctl -u newslettar.service -f ;;
    edit)    nano /opt/newslettar/.env; systemctl restart newslettar.service ;;
    web)     echo "Web UI: http://$(hostname -I | awk '{print $1}'):8080" ;;
    test)    cd /opt/newslettar && ./newslettar ;;
    update)  cd /opt/newslettar && git fetch origin main -q && git reset --hard origin/main -q && go build -ldflags="-s -w" -trimpath -o newslettar main.go && systemctl restart newslettar.service && echo "✓ Updated" ;;
    memory)  ps aux | grep newslettar | grep -v grep | awk '{print "Memory: "$6/1024" MB"}' ;;
    *)       echo "Usage: newslettar-ctl {start|stop|restart|status|logs|edit|web|test|update|memory}" ;;
esac
EOF

chmod +x /usr/local/bin/newslettar-ctl
```

### Step 9: Start the Service
```bash
systemctl start newslettar.service

# Check if it started
systemctl status newslettar.service

# View recent logs
journalctl -u newslettar.service -n 20
```

### Step 10: Access the Web UI
```bash
# Get your container IP
hostname -I

# Open browser to: http://<IP>:8080
```

## Troubleshooting

### Check service status
```bash
systemctl status newslettar.service
journalctl -u newslettar.service -f
```

### View logs
```bash
newslettar-ctl logs
```

### Rebuild manually
```bash
cd /opt/newslettar
go mod tidy
go build -ldflags="-s -w" -trimpath -o newslettar main.go
systemctl restart newslettar.service
```

### Check if port 8080 is accessible
```bash
curl http://localhost:8080
```

## Management Commands

Once installed, you can use:

```bash
newslettar-ctl start      # Start service
newslettar-ctl stop       # Stop service
newslettar-ctl restart    # Restart service
newslettar-ctl status     # Check status
newslettar-ctl logs       # View logs (live)
newslettar-ctl edit       # Edit configuration (.env)
newslettar-ctl web        # Show Web UI URL
newslettar-ctl test       # Send test newsletter
newslettar-ctl update     # Update to latest version
newslettar-ctl memory     # Show memory usage
```

## Next Steps

1. **Edit configuration:**
   ```bash
   newslettar-ctl edit
   ```

2. **Test the setup:**
   ```bash
   newslettar-ctl test
   ```

3. **View logs:**
   ```bash
   newslettar-ctl logs
   ```

4. **Access Web UI:**
   ```bash
   newslettar-ctl web
   ```

That's it! Your Newslettar is now installed and running.
