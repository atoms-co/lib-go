package statshandlerx

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/metric/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

var otelClientHandler = otelgrpc.NewClientHandler(
	// Disable the standard OTel gRPC metrics. This package records the legacy
	// gRPC metric schemas separately with native OTel instruments.
	otelgrpc.WithMeterProvider(noop.NewMeterProvider()),
)

type ClientHandler struct {
}

// WithClientGRPCStatsHandler sets up the gRPC stats handler for the client with metrics and tracing support
func WithClientGRPCStatsHandler() grpc.DialOption {
	return grpc.WithStatsHandler(&ClientHandler{})
}

func (h *ClientHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	otelClientHandler.HandleConn(ctx, cs)
}

func (h *ClientHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	return otelClientHandler.TagConn(ctx, cti)
}

func (h *ClientHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	otelClientHandler.HandleRPC(ctx, rs)
}

func (h *ClientHandler) TagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	return otelClientHandler.TagRPC(ctx, rti)
}
