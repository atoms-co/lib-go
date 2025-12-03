package statshandlerx

import (
	"context"

	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

type ServerHandler struct {
	handler ocgrpc.ServerHandler
}

// WithServerGRPCStatsHandler sets up the gRPC stats handler for the server with metrics and tracing support.
func WithServerGRPCStatsHandler() grpc.ServerOption {
	return grpc.StatsHandler(&ServerHandler{})
}

func (h *ServerHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	h.handler.HandleConn(ctx, cs)
}

func (h *ServerHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	return h.handler.TagConn(ctx, cti)
}

func (h *ServerHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	h.handler.HandleRPC(ctx, rs)
}

func (h *ServerHandler) TagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	return h.handler.TagRPC(ctx, rti)
}
