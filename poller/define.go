package poller

import (
	"errors"
	"time"
)

var (
	ErrReadTimeout     = errors.New("tcp read timeout")
	ErrBufferNotEnough = errors.New("buffer not enough")

	ErrIOUringFeaturesUnAvailable = errors.New("required IORING_FEAT_SINGLE_MMAP | IORING_FEAT_FAST_POLL | IORING_FEAT_NODROP not available in the kernel")
	ErrIOUringRegisterFDFail      = errors.New("iouring register fd failed")
	ErrIOUringSubmitFail          = errors.New("iouring submit failed")
	ErrIOUringSubmitedNoFull      = errors.New("iouring submited no full")
	ErrIOUringWaitCqeFail         = errors.New("iouring wait cqe failed")
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
