package push_gateway

import (
	"errors"
	"github.com/Snowlights/push/gateway/pkg/encoding/binary"
	"github.com/Snowlights/push/gateway/websocket"
)

const (
	// MaxBodySize max proto body size
	MaxBodySize = int32(1 << 12)
)

const (
	// size
	_packSize      = 4
	_headerSize    = 2
	_verSize       = 2
	_opSize        = 4
	_seqSize       = 4
	_heartSize     = 4
	_rawHeaderSize = _packSize + _headerSize + _verSize + _opSize + _seqSize
	_maxPackSize   = MaxBodySize + int32(_rawHeaderSize)
	// offset
	_packOffset   = 0
	_headerOffset = _packOffset + _packSize
	_verOffset    = _headerOffset + _headerSize
	_opOffset     = _verOffset + _verSize
	_seqOffset    = _opOffset + _opSize
	_heartOffset  = _seqOffset + _seqSize
)

var (
	// ErrProtoPackLen proto packet len error
	ErrProtoPackLen = errors.New("default server codec pack length error")
	// ErrProtoHeaderLen proto header len error
	ErrProtoHeaderLen = errors.New("default server codec header length error")
)

func (p *Proto) ReadWebsocket(ws *websocket.Conn) (err error) {
	var (
		bodyLen   int
		headerLen int16
		packLen   int32
		buf       []byte
	)
	if _, buf, err = ws.ReadMessage(); err != nil {
		return
	}
	if len(buf) < _rawHeaderSize {
		return ErrProtoPackLen
	}
	packLen = binary.BigEndian.Int32(buf[_packOffset:_headerOffset])
	headerLen = binary.BigEndian.Int16(buf[_headerOffset:_verOffset])
	p.Ver = int64(binary.BigEndian.Int16(buf[_verOffset:_opOffset]))
	p.Op = int64(binary.BigEndian.Int32(buf[_opOffset:_seqOffset]))
	p.Seq = int64(binary.BigEndian.Int32(buf[_seqOffset:]))
	if packLen < 0 || packLen > _maxPackSize {
		return ErrProtoPackLen
	}
	if headerLen != _rawHeaderSize {
		return ErrProtoHeaderLen
	}
	if bodyLen = int(packLen - int32(headerLen)); bodyLen > 0 {
		p.Body = buf[headerLen:packLen]
	} else {
		p.Body = nil
	}
	return
}

// WriteWebsocket write a proto to websocket connection.
func (p *Proto) WriteWebsocket(ws *websocket.Conn) (err error) {
	var (
		buf     []byte
		packLen int
	)
	packLen = _rawHeaderSize + len(p.Body)
	if err = ws.WriteHeader(websocket.BinaryMessage, packLen); err != nil {
		return
	}
	if buf, err = ws.Peek(_rawHeaderSize); err != nil {
		return
	}
	binary.BigEndian.PutInt32(buf[_packOffset:], int32(packLen))
	binary.BigEndian.PutInt16(buf[_headerOffset:], int16(_rawHeaderSize))
	binary.BigEndian.PutInt16(buf[_verOffset:], int16(p.Ver))
	binary.BigEndian.PutInt32(buf[_opOffset:], int32(p.Op))
	binary.BigEndian.PutInt32(buf[_seqOffset:], int32(p.Seq))
	if p.Body != nil {
		err = ws.WriteBody(p.Body)
	}
	return
}

func (p *Proto) WriteWebsocketHeart(wr *websocket.Conn, online int64) (err error) {
	var (
		buf     []byte
		packLen int
	)
	packLen = _rawHeaderSize + _heartSize
	// websocket header
	if err = wr.WriteHeader(websocket.BinaryMessage, packLen); err != nil {
		return
	}
	if buf, err = wr.Peek(packLen); err != nil {
		return
	}
	// proto header
	binary.BigEndian.PutInt32(buf[_packOffset:], int32(packLen))
	binary.BigEndian.PutInt16(buf[_headerOffset:], int16(_rawHeaderSize))
	binary.BigEndian.PutInt16(buf[_verOffset:], int16(p.Ver))
	binary.BigEndian.PutInt32(buf[_opOffset:], int32(p.Op))
	binary.BigEndian.PutInt32(buf[_seqOffset:], int32(p.Seq))
	// proto body
	binary.BigEndian.PutInt32(buf[_heartOffset:], int32(online))
	return
}
