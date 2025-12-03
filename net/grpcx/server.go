package grpcx

import (
	"context"
	"net"
	"sync"

	"google.golang.org/grpc"
)

// Serve starts the server on the given listener. Blocking. Exits gracefully on context cancellation.
func Serve(ctx context.Context, gs *grpc.Server, listener net.Listener) error {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-ctx.Done()
		gs.GracefulStop()
	}()

	if err := gs.Serve(listener); err != nil {
		return err
	}
	wg.Wait()
	return nil
}
