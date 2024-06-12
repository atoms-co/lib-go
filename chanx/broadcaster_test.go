package chanx_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.atoms.co/lib/chanx"
)

func TestBroadcaster_SingleConsumer(t *testing.T) {
	ctx := context.Background()

	b := chanx.NewBroadcaster[int]()
	defer b.Close()

	// (1) Connect consumer
	out, quit := b.Connect()

	// (2) Forward channel
	in := make(chan int, 1)
	b.Forward(ctx, in)

	// (3) Read first update
	in <- 1
	m := read(t, out)
	assert.Equal(t, 1, m)

	// (4) Disconnect
	quit.Close()
	time.Sleep(100 * time.Millisecond)

	// (5) Don't read update
	in <- 2
	dontRead(t, out)
}

func TestBroadcaster_MultiConsumer(t *testing.T) {
	ctx := context.Background()

	b := chanx.NewBroadcaster[int]()
	defer b.Close()

	// (1) Connect consumer 1
	out1, quit1 := b.Connect()

	// (2) Forward channel
	in := make(chan int, 1)
	b.Forward(ctx, in)

	// (3) Read first update
	in <- 1
	m1 := read(t, out1)
	assert.Equal(t, 1, m1)

	// (4) Connect consumer 2, receives first message
	out2, quit2 := b.Connect()
	defer quit2.Close()
	m2 := read(t, out2)
	assert.Equal(t, 1, m2)

	// (5) Disconnect first consumer
	quit1.Close()
	time.Sleep(100 * time.Millisecond)

	// (6) First consumer doesn't read update
	in <- 2
	dontRead(t, out1)
	m3 := read(t, out2)
	assert.Equal(t, 2, m3)
}

func TestBroadcaster_ChangeForward(t *testing.T) {
	ctx := context.Background()

	b := chanx.NewBroadcaster[int]()
	defer b.Close()

	// (1) Connect consumer
	out, quit := b.Connect()

	// (2) Forward channel
	in := make(chan int, 1)
	b.Forward(ctx, in)

	// (3) Read first update
	in <- 1
	m := read(t, out)
	assert.Equal(t, 1, m)

	// (4) Forward new channel
	in2 := make(chan int, 1)
	b.Forward(ctx, in2)

	// (5) Don't read update from first channel
	in <- 2
	dontRead(t, out)

	// (6) Read update from second channel
	in2 <- 3
	m = read(t, out)
	assert.Equal(t, 3, m)

	// (7) Disconnect
	quit.Close()
	time.Sleep(100 * time.Millisecond)

	// (8) Don't read update
	in2 <- 2
	dontRead(t, out)
}

func read(t *testing.T, in <-chan int) int {
	t.Helper()
	select {
	case m := <-in:
		return m
	case <-time.After(1 * time.Second):
		t.Fatalf("failed to read message")
		return -1
	}
}

func dontRead(t *testing.T, in <-chan int) {
	t.Helper()
	select {
	case m, ok := <-in:
		if ok {
			t.Fatalf("unexpected message, %v", m)
		}
	case <-time.After(1 * time.Second):
	}
}
