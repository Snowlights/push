package logic

import (
	"github.com/Snowlights/push/gateway/config"
	push_gateway "github.com/Snowlights/push/gateway/protocol"
	"sync"
	"sync/atomic"
)

type Bucket struct {
	mu           sync.RWMutex
	keyToChannel map[string]*Channel // channel list
	roomIDToRoom map[string]*Room    // room list

	routines    []chan *push_gateway.BroadCastRoomReq // push msg handler
	routinesNum uint64
	curRoutine  uint64
}

func NewBucket(c *config.Bucket) *Bucket {
	b := new(Bucket)
	b.keyToChannel = make(map[string]*Channel, c.Channel)
	b.roomIDToRoom = make(map[string]*Room, c.Room)
	for i := uint64(0); i < c.Routine; i++ {
		c := make(chan *push_gateway.BroadCastRoomReq, c.RoutineSize)
		b.routines[i] = c
		go b.procRoomRoutine(c)
	}
	return b
}

func (b *Bucket) Room(roomID string) (*Room, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	r, ok := b.roomIDToRoom[roomID]
	if ok {
		return r, true
	}
	return nil, false
}

func (b *Bucket) RemoveRoom(room *Room) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.roomIDToRoom, room.ID)
	go room.Close()
}

func (b *Bucket) Channel(key string) (*Channel, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	c, ok := b.keyToChannel[key]
	if ok {
		return c, true
	}
	return nil, false
}

func (b *Bucket) PutChannel(roomID string, channel *Channel) {
	b.mu.Lock()
	defer b.mu.Unlock()
	c, ok := b.keyToChannel[channel.Key]
	if ok {
		c.Close()
	}
	b.keyToChannel[channel.Key] = channel

	if len(roomID) > 0 {
		r, ok := b.roomIDToRoom[roomID]
		if !ok {
			r = NewRoom(roomID)
			b.roomIDToRoom[roomID] = r
		}
		channel.Room = r
		r.PutChannel(channel)
	}
}

func (b *Bucket) RemoveChannel(channel *Channel) {
	b.mu.Lock()
	defer b.mu.Unlock()
	c, ok := b.keyToChannel[channel.Key]
	if ok && c == channel {
		delete(b.keyToChannel, channel.Key)
	}

	go func() {
		if c.Room != nil {
			drop := c.Room.RemoveChannel(channel.Key)
			if drop {
				c.Room.Close()
			}
		}
	}()
}

func (b *Bucket) BroadCast(proto *push_gateway.Proto) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, v := range b.keyToChannel {
		v.Push(proto)
	}
}

func (b *Bucket) BroadCastRoom(proto *push_gateway.BroadCastRoomReq) {
	idx := atomic.AddUint64(&b.curRoutine, 1) % b.routinesNum
	b.routines[idx] <- proto
}

func (b *Bucket) procRoomRoutine(c chan *push_gateway.BroadCastRoomReq) {
	for {
		// todo: speed limit
		v := <-c
		room, ok := b.Room(v.RoomID)
		if ok {
			room.PushRoom(v.Msg)
		}
	}
}
