// Package grpcx contain utilities for working with grpc.
package grpcx

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.atoms.co/lib/statshandlerx"
)

// WithInsecure returns an insecure transport credential option. Convenience replacement for the deprecated
// grpc.WithInsecure.
func WithInsecure() grpc.DialOption {
	return grpc.WithTransportCredentials(insecure.NewCredentials())
}

// Dial makes a blocking grpc dial with a timeout & tracing.
func Dial(ctx context.Context, endpoint string, timeout time.Duration, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return DialNonBlocking(ctx, endpoint, append(opts, grpc.WithBlock())...)
}

// DialNonBlocking makes a non-blocking grpc dial with tracing.
func DialNonBlocking(ctx context.Context, endpoint string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	cc, err := grpc.DialContext(ctx, endpoint, append(opts, statshandlerx.WithClientGRPCStatsHandler())...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to dial server at %v", endpoint)
	}
	return cc, nil
}
