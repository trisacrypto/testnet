package rvasp

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// This unary interceptor traces gRPC requests, adds zerolog logging and panic recovery.
func UnaryTraceInterceptor(ctx context.Context, in interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (out interface{}, err error) {
	// Track how long the method takes to execute.
	start := time.Now()
	panicked := true

	// Recover from panics in the handler.
	defer func() {
		if r := recover(); r != nil || panicked {
			log.WithLevel(zerolog.PanicLevel).
				Err(fmt.Errorf("%v", r)).
				Str("stack_trace", string(debug.Stack())).
				Msg("grpc server has recovered from a panic")
			err = status.Error(codes.Internal, "an unhandled exception occurred")
		}
	}()

	// Call the handler to finalize the request and get the response.
	out, err = handler(ctx, in)
	panicked = false

	// Log with zerolog - checkout grpclog.LoggerV2 for default logging.
	log.Debug().
		Err(err).
		Str("method", info.FullMethod).
		Str("latency", time.Since(start).String()).
		Msg("gRPC request complete")
	return out, err
}

// This streaming interceptor traces gRPC requests, adds zerolog logging and panic recovery.
func StreamTraceInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	// Track how long the method takes to execute.
	start := time.Now()
	panicked := true

	defer func() {
		if r := recover(); r != nil || panicked {
			log.WithLevel(zerolog.PanicLevel).
				Err(fmt.Errorf("%v", r)).
				Str("stack_trace", string(debug.Stack())).
				Msg("grpc server has recovered from a panic")
			err = status.Error(codes.Internal, "an unhandled exception occurred")
		}
	}()

	err = handler(srv, stream)
	panicked = false

	// Log with zerolog - checkout grpclog.LoggerV2 for default logging.
	log.Debug().
		Err(err).
		Str("method", info.FullMethod).
		Str("duration", time.Since(start).String()).
		Msg("gRPC stream closed")
	return err
}
