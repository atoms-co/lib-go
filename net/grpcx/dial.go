// Package grpx contain utilities for working with grpc.
package grpcx

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Dial makes a blocking grpc dial with a timeout.
func Dial(ctx context.Context, endpoint string, timeout time.Duration, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cc, err := grpc.DialContext(ctx, endpoint, append(opts, grpc.WithBlock())...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to dial server at %v", endpoint)
	}
	return cc, nil
}
