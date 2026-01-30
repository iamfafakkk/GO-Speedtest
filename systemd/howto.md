# GO-Speedtest Systemd Integration Guide

## Prerequisites

- Ubuntu 18.04+ atau Debian 10+
- Go 1.21+ (untuk build)
- Root access

## Quick Install

```bash
# Clone repository
git clone https://github.com/your-repo/GO-Speedtest.git
cd GO-Speedtest

# Build dan install
sudo ./update.sh
```

## Manual Installation

### 1. Build Binary

```bash
./build.sh
```

### 2. Create Install Directory

```bash
sudo mkdir -p /opt/speedtest
sudo chown www-data:www-data /opt/speedtest
```

### 3. Copy Binary

```bash
sudo cp ./build/speedtest /opt/speedtest/
sudo chmod +x /opt/speedtest/speedtest
```

### 4. Install Service

```bash
sudo cp ./systemd/speedgo.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable speedgo
```

### 5. Start Service

```bash
sudo systemctl start speedgo
sudo systemctl status speedgo
```

## Service Management

```bash
# Start
sudo systemctl start speedgo

# Stop
sudo systemctl stop speedgo

# Restart
sudo systemctl restart speedgo

# Status
sudo systemctl status speedgo

# Logs
sudo journalctl -u speedgo -f
```

## Configuration

Edit service file untuk mengubah port atau user:

```bash
sudo nano /etc/systemd/system/speedgo.service
```

Ubah environment variable:
```ini
Environment=PORT=8645
```

Reload setelah edit:
```bash
sudo systemctl daemon-reload
sudo systemctl restart speedgo
```

## Update Deployment

```bash
cd /path/to/GO-Speedtest
sudo ./update.sh
```

Script akan:
1. Git pull
2. Stop service
3. Build
4. Copy binary
5. Start service

## Troubleshooting

### Service Gagal Start

```bash
# Cek logs
sudo journalctl -u speedgo -n 50 --no-pager

# Cek permission
ls -la /opt/speedtest/
```

### Port Already in Use

```bash
# Cek port
sudo lsof -i :8645

# Kill process
sudo kill -9 <PID>
```

### Permission Denied

```bash
sudo chown -R www-data:www-data /opt/speedtest
sudo chmod +x /opt/speedtest/speedtest
```

## Nginx Reverse Proxy (Optional)

```nginx
server {
    listen 80;
    server_name speedtest.example.com;

    location / {
        proxy_pass http://127.0.0.1:8645;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_buffering off;
    }
}
```

## Uninstall

```bash
sudo systemctl stop speedgo
sudo systemctl disable speedgo
sudo rm /etc/systemd/system/speedgo.service
sudo rm -rf /opt/speedtest
sudo systemctl daemon-reload
```
