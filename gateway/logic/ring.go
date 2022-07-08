package logic

import push_gateway "github.com/Snowlights/push/gateway/protocol"

type ProtoRing struct {
	num               uint64               // number of elements
	writePos, readPos uint64               // position of element
	mask              uint64               // masks to
	data              []push_gateway.Proto // data
}

func NewRing(num uint64) *ProtoRing {
	r := new(ProtoRing)
	r.Init(num)
	return r
}

func (r *ProtoRing) Init(num uint64) {
	for num&(num-1) != 0 {
		num &= num - 1
	}
	num <<= 1
	r.num = num
	r.mask = num - 1
	r.data = make([]push_gateway.Proto, num)
}

func (r *ProtoRing) GetProto() (*push_gateway.Proto, error) {
	if r.readPos == r.writePos {
		return nil, ErrRingEmpty
	}
	proto := &r.data[r.readPos&r.mask]
	return proto, nil
}

func (r *ProtoRing) GetDone() {
	r.readPos++
}

func (r *ProtoRing) SetProto() (*push_gateway.Proto, error) {
	if r.writePos-r.readPos >= r.num {
		return nil, ErrRingFull
	}
	proto := &r.data[r.writePos&r.mask]
	return proto, nil
}

func (r *ProtoRing) SetDone() {
	r.writePos++
}
