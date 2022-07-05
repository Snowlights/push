package logic

import (
	"bufio"
	push_gateway "github.com/Snowlights/push/gateway/protocol"
)

// Channel to one conn
type Channel struct {
	ClientProto ProtoRing

	room *Room

	signal chan *push_gateway.Proto
	Writer bufio.Writer
	Reader bufio.Reader
	Next   *Channel
	Prev   *Channel

	uid     string
	ip      string  // ip address
	accepts []int64 // accept message op
}
