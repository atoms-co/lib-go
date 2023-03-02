package signalx

import (
	"go.cloudkitchens.org/lib/iox"
	"os"
	"os/signal"
	"syscall"
)

// InterruptChan returns a channel that receives a signal when the process receives either syscall.SIGTERM or syscall.SIGINT
func InterruptChan() <-chan os.Signal {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	return stopChan
}

// InterruptCloser returns a closer that is closed when the process receives either syscall.SIGTERM or syscall.SIGINT
func InterruptCloser() iox.AsyncCloser {
	closer := iox.NewAsyncCloser()
	stopChan := InterruptChan()
	go func() {
		<-stopChan
		closer.Close()
	}()
	return closer
}
