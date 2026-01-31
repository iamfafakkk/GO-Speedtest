#!/bin/bash
# Update script for GO-Speedtest on Ubuntu
# Usage: sudo ./update.sh

set -e

SERVICE_NAME="speedgo"
APP_NAME="speedtest"
INSTALL_DIR="/opt/speedtest"
SYSTEMD_DIR="/etc/systemd/system"

echo "=========================================="
echo "  GO-Speedtest Update Script"
echo "=========================================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root: sudo ./update.sh"
    exit 1
fi

# 1. Git pull
echo ""
echo "[1/6] Pulling latest changes..."
git pull

# 2. Stop service
echo ""
echo "[2/6] Stopping $SERVICE_NAME service..."
systemctl stop $SERVICE_NAME || echo "Service not running"

# 3. Build
echo ""
echo "[3/6] Building..."
./build.sh

# 4. Move to install directory
echo ""
echo "[4/6] Installing to $INSTALL_DIR..."
mkdir -p $INSTALL_DIR
cp ./build/$APP_NAME $INSTALL_DIR/
chown root:root $INSTALL_DIR/$APP_NAME
chmod +x $INSTALL_DIR/$APP_NAME

# 5. Update systemd service
echo ""
echo "[5/6] Updating systemd service..."
cp ./systemd/$SERVICE_NAME.service $SYSTEMD_DIR/
systemctl daemon-reload
systemctl enable $SERVICE_NAME

# 6. Start and show status
echo ""
echo "[6/6] Starting service..."
systemctl start $SERVICE_NAME

echo ""
echo "=========================================="
echo "  Update Complete!"
echo "=========================================="
systemctl status $SERVICE_NAME --no-pager
