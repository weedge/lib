package main

import (
	"fmt"
	"time"

	"github.com/weedge/lib/log"
	"github.com/weedge/lib/poller"
	"github.com/weedge/lib/poller/cmd/common"
)

type MockServerHandler struct {
}

func (m *MockServerHandler) OnConnect(c *poller.Conn) {
	log.Info("connect:", c.GetFd(), c.GetAddr())
}

func (m *MockServerHandler) OnMessage(c *poller.Conn, bytes []byte) {
	log.Info("read:", string(bytes), "from fd:", c.GetFd())
	encoder := poller.NewHeaderLenEncoder(common.HeaderLen, common.MaxLen)
	res := fmt.Sprintf("got it: %s", bytes)
	err := encoder.EncodeToWriter(c, []byte(res))
	if err != nil {
		log.Error("res", res, "EncodeToFD err:", err.Error())
	}
	log.Info("res", res, "EncodeToFD", c.GetFd(), "ok")
}

func (m *MockServerHandler) OnClose(c *poller.Conn, err error) {
	log.Info("close:", c.GetFd(), err)
}

func main() {
	server, err := poller.NewServer(":8085", &MockServerHandler{}, poller.WithDecoder(poller.NewHeaderLenDecoder(common.HeaderLen)),
		poller.WithTimeout(10*time.Second, 3600*time.Second), poller.WithReadBufferLen(10))
	if err != nil {
		log.Info("err")
		return
	}

	server.Run()
}
