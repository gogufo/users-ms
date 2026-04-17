package middleware

import (
	"context"
	"time"

	. "github.com/gogufo/gufo-api-gateway/gufodao"
	"google.golang.org/grpc"
)

func LoggingUnary(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {

	start := time.Now()
	SetLog("Unary call: " + info.FullMethod)

	resp, err := handler(ctx, req)

	SetLog("Unary done: " + info.FullMethod + " in " + time.Since(start).String())
	return resp, err
}

func LoggingStream(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {

	start := time.Now()
	SetLog("Stream call: " + info.FullMethod)

	err := handler(srv, ss)

	SetLog("Stream done: " + info.FullMethod + " in " + time.Since(start).String())
	return err
}
