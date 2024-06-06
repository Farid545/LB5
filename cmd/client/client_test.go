package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	}))
	defer testServer.Close()

	*target = testServer.URL

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)

	go main()

	time.Sleep(2 * time.Second)

	logOutput := logBuf.String()
	if !contains(logOutput, "response 200") {
		t.Errorf("Expected log to contain 'response 200', but got %s", logOutput)
	}
}

func contains(str, substr string) bool {
	return bytes.Contains([]byte(str), []byte(substr))
}
