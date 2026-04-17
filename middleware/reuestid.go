package middleware

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type requestIDKey struct{}

func RequestIDUnary(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {

	id := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey{}, id)

	return handler(ctx, req)
}

func RequestIDStream(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {

	id := uuid.New().String()
	wrapped := &wrappedStream{ServerStream: ss, ctx: context.WithValue(ss.Context(), requestIDKey{}, id)}

	return handler(srv, wrapped)
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context { return w.ctx }
