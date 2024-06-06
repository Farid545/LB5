package integration

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

const baseAddress = "http://balancer:8090"

var client = http.Client{
	Timeout: 3 * time.Second,
}

func TestBalancer(t *testing.T) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		t.Skip("Integration test is not enabled")
	}

	url := fmt.Sprintf("%s/api/v1/some-data", baseAddress)
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Failed to get response: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, but got %d", resp.StatusCode)
	}

	lbFrom := resp.Header.Get("lb-from")
	if lbFrom == "" {
		t.Fatalf("Expected lb-from header, but got empty")
	}
	t.Logf("response from [%s]", lbFrom)
}

func BenchmarkBalancer(b *testing.B) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		b.Skip("Integration test is not enabled")
	}

	url := fmt.Sprintf("%s/api/v1/some-data", baseAddress)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(url)
		if err != nil {
			b.Fatalf("Failed to get response: %v", err)
		}
		resp.Body.Close()
	}
}
