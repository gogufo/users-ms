package post

import (
	"encoding/json"
	"fmt"
	"users/cron"

	. "github.com/gogufo/gufo-api-gateway/gufodao"
	pb "github.com/gogufo/gufo-api-gateway/proto/go"
	"github.com/microcosm-cc/bluemonday"
	"github.com/spf13/viper"
	. "users/global"
)

var body struct {
	Action string `json:"action"`
}

func UpdateCron(t *pb.Request) (response *pb.Response) {
	ans := make(map[string]interface{})
	p := bluemonday.UGCPolicy()

	// Validate body presence
	if t == nil || t.Body == nil || len(t.Body.Value) == 0 {
		return ErrorReturn(t, 400, "000012", "Missing body")
	}

	// Define request payload structure

	// Parse JSON body
	if err := json.Unmarshal(t.Body.Value, &body); err != nil {
		return ErrorReturn(t, 400, "000013", "Invalid JSON body")
	}

	// Validate required field
	if body.Action == "" {
		return ErrorReturn(t, 400, "000012", "Missing action")
	}

	// Sanitize input
	action := p.Sanitize(body.Action)

	settingsKey := fmt.Sprintf("%s.cron", MicroServiceName)

	if action == "true" {
		viper.Set(settingsKey, true)
		// Run Cron
		go cron.Init()
	} else {
		viper.Set(settingsKey, false)
	}

	ans["answer"] = action
	return Interfacetoresponse(t, ans)
}
