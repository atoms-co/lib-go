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

// Txn is a convenience wrapper for txn where we want to return one value of type T.
// TODO (styurin) 10/27/2023: rename to be consistent with Txn1
func Txn[T any](ctx context.Context, txn TxnFn, fn func() T) (T, error) {
	var ret T
	err := txn(ctx, func() error {
		ret = fn()
		return nil
	})
	return ret, err
}

// Txn1 is a convenience wrapper for txn with 1 return value of type T.
func Txn1[T any](ctx context.Context, txn TxnFn, fn func() (T, error)) (T, error) {
	var ret T
	var err error

	Txn0(ctx, txn, func() {
		ret, err = fn()
	})
	return ret, err
}

// Txn2 is a convenience wrapper for txn with 2 return values of type T1 and T2.
func Txn2[T1, T2 any](ctx context.Context, txn TxnFn, fn func() (T1, T2, error)) (T1, T2, error) {
	var ret1 T1
	var ret2 T2
	var err error

	Txn0(ctx, txn, func() {
		ret1, ret2, err = fn()
	})
	return ret1, ret2, err
}
