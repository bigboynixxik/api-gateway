package interceptor

import (
	"context"
	"log/slog"
	"time"

	googleGrpc "google.golang.org/grpc"
)

func LoggingClientInterceptor(log *slog.Logger) googleGrpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *googleGrpc.ClientConn, invoker googleGrpc.UnaryInvoker, opts ...googleGrpc.CallOption) error {
		log.Info("outgoing gRPC request", "method", method)
		start := time.Now()

		err := invoker(ctx, method, req, reply, cc, opts...)

		log.Info("gRPC response received",
			"method", method,
			"duration", time.Since(start),
			"error", err,
		)
		return err
	}
}
