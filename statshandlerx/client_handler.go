package statshandlerx

import (
	"context"

	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

type ClientHandler struct {
	handler ocgrpc.ClientHandler
}

// WithClientGRPCStatsHandler sets up the gRPC stats handler for the client with metrics and tracing support
func WithClientGRPCStatsHandler() grpc.DialOption {
	return grpc.WithStatsHandler(&ClientHandler{})
}

func (h *ClientHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	h.handler.HandleConn(ctx, cs)
}

func (h *ClientHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	return h.handler.TagConn(ctx, cti)
}

func (h *ClientHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	h.handler.HandleRPC(ctx, rs)
}

func (h *ClientHandler) TagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	return h.handler.TagRPC(ctx, rti)
}
