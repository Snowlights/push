package websocket

import (
	"context"
	"fmt"
	"github.com/Snowlights/tool/vlog"
	"net"
	"runtime"
)

// 维护用户的conn链接信息，以及接受用户发送的消息推送给service部分
// 服务发现部分将依赖etcd
// 推送service部分将使用rpc的方式

func InitWebSocketServer(ctx context.Context, addr string) error {
	fun := "websocket.InitWebSocketServer -->"

	bind, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return fmt.Errorf("%s ResolveTCPAddr failed: %v", fun, err)
	}

	listener, err := net.ListenTCP("tcp", bind)
	if err != nil {
		return fmt.Errorf("%s ListenServAddr failed, error: %v", fun, err)
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go acceptWebSocket(ctx, listener)
	}

	return nil
}

func acceptWebSocket(ctx context.Context, listener *net.TCPListener) {
	fun := "websocket.acceptWebSocket -->"
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			vlog.ErrorF(ctx, "%s AcceptTCP failed, listen addr: %v, error: %v", fun, listener.Addr().String(), err)
			return
		}
		if err = conn.SetKeepAlive(false); err != nil {
			vlog.ErrorF(ctx, "%s conn.SetKeepAlive() error(%v)", fun, err)
			return
		}
		if err = conn.SetReadBuffer(4096); err != nil {
			vlog.ErrorF(ctx, "%s conn.SetReadBuffer() error(%v)", fun, err)
			return
		}
		if err = conn.SetWriteBuffer(4096); err != nil {
			vlog.ErrorF(ctx, "%s conn.SetWriteBuffer() error(%v)", fun, err)
			return
		}
		go servWebsocket(conn)
	}

}

func servWebsocket(conn net.Conn) {

}
