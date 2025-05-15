package sse

import (
	"context"
	"multiple-sse/internal/service"

	"multiple-sse/api/sse/v1"
)

func (c *ControllerV1) BroadcastMsg(ctx context.Context, req *v1.BroadcastMsgReq) (res *v1.BroadcastMsgRes, err error) {
	err = service.Sse().BroadcastMsg(ctx, req.EventType, req.Data)
	return &v1.BroadcastMsgRes{}, err
}
