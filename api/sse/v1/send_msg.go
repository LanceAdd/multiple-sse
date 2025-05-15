package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type SendMsgReq struct {
	g.Meta    `path:"/sse/send/msg" method:"post" tags:"SSE" summary:"发送信息"`
	ClientId  string `json:"clientId" v:"required#客户端id不能为空"`
	EventType string `json:"eventType" v:"required#事件类型不能为空"`
	Data      string `json:"data" v:"required#数据不能为空"`
}

type SendMsgRes struct {
}
