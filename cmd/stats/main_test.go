package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	servers := make([]*httptest.Server, len(serversPool))
	for i := range servers {
		servers[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data := report{
				"data": []string{"test1", "test2", "test3", "test4", "test5", "test6"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(data)
		}))
		defer servers[i].Close()
	}

	originalServersPool := serversPool
	for i, server := range servers {
		serversPool[i] = server.Listener.Addr().String()
	}
	defer func() { serversPool = originalServersPool }()

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	go main()

	time.Sleep(2 * time.Second)

	logOutput := logBuf.String()
	for _, server := range servers {
		if !contains(logOutput, server.Listener.Addr().String()) {
			t.Errorf("Expected logs to contain server address %s, but got %s", server.Listener.Addr().String(), logOutput)
		}
	}
}

func contains(str, substr string) bool {
	return bytes.Contains([]byte(str), []byte(substr))
}
