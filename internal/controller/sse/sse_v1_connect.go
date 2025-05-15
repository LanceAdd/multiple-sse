package sse

import (
	"context"
	"multiple-sse/api/sse/v1"
	"multiple-sse/internal/service"
)

func (c *ControllerV1) Connect(ctx context.Context, req *v1.ConnectReq) (res *v1.ConnectRes, err error) {
	service.Sse().Connect(ctx)
	return &v1.ConnectRes{}, nil
}
