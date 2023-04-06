package poller

import (
	"errors"
	"io"
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
	ErrIOUringReadFail            = errors.New("iouring read event op failed")
	ErrIOUringWriteFail           = errors.New("iouring write event op failed")
)

// Handler Server for biz logic
type Handler interface {
	OnConnect(c *Conn)
	OnMessage(c *Conn, bytes []byte)
	OnClose(c *Conn, err error)
}

// Decoder
type Decoder interface {
	Decode(c *Conn) error
}

// Encoder
type Encoder interface {
	EncodeToWriter(w io.Writer, bytes []byte) error
}

// defaultTCPKeepAlive is a default constant value for TCPKeepAlive times
// See golang.org/issue/31510
const (
	defaultTCPKeepAlive = 15 * time.Second
)

type Poller interface {
}
