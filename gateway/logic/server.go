package logic

import (
	push_service "github.com/Snowlights/push/service/protocol"
	"runtime"
	"sync/atomic"
)

type Server struct {
	buckets                  []*Bucket
	bucketSize, curBucketPos uint64

	reader, writer             []*Pool
	readerSize, writerSize     uint64
	curReaderPos, curWriterPos uint64

	// todo 依赖etcd
	serviceRpcClient []push_service.ServiceClient
}

func NewServer(poolSize, bufSize, bufLength uint64) *Server {
	s := &Server{
		bucketSize: uint64(runtime.NumCPU()),
		reader:     nil,
		writer:     nil,
		readerSize: poolSize,
		writerSize: poolSize,
	}

	s.buckets = make([]*Bucket, s.bucketSize)
	for i := uint64(0); i < s.bucketSize; i++ {
		bucket := NewBucket(10)
		s.buckets[i] = bucket
	}

	s.reader = make([]*Pool, s.readerSize)
	for i := uint64(0); i < s.readerSize; i++ {
		s.reader[i] = NewPool(bufLength, bufSize)
	}
	s.writer = make([]*Pool, s.writerSize)
	for i := uint64(0); i < s.writerSize; i++ {
		s.writer[i] = NewPool(bufLength, bufSize)
	}

	return s
}

func (s *Server) GetReader() *Buffer {
	idx := atomic.AddUint64(&s.curReaderPos, 1) % s.readerSize
	return s.reader[idx].Get()
}

func (s *Server) GetWriter() *Buffer {
	idx := atomic.AddUint64(&s.curWriterPos, 1) % s.writerSize
	return s.writer[idx].Get()
}

func (s *Server) GetBucket() []*Bucket {
	return s.buckets
}

func (s *Server) GetRandomBucket() *Bucket {
	idx := atomic.AddUint64(&s.curBucketPos, 1) % s.bucketSize
	return s.buckets[idx]
}
