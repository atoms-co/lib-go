package statshandlerx

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/metric/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

var otelServerHandler = otelgrpc.NewServerHandler(
	// Disable the standard OTel gRPC metrics. This package records the legacy
	// gRPC metric schemas separately with native OTel instruments.
	otelgrpc.WithMeterProvider(noop.NewMeterProvider()),
)

type ServerHandler struct {
}

// WithServerGRPCStatsHandler sets up the gRPC stats handler for the server with metrics and tracing support.
func WithServerGRPCStatsHandler() grpc.ServerOption {
	return grpc.StatsHandler(&ServerHandler{})
}

func (h *ServerHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	otelServerHandler.HandleConn(ctx, cs)
}

func (h *ServerHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	return otelServerHandler.TagConn(ctx, cti)
}

func (h *ServerHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	otelServerHandler.HandleRPC(ctx, rs)
}

func (h *ServerHandler) TagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	return otelServerHandler.TagRPC(ctx, rti)
}
