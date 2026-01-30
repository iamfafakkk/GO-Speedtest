# GO-Speedtest

REST API server untuk internet speed test dengan **realtime streaming** seperti Ookla Speedtest. Pure Go, zero dependencies.

## Features

- 🏓 **Ping Test** - Mengukur latency ke server
- ⬇️ **Download Test** - Chunked streaming untuk realtime progress
- ⬆️ **Upload Test** - Speed calculation dari upload stream
- 🌐 **CORS Support** - Browser-compatible
- ⚡ **Zero Dependencies** - Pure Go standard library

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

### GET /ping
Cek latency ke server.

**Response:**
```json
{
  "status": "pong",
  "timestamp": 1706688000000,
  "server_id": "hostname"
}
```

**Cara hitung latency:**
```javascript
const start = Date.now();
const res = await fetch('http://localhost:8645/ping');
const latency = Date.now() - start;
console.log(`Latency: ${latency}ms`);
```

---

### GET /download
Download test dengan chunked streaming.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| size | int | 25 | Ukuran download dalam MB (max: 100) |

**Response:** Binary stream dengan `Content-Length` header

**Cara hitung speed (realtime):**
```javascript
const response = await fetch('http://localhost:8645/download?size=25');
const reader = response.body.getReader();
const contentLength = +response.headers.get('Content-Length');

let receivedBytes = 0;
const startTime = Date.now();

while (true) {
  const { done, value } = await reader.read();
  if (done) break;
  
  receivedBytes += value.length;
  const elapsed = (Date.now() - startTime) / 1000;
  const speedMbps = (receivedBytes * 8) / (elapsed * 1_000_000);
  const progress = (receivedBytes / contentLength) * 100;
  
  console.log(`Progress: ${progress.toFixed(1)}% | Speed: ${speedMbps.toFixed(2)} Mbps`);
}
```

---

### POST /upload
Upload test - kirim data, server hitung speed.

**Request:** Binary body (application/octet-stream)

**Response:**
```json
{
  "bytes": 10485760,
  "duration_ms": 1500,
  "speed_mbps": 55.92
}
```

**Cara test upload (realtime):**
```javascript
const data = new Uint8Array(10 * 1024 * 1024); // 10MB
crypto.getRandomValues(data);

const startTime = Date.now();
const response = await fetch('http://localhost:8645/upload', {
  method: 'POST',
  body: data,
  headers: { 'Content-Type': 'application/octet-stream' }
});

const result = await response.json();
console.log(`Uploaded: ${result.bytes} bytes | Speed: ${result.speed_mbps.toFixed(2)} Mbps`);
```

---

### GET /
Health check endpoint.

**Response:**
```json
{
  "service": "GO-Speedtest",
  "version": "1.0.0",
  "status": "running"
}
```

## Testing dengan cURL

```bash
# 1. Ping test
curl -w "\nLatency: %{time_total}s\n" http://localhost:8645/ping

# 2. Download test (10MB)
curl -o /dev/null -w "Speed: %{speed_download} B/s\nTime: %{time_total}s\n" \
  "http://localhost:8645/download?size=10"

# 3. Upload test (10MB random data)
dd if=/dev/urandom bs=1M count=10 2>/dev/null | \
  curl -X POST -H "Content-Type: application/octet-stream" \
  --data-binary @- http://localhost:8645/upload
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8645 | Port server |

## Realtime Client Example (HTML)

```html
<!DOCTYPE html>
<html>
<head><title>Speedtest</title></head>
<body>
  <button onclick="runTest()">Start Test</button>
  <div id="result"></div>
  <script>
    async function runTest() {
      const result = document.getElementById('result');
      
      // Ping
      const pingStart = Date.now();
      await fetch('http://localhost:8645/ping');
      result.innerHTML = `Ping: ${Date.now() - pingStart}ms<br>`;
      
      // Download
      const dlStart = Date.now();
      const res = await fetch('http://localhost:8645/download?size=10');
      const reader = res.body.getReader();
      const total = +res.headers.get('Content-Length');
      let received = 0;
      
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        received += value.length;
        const speed = (received * 8) / ((Date.now() - dlStart) / 1000) / 1e6;
        result.innerHTML += `Download: ${speed.toFixed(2)} Mbps (${((received/total)*100).toFixed(0)}%)<br>`;
      }
    }
  </script>
</body>
</html>
```

## Architecture Notes

- **Chunked Transfer**: Server flush setiap 64KB chunk untuk realtime progress
- **Random Data**: Download menggunakan crypto/rand untuk data realistis
- **Memory Efficient**: Upload menggunakan streaming read, tidak buffer seluruh body
- **CORS Enabled**: Support browser-based speedtest clients

## License

MIT
