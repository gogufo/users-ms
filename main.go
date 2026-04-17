// Generated with GRPC Microservice Creator v1.10.2

package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"users/entrypoint"
	"users/middleware"

	"github.com/certifi/gocertifi"
	"github.com/getsentry/sentry-go"
	. "github.com/gogufo/gufo-api-gateway/gufodao"
	pb "github.com/gogufo/gufo-api-gateway/proto/go"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	. "users/global"
	. "users/version"
)

// fileState describes a file being received via streaming
type fileState struct {
	path    string
	f       *os.File
	written int64
}

// Server implements pb.ReverseServer
type Server struct{}

// ======================= MAIN =======================
func main() {
	fmt.Fprintln(os.Stderr, ">>> process started")
	fmt.Fprintln(os.Stderr, ">>> working dir:", getCwdDebug())

	loadConfig()
	fmt.Fprintln(os.Stderr, ">>> config loaded OK")

	initSentry()
	fmt.Fprintln(os.Stderr, ">>> config loaded OK")

	validateSecurityOrExit()
	fmt.Fprintln(os.Stderr, ">>> security OK")

	port := resolvePort()
	fmt.Fprintln(os.Stderr, ">>> ABOUT TO LISTEN ON:", port)

	listener, err := net.Listen("tcp", port)
	//listener, err := net.Listen("tcp4", "0.0.0.0"+port)
	if err != nil {
		fmt.Fprintln(os.Stderr, ">>> failed to listen: ", err.Error())
		grpclog.Fatalf("failed to listen: %v", err)
		os.Exit(10)
	}

	fmt.Fprintln(os.Stderr, ">>> LISTENER OK, STARTING SERVE")

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.RequestIDUnary,
			middleware.LoggingUnary,
			middleware.SecurityUnary,
			middleware.MetricsUnary,
			middleware.RecoveryUnary,
		),
		grpc.ChainStreamInterceptor(
			middleware.RequestIDStream,
			middleware.LoggingStream,
			middleware.SecurityStream,
			middleware.MetricsStream,
			middleware.RecoveryStream,
		),
	)
	s := &Server{}
	pb.RegisterReverseServer(grpcServer, s)

	InitRoutes()
	fmt.Fprintln(os.Stderr, ">>> Routes Init")

	runEntrypoint()
	fmt.Fprintln(os.Stderr, ">>> Entrypoint OK")

	go StartHeartbeat()
	fmt.Fprintln(os.Stderr, ">>> StartHeartbeat OK")

	SetLog(fmt.Sprintf("%s microservice listening on %s", MicroServiceName, port))
	fmt.Fprintln(os.Stderr, ">>> microservice listening on: ", port)
	// Graceful shutdown listener
	waitForShutdown(grpcServer)

	if err := grpcServer.Serve(listener); err != nil {
		SetErrorLog("gRPC server error: " + err.Error())
		fmt.Fprintln(os.Stderr, ">>> gRPC server error:  ", err.Error())
	}
}

func getCwdDebug() string {
	d, err := os.Getwd()
	if err != nil {
		return "cwd_error: " + err.Error()
	}
	return d
}

//
// ======================= INITIALIZATION =======================
//

func loadConfig() {
	// Allow env variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigName("settings")
	viper.AddConfigPath("./config/")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		SetErrorLog("Config load error: " + err.Error())
		if viper.GetBool("server.sentry") {
			sentry.CaptureException(err)
		}
		fmt.Fprintln(os.Stderr, ">>> CONFIG ERROR:", err.Error())
		os.Exit(3)
	}
}

func initSentry() {
	if !viper.GetBool("server.sentry") {
		return
	}

	SetLog("Connecting to Sentry...")

	sentryOptions := sentry.ClientOptions{
		Dsn:              viper.GetString("sentry.dsn"),
		EnableTracing:    viper.GetBool("sentry.tracing"),
		Debug:            viper.GetBool("sentry.debug"),
		TracesSampleRate: viper.GetFloat64("sentry.trace"),
	}

	rootCAs, err := gocertifi.CACerts()
	if err == nil {
		sentryOptions.CaCerts = rootCAs
	}

	if err := sentry.Init(sentryOptions); err != nil {
		SetErrorLog("sentry.Init error: " + err.Error())
	}

	flush := viper.GetDuration("sentry.flush")
	if flush <= 0 {
		flush = 2 * time.Second
	}
	defer sentry.Flush(flush)
}

func validateSecurityOrExit() {
	if err := validateSecurityConfig(); err != nil {
		SetErrorLog("Security config error: " + err.Error())
		if viper.GetBool("server.sentry") {
			sentry.CaptureException(err)
		}
		fmt.Fprintln(os.Stderr, ">>> validateSecurityOrExit error:", err.Error())
		os.Exit(3)
	}
}

func resolvePort() string {
	key := fmt.Sprintf("microservices.%s.port", MicroServiceName)
	p := viper.GetString(key)
	if p != "" {
		return ":" + p
	}
	return ":5300"
}

func runEntrypoint() {
	key := fmt.Sprintf("microservices.%s.entrypointversion", MicroServiceName)
	last := viper.GetString(key)
	if last != VERSIONPLUGIN {
		go entrypoint.Init()
	}
	go entrypoint.EntryPoint()
}

func waitForShutdown(grpcServer *grpc.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-stop
		SetLog(fmt.Sprintf("Shutdown signal received (%s), stopping gRPC server gracefully...", sig.String()))

		done := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()

		shutdownTimeout := viper.GetDuration("server.shutdown_timeout")
		if shutdownTimeout <= 0 {
			shutdownTimeout = 10 * time.Second
		}

		select {
		case <-done:
			SetLog("gRPC server stopped gracefully")
		case <-time.After(shutdownTimeout):
			SetErrorLog("Graceful shutdown timeout exceeded, forcing Stop()")
			grpcServer.Stop()
		}

		if CachePool != nil {
			if err := CachePool.Close(); err != nil {
				SetErrorLog("Redis pool close error: " + err.Error())
			} else {
				SetLog("Redis pool closed")
			}
		}

		SetLog("Shutdown sequence completed")
	}()
}

//
// ======================= G R P C —   D O  =======================
//

func (s *Server) Do(parentCtx context.Context, req *pb.Request) (*pb.Response, error) {
	// Deadline is applied here; all other processing happens in middleware
	fmt.Fprintln(os.Stderr, ">>> users GRPC IN")
	if req.Param != "" {
		fmt.Fprintln(os.Stderr, ">>> users param=", req.Param)
	}

	timeout := viper.GetDuration("server.grpc_timeout")
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	// Local heartbeat handler (ping from Gateway / masterservice)
	if req.Param == "heartbeat" {
		return heartbeatLocal(req), nil
	}

	// Everything else goes through router
	handler := routerLookup(req)
	fmt.Fprintln(os.Stderr, ">>>  GRPC OUT")

	return handler(ctx, req), nil
}

// Helper for metrics / logging
func getRouteKey(req *pb.Request) string {
	if req == nil {
		return "nil"
	}

	param := "unknown"
	if req.Param != "" {
		param = strings.ToLower(req.Param)
	}

	method := strings.ToUpper(ProtoMethodToString(req.Method))
	if method == "" {
		return param
	}

	return fmt.Sprintf("%s:%s", method, param)
}

//
// ======================= G R P C —   S T R E A M =======================
//

func (s *Server) Stream(stream pb.Reverse_StreamServer) error {
	// All security / logging / metrics are handled by stream middleware here

	first, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		SetErrorLog("Stream initial recv: " + err.Error())
		return err
	}

	// timeout for entire stream
	streamTimeout := viper.GetDuration("server.stream_timeout")
	if streamTimeout <= 0 {
		streamTimeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(stream.Context(), streamTimeout)
	defer cancel()

	files := make(map[string]*fileState)
	module := first.Module

	if err := handleStreamRequest(ctx, stream, first, files, module); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			SetErrorLog("Stream timeout exceeded")
			return ctx.Err()
		default:
		}

		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			SetErrorLog("Stream recv error: " + err.Error())
			return err
		}

		if err := handleStreamRequest(ctx, stream, req, files, module); err != nil {
			return err
		}
	}

	for _, st := range files {
		if st.f != nil {
			st.f.Close()
		}
	}

	return nil
}

//
// ======================= Stream Frame Handler =======================
//

func handleStreamRequest(
	ctx context.Context,
	stream pb.Reverse_StreamServer,
	req *pb.Request,
	files map[string]*fileState,
	module string,
) error {

	_ = ctx

	// META frames: start/end
	if req.Body != nil {
		var metaCtx pb.RequestContext
		if err := req.Body.UnmarshalTo(&metaCtx); err == nil && metaCtx.Meta != nil {
			filename := metaCtx.Meta["filename"]
			phase := metaCtx.Meta["phase"]

			if filename != "" && phase != "" {
				switch phase {
				case "start":
					dir := "/tmp"
					if module != "" {
						dir = filepath.Join(dir, module)
					}
					if err := os.MkdirAll(dir, 0o755); err != nil {
						SetErrorLog("dir create: " + err.Error())
						return err
					}

					fp := filepath.Join(dir, filename)
					f, err := os.Create(fp)
					if err != nil {
						SetErrorLog("file create: " + err.Error())
						return err
					}

					files[filename] = &fileState{path: fp, f: f, written: 0}
					SetLog("Start receiving: " + fp)

				case "end":
					if st, ok := files[filename]; ok && st.f != nil {
						st.f.Close()
						fp := st.path
						SetLog("Finished receiving: " + fp)
						delete(files, filename)
						go processUploadedFile(fp)
					}
				}

				val, _ := anypb.New(wrapperspb.String("ok"))
				resp := &pb.Response{Data: map[string]*anypb.Any{"status": val}}
				return stream.Send(resp)
			}
		}
	}

	// CHUNK frames
	if req.Body != nil {
		var chunk pb.FileChunk
		if err := req.Body.UnmarshalTo(&chunk); err == nil {
			st, ok := files[chunk.Name]
			if !ok || st.f == nil {
				return fmt.Errorf("chunk for unknown file")
			}

			n, err := st.f.Write(chunk.Data)
			if err != nil {
				SetErrorLog("write chunk: " + err.Error())
				return err
			}
			st.written += int64(n)

			return nil
		}
	}

	return nil
}

//
// ======================= Security Checks =======================
//

func validateSecurityConfig() error {
	mode := strings.ToLower(viper.GetString("security.mode"))

	switch mode {
	case "hmac":
		if viper.GetString("security.hmac_secret") == "" {
			return fmt.Errorf("security.hmac_secret required")
		}
		if viper.GetInt("security.max_age") <= 0 {
			return fmt.Errorf("security.max_age must be >0")
		}
	case "sign":
		if viper.GetString("server.sign") == "" {
			return fmt.Errorf("server.sign required")
		}
	case "mtls":
		if viper.GetString("tls.cert") == "" ||
			viper.GetString("tls.key") == "" ||
			viper.GetString("tls.ca") == "" {
			return fmt.Errorf("mTLS requires tls.cert, tls.key, tls.ca")
		}
	default:
		return fmt.Errorf("unknown or missing security.mode")
	}

	return nil
}

//
// ======================= Helpers =======================
//

func heartbeatLocal(t *pb.Request) *pb.Response {
	return Interfacetoresponse(t, map[string]interface{}{
		"status": "alive",
	})
}
