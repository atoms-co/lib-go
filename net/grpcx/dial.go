// Package grpcx contain utilities for working with grpc.
package grpcx

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.atoms.co/lib/statshandlerx"
)

// MaxMessageSize is a proposed default max message size of 64MB, instead of the 4MB.
const MaxMessageSize = 64 * 1024 * 1024

// WithInsecure returns an insecure transport credential option. Convenience replacement for the deprecated
// grpc.WithInsecure.
func WithInsecure() grpc.DialOption {
	return grpc.WithTransportCredentials(insecure.NewCredentials())
}

// WithMaxMessageSize returns a default call option for both send and receive with the given limit.
// Convenience wrapper.
func WithMaxMessageSize(limit int) grpc.DialOption {
	return grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(limit), grpc.MaxCallSendMsgSize(limit))
}

// Dial makes a blocking grpc dial with a timeout and tracing.
func Dial(ctx context.Context, endpoint string, timeout time.Duration, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return DialNonBlocking(ctx, endpoint, append(opts, grpc.WithBlock())...)
}

// DialNonBlocking makes a non-blocking grpc dial with tracing.
func DialNonBlocking(ctx context.Context, endpoint string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	cc, err := grpc.DialContext(ctx, endpoint, append(opts, statshandlerx.WithClientGRPCStatsHandler())...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server at %v: %w", endpoint, err)
	}
	return cc, nil
}

// Dial64 makes a blocking grpc dial with a timeout, tracing and 64mb limit.
func Dial64(ctx context.Context, endpoint string, timeout time.Duration, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return Dial(ctx, endpoint, timeout, append([]grpc.DialOption{WithMaxMessageSize(MaxMessageSize)}, opts...)...)
}

// DialNonBlocking64 makes a non-blocking grpc dial with a timeout, tracing and 64mb limit.
func DialNonBlocking64(ctx context.Context, endpoint string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return DialNonBlocking(ctx, endpoint, append([]grpc.DialOption{WithMaxMessageSize(MaxMessageSize)}, opts...)...)
}
