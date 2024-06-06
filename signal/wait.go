package signal

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// WaitForTerminationSignal waits for SIGINT or SIGTERM signals.
func WaitForTerminationSignal() {
	intChannel := make(chan os.Signal, 1)
	signal.Notify(intChannel, syscall.SIGINT, syscall.SIGTERM)
	<-intChannel
	log.Println("Shutting down...")
}
