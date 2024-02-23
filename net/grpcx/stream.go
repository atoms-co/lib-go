package grpcx

import (
	"context"
	"time"

	"go.uber.org/atomic"
	"google.golang.org/grpc"

	"go.atoms.co/lib/log"
	"go.atoms.co/lib/chanx"
	"go.atoms.co/lib/contextx"
	"go.atoms.co/lib/iox"
)

const (
	// bufChanSize is the interal buffer size for streaming handler utilities.
	bufChanSize = 20
	// contextCancelDelay is the delay on cancelling the gRPC context when a user initiates a cancel.
	// This gives time for messages to flush properly.
	// TODO(jhhurwitz): 12/21/24 A more elegant solution to this problem
	contextCancelDelay = 100 * time.Millisecond
)

// Handler is a low-level bidirectional protobuf message handler. Connection closure is realized as a chan closure
// in either direction. The hander _may_ consume input before determining whether to accept, but some side must
// obviously take the initiative. The handler must close the output chan.
type Handler[A, B any] func(ctx context.Context, in <-chan A) (<-chan B, error)

// Stream is the high-level grpcx streaming interface, which matches the generated types for streaming methods.
type Stream[A, B any] interface {
	Send(B) error
	Recv() (A, error)
}

// Receive is a server handler for the given stream. The context must be the server context. Blocking.
func Receive[A, B any, S Stream[A, B]](octx context.Context, server S, fn Handler[A, B]) error {
	ctx, _ := contextx.WithQuitCancelDelay(context.Background(), octx.Done(), contextCancelDelay)

	quit := iox.WithCancel(ctx, iox.NewAsyncCloser())                               // ctx closed => quit
	wctx, _ := contextx.WithQuitCancelDelay(ctx, quit.Closed(), contextCancelDelay) // quit -> wctx closed
	defer quit.Close()

	// Start reading async first to allow the server to see any registration messages.

	in := make(chan A, bufChanSize)

	go func() {
		defer quit.Close()
		defer close(in)

		for !quit.IsClosed() {
			msg, err := server.Recv()
			if err != nil {
				if quit.IsClosed() || contextx.IsCancelled(ctx) {
					return
				}
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

	// Run server function and ensure all messages are emitted, unless there is a failure.

	out, err := fn(wctx, in)
	if err != nil {
		return err
	}

	for msg := range out {
		if err := server.Send(msg); err != nil {
			if quit.IsClosed() || contextx.IsCancelled(ctx) {
				break
			}
			log.Warnf(ctx, "Send failed: %v", err)
			break
		}
	}

	go chanx.Drain(out)
	return nil
}

// Connect connects the client handler to a compatible grpc streaming service method. Stopped by context
// cancellation or either side. Blocking.
func Connect[A, B any, S Stream[A, B]](octx context.Context, con func(context.Context, ...grpc.CallOption) (S, error), fn Handler[A, B], opts ...grpc.CallOption) error {
	ctx, _ := contextx.WithQuitCancelDelay(context.Background(), octx.Done(), contextCancelDelay)

	quit := iox.WithCancel(ctx, iox.NewAsyncCloser())                               // ctx closed => quit
	wctx, _ := contextx.WithQuitCancelDelay(ctx, quit.Closed(), contextCancelDelay) // quit -> wctx closed
	defer quit.Close()

	client, err := con(wctx, opts...)
	if err != nil {
		return err
	}

	in := make(chan A)

	out, err := fn(wctx, chanx.Breaker(in, quit, bufChanSize))
	if err != nil {
		return err
	}

	var recvErr atomic.Error

	go func() {
		defer close(in)

		for !quit.IsClosed() {
			msg, err := client.Recv()
			if err != nil {
				if quit.IsClosed() {
					return
				}
				recvErr.Store(err)

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

	// Ensure all messages are emitted to the client, unless there is a failure.

	for msg := range out {
		if err := client.Send(msg); err != nil {
			if quit.IsClosed() || contextx.IsCancelled(ctx) {
				break
			}
			log.Warnf(ctx, "Send failed: %v", err)
			break
		}
	}

	go chanx.Drain(out)
	return recvErr.Load()
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
		} else if err2 != nil {
			log.Warnf(ctx, "Stream failed after connecting: %v", err2)
		}
	}()

	<-wait.Closed()

	return ret, err
}

// ShortCircuit connects two handlers directly, without any grpc server. The client is assumed to initiate the
// exchange. Stopped by context cancellation or any of the handlers. Blocking.
func ShortCircuit[A, B any](octx context.Context, client Handler[A, B], server Handler[B, A]) error {
	ctx, cancel := contextx.WithQuitCancelDelay(context.Background(), octx.Done(), contextCancelDelay)
	defer cancel()

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

	in, err := server(ctx, chanx.Breaker(out, quit, bufChanSize))
	if err != nil {
		return err
	}

	// (3) Forward server -> client message sync to be blocking.

	for msg := range in {
		select {
		case a <- msg:
			// ok
		case <-quit.Closed():
			break
		}
	}

	go chanx.Drain(in)
	return nil
}
