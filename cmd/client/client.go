package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var target = flag.String("target", "http://localhost:8090", "request target") // Define a flag for request target

func main() {
	flag.Parse() // Parse command-line flags

	// Create an HTTP client with a timeout of 10 seconds
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Perform HTTP GET requests periodically
	for range time.Tick(1 * time.Second) {
		// Send GET request to the specified target URL
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", *target))
		if err == nil {
			// Log the response status code if request is successful
			log.Printf("response %d", resp.StatusCode)
		} else {
			// Log the error if request fails
			log.Printf("error %s", err)
		}
	}
}
