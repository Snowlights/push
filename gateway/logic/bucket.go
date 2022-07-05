package logic

import (
	push_gateway "github.com/Snowlights/push/gateway/protocol"
	"sync"
)

type Bucket struct {
	channelMu    sync.RWMutex
	uidToChannel map[string]*Channel // channel list

	roomMu       sync.RWMutex
	roomIDToRoom map[string]*Room // room list

	routines    []*chan push_gateway.BroadCastRoomReq // push msg handler
	routinesNum uint64
}
