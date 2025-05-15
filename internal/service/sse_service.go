package service

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
	"sync"
	"time"
)

const RedisChannel = "SSE:MSG"

var (
	clientOnce sync.Once
	sseService *SseService
)

type Client struct {
	Id             string
	Request        *ghttp.Request
	msgChan        chan string
	lastActiveTime time.Time
	cancelFunc     context.CancelFunc
}

type SseService struct {
	clientMap                 *gmap.StrAnyMap
	stopHeartbeat             chan struct{}
	stopIdleConnectionCleaner chan struct{}
	stopSubscriber            chan struct{}
	wg                        sync.WaitGroup
}

type RedisMsg struct {
	ClientId  string
	Event     string
	Data      string
	BroadCast bool
}

func init() {
	clientOnce.Do(func() {
		sseService = NewSseService()
	})
}

func NewSseService() *SseService {
	return &SseService{
		clientMap: gmap.NewStrAnyMap(true),
	}
}

func Sse() *SseService {
	return sseService
}

func (s *SseService) Connect(ctx context.Context) {
	r := ghttp.RequestFromCtx(ctx)
	r.Response.Header().Set("Access-Control-Allow-Origin", "*")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Connection", "keep-alive")
	clientId := r.Get("client_id", guid.S()).String()
	_, cancelFunc := context.WithCancel(ctx)
	client := &Client{
		Id:             clientId,
		Request:        r,
		msgChan:        make(chan string, 100),
		lastActiveTime: time.Now(),
		cancelFunc:     cancelFunc,
	}
	s.clientMap.Set(clientId, client)
	defer func() {
		close(client.msgChan)
		s.clientMap.Remove(clientId)
	}()
	r.Response.Writefln("id: %s\nevent: connected\ndata: {\"status\": \"connected\", \"client_id\": \"%s\"}\n", clientId, clientId)
	r.Response.Flush()

	for {
		select {
		case msg, ok := <-client.msgChan:
			if !ok {
				return
			}
			r.Response.Writefln(msg)
			r.Response.Flush()
			client.lastActiveTime = time.Now()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *SseService) SendMsgToClient(clientId, event, data string) bool {
	if client := s.clientMap.Get(clientId); client != nil {
		c := client.(*Client)
		msg := fmt.Sprintf("id: %d\nevent: %s\ndata: %s\n\n", time.Now().UnixNano(), event, data)
		select {
		case c.msgChan <- msg:
			return true
		default:
			return false
		}
	}
	return false
}

func (s *SseService) BroadcastMsgToClients(event, data string) int {
	count := 0
	s.clientMap.Iterator(func(k string, v interface{}) bool {
		if s.SendMsgToClient(k, event, data) {
			count++
		}
		return true
	})
	return count
}

func (s *SseService) StartHeartBeat(tickerTime time.Duration) {
	ticker := time.NewTicker(tickerTime)
	defer ticker.Stop()
	for range ticker.C {
		s.clientMap.Iterator(func(k string, v interface{}) bool {
			client := v.(*Client)
			select {
			case client.msgChan <- fmt.Sprintf("id: %d\nevent: %s\ndata: {\"status\": \"alive\", \"client_id\": \"%s\"}\n", time.Now().UnixNano(), "heartbeat", k):
				client.lastActiveTime = time.Now()
			default:
				g.Log().Infof(context.TODO(), "client %s is idle, close it", k)
				close(client.msgChan)
				s.clientMap.Remove(k)
			}
			return true
		})
	}
}

func (s *SseService) StartIdleConnectionCleaner(tickerTime time.Duration, aliveTime time.Duration) {
	ticker := time.NewTicker(tickerTime)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.clientMap.LockFunc(func(m map[string]interface{}) {
			for k, v := range m {
				client := v.(*Client)
				if now.Sub(client.lastActiveTime) > aliveTime {
					select {
					case <-client.msgChan:
					default:
						close(client.msgChan)
					}
					delete(m, k)
				}
			}
		})
	}
}

func (s *SseService) PublishToRedis(ctx context.Context, clientId, event, data string, broadcast bool) error {
	msg := RedisMsg{
		ClientId:  clientId,
		Event:     event,
		Data:      data,
		BroadCast: broadcast,
	}

	bytes, err := gjson.Encode(msg)
	if err != nil {
		return err
	}
	_, err = g.Redis().Publish(ctx, RedisChannel, bytes)
	return err
}

func (s *SseService) StartRedisSubscriber() error {
	ctx := context.Background()
	con, _, err := g.Redis().Subscribe(ctx, RedisChannel)
	if err != nil {
		return err
	}
	defer con.Close(ctx)
	for {
		msg, err := con.ReceiveMessage(ctx)
		if err != nil {
			g.Log().Fatal(ctx, err)
			continue
		}
		var redisMsg RedisMsg
		err = gjson.DecodeTo([]byte(msg.Payload), &redisMsg)
		if err != nil {
			g.Log().Fatal(ctx, err)
			continue
		}
		if redisMsg.BroadCast {
			s.BroadcastMsgToClients(redisMsg.Event, redisMsg.Data)
		} else {
			s.SendMsgToClient(redisMsg.ClientId, redisMsg.Event, redisMsg.Data)
		}
	}
	return nil
}

func (s *SseService) SendMsg(ctx context.Context, clientId, eventType, data string) error {
	return s.PublishToRedis(ctx, clientId, eventType, data, false)
}

func (s *SseService) BroadcastMsg(ctx context.Context, eventType, data string) error {
	return s.PublishToRedis(ctx, "", eventType, data, true)
}
