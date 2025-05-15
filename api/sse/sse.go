// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package sse

import (
	"context"

	"multiple-sse/api/sse/v1"
)

type ISseV1 interface {
	BroadcastMsg(ctx context.Context, req *v1.BroadcastMsgReq) (res *v1.BroadcastMsgRes, err error)
	Connect(ctx context.Context, req *v1.ConnectReq) (res *v1.ConnectRes, err error)
	SendMsg(ctx context.Context, req *v1.SendMsgReq) (res *v1.SendMsgRes, err error)
}
