package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/weedge/lib/log"
	"github.com/weedge/lib/poller"
)

type MockDecoder struct {
}

func (*MockDecoder) Decode(c *poller.Conn) (err error) {
	buff := c.GetBuff()
	bytes := buff.ReadAll()
	//log.Infof("read:%s len:%d bytes from fd:%d", bytes, len(bytes), c.GetFd())
	_, err = c.Write(bytes)

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
	server, err := poller.NewServer(":8081", &MockServerHandler{}, poller.WithDecoder(&MockDecoder{}),
		poller.WithTimeout(10*time.Second, 3600*time.Second), poller.WithReadBufferLen(128))
	if err != nil {
		log.Info("err")
		return
	}

	go server.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("server stop")
	server.Stop()
}
