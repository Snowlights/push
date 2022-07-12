package logic

import (
	"context"
	"fmt"
	"github.com/Snowlights/push/gateway/config"
	push_gateway "github.com/Snowlights/push/gateway/protocol"
	"github.com/Snowlights/push/gateway/websocket"
	"github.com/Snowlights/tool/vlog"
	"io"
	"net"
	"runtime"
)

// 维护用户的conn链接信息，以及接受用户发送的消息推送给service部分
// 服务发现部分将依赖etcd
// 推送service部分将使用rpc的方式

type WebSocketServer struct {
	c      *config.Config
	server *Server
}

func InitWebSocketServer(ctx context.Context, c *config.Config, s *Server) (*WebSocketServer, error) {
	fun := "websocket.InitWebSocketServer -->"

	serv := &WebSocketServer{
		c:      c,
		server: s,
	}

	bind, err := net.ResolveTCPAddr("tcp", serv.c.WebSocket.Addr)
	if err != nil {
		return nil, fmt.Errorf("%s ResolveTCPAddr failed: %v", fun, err)
	}

	listener, err := net.ListenTCP("tcp", bind)
	if err != nil {
		return nil, fmt.Errorf("%s ListenServAddr failed, error: %v", fun, err)
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go serv.acceptWebSocket(ctx, listener)
	}

	return serv, nil
}

func (s *WebSocketServer) acceptWebSocket(ctx context.Context, listener *net.TCPListener) {
	fun := "websocket.acceptWebSocket -->"
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			vlog.ErrorF(ctx, "%s AcceptTCP failed, listen addr: %v, error: %v", fun, listener.Addr().String(), err)
			return
		}
		if err = conn.SetKeepAlive(s.c.WebSocket.Keepalive); err != nil {
			vlog.ErrorF(ctx, "%s conn.SetKeepAlive error is %v", fun, err)
			return
		}
		if err = conn.SetReadBuffer(s.c.Server.Conn.RecvBuf); err != nil {
			vlog.ErrorF(ctx, "%s conn.SetReadBuffer error is %v", fun, err)
			return
		}
		if err = conn.SetWriteBuffer(s.c.Server.Conn.SendBuf); err != nil {
			vlog.ErrorF(ctx, "%s conn.SetWriteBuffer error is %v", fun, err)
			return
		}
		go s.servWebsocket(conn)
	}
}

func (s *WebSocketServer) servWebsocket(conn net.Conn) {
	reader, writer, bucket := s.server.getReader(), s.server.getWriter(), s.server.getRandomBucket()
	channel := NewChannel(s.c.Server.Conn.ClientProto, s.c.Server.Conn.ServerProto)
	channel.Reader.ResetBuffer(conn, reader.Bytes())
	channel.Writer.ResetBuffer(conn, writer.Bytes())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		conn.Close()
		s.server.putReader(reader)
		s.server.putWriter(writer)
	}()

	channel.IP, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
	step := 1
	req, err := websocket.ReadRequest(&channel.Reader)
	if err != nil {
		if err != io.EOF {
			vlog.ErrorF(ctx, "http.ReadRequest  error is %v", err)
		}
		return
	}

	step = 2
	ws, err := websocket.Upgrade(conn, &channel.Reader, &channel.Writer, req)
	if err != nil {
		if err != io.EOF {
			vlog.ErrorF(ctx, "websocket.NewServerConn error is %v", err)
		}
		return
	}
	defer func() {
		ws.Close()
	}()
	// must not setadv, only used in auth
	step = 3
	p, err := channel.ClientProto.SetProto()
	if err == nil {
		uData, err := s.server.authWebsocket(ctx, ws, p, req.Header.Get("Cookie"))
		if err == nil {
			channel.Accept(uData.Accepts...)
			bucket.PutChannel("", channel)
			channel.Key = uData.UserInfo.Key
		}
	}
	if err != nil {
		if err != io.EOF && err != websocket.ErrMessageClose {
			vlog.ErrorF(ctx, "key: %s remoteIP: %s step: %d ws handshake failed error(%v)", channel.Key, conn.RemoteAddr().String(), step, err)
		}
		return
	}
	go s.dispatchWebsocket(ctx, ws, channel)
	for {
		if p, err = channel.ClientProto.SetProto(); err != nil {
			break
		}
		if err = p.ReadWebsocket(ws); err != nil {
			break
		}
		if p.Op == int64(push_gateway.OpType_OpType_HeartBeat) {
			// todo expire token
			// NOTE: send server heartbeat for a long time
			p.Op = int64(push_gateway.OpType_OpType_HeartBeatReply)
			p.Body = nil
		} else {
			if err = s.server.Operate(ctx, p, channel, bucket); err != nil {
				break
			}
		}
		channel.ClientProto.SetDone()
	}

	ws.Close()
	if err = s.server.DisConnect(ctx, channel.Key); err != nil {
		vlog.ErrorF(ctx, "key: %s operator do disconnect error is %v", channel.Key, err)
	}
}

func (s *WebSocketServer) dispatchWebsocket(ctx context.Context, ws *websocket.Conn, ch *Channel) {
	var err error
	for {
		var p = ch.Ready()
		switch p.Op {
		case int64(push_gateway.OpType_OpType_Disconnect):
			goto failed
			// fetch message from svrbox(client send)
		default:
			for {
				if p, err = ch.ClientProto.GetProto(); err != nil {
					break
				}
				if p.Op == int64(push_gateway.OpType_OpType_HeartBeat) {
					var online int64
					if ch.Room != nil {
						online = ch.Room.Online()
					}
					if err = p.WriteWebsocketHeart(ws, online); err != nil {
						goto failed
					}
				} else {
					if err = p.WriteWebsocket(ws); err != nil {
						goto failed
					}
				}
				p.Body = nil // avoid memory leak
				ch.ClientProto.GetDone()
			}
		}
		// only hungry flush response
		if err = ws.Flush(); err != nil {
			break
		}
	}
failed:
	if err != nil && err != io.EOF && err != websocket.ErrMessageClose {
		vlog.ErrorF(ctx, "key: %s dispatch ws error is %v ", ch.Key, err)
	}
	ws.Close()
}
