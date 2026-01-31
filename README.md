# GO-Speedtest

REST API server untuk internet speed test menggunakan [speedtest-go](https://github.com/iamfafakkk/speedtest-go) library. Test langsung ke server Ookla Speedtest.

## Features

- 🏓 **Ping Test** - Latency ke Ookla server terdekat
- ⬇️ **Download Test** - Real download speed via Ookla server
- ⬆️ **Upload Test** - Real upload speed via Ookla server
- 🌐 **Server List** - Daftar server Ookla terdekat
- 🎯 **Specific Server** - Test ke server tertentu via `server_id`

## Quick Start

```bash
# Run server
go run main.go

# Atau build dulu
go build -o speedtest main.go
./speedtest
```

Server berjalan di `http://localhost:8645` (atau custom via env `PORT`)

## API Reference

### GET /speedtest/ping
Test latency ke server Ookla terdekat.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| server_id | string | - | Optional, ID server Ookla tertentu |

**Response:**
```json
{
  "status": "success",
  "latency_ms": 15.5,
  "server_id": "12345",
  "server_name": "MyISP - Jakarta",
  "server_host": "speedtest.myisp.co.id:8080",
  "country": "Indonesia",
  "distance_km": 5.2,
  "timestamp": 1706688000000
}
```

---

### GET /speedtest/download
Test download speed ke server Ookla terdekat.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| server_id | string | - | Optional, ID server Ookla tertentu |

**Response:**
```json
{
  "speed_mbps": 95.5,
  "server_id": "12345",
  "server_name": "MyISP - Jakarta",
  "server_host": "speedtest.myisp.co.id:8080",
  "country": "Indonesia",
  "latency_ms": 15.5,
  "duration_ms": 8500,
  "timestamp": 1706688000000
}
```

---

### GET /speedtest/upload
Test upload speed ke server Ookla terdekat.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| server_id | string | - | Optional, ID server Ookla tertentu |

**Response:**
```json
{
  "speed_mbps": 45.2,
  "server_id": "12345",
  "server_name": "MyISP - Jakarta",
  "server_host": "speedtest.myisp.co.id:8080",
  "country": "Indonesia",
  "latency_ms": 15.5,
  "duration_ms": 8500,
  "timestamp": 1706688000000
}
```

---

### GET /speedtest/servers
Daftar 10 server Ookla terdekat.

**Response:**
```json
{
  "count": 10,
  "servers": [
    {
      "id": "12345",
      "name": "MyISP - Jakarta",
      "host": "speedtest.myisp.co.id:8080",
      "country": "Indonesia",
      "distance_km": 5.2
    }
  ]
}
```

---

### GET /
Health check endpoint.

**Response:**
```json
{
  "service": "GO-Speedtest",
  "version": "2.0.0",
  "status": "running",
  "library": "speedtest-go v1.7.10"
}
```

## Testing dengan cURL

```bash
# 1. Health check
curl http://localhost:8645/

# 2. List servers
curl http://localhost:8645/speedtest/servers

# 3. Ping test
curl http://localhost:8645/speedtest/ping

# 4. Download test
curl http://localhost:8645/speedtest/download

# 5. Upload test
curl http://localhost:8645/speedtest/upload

# 6. Test dengan server tertentu
curl "http://localhost:8645/speedtest/ping?server_id=12345"
curl "http://localhost:8645/speedtest/download?server_id=12345"
curl "http://localhost:8645/speedtest/upload?server_id=12345"
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8645 | Port server |

## Dependencies

- [speedtest-go](https://github.com/showwin/speedtest-go) v1.7.10 - Library untuk Ookla Speedtest

---

## Realtime SSE Streaming

Untuk mendapatkan progress realtime selama test, gunakan SSE (Server-Sent Events) endpoints.

### GET /speedtest/download/stream
SSE streaming untuk download test dengan progress realtime.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| server_id | string | - | Optional, ID server Ookla tertentu |
| duration | int | 10 | Test duration in seconds (max: 30) |

**SSE Events:**
```javascript
// Event types: "start", "progress", "complete", "error"

// start event
{"type":"start","server_id":"12345","server_name":"MyISP - Jakarta","latency_ms":15.5}

// progress event (setiap 200ms)
{"type":"progress","speed_mbps":85.5,"elapsed_sec":2.4}

// complete event
{"type":"complete","speed_mbps":95.5,"elapsed_sec":10.2,"server_id":"12345","server_name":"MyISP - Jakarta","latency_ms":15.5}

// error event
{"type":"error","message":"failed to connect to server"}
```

### GET /speedtest/upload/stream
SSE streaming untuk upload test dengan progress realtime. Format sama dengan download.

### JavaScript Usage Example

```javascript
// Realtime download test dengan progress
const eventSource = new EventSource('http://localhost:8645/speedtest/download/stream?duration=15');

eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  
  switch(data.type) {
    case 'start':
      console.log(`Testing server: ${data.server_name}, Latency: ${data.latency_ms}ms`);
      break;
    case 'progress':
      console.log(`Speed: ${data.speed_mbps.toFixed(2)} Mbps (${data.elapsed_sec.toFixed(1)}s)`);
      break;
    case 'complete':
      console.log(`Final: ${data.speed_mbps.toFixed(2)} Mbps`);
      eventSource.close();
      break;
    case 'error':
      console.error(data.message);
      eventSource.close();
      break;
  }
};

eventSource.onerror = () => {
  console.error('SSE connection failed');
  eventSource.close();
};
```

### cURL Test (SSE)
```bash
# Download stream (realtime, 15 seconds)
curl -N "http://localhost:8645/speedtest/download/stream?duration=15"

# Upload stream (realtime, 10 seconds)
curl -N "http://localhost:8645/speedtest/upload/stream?duration=10"
```

---

## Notes

1. **Execution Time**: Download dan upload test membutuhkan waktu 5-15 detik karena menjalankan test actual ke server Ookla
2. **Network Required**: Server harus terkoneksi ke internet untuk bisa melakukan speedtest
3. **Server Selection**: Secara default akan memilih server terdekat, gunakan `server_id` untuk memilih server tertentu

## Reverse Proxy Configuration (Nginx/Cloudflare)

Jika menjalankan server di belakang Nginx atau Cloudflare, pastikan buffering dimatikan untuk path `/speedtest/*/stream`.

Aplikasi ini sudah menambahkan header `X-Accel-Buffering: no` secara otomatis. Namun pastikan konfigurasi Nginx Anda mendukungnya:

```nginx
location @reverse_proxy {
    proxy_pass {{reverse_proxy_url}};
    proxy_http_version 1.1;

    # Headers
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Forwarded-Server $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header Host $host;

    # Connection Setup for SSE (Keep-Alive)
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "";

    # === FIX REALTIME SPEED (BUFFERING OFF) ===
    # Matikan buffering agar output realtime tidak ditahan di memory Nginx
    proxy_buffering off;
    proxy_cache off;
    proxy_max_temp_file_size 0;
    
    # Timeout Settings
    proxy_connect_timeout 900;
    proxy_send_timeout 900;
    proxy_read_timeout 900;

    # SSL (If needed)
    proxy_ssl_server_name on;
    proxy_ssl_name $host;
    proxy_pass_request_headers on;
}
```

## License

MIT
