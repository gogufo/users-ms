package main

import (
	"context"
	"fmt"
	"strings"
	ad "users/admin"

	. "github.com/gogufo/gufo-api-gateway/gufodao"
	pb "github.com/gogufo/gufo-api-gateway/proto/go"
	"github.com/spf13/viper"
	. "users/global"
	. "users/version"
)

type handlerFuncCtx func(context.Context, *pb.Request) *pb.Response

// ===============================================================
// ROUTER TABLE
// ===============================================================

var routeTable = map[string]handlerFuncCtx{}

func InitRoutes() {

	routeTable["GET:info"] = infoHandler
	routeTable["GET:health"] = healthHandler

	if viper.GetBool(fmt.Sprintf("microservices.%s.admin", MicroServiceName)) {
		routeTable["POST:admin"] = adminHandler
		routeTable["GET:admin"] = adminHandler
		SetLog("[ROUTER] admin routes enabled")
	} else {
		SetLog("[ROUTER] admin routes disabled")
	}
}

// routerLookup selects handler based on METHOD:param
// (Do() guarantees heartbeatLocal is checked separately)
func routerLookup(t *pb.Request) handlerFuncCtx {
	if t == nil {
		return unknownRouteHandler
	}

	method := ProtoMethodToString(t.Method)
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	param := "unknown"
	if t.Param != "" {
		param = strings.ToLower(t.Param)
	}

	key := fmt.Sprintf("%s:%s", method, param)

	if h, ok := routeTable[key]; ok {
		return h
	}

	return unknownRouteHandler
}

// ===============================================================
// HANDLERS
// ===============================================================

// infoHandler returns basic microservice info.
func infoHandler(ctx context.Context, t *pb.Request) *pb.Response {
	select {
	case <-ctx.Done():
		// Unified timeout error
		return ErrorReturn(t, 408, "000408", "Request timeout")
	default:
	}

	ans := map[string]interface{}{
		"pluginname":  "users",
		"version":     VERSIONPLUGIN,
		"description": "",
	}
	return Interfacetoresponse(t, ans)
}

// healthHandler returns simple health status.
func healthHandler(ctx context.Context, t *pb.Request) *pb.Response {
	select {
	case <-ctx.Done():
		return ErrorReturn(t, 408, "000408", "Request timeout")
	default:
	}

	ans := map[string]interface{}{
		"health": "OK",
	}
	return Interfacetoresponse(t, ans)
}

// adminHandler checks admin rights.
func adminHandler(ctx context.Context, t *pb.Request) *pb.Response {
	select {
	case <-ctx.Done():
		return ErrorReturn(t, 408, "000408", "Request timeout")
	default:
	}

	if t.Auth == nil || !t.Auth.IsAdmin {
		return ErrorReturn(t, 401, "000012", "You have no admin rights")
	}

	return ad.Init(t) // Delegate to admin module
}

// unknownRouteHandler handles nonexistent routes.
func unknownRouteHandler(ctx context.Context, t *pb.Request) *pb.Response {
	select {
	case <-ctx.Done():
		return ErrorReturn(t, 408, "000408", "Request timeout")
	default:
	}

	param := "unknown"
	if t != nil && t.Param != "" {
		param = t.Param
	}
	return ErrorReturn(t, 404, "000404", "Unknown route: "+param)
}
