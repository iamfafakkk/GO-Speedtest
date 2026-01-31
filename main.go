// GO-Speedtest: Realtime Speedtest API Server using speedtest-go library
// Menggunakan showwin/speedtest-go untuk actual internet speed testing
// Endpoints: /speedtest/ping, /speedtest/download, /speedtest/upload

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

// ==================== Constants ====================

const (
	DefaultPort = "8645"
)

// ==================== Response Structs ====================

// PingResponse represents ping test result
type PingResponse struct {
	Status     string  `json:"status"`
	Latency    float64 `json:"latency_ms"`
	ServerID   string  `json:"server_id"`
	ServerName string  `json:"server_name"`
	ServerHost string  `json:"server_host"`
	Country    string  `json:"country"`
	Distance   float64 `json:"distance_km"`
	Timestamp  int64   `json:"timestamp"`
}

// DownloadResponse represents download test result
type DownloadResponse struct {
	SpeedMbps  float64 `json:"speed_mbps"`
	ServerID   string  `json:"server_id"`
	ServerName string  `json:"server_name"`
	ServerHost string  `json:"server_host"`
	Country    string  `json:"country"`
	Latency    float64 `json:"latency_ms"`
	DurationMs int64   `json:"duration_ms"`
	Timestamp  int64   `json:"timestamp"`
}

// UploadResponse represents upload test result
type UploadResponse struct {
	SpeedMbps  float64 `json:"speed_mbps"`
	ServerID   string  `json:"server_id"`
	ServerName string  `json:"server_name"`
	ServerHost string  `json:"server_host"`
	Country    string  `json:"country"`
	Latency    float64 `json:"latency_ms"`
	DurationMs int64   `json:"duration_ms"`
	Timestamp  int64   `json:"timestamp"`
}

// ServerInfo represents minimal server info for response
type ServerInfo struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Host     string  `json:"host"`
	Country  string  `json:"country"`
	Distance float64 `json:"distance_km"`
}

// StreamEvent represents SSE event for realtime progress
type StreamEvent struct {
	Type       string  `json:"type"` // "progress", "complete", "error"
	SpeedMbps  float64 `json:"speed_mbps"`
	Elapsed    float64 `json:"elapsed_sec"`
	ServerID   string  `json:"server_id,omitempty"`
	ServerName string  `json:"server_name,omitempty"`
	Latency    float64 `json:"latency_ms,omitempty"`
	Message    string  `json:"message,omitempty"`
}

// ErrorResponse for API errors
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ==================== CORS Middleware ====================

// corsMiddleware adds CORS headers untuk browser-based clients
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// ==================== Helper Functions ====================

// writeJSON writes JSON response with proper headers
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func writeError(w http.ResponseWriter, status int, errType, message string) {
	writeJSON(w, status, ErrorResponse{
		Error:   errType,
		Message: message,
	})
}

// getClosestServer fetches and returns the closest speedtest server
// Optional serverID query param untuk specific server
func getClosestServer(r *http.Request) (*speedtest.Server, error) {
	client := speedtest.New()

	// Check if specific server ID requested
	serverID := r.URL.Query().Get("server_id")

	if serverID != "" {
		// Fetch specific server by ID
		servers, err := client.FetchServers()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch servers: %w", err)
		}

		id, err := strconv.Atoi(serverID)
		if err != nil {
			return nil, fmt.Errorf("invalid server_id: %w", err)
		}

		targets, err := servers.FindServer([]int{id})
		if err != nil || len(targets) == 0 {
			return nil, fmt.Errorf("server with ID %s not found", serverID)
		}

		return targets[0], nil
	}

	// Get list of servers and find closest
	servers, err := client.FetchServers()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch servers: %w", err)
	}

	// FindServer with empty slice returns closest servers
	targets, err := servers.FindServer([]int{})
	if err != nil || len(targets) == 0 {
		return nil, fmt.Errorf("no available servers found")
	}

	// Return the closest server (first in list)
	return targets[0], nil
}

// ==================== Ping Handler ====================

// speedtestPingHandler - GET /speedtest/ping
// Tests latency to the closest speedtest server
// Optional query: ?server_id=12345 untuk specific server
func speedtestPingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed")
		return
	}

	log.Printf("[PING] Starting ping test...")

	server, err := getClosestServer(r)
	if err != nil {
		log.Printf("[PING] Error finding server: %v", err)
		writeError(w, http.StatusServiceUnavailable, "server_error", err.Error())
		return
	}

	// Perform ping test
	err = server.PingTest(nil)
	if err != nil {
		log.Printf("[PING] Ping test failed: %v", err)
		writeError(w, http.StatusServiceUnavailable, "ping_failed", err.Error())
		return
	}

	response := PingResponse{
		Status:     "success",
		Latency:    float64(server.Latency.Milliseconds()),
		ServerID:   server.ID,
		ServerName: server.Name,
		ServerHost: server.Host,
		Country:    server.Country,
		Distance:   server.Distance,
		Timestamp:  time.Now().UnixMilli(),
	}

	log.Printf("[PING] Complete - Server: %s (%s), Latency: %.2fms",
		server.Name, server.Country, response.Latency)

	writeJSON(w, http.StatusOK, response)
}

// ==================== Download Handler ====================

// speedtestDownloadHandler - GET /speedtest/download
// Tests download speed to the closest speedtest server
// Optional query: ?server_id=12345 untuk specific server
func speedtestDownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed")
		return
	}

	log.Printf("[DOWNLOAD] Starting download test...")

	server, err := getClosestServer(r)
	if err != nil {
		log.Printf("[DOWNLOAD] Error finding server: %v", err)
		writeError(w, http.StatusServiceUnavailable, "server_error", err.Error())
		return
	}

	// Ping first untuk get latency
	err = server.PingTest(nil)
	if err != nil {
		log.Printf("[DOWNLOAD] Ping test failed: %v", err)
		writeError(w, http.StatusServiceUnavailable, "ping_failed", err.Error())
		return
	}

	startTime := time.Now()

	// Perform download test
	err = server.DownloadTest()
	if err != nil {
		log.Printf("[DOWNLOAD] Download test failed: %v", err)
		writeError(w, http.StatusServiceUnavailable, "download_failed", err.Error())
		return
	}

	duration := time.Since(startTime)

	// DLSpeed is in bytes per second, convert to Mbps
	speedMbps := float64(server.DLSpeed) / 1_000_000 * 8

	response := DownloadResponse{
		SpeedMbps:  speedMbps,
		ServerID:   server.ID,
		ServerName: server.Name,
		ServerHost: server.Host,
		Country:    server.Country,
		Latency:    float64(server.Latency.Milliseconds()),
		DurationMs: duration.Milliseconds(),
		Timestamp:  time.Now().UnixMilli(),
	}

	log.Printf("[DOWNLOAD] Complete - Server: %s, Speed: %.2f Mbps, Duration: %dms",
		server.Name, speedMbps, duration.Milliseconds())

	// Reset server context untuk cleanup
	server.Context.Reset()

	writeJSON(w, http.StatusOK, response)
}

// ==================== Upload Handler ====================

// speedtestUploadHandler - GET /speedtest/upload
// Tests upload speed to the closest speedtest server
// Note: Using GET for simplicity (actual upload data handled by speedtest-go)
// Optional query: ?server_id=12345 untuk specific server
func speedtestUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed")
		return
	}

	log.Printf("[UPLOAD] Starting upload test...")

	server, err := getClosestServer(r)
	if err != nil {
		log.Printf("[UPLOAD] Error finding server: %v", err)
		writeError(w, http.StatusServiceUnavailable, "server_error", err.Error())
		return
	}

	// Ping first untuk get latency
	err = server.PingTest(nil)
	if err != nil {
		log.Printf("[UPLOAD] Ping test failed: %v", err)
		writeError(w, http.StatusServiceUnavailable, "ping_failed", err.Error())
		return
	}

	startTime := time.Now()

	// Perform upload test
	err = server.UploadTest()
	if err != nil {
		log.Printf("[UPLOAD] Upload test failed: %v", err)
		writeError(w, http.StatusServiceUnavailable, "upload_failed", err.Error())
		return
	}

	duration := time.Since(startTime)

	// ULSpeed is in bytes per second, convert to Mbps
	speedMbps := float64(server.ULSpeed) / 1_000_000 * 8

	response := UploadResponse{
		SpeedMbps:  speedMbps,
		ServerID:   server.ID,
		ServerName: server.Name,
		ServerHost: server.Host,
		Country:    server.Country,
		Latency:    float64(server.Latency.Milliseconds()),
		DurationMs: duration.Milliseconds(),
		Timestamp:  time.Now().UnixMilli(),
	}

	log.Printf("[UPLOAD] Complete - Server: %s, Speed: %.2f Mbps, Duration: %dms",
		server.Name, speedMbps, duration.Milliseconds())

	// Reset server context untuk cleanup
	server.Context.Reset()

	writeJSON(w, http.StatusOK, response)
}

// ==================== Server List Handler ====================

// speedtestServersHandler - GET /speedtest/servers
// Returns list of available speedtest servers
func speedtestServersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed")
		return
	}

	log.Printf("[SERVERS] Fetching server list...")

	client := speedtest.New()
	servers, err := client.FetchServers()
	if err != nil {
		log.Printf("[SERVERS] Error fetching servers: %v", err)
		writeError(w, http.StatusServiceUnavailable, "server_error", err.Error())
		return
	}

	// Limit to top 10 closest servers
	limit := 10
	if len(servers) < limit {
		limit = len(servers)
	}

	var serverList []ServerInfo
	for i := 0; i < limit; i++ {
		s := servers[i]
		serverList = append(serverList, ServerInfo{
			ID:       s.ID,
			Name:     s.Name,
			Host:     s.Host,
			Country:  s.Country,
			Distance: s.Distance,
		})
	}

	log.Printf("[SERVERS] Found %d servers", len(serverList))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count":   len(serverList),
		"servers": serverList,
	})
}

// ==================== SSE Helper ====================

// sendSSE sends a Server-Sent Event
func sendSSE(w http.ResponseWriter, flusher http.Flusher, event StreamEvent) {
	data, _ := json.Marshal(event)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// ==================== Download Stream Handler ====================

// SSE streaming untuk realtime download progress
// Query params:
//   - server_id: optional server ID
//   - duration: test duration in seconds (default: 10, max: 30)
func speedtestDownloadStreamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed")
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Parse duration parameter (in seconds)
	durationStr := r.URL.Query().Get("duration")
	testDuration := 10 // default 10 seconds
	if durationStr != "" {
		if parsed, err := strconv.Atoi(durationStr); err == nil && parsed > 0 {
			testDuration = parsed
		}
	}
	if testDuration > 30 {
		testDuration = 30 // max 30 seconds
	}

	log.Printf("[DOWNLOAD STREAM] Starting %ds download test...", testDuration)

	server, err := getClosestServer(r)
	if err != nil {
		sendSSE(w, flusher, StreamEvent{Type: "error", Message: err.Error()})
		return
	}

	// Ping test first
	err = server.PingTest(nil)
	if err != nil {
		sendSSE(w, flusher, StreamEvent{Type: "error", Message: "Ping failed: " + err.Error()})
		return
	}

	latency := float64(server.Latency.Milliseconds())
	startTime := time.Now()

	// Send initial event
	sendSSE(w, flusher, StreamEvent{
		Type:       "start",
		ServerID:   server.ID,
		ServerName: server.Name,
		Latency:    latency,
	})

	// Channel for realtime speed updates from callback
	speedChan := make(chan float64, 100)
	done := make(chan struct{})

	// Set callback for realtime download speed
	server.Context.SetCallbackDownload(func(rate speedtest.ByteRate) {
		// ByteRate is bytes per second, convert to Mbps
		speedMbps := float64(rate) / 1_000_000 * 8
		select {
		case speedChan <- speedMbps:
		default:
			// Channel full, skip this update
		}
	})

	// Run download test in goroutine
	go func() {
		defer close(done)
		server.DownloadTest()
	}()

	// Stream progress
	ctx := r.Context()
	testDeadline := time.Now().Add(time.Duration(testDuration) * time.Second)
	var lastSpeed float64

	for {
		select {
		case <-ctx.Done():
			server.Context.Reset()
			return
		case <-done:
			// Test completed
			finalSpeed := float64(server.DLSpeed) / 1_000_000 * 8
			if finalSpeed == 0 {
				finalSpeed = lastSpeed
			}
			elapsed := time.Since(startTime).Seconds()
			sendSSE(w, flusher, StreamEvent{
				Type:       "complete",
				SpeedMbps:  finalSpeed,
				Elapsed:    elapsed,
				ServerID:   server.ID,
				ServerName: server.Name,
				Latency:    latency,
			})
			server.Context.Reset()
			return
		case speed := <-speedChan:
			lastSpeed = speed
			if time.Now().After(testDeadline) {
				// Force stop after duration
				elapsed := time.Since(startTime).Seconds()
				sendSSE(w, flusher, StreamEvent{
					Type:       "complete",
					SpeedMbps:  speed,
					Elapsed:    elapsed,
					ServerID:   server.ID,
					ServerName: server.Name,
					Latency:    latency,
				})
				server.Context.Reset()
				return
			}
			// Send progress update
			elapsed := time.Since(startTime).Seconds()
			sendSSE(w, flusher, StreamEvent{
				Type:      "progress",
				SpeedMbps: speed,
				Elapsed:   elapsed,
			})
		case <-time.After(200 * time.Millisecond):
			// Timeout waiting for speed, check deadline
			if time.Now().After(testDeadline) {
				elapsed := time.Since(startTime).Seconds()
				sendSSE(w, flusher, StreamEvent{
					Type:       "complete",
					SpeedMbps:  lastSpeed,
					Elapsed:    elapsed,
					ServerID:   server.ID,
					ServerName: server.Name,
					Latency:    latency,
				})
				server.Context.Reset()
				return
			}
		}
	}
}

// ==================== Upload Stream Handler ====================

// speedtestUploadStreamHandler - GET /speedtest/upload/stream
// SSE streaming untuk realtime upload progress
// Query params:
//   - server_id: optional server ID
//   - duration: test duration in seconds (default: 10, max: 30)
func speedtestUploadStreamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed")
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Parse duration parameter (in seconds)
	durationStr := r.URL.Query().Get("duration")
	testDuration := 10 // default 10 seconds
	if durationStr != "" {
		if parsed, err := strconv.Atoi(durationStr); err == nil && parsed > 0 {
			testDuration = parsed
		}
	}
	if testDuration > 30 {
		testDuration = 30 // max 30 seconds
	}

	log.Printf("[UPLOAD STREAM] Starting %ds upload test...", testDuration)

	server, err := getClosestServer(r)
	if err != nil {
		sendSSE(w, flusher, StreamEvent{Type: "error", Message: err.Error()})
		return
	}

	// Ping test first
	err = server.PingTest(nil)
	if err != nil {
		sendSSE(w, flusher, StreamEvent{Type: "error", Message: "Ping failed: " + err.Error()})
		return
	}

	latency := float64(server.Latency.Milliseconds())
	startTime := time.Now()

	// Send initial event
	sendSSE(w, flusher, StreamEvent{
		Type:       "start",
		ServerID:   server.ID,
		ServerName: server.Name,
		Latency:    latency,
	})

	// Channel for realtime speed updates from callback
	speedChan := make(chan float64, 100)
	done := make(chan struct{})

	// Set callback for realtime upload speed
	server.Context.SetCallbackUpload(func(rate speedtest.ByteRate) {
		// ByteRate is bytes per second, convert to Mbps
		speedMbps := float64(rate) / 1_000_000 * 8
		select {
		case speedChan <- speedMbps:
		default:
			// Channel full, skip this update
		}
	})

	// Run upload test in goroutine
	go func() {
		defer close(done)
		server.UploadTest()
	}()

	// Stream progress
	ctx := r.Context()
	testDeadline := time.Now().Add(time.Duration(testDuration) * time.Second)
	var lastSpeed float64

	for {
		select {
		case <-ctx.Done():
			server.Context.Reset()
			return
		case <-done:
			// Test completed
			finalSpeed := float64(server.ULSpeed) / 1_000_000 * 8
			if finalSpeed == 0 {
				finalSpeed = lastSpeed
			}
			elapsed := time.Since(startTime).Seconds()
			sendSSE(w, flusher, StreamEvent{
				Type:       "complete",
				SpeedMbps:  finalSpeed,
				Elapsed:    elapsed,
				ServerID:   server.ID,
				ServerName: server.Name,
				Latency:    latency,
			})
			server.Context.Reset()
			return
		case speed := <-speedChan:
			lastSpeed = speed
			if time.Now().After(testDeadline) {
				// Force stop after duration
				elapsed := time.Since(startTime).Seconds()
				sendSSE(w, flusher, StreamEvent{
					Type:       "complete",
					SpeedMbps:  speed,
					Elapsed:    elapsed,
					ServerID:   server.ID,
					ServerName: server.Name,
					Latency:    latency,
				})
				server.Context.Reset()
				return
			}
			// Send progress update
			elapsed := time.Since(startTime).Seconds()
			sendSSE(w, flusher, StreamEvent{
				Type:      "progress",
				SpeedMbps: speed,
				Elapsed:   elapsed,
			})
		case <-time.After(200 * time.Millisecond):
			// Timeout waiting for speed, check deadline
			if time.Now().After(testDeadline) {
				elapsed := time.Since(startTime).Seconds()
				sendSSE(w, flusher, StreamEvent{
					Type:       "complete",
					SpeedMbps:  lastSpeed,
					Elapsed:    elapsed,
					ServerID:   server.ID,
					ServerName: server.Name,
					Latency:    latency,
				})
				server.Context.Reset()
				return
			}
		}
	}
}

// ==================== Main ====================

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	// Speedtest endpoints dengan CORS
	http.HandleFunc("/speedtest/ping", corsMiddleware(speedtestPingHandler))
	http.HandleFunc("/speedtest/download", corsMiddleware(speedtestDownloadHandler))
	http.HandleFunc("/speedtest/upload", corsMiddleware(speedtestUploadHandler))
	http.HandleFunc("/speedtest/servers", corsMiddleware(speedtestServersHandler))

	// SSE Streaming endpoints
	http.HandleFunc("/speedtest/download/stream", corsMiddleware(speedtestDownloadStreamHandler))
	http.HandleFunc("/speedtest/upload/stream", corsMiddleware(speedtestUploadStreamHandler))

	// Health check endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"service": "GO-Speedtest",
			"version": "2.1.0",
			"status":  "running",
			"library": "speedtest-go v1.7.10",
		})
	})

	fmt.Printf(`
╔═══════════════════════════════════════════════════════════════════╗
║           GO-Speedtest Server v2.1.0                              ║
║           Powered by speedtest-go library                         ║
╠═══════════════════════════════════════════════════════════════════╣
║  Endpoints:                                                       ║
║    GET  /                         - Server status                 ║
║    GET  /speedtest/ping           - Latency test                  ║
║    GET  /speedtest/download       - Download speed (JSON)         ║
║    GET  /speedtest/upload         - Upload speed (JSON)           ║
║    GET  /speedtest/servers        - List available servers        ║
║                                                                   ║
║  Realtime SSE Streaming:                                          ║
║    GET  /speedtest/download/stream - Download (SSE)               ║
║    GET  /speedtest/upload/stream   - Upload (SSE)                 ║
║                                                                   ║
║  Query Parameters:                                                ║
║    ?server_id=12345  - Test against specific server               ║
║    ?duration=15      - Test duration in seconds (SSE, max 30)     ║
╠═══════════════════════════════════════════════════════════════════╣
║  Server running on http://localhost:%s                           ║
╚═══════════════════════════════════════════════════════════════════╝
`, port)

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
