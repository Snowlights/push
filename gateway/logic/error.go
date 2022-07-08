package logic

import "errors"

var (
	// ring
	ErrRingEmpty = errors.New("ring buffer empty")
	ErrRingFull  = errors.New("ring buffer full")

	// channel
	ErrChannelFull = errors.New("channel buffer full")

	// room
	ErrRoomDrop = errors.New("room drop")

	// bucket

	// pool & buffer

)
