package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/gogufo/gufo-api-gateway/gufodao"
	pb "github.com/gogufo/gufo-api-gateway/proto/go"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func SecurityUnary(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {

	pbReq, ok := req.(*pb.Request)
	if ok {
		if err := verifyRequestSecurity(pbReq); err != nil {
			return ErrorReturn(pbReq, 401, "00001", "You are not authorized"), nil
		}
	}

	return handler(ctx, req)
}

func SecurityStream(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	// We can wrap RecvMsg but simplest variant:
	return handler(srv, ss)
}

func verifyRequestSecurity(request *pb.Request) error {
	mode := strings.ToLower(viper.GetString("security.mode"))

	switch mode {

	case "hmac":
		secret := viper.GetString("security.hmac_secret")
		maxAge := time.Duration(viper.GetInt("security.max_age")) * time.Second

		if request == nil ||
			request.Auth == nil ||
			request.Auth.Sign == "" ||
			request.Module == "" ||
			!VerifyHMAC(secret, request.Module, request.Auth.Sign, maxAge) {

			return fmt.Errorf("invalid HMAC signature")
		}

	case "sign":
		if request == nil ||
			request.Auth == nil ||
			request.Auth.Sign == "" ||
			viper.GetString("server.sign") != request.Auth.Sign {

			return fmt.Errorf("invalid static signature")
		}

	case "mtls":
		// trust TLS layer

	default:
		return fmt.Errorf("unknown security.mode")
	}

	return nil
}
