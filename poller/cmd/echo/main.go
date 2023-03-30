package main

import (
	"syscall"
	"time"

	"github.com/weedge/lib/log"
	"github.com/weedge/lib/poller"
)

type MockDecoder struct {
}

func (*MockDecoder) Decode(c *poller.Conn) (err error) {
	buff := c.GetBuff()
	bytes := buff.Buf[:buff.Len()]
	log.Info("read:", len(bytes), " bytes from fd:", c.GetFd())
	_, err = syscall.Write(int(c.GetFd()), bytes)

	return
}

type MockServerHandler struct {
}

func (m *MockServerHandler) OnConnect(c *poller.Conn) {
	log.Infof("connect fd %d addr %s", c.GetFd(), c.GetAddr())
}

func (m *MockServerHandler) OnMessage(c *poller.Conn, bytes []byte) {
}

func (m *MockServerHandler) OnClose(c *poller.Conn, err error) {
	log.Infof("close: %d err: %s", c.GetFd(), err.Error())
}

func main() {
	server, err := poller.NewServer(":8081", &MockServerHandler{}, &MockDecoder{},
		poller.WithTimeout(10*time.Second, 3600*time.Second), poller.WithReadBufferLen(1024))
	if err != nil {
		log.Info("err")
		return
	}

	server.Run()
}
