package poller

import (
	"errors"
)

var (
	ErrReadTimeout     = errors.New("tcp read timeout")
	ErrBufferNotEnough = errors.New("buffer not enough")
)

// Handler Server 注册接口
type Handler interface {
	OnConnect(c *Conn)               // OnConnect 当TCP长连接建立成功是回调
	OnMessage(c *Conn, bytes []byte) // OnMessage 当客户端有数据写入是回调
	OnClose(c *Conn, err error)      // OnClose 当客户端主动断开链接或者超时时回调,err返回关闭的原因
}

