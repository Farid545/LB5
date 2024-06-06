package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/roman-mazur/architecture-practice-4-template/httptools"
	"github.com/roman-mazur/architecture-practice-4-template/signal"
)

var port = flag.Int("port", 8080, "server port") // Define a flag for server port

// Constants for configuration keys
const (
	confResponseDelaySec = "CONF_RESPONSE_DELAY_SEC"
	confHealthFailure    = "CONF_HEALTH_FAILURE"
)

func main() {
	// Initialize HTTP server mux
	h := new(http.ServeMux)

	// Handle health check endpoint
	h.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("content-type", "text/plain")
		// Check if health failure configuration is enabled
		if failConfig := os.Getenv(confHealthFailure); failConfig == "true" {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte("FAILURE"))
		} else {
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte("OK"))
		}
	})

	// Initialize report
	report := make(Report)

	// Handle API endpoint for processing some data
	h.HandleFunc("/api/v1/some-data", func(rw http.ResponseWriter, r *http.Request) {
		// Get response delay configuration
		respDelayString := os.Getenv(confResponseDelaySec)
		// Parse response delay and sleep if valid
		if delaySec, parseErr := strconv.Atoi(respDelayString); parseErr == nil && delaySec > 0 && delaySec < 300 {
			time.Sleep(time.Duration(delaySec) * time.Second)
		}

		// Process the request and update the report
		report.Process(r)

		// Set response headers and status
		rw.Header().Set("content-type", "application/json")
		rw.WriteHeader(http.StatusOK)
		// Encode response data as JSON and write to response writer
		_ = json.NewEncoder(rw).Encode([]string{"1", "2"})
	})

	// Mount report handler
	h.Handle("/report", report)

	// Create and start HTTP server
	server := httptools.CreateServer(*port, h)
	server.Start()

	// Wait for termination signal
	signal.WaitForTerminationSignal()
}
