package signal

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

// TestWaitForTerminationSignal tests the WaitForTerminationSignal function.
func TestWaitForTerminationSignal(t *testing.T) {
	// Setup a channel to receive OS signals.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Simulate sending a signal after a short delay.
	go func() {
		time.Sleep(1 * time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	// Call the function that waits for the termination signal.
	go WaitForTerminationSignal()

	// Wait for the signal.
	sig := <-sigChan
	if sig != syscall.SIGINT {
		t.Fatalf("Expected SIGINT, but got %v", sig)
	}
}
