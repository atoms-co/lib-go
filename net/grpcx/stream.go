package grpcx

import (
	"context"
	"go.cloudkitchens.org/lib/log"
	"go.cloudkitchens.org/lib/contextx"
	"go.cloudkitchens.org/lib/iox"
	"google.golang.org/grpc"
)

const (
	// bufChanSize is the interal buffer size for streaming handler utilities.
	bufChanSize = 20
)

// Handler is a low-level bidirectional protobuf message handler. Connection closure is realized as a chan closure
// in either direction. The hander _may_ consume input before determining whether to accept, but some side must
// obviously take the initiative.
type Handler[A, B any] func(ctx context.Context, in <-chan A) (<-chan B, error)

// Receive is a server handler for the given stream. The context must be the server context. Blocking.
func Receive[A, B any, S Stream[A, B]](ctx context.Context, server S, fn Handler[A, B]) error {
	quit := iox.WithCancel(ctx, iox.NewAsyncCloser())      // ctx closed => quit
	wctx, _ := contextx.WithQuitCancel(ctx, quit.Closed()) // quit -> wctx closed
	defer quit.Close()

	// Start reading async first to allow the server to see any registration messages.

	in := make(chan A, bufChanSize)
	defer close(in)

	go func() {
		defer quit.Close()

		for !quit.IsClosed() {
			msg, err := server.Recv()
			if err != nil {
				if quit.IsClosed() {
					return
				}
				log.Warnf(ctx, "Recv failed: %v", err)
				return
			}

			select {
			case in <- msg:
				// ok
			case <-quit.Closed():
				return
			}
		}
	}()

	out, err := fn(wctx, in)
	if err != nil {
		return err
	}

	for {
		select {
		case msg, ok := <-out:
			if !ok {
				return nil
			}

			if err := server.Send(msg); err != nil {
				if quit.IsClosed() {
					return nil
				}
				log.Warnf(ctx, "Send failed: %v", err)
				return nil
			}
		case <-quit.Closed():
			return nil
		}
	}
}

// Stream is the high-level grpx streaming interface, which matches the generated types for streaming methods.
type Stream[A, B any] interface {
	Send(B) error
	Recv() (A, error)
}

// Connect connects the client handler to a compatible grpc streaming service method. Stopped by context
// cancellation or either side. Blocking.
func Connect[A, B any, S Stream[A, B]](ctx context.Context, con func(context.Context, ...grpc.CallOption) (S, error), fn Handler[A, B], opts ...grpc.CallOption) error {
	quit := iox.WithCancel(ctx, iox.NewAsyncCloser())      // ctx closed => quit
	wctx, _ := contextx.WithQuitCancel(ctx, quit.Closed()) // quit -> wctx closed
	defer quit.Close()

	client, err := con(wctx, opts...)
	if err != nil {
		return err
	}

	in := make(chan A, bufChanSize)
	defer close(in)

	out, err := fn(wctx, in)
	if err != nil {
		return err
	}

	go func() {
		defer quit.Close()

		for {
			select {
			case msg, ok := <-out:
				if !ok {
					return
				}

				if err := client.Send(msg); err != nil {
					if quit.IsClosed() {
						return
					}
					log.Warnf(ctx, "Send failed: %v", err)
					return
				}
			case <-quit.Closed():
				return
			}
		}
	}()

	for !quit.IsClosed() {
		msg, err := client.Recv()
		if err != nil {
			if quit.IsClosed() {
				return nil
			}
			log.Warnf(ctx, "Send failed: %v", err)
			return nil
		}

		select {
		case in <- msg:
			// ok
		case <-quit.Closed():
			return nil
		}
	}
	return nil
}

// ConnectNonBlocking returns a connection type T, if the connection is successful. Non-blocking.
func ConnectNonBlocking[T, A, B any, S Stream[A, B]](ctx context.Context, con func(context.Context, ...grpc.CallOption) (S, error), fn func(context.Context, <-chan A) (T, <-chan B, error), opts ...grpc.CallOption) (T, error) {
	var ret T
	var err error

	wait := iox.NewAsyncCloser()

	go func() {
		defer wait.Close()

		err2 := Connect(ctx, con, func(ctx context.Context, in <-chan A) (<-chan B, error) {
			defer wait.Close()

			var out <-chan B
			ret, out, err = fn(ctx, in)
			return out, err
		}, opts...)

		if !wait.IsClosed() {
			err = err2 // propagate err: we did not get to handler call
		}
	}()

	<-wait.Closed()

	return ret, err
}

// ShortCircuit connects two handlers directly, without any grpc server. The client is assumed to initiate the
// exchange. Stopped by context cancellation or any of the handlers. Blocking.
func ShortCircuit[A, B any](ctx context.Context, client Handler[A, B], server Handler[B, A]) error {
	quit := iox.WithCancel(ctx, iox.NewAsyncCloser())
	defer quit.Close()

	// (1) Create buffer for client and connect

	a := make(chan A, bufChanSize)
	defer close(a)

	out, err := client(ctx, a)
	if err != nil {
		return err
	}

	// (2) Create buffer for server and connect. Fill buffer async to let the server peek at messages before
	// accepting the connection -- this assumes the client sends messages first. Otherwise, reverse the roles.

	b := make(chan B, bufChanSize)

	go func() {
		defer quit.Close()
		defer close(b)

		for {
			select {
			case msg, ok := <-out:
				if !ok {
					return
				}

				select {
				case b <- msg:
					// ok
				case <-quit.Closed():
					return
				}
			case <-quit.Closed():
				return
			}
		}
	}()

	in, err := server(ctx, b)
	if err != nil {
		return err
	}

	// (3) Forward server -> client message sync to be blocking.

	for {
		select {
		case msg, ok := <-in:
			if !ok {
				return nil
			}

			select {
			case a <- msg:
				// ok
			case <-quit.Closed():
				return nil
			}

		case <-quit.Closed():
			return nil
		}
	}
}
