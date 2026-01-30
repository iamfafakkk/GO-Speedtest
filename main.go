// GO-Speedtest: Realtime Speedtest API Server
// Seperti behavior Ookla Speedtest dengan chunked streaming
// Pure Go standard library - zero dependencies

package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// ==================== Constants ====================

const (
	DefaultPort       = "8645"
	DefaultDownloadMB = 25        // Default download size
	MaxDownloadMB     = 100       // Max download size
	ChunkSize         = 64 * 1024 // 64KB per chunk - optimal for streaming
	FlushInterval     = 50 * time.Millisecond
)

// ==================== Response Structs ====================

// PingResponse represents ping endpoint response
type PingResponse struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	ServerID  string `json:"server_id"`
}

// UploadResponse represents upload test result
type UploadResponse struct {
	Bytes      int64   `json:"bytes"`
	DurationMs int64   `json:"duration_ms"`
	SpeedMbps  float64 `json:"speed_mbps"`
}

// ==================== CORS Middleware ====================

// corsMiddleware adds CORS headers untuk browser-based clients
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins for speedtest (adjust for production)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// ==================== Ping Handler ====================

// pingHandler - GET /ping
// Mengecek latency ke server dengan timestamp response
func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := PingResponse{
		Status:    "pong",
		Timestamp: time.Now().UnixMilli(),
		ServerID:  getServerID(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	json.NewEncoder(w).Encode(response)
}

// ==================== Download Handler ====================

// downloadHandler - GET /download?size=50
// Streaming download dengan chunked transfer encoding
// Query params:
//   - size: ukuran dalam MB (default 25, max 100)
//
// Client menghitung: downloaded_bytes / elapsed_time = speed
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse size parameter
	sizeStr := r.URL.Query().Get("size")
	sizeMB := DefaultDownloadMB
	if sizeStr != "" {
		if parsed, err := strconv.Atoi(sizeStr); err == nil {
			sizeMB = parsed
		}
	}

	// Clamp size
	if sizeMB < 1 {
		sizeMB = 1
	}
	if sizeMB > MaxDownloadMB {
		sizeMB = MaxDownloadMB
	}

	totalBytes := int64(sizeMB) * 1024 * 1024

	// Set headers untuk realtime streaming
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(totalBytes, 10))
	w.Header().Set("Cache-Control", "no-cache, no-store")
	w.Header().Set("X-Content-Size-MB", strconv.Itoa(sizeMB))

	// Get flusher for realtime streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Generate random data per chunk dan stream ke client
	chunk := make([]byte, ChunkSize)
	var written int64

	for written < totalBytes {
		// Calculate remaining bytes
		remaining := totalBytes - written
		toWrite := int64(ChunkSize)
		if remaining < toWrite {
			toWrite = remaining
			chunk = chunk[:toWrite] // resize chunk
		}

		// Generate random bytes (simulate real data)
		rand.Read(chunk)

		// Write chunk
		n, err := w.Write(chunk)
		if err != nil {
			// Client disconnected
			log.Printf("Download interrupted: %v (sent %d/%d bytes)", err, written, totalBytes)
			return
		}
		written += int64(n)

		// Flush immediately untuk realtime
		flusher.Flush()
	}

	log.Printf("Download complete: %d bytes (%.2f MB)", written, float64(written)/(1024*1024))
}

// ==================== Upload Handler ====================

// uploadHandler - POST /upload
// Menerima data stream dan menghitung speed
// Response: bytes, duration_ms, speed_mbps
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startTime := time.Now()

	// Read all uploaded data
	// Gunakan io.Copy ke io.Discard untuk efisiensi memory
	var totalBytes int64
	buf := make([]byte, ChunkSize)

	for {
		n, err := r.Body.Read(buf)
		totalBytes += int64(n)

		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "Read error: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	duration := time.Since(startTime)
	durationMs := duration.Milliseconds()
	if durationMs == 0 {
		durationMs = 1 // Avoid division by zero
	}

	// Calculate speed: bytes/second -> Mbps
	// Mbps = (bytes * 8) / (duration_seconds * 1_000_000)
	durationSeconds := duration.Seconds()
	if durationSeconds == 0 {
		durationSeconds = 0.001
	}
	speedMbps := (float64(totalBytes) * 8) / (durationSeconds * 1_000_000)

	response := UploadResponse{
		Bytes:      totalBytes,
		DurationMs: durationMs,
		SpeedMbps:  speedMbps,
	}

	log.Printf("Upload complete: %d bytes, %.2f Mbps", totalBytes, speedMbps)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	json.NewEncoder(w).Encode(response)
}

// ==================== Utility Functions ====================

// getServerID returns unique server identifier
func getServerID() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "go-speedtest"
	}
	return hostname
}

// ==================== Main ====================

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	// Register handlers dengan CORS
	http.HandleFunc("/ping", corsMiddleware(pingHandler))
	http.HandleFunc("/download", corsMiddleware(downloadHandler))
	http.HandleFunc("/upload", corsMiddleware(uploadHandler))

	// Health check endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"service": "GO-Speedtest",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	fmt.Printf(`
╔═══════════════════════════════════════════════════════════╗
║             GO-Speedtest Server v1.0.0                    ║
╠═══════════════════════════════════════════════════════════╣
║  Endpoints:                                               ║
║    GET  /           - Server status                       ║
║    GET  /ping       - Latency check                       ║
║    GET  /download   - Download speed test (?size=MB)      ║
║    POST /upload     - Upload speed test                   ║
╠═══════════════════════════════════════════════════════════╣
║  Server running on http://localhost:%s                   ║
╚═══════════════════════════════════════════════════════════╝
`, port)

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
