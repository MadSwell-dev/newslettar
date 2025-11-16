#!/bin/bash
# Docker Setup Script for Newslettar
# This script handles Docker setup for environments without docker-compose

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║    Newslettar Docker Setup v1.1.0      ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker is not installed. Please install Docker first.${NC}"
    echo "Visit: https://docs.docker.com/get-docker/"
    exit 1
fi

echo -e "${GREEN}✓ Docker is installed${NC}"
echo ""

# Create data directory
echo -e "${YELLOW}[1/4] Setting up directories...${NC}"
mkdir -p data
echo -e "${GREEN}✓ Data directory created${NC}"

# Check if .env.example exists
if [ ! -f ".env.example" ]; then
    echo -e "${RED}ERROR: .env.example not found in current directory${NC}"
    echo "Make sure you're in the newslettar repository directory"
    exit 1
fi

# Create .env if it doesn't exist
if [ ! -f "data/.env" ]; then
    echo -e "${YELLOW}[2/4] Creating configuration...${NC}"
    cp .env.example data/.env
    echo -e "${GREEN}✓ Configuration template created at data/.env${NC}"
    echo ""
    echo -e "${YELLOW}Please edit data/.env with your settings:${NC}"
    echo "  - SONARR_URL and SONARR_API_KEY"
    echo "  - RADARR_URL and RADARR_API_KEY"
    echo "  - Email configuration (MAILGUN_*)"
    echo "  - TIMEZONE and SCHEDULE settings"
    echo ""
    echo -e "${BLUE}Run: nano data/.env${NC}"
    echo ""
    exit 0
else
    echo -e "${YELLOW}[2/4] Configuration found at data/.env${NC}"
fi

# Build Docker image
echo -e "${YELLOW}[3/4] Building Docker image...${NC}"
docker build -t newslettar:latest .
echo -e "${GREEN}✓ Docker image built${NC}"

# Run container
echo -e "${YELLOW}[4/4] Starting container...${NC}"

# Check if container already exists
if docker ps -a --format '{{.Names}}' | grep -q "^newslettar$"; then
    echo -e "${YELLOW}Removing old container...${NC}"
    docker stop newslettar 2>/dev/null || true
    docker rm newslettar 2>/dev/null || true
fi

# Run the container
docker run -d \
  --name newslettar \
  -p 8080:8080 \
  -v "$(pwd)/data/.env:/opt/newslettar/.env" \
  --restart unless-stopped \
  newslettar:latest

echo -e "${GREEN}✓ Container started${NC}"

# Wait for container to be ready
sleep 2

# Get container status
if docker ps --format '{{.Names}}' | grep -q "^newslettar$"; then
    echo -e "${GREEN}✓ Container is running${NC}"
else
    echo -e "${RED}ERROR: Container failed to start${NC}"
    echo "Run: docker logs newslettar"
    exit 1
fi

echo ""
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║     Installation Complete! 🚀          ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Web UI: http://localhost:8080${NC}"
echo ""
echo -e "${YELLOW}Quick commands:${NC}"
echo "  View logs:       docker logs -f newslettar"
echo "  Restart:         docker restart newslettar"
echo "  Stop:            docker stop newslettar"
echo "  Start:           docker start newslettar"
echo "  Shell:           docker exec -it newslettar /bin/bash"
echo ""
echo -e "${GREEN}Visit http://localhost:8080 to configure and use Newslettar!${NC}"
