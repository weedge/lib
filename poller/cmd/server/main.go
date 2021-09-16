package main

import (
	"github.com/weedge/lib/log"
	"github.com/weedge/lib/poller"
	"github.com/weedge/lib/poller/cmd/common"
	"time"
)

type MockServerHandler struct {
}

func (m *MockServerHandler) OnConnect(c *poller.Conn) {
	log.Info("connect:", c.GetFd(), c.GetAddr())
}

func (m *MockServerHandler) OnMessage(c *poller.Conn, bytes []byte) {
	err := poller.NewHeaderLenEncoder(common.HeaderLen, common.MaxLen).EncodeToFD(c.GetFd(), bytes)
	log.Info("read:", string(bytes), err)
}

func (m *MockServerHandler) OnClose(c *poller.Conn, err error) {
	log.Info("close:", c.GetFd(), err)
}

func main() {
	server, err := poller.NewServer(":8085", &MockServerHandler{}, poller.NewHeaderLenDecoder(common.HeaderLen),
		poller.WithTimeout(10*time.Second, 3600*time.Second), poller.WithReadBufferLen(10))
	if err != nil {
		log.Info("err")
		return
	}

	server.Run()
}
