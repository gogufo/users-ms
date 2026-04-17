package main

import (
	"fmt"
	"time"

	. "github.com/gogufo/gufo-api-gateway/gufodao"
	pb "github.com/gogufo/gufo-api-gateway/proto/go"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	. "users/cron"
	. "users/global"
	. "users/version"
)

func StartHeartbeat() {
	interval := viper.GetInt("server.heartbeat")
	if interval < 2 {
		interval = 10 // default
	}

	for {
		sendHeartbeat()
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func sendHeartbeat() {

	args := map[string]interface{}{
		"name":        MicroServiceName,
		"host":        viper.GetString(fmt.Sprintf("microservices.%s.host", MicroServiceName)),
		"port":        viper.GetString(fmt.Sprintf("microservices.%s.port", MicroServiceName)),
		"group":       viper.GetString(fmt.Sprintf("microservices.%s.group", MicroServiceName)),
		"isinternal":  viper.GetBool(fmt.Sprintf("microservices.%s.isinternal", MicroServiceName)),
		"issession":   viper.GetBool(fmt.Sprintf("microservices.%s.issession", MicroServiceName)),
		"description": viper.GetString(fmt.Sprintf("microservices.%s.description", MicroServiceName)),
		"author":      viper.GetString(fmt.Sprintf("microservices.%s.author", MicroServiceName)),
		"version":     VERSIONPLUGIN,
	}

	// Convert args → StructPB
	pbStruct, err := structpb.NewStruct(args)
	if err != nil {
		SetErrorLog(fmt.Sprintf("[HEARTBEAT] structpb error: %v", err))
		return
	}

	// Wrap Struct into Any (this becomes Body)
	anyPayload, err := anypb.New(pbStruct)
	if err != nil {
		SetErrorLog(fmt.Sprintf("[HEARTBEAT] anypb error: %v", err))
		return
	}

	module := "heartbeat"
	param := "send"

	req := &pb.Request{
		Module: module,
		Param:  param,
		Method: pb.Method_METHOD_POST,
		Body:   anyPayload,
	}

	req = Gufosign(req)

	// Call GRPCConnect
	host := viper.GetString("server.internal_host")
	port := viper.GetString("server.port")

	resp := GRPCConnect(host, port, req)

	// Check if this is an error (GRPCConnect returns httpcode only in case of error)
	if codeObj, exists := resp["httpcode"]; exists {
		httpcode, _ := codeObj.(int)
		if httpcode != 200 {
			SetErrorLog(fmt.Sprintf("[HEARTBEAT] ERROR: %+v", resp))
			return
		}
	}

	// resp IS the response data
	cronFlag, _ := resp["cron"].(bool)
	leaderFlag, _ := resp["leader"].(bool)
	ttl, _ := resp["ttl"].(float64)

	ApplyCronState(cronFlag)

	SetLog(fmt.Sprintf(
		"[HEARTBEAT] OK | cron=%v | leader=%v | ttl=%v",
		cronFlag, leaderFlag, ttl,
	))
}
