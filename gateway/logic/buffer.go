package logic

import "sync"

type Buffer struct {
	buf  []byte
	next *Buffer
}

type Pool struct {
	mu                    sync.Mutex
	next                  *Buffer
	size                  uint64
	nextLength, maxLength uint64
}

func NewPool(size, maxLength uint64) *Pool {
	p := new(Pool)
	p.size = size
	p.maxLength = maxLength
	p.nextLength = maxLength

	for i := uint64(0); i < maxLength/4; i++ {
		b := p.newBuffer()
		b.next = p.next
		p.next = b
	}
	return p
}

func (p *Pool) newBuffer() *Buffer {
	buf := make([]byte, p.size)
	return &Buffer{buf: buf}
}

func (p *Pool) Get() *Buffer {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.next != nil {
		cur := p.next
		p.next = p.next.next
		p.nextLength--
		cur.next = nil
		return cur
	}

	return p.newBuffer()
}

func (p *Pool) Put(b *Buffer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.nextLength >= p.maxLength {
		return
	}

	b.next = p.next
	p.next = b
	p.nextLength++
}
