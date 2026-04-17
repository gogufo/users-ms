package admin

import (
	gt "users/admin/get"
	pt "users/admin/post"

	. "github.com/gogufo/gufo-api-gateway/gufodao"
	pb "github.com/gogufo/gufo-api-gateway/proto/go"
)

func Init(t *pb.Request) *pb.Response {
	if t == nil || t.ParamId == "" {
		return ErrorReturn(t, 400, "000400", "Missing action")
	}

	switch t.ParamId {
	case "cronstatus":
		return gt.CheckCron(t)

	case "cron":
		return pt.UpdateCron(t)

	default:
		return ErrorReturn(t, 404, "000404", "Unknown admin action")
	}
}
