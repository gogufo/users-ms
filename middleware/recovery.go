package middleware

import (
	"context"
	"fmt"

	. "github.com/gogufo/gufo-api-gateway/gufodao"
	"google.golang.org/grpc"
)

func RecoveryUnary(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {

	defer func() {
		if r := recover(); r != nil {
			SetErrorLog(fmt.Sprintf("panic recovered in %s: %v", info.FullMethod, r))
		}
	}()

	return handler(ctx, req)
}

func RecoveryStream(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {

	defer func() {
		if r := recover(); r != nil {
			SetErrorLog(fmt.Sprintf("panic recovered in stream %s: %v", info.FullMethod, r))
		}
	}()

	return handler(srv, ss)
}
