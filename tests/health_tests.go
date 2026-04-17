package tests

import (
	"context"
	"testing"
	"time"

	pb "github.com/gogufo/gufo-api-gateway/proto/go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestHealthGRPC(t *testing.T) {

	conn, err := grpc.Dial(
		"localhost:5300",
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("grpc connection failed: %v", err)
	}
	defer conn.Close()

	client := pb.NewReverseClient(conn)

	req := &pb.Request{
		Method: pb.Method_METHOD_GET,
		Param:  "health",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Do(ctx, req)
	if err != nil {
		t.Fatalf("grpc health request failed: %v", err)
	}

	data := resp.GetData()
	if data == nil {
		t.Fatalf("empty response data")
	}

	valAny, ok := data["health"]
	if !ok {
		t.Fatalf("health key not found in response")
	}

	var val wrapperspb.StringValue
	if err := anypb.UnmarshalTo(valAny, &val, proto.UnmarshalOptions{}); err != nil {
		t.Fatalf("failed to unmarshal health value: %v", err)
	}

	if val.Value != "OK" {
		t.Fatalf("unexpected health value: %s", val.Value)
	}
}
