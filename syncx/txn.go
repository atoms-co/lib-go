package syncx

import "context"

// TxnFn attempts to inject a function into a transactional context. Used in single-threaded components.
type TxnFn func(context.Context, func() error) error

// AsyncTxn is a convenience wrapper for txn where we want to ignore the error without context cancellation.
func AsyncTxn(txn TxnFn, fn func()) {
	_ = txn(context.Background(), func() error {
		fn()
		return nil
	})
}

// Txn0 is a convenience wrapper for txn where we want to ignore the error.
func Txn0(ctx context.Context, txn TxnFn, fn func()) {
	_ = txn(ctx, func() error {
		fn()
		return nil
	})
}

// Txn1 is a convenience wrapper for txn with 1 return value of type T.
func Txn1[T any](ctx context.Context, txn TxnFn, fn func() (T, error)) (T, error) {
	var ret T
	var err error

	err = txn(ctx, func() error {
		ret, err = fn()
		return err
	})
	return ret, err
}
