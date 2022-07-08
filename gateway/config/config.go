package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Server    *Server    `json:"server"`
	WebSocket *WebSocket `json:"websocket"`
}

type Server struct {
	Conn   Conn    `json:"conn"`
	Pool   *Pool   `json:"pool"`
	Bucket *Bucket `json:"bucket"`
}

type Conn struct {
	SendBuf     int    `json:"sendBuf"`
	RecvBuf     int    `json:"recvBuf"`
	ClientProto uint64 `json:"clientProto"`
	ServerProto uint64 `json:"serverProto"`
}

type Pool struct {
	ReaderCap  uint64 `json:"readerCap"`
	ReaderSize uint64 `json:"readerSize"`
	ReaderBuf  uint64 `json:"readerBuf"`
	WriterCap  uint64 `json:"writerCap"`
	WriterSize uint64 `json:"writerSize"`
	WriterBuf  uint64 `json:"writerBuf"`
}

type Bucket struct {
	Cap         uint64 `json:"cap"`
	Channel     uint64 `json:"channel"`
	Room        uint64 `json:"room"`
	Routine     uint64 `json:"routine"`
	RoutineSize uint64 `json:"routineSize"`
}

type WebSocket struct {
	Addr      string `json:"addr"`
	Keepalive bool   `json:"keepalive"`
}

func InitConfig(path string) (*Config, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	c := new(Config)
	return c, json.Unmarshal(body, c)
}
