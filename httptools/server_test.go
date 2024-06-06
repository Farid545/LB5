package httptools

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateServer(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	server := CreateServer(8080, handler) // 
	ts := httptest.NewServer(handler)
	defer ts.Close()

	go server.Start()

	time.Sleep(1 * time.Second)

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestServer_Start(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := CreateServer(8081, handler) // 

	go server.Start()

	time.Sleep(1 * time.Second)

	resp, err := http.Get("http://localhost:8081") //
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}
