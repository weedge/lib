package poller

import (
	"errors"
	"time"
)

var (
	ErrReadTimeout     = errors.New("tcp read timeout")
	ErrBufferNotEnough = errors.New("buffer not enough")

	ErrIOUringParamsFastPollUnAvailable = errors.New("IORING_FEAT_FAST_POLL not available in the kernel, quiting...")
	ErrIOUringSubmitFail                = errors.New("iouring submit fail")
	ErrIOUringWaitCqe                   = errors.New("iouring wait cqe fail")
)

// Handler Server for biz logic
type Handler interface {
	OnConnect(c *Conn)
	OnMessage(c *Conn, bytes []byte)
	OnClose(c *Conn, err error)
}

// defaultTCPKeepAlive is a default constant value for TCPKeepAlive times
// See golang.org/issue/31510
const (
	defaultTCPKeepAlive = 15 * time.Second
)

type Poller interface {
}
