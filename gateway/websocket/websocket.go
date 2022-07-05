package websocket

import (
	"bufio"
	"errors"
	"io"
)

const (
	// Frame header byte 0 bits from Section 5.2 of RFC 6455
	finBit  = 1 << 7
	rsv1Bit = 1 << 6
	rsv2Bit = 1 << 5
	rsv3Bit = 1 << 4
	opBit   = 0x0f

	// Frame header byte 1 bits from Section 5.2 of RFC 6455
	maskBit = 1 << 7
	lenBit  = 0x7f

	continuationFrame        = 0
	continuationFrameMaxRead = 100
)

// The message types are defined in RFC 6455, section 11.8.
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

var (
	// ErrMessageClose close control message
	ErrMessageClose = errors.New("close control message")
	// ErrMessageMaxRead continuation frame max read
	ErrMessageMaxRead = errors.New("continuation frame max read")
)

type Conn struct {
	rwc io.ReadWriteCloser
	r   *bufio.Reader
	w   *bufio.Writer
}

func newConn(rwc io.ReadWriteCloser, r *bufio.Reader, w *bufio.Writer) *Conn {
	return &Conn{rwc: rwc, r: r, w: w}
}

// WriteMessage write a message by type.
func (c *Conn) WriteMessage(msgType int, msg []byte) (err error) {
	if err = c.WriteHeader(msgType, len(msg)); err != nil {
		return
	}
	err = c.WriteBody(msg)
	return
}

// WriteHeader write header frame.
func (c *Conn) WriteHeader(msgType int, length int) (err error) {

	return nil
}

// WriteBody write a message body.
func (c *Conn) WriteBody(b []byte) (err error) {
	if len(b) > 0 {
		_, err = c.w.Write(b)
	}
	return
}

// Flush flush writer buffer
func (c *Conn) Flush() error {
	return c.w.Flush()
}

// ReadMessage read a message.
func (c *Conn) ReadMessage() (op int, payload []byte, err error) {
	return 0, nil, nil
}

// Close close the connection.
func (c *Conn) Close() error {
	return c.rwc.Close()
}
