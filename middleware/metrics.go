package middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"users/metrics"
)

func MetricsUnary(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {

	start := time.Now()
	resp, err = handler(ctx, req)

	metrics.RecordRequestMetrics(info.FullMethod, time.Since(start), err != nil)

	return
}

func MetricsStream(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {

	start := time.Now()
	err := handler(srv, ss)

	metrics.RecordRequestMetrics(info.FullMethod, time.Since(start), err != nil)

	return err
}
