package logic

import (
	"github.com/Snowlights/push/gateway/pkg/buf"
	push_gateway "github.com/Snowlights/push/gateway/protocol"
	"sync"
)

// Channel to one conn
type Channel struct {
	ClientProto ProtoRing
	ServerProto chan *push_gateway.Proto
	Room        *Room

	Writer buf.Writer
	Reader buf.Reader
	Next   *Channel
	Prev   *Channel

	Key string
	IP  string // ip address

	acceptMu sync.RWMutex
	accepts  map[int64]struct{} // accept message op
}

func NewChannel(serv, cli uint64) *Channel {
	c := new(Channel)
	c.ClientProto.Init(cli)
	c.ServerProto = make(chan *push_gateway.Proto, serv)
	c.accepts = make(map[int64]struct{}, 2)
	return c
}

func (c *Channel) Accept(accepts ...int64) {
	c.acceptMu.Lock()
	defer c.acceptMu.Unlock()

	for _, v := range accepts {
		c.accepts[v] = struct{}{}
	}
}

func (c *Channel) UnAccept(accepts ...int64) {
	c.acceptMu.Lock()
	defer c.acceptMu.Unlock()

	for _, v := range accepts {
		delete(c.accepts, v)
	}
}

func (c *Channel) NeedPush(v int64) bool {
	c.acceptMu.RLock()
	defer c.acceptMu.RUnlock()

	_, ok := c.accepts[v]
	return ok
}

func (c *Channel) Push(v *push_gateway.Proto) error {
	if c.NeedPush(v.Op) {
		return nil
	}
	select {
	case c.ServerProto <- v:
		return nil
	default:
		return ErrChannelFull
	}
}

func (c *Channel) Ready() *push_gateway.Proto {
	return <-c.ServerProto
}

func (c *Channel) Close() {
	// todo
}
