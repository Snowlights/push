package logic

import (
	"github.com/Snowlights/push/gateway/config"
	push_service "github.com/Snowlights/push/service/protocol"
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

func (s *Server) GetReader() *Buffer {
	idx := atomic.AddUint64(&s.curReaderPos, 1) % s.c.Server.Bucket.Cap
	return s.reader[idx].Get()
}

func (s *Server) GetWriter() *Buffer {
	idx := atomic.AddUint64(&s.curWriterPos, 1) % s.c.Server.Bucket.Cap
	return s.writer[idx].Get()
}

func (s *Server) GetBucket() []*Bucket {
	return s.buckets
}

func (s *Server) GetRandomBucket() *Bucket {
	idx := atomic.AddUint64(&s.curBucketPos, 1) % s.bucketSize
	return s.buckets[idx]
}
