package cmd

import (
	"context"
	"multiple-sse/internal/controller/sse"
	"multiple-sse/internal/service"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			go service.Sse().StartRedisSubscriber()
			go service.Sse().StartHeartBeat(10 * time.Second)
			go service.Sse().StartIdleConnectionCleaner(30*time.Minute, 10*time.Second)
			s := g.Server("multiple-sse")
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Bind(sse.NewV1())
			})
			s.Run()
			return nil
		},
	}
)
