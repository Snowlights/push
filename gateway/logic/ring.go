package logic

import push_gateway "github.com/Snowlights/push/gateway/protocol"

type ProtoRing struct {
	num               uint64               // number of elements
	writePos, readPos uint64               // position of element
	mask              uint64               // masks to
	data              []push_gateway.Proto // data
}
