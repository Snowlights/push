package logic

import (
	push_gateway "github.com/Snowlights/push/gateway/protocol"
	"sync"
)

type Room struct {
	ID string `json:"id"`

	cMu        sync.RWMutex
	channelMap map[string]*Channel

	drop   bool
	online int64
}

func NewRoom(id string) *Room {
	r := new(Room)
	r.ID = id
	r.channelMap = make(map[string]*Channel, 16)
	return r
}

func (r *Room) PutChannel(channel *Channel) error {
	r.cMu.Lock()
	defer r.cMu.Unlock()

	if r.drop {
		return ErrRoomDrop
	}

	r.channelMap[channel.Key] = channel
	r.online++
	r.drop = len(r.channelMap) == 0
	return nil
}

func (r *Room) RemoveChannel(key string) bool {
	r.cMu.Lock()
	defer r.cMu.Unlock()

	_, ok := r.channelMap[key]
	if ok {
		delete(r.channelMap, key)
	}
	r.online--
	r.drop = len(r.channelMap) == 0
	return r.drop
}

func (r *Room) PushRoom(proto *push_gateway.Proto) {
	r.cMu.RLock()
	defer r.cMu.RUnlock()

	for _, v := range r.channelMap {
		v.Push(proto)
	}
}

func (r *Room) Online() int64 {
	r.cMu.RLock()
	defer r.cMu.RUnlock()

	return r.online
}

func (r *Room) Close() {
	r.cMu.Lock()
	defer r.cMu.Unlock()
	r.online = 0
	for _, v := range r.channelMap {
		v.Close()
	}
}
