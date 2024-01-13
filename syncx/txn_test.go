package syncx_test

import (
	"context"
	"errors"
	"testing"

	"go.atoms.co/lib/testing/assertx"
	"go.atoms.co/lib/syncx"
)

var (
	ErrSkipped = errors.New("skipped")
	ErrFailed  = errors.New("failed")
)

func TestTxn(t *testing.T) {
	ctx := context.Background()

	callTxn := func(ctx context.Context, fn func() error) error {
		return fn()
	}
	skipTxn := func(ctx context.Context, fn func() error) error {
		return ErrSkipped // the function may never be called
	}

	t.Run("txn1", func(t *testing.T) {
		one := func() (int, error) { return 1, nil }
		fail := func() (int, error) { return 0, ErrFailed }

		tests := []struct {
			txn syncx.TxnFn
			fn  func() (int, error)
			n   int
			err error
		}{
			{callTxn, one, 1, nil},
			{callTxn, fail, 0, ErrFailed},
			{skipTxn, one, 0, ErrSkipped},
			{skipTxn, fail, 0, ErrSkipped},
		}

		for _, tt := range tests {
			actual, err := syncx.Txn1(ctx, tt.txn, tt.fn)
			assertx.Equal(t, actual, tt.n)
			assertx.Equal(t, err, tt.err)
		}

		for _, tt := range tests {
			actual, _, err := syncx.Txn2(ctx, tt.txn, func() (int, int, error) {
				n, err := tt.fn()
				return n, n, err
			})
			assertx.Equal(t, actual, tt.n)
			assertx.Equal(t, err, tt.err)
		}

	})

	t.Run("txn0", func(t *testing.T) {
		v := 1
		syncx.Txn0(ctx, callTxn, func() { v++ })
		assertx.Equal(t, v, 2)

		syncx.Txn0(ctx, skipTxn, func() { v++ })
		assertx.Equal(t, v, 2)
	})
}
