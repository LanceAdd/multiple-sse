package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type ConnectReq struct {
	g.Meta `path:"/sse/create" method:"get" tags:"SSE" summary:"创建sse"`
}

type ConnectRes struct {
}
