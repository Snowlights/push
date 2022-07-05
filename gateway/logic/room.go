package logic

import "sync"

type Room struct {
	ID string `json:"id"`

	cMu         sync.RWMutex
	channelList *Channel

	drop   bool
	online int64
}

