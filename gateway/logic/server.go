package logic

import (
	"context"
	"fmt"
	"github.com/Snowlights/push/gateway/config"
	push_gateway "github.com/Snowlights/push/gateway/protocol"
	"github.com/Snowlights/push/gateway/websocket"
	push_service "github.com/Snowlights/push/service/protocol"
	"github.com/Snowlights/tool/vlog"
	"sync/atomic"
)

type Server struct {
	c *config.Config

	buckets                  []*Bucket
	bucketSize, curBucketPos uint64

	reader, writer             []*Pool
	curReaderPos, curWriterPos uint64

	// todo 依赖etcd
	serviceRpcClient []push_service.ServiceClient
}

func NewServer(c *config.Config) *Server {
	s := &Server{
		c:      c,
		reader: nil,
		writer: nil,
	}

	bucketConfig := s.c.Server.Bucket
	s.buckets = make([]*Bucket, bucketConfig.Cap)
	for i := uint64(0); i < s.bucketSize; i++ {
		bucket := NewBucket(bucketConfig)
		s.buckets[i] = bucket
	}

	poolConfig := s.c.Server.Pool
	s.reader = make([]*Pool, s.c.Server.Pool.ReaderCap)
	for i := uint64(0); i < poolConfig.ReaderCap; i++ {
		s.reader[i] = NewPool(poolConfig.ReaderBuf, poolConfig.ReaderSize)
	}
	s.writer = make([]*Pool, poolConfig.WriterCap)
	for i := uint64(0); i < poolConfig.WriterCap; i++ {
		s.writer[i] = NewPool(poolConfig.WriterBuf, poolConfig.WriterSize)
	}

	return s
}

// auth for goim handshake with client, use rsa & aes.
func (s *Server) authWebsocket(ctx context.Context, ws *websocket.Conn, p *push_gateway.Proto, cookie string) (uData *push_service.ConnectData, err error) {
	for {
		if err = p.ReadWebsocket(ws); err != nil {
			return
		}
		if p.Op == int64(push_gateway.OpType_OpType_Auth) {
			break
		} else {
			vlog.ErrorF(ctx, "ws request operation(%d) not auth", p.Op)
		}
	}
	uData, err = s.Connect(ctx, p, cookie)
	if err != nil {
		return
	}
	p.Op = int64(push_gateway.OpType_OpType_AuthReply)
	p.Body = nil
	if err = p.WriteWebsocket(ws); err != nil {
		return
	}
	err = ws.Flush()
	return
}

func (s *Server) Connect(ctx context.Context, p *push_gateway.Proto, cookie string) (*push_service.ConnectData, error) {
	fun := "Server.Connect -->"
	res, _ := s.serviceRpcClient[0].Connect(ctx, &push_service.ConnectReq{
		Token:  p.Body,
		Cookie: cookie,
	})
	if res.ErrInfo != nil {
		return nil, fmt.Errorf("%s connect to service failed, error is %v", fun, res.ErrInfo.Msg)
	}
	return res.Data, nil
}

func (s *Server) DisConnect(ctx context.Context, key string) error {
	fun := "Server.Connect -->"
	res, _ := s.serviceRpcClient[0].DisConnect(ctx, &push_service.DisConnectReq{
		ID: key,
	})
	if res.ErrInfo != nil {
		return fmt.Errorf("%s disconnect to service failed, error is %v", fun, res.ErrInfo.Msg)
	}
	return nil
}

func (s *Server) Operate(ctx context.Context, p *push_gateway.Proto, ch *Channel, b *Bucket) error {
	switch p.Op {
	case int64(push_gateway.OpType_OpType_ChangeRoom):
		// todo change room, use channel and bucket
		p.Op = int64(push_gateway.OpType_OpType_ChangeRoomReply)
	default:
		p.Op = int64(push_gateway.OpType_OpType_Default)
		p.Body = nil
	}
	return nil
}

func (s *Server) getReader() *Buffer {
	idx := atomic.AddUint64(&s.curReaderPos, 1) % s.c.Server.Bucket.Cap
	return s.reader[idx].Get()
}

func (s *Server) putReader(b *Buffer) {
	idx := atomic.AddUint64(&s.curReaderPos, 1) % s.c.Server.Bucket.Cap
	s.reader[idx].Put(b)
}

func (s *Server) getWriter() *Buffer {
	idx := atomic.AddUint64(&s.curWriterPos, 1) % s.c.Server.Bucket.Cap
	return s.writer[idx].Get()
}

func (s *Server) putWriter(b *Buffer) {
	idx := atomic.AddUint64(&s.curReaderPos, 1) % s.c.Server.Bucket.Cap
	s.reader[idx].Put(b)
}

func (s *Server) getBucket() []*Bucket {
	return s.buckets
}

func (s *Server) getRandomBucket() *Bucket {
	idx := atomic.AddUint64(&s.curBucketPos, 1) % s.bucketSize
	return s.buckets[idx]
}
