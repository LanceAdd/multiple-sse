package sse

import (
	"context"
	"multiple-sse/api/sse/v1"
	"multiple-sse/internal/service"
)

func (c *ControllerV1) SendMsg(ctx context.Context, req *v1.SendMsgReq) (res *v1.SendMsgRes, err error) {
	err = service.Sse().SendMsg(ctx, req.ClientId, req.EventType, req.Data)
	return &v1.SendMsgRes{}, err
}
