#!/bin/bash
set -e

# Debian Package Builder for Newslettar
# This script creates a .deb package for easy installation on Debian/Ubuntu systems

VERSION=$(cat version.json | grep version | cut -d'"' -f4)
ARCH=$(dpkg --print-architecture 2>/dev/null || echo "amd64")
PACKAGE_NAME="newslettar"
PACKAGE_DIR="build/deb/${PACKAGE_NAME}_${VERSION}_${ARCH}"

echo "Building Debian package: ${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"

# Clean previous builds
rm -rf build/deb
mkdir -p "$PACKAGE_DIR"

# Create directory structure
mkdir -p "$PACKAGE_DIR/DEBIAN"
mkdir -p "$PACKAGE_DIR/opt/newslettar"
mkdir -p "$PACKAGE_DIR/etc/systemd/system"
mkdir -p "$PACKAGE_DIR/usr/local/bin"

# Build the binary (templates and assets are embedded)
echo "Building binary..."
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o "$PACKAGE_DIR/opt/newslettar/newslettar" ./cmd/newslettar

# Copy version file
cp version.json "$PACKAGE_DIR/opt/newslettar/"

# Copy environment example
cp .env.example "$PACKAGE_DIR/opt/newslettar/.env.example"

# Create systemd service file
cat > "$PACKAGE_DIR/etc/systemd/system/newslettar.service" << 'EOF'
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
MemoryMax=200M
CPUQuota=50%

[Install]
WantedBy=multi-user.target
EOF

# Create management script
cat > "$PACKAGE_DIR/usr/local/bin/newslettar-ctl" << 'EOF'
#!/bin/bash
case "$1" in
    start)   systemctl start newslettar.service ;;
    stop)    systemctl stop newslettar.service ;;
    restart) systemctl restart newslettar.service ;;
    status)  systemctl status newslettar.service ;;
    logs)    journalctl -u newslettar.service -f ;;
    edit)    ${EDITOR:-nano} /opt/newslettar/.env; systemctl restart newslettar.service ;;
    web)     echo "Web UI: http://$(hostname -I | awk '{print $1}'):8080" ;;
    test)    cd /opt/newslettar && ./newslettar ;;
    memory)  ps aux | grep newslettar | grep -v grep | awk '{print "Memory: "$6/1024" MB"}' ;;
    version) cat /opt/newslettar/version.json ;;
    *)       echo "Usage: newslettar-ctl {start|stop|restart|status|logs|edit|web|test|memory|version}" ;;
esac
EOF

chmod +x "$PACKAGE_DIR/usr/local/bin/newslettar-ctl"
chmod +x "$PACKAGE_DIR/opt/newslettar/newslettar"

# Create control file
cat > "$PACKAGE_DIR/DEBIAN/control" << EOF
Package: newslettar
Version: $VERSION
Section: web
Priority: optional
Architecture: $ARCH
Maintainer: MadSwell <madswell.dev@gmail.com>
Description: Automated newsletter generator for Sonarr and Radarr
 Newslettar automatically generates beautiful email newsletters
 summarizing new content from your Sonarr (TV) and Radarr (Movie)
 installations. Features include:
  - Scheduled weekly newsletters
  - Beautiful HTML templates with posters
  - Web UI for configuration
  - Trakt.tv integration for trending content
  - Low resource usage (~12MB RAM)
Homepage: https://github.com/MadSwell-dev/newslettar
Depends: systemd
EOF

# Create postinst script (runs after installation)
cat > "$PACKAGE_DIR/DEBIAN/postinst" << 'EOF'
#!/bin/bash
set -e

# Create .env if it doesn't exist
if [ ! -f /opt/newslettar/.env ]; then
    cp /opt/newslettar/.env.example /opt/newslettar/.env
    echo "Created /opt/newslettar/.env - please configure before starting"
fi

# Reload systemd
systemctl daemon-reload

# Enable service (but don't start it - user needs to configure first)
systemctl enable newslettar.service

echo ""
echo "✓ Newslettar installed successfully!"
echo ""
echo "Next steps:"
echo "  1. Configure: newslettar-ctl edit"
echo "  2. Start service: newslettar-ctl start"
echo "  3. Check status: newslettar-ctl status"
echo "  4. Access Web UI: http://$(hostname -I | awk '{print $1}'):8080"
echo ""
echo "Run 'newslettar-ctl' for available commands"
echo ""

exit 0
EOF

chmod +x "$PACKAGE_DIR/DEBIAN/postinst"

# Create prerm script (runs before removal)
cat > "$PACKAGE_DIR/DEBIAN/prerm" << 'EOF'
#!/bin/bash
set -e

# Stop and disable service
systemctl stop newslettar.service || true
systemctl disable newslettar.service || true

exit 0
EOF

chmod +x "$PACKAGE_DIR/DEBIAN/prerm"

# Create postrm script (runs after removal)
cat > "$PACKAGE_DIR/DEBIAN/postrm" << 'EOF'
#!/bin/bash
set -e

# Reload systemd
systemctl daemon-reload || true

echo "Newslettar has been removed."
echo "Configuration files remain in /opt/newslettar/.env (remove manually if desired)"

exit 0
EOF

chmod +x "$PACKAGE_DIR/DEBIAN/postrm"

# Build the package
echo "Creating .deb package..."
dpkg-deb --build "$PACKAGE_DIR"

# Move to dist directory
mkdir -p dist
mv "build/deb/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb" dist/

echo ""
echo "✓ Debian package created: dist/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
echo ""
echo "To install:"
echo "  sudo dpkg -i dist/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
echo ""
echo "To remove:"
echo "  sudo dpkg -r newslettar"
echo ""
