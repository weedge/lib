package poller

import (
	"runtime"
	"time"

	"github.com/ii64/gouring"
)

type IOMode int32

const (
	IOModeUnkonw IOMode = iota
	IOModeUring
)

// options Server opt config
type options struct {
	readBufferLen     int                    // 所读取的客户端包的最大长度，客户端发送的包不能超过这个长度，默认值是1024字节
	acceptGNum        int                    // 处理接受请求的goroutine数量
	ioGNum            int                    // 处理io的goroutine数量
	ioEventQueueLen   int                    // io事件队列长度
	timeoutTicker     time.Duration          // 超时时间检查间隔
	timeout           time.Duration          // 超时时间
	keepaliveInterval time.Duration          // tcp keepalive interval
	listenBacklog     int                    // listen bakklog size
	ioMode            IOMode                 // io event mode: io_uring, other
	ioUringParams     *gouring.IoUringParams // io_uring_setup params
	ioUringEntries    uint32                 // io_uring setup sqe entry array size
}

type Option interface {
	apply(*options)
}

type funcServerOption struct {
	f func(*options)
}

func (fdo *funcServerOption) apply(do *options) {
	fdo.f(do)
}

func newFuncServerOption(f func(*options)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

func WithReadBufferLen(len int) Option {
	return newFuncServerOption(func(o *options) {
		if len <= 0 {
			panic("acceptGNum must greater than 0")
		}
		o.readBufferLen = len
	})
}

func WithAcceptGNum(num int) Option {
	return newFuncServerOption(func(o *options) {
		if num <= 0 {
			panic("acceptGNum must greater than 0")
		}
		o.acceptGNum = num
	})
}

func WithIOGNum(num int) Option {
	return newFuncServerOption(func(o *options) {
		if num <= 0 {
			panic("IOGNum must greater than 0")
		}
		o.ioGNum = num
	})
}

func WithIOEventQueueLen(num int) Option {
	return newFuncServerOption(func(o *options) {
		if num <= 0 {
			panic("ioEventQueueLen must greater than 0")
		}
		o.ioEventQueueLen = num
	})
}

func WithTimeout(timeoutTicker, timeout time.Duration) Option {
	return newFuncServerOption(func(o *options) {
		if timeoutTicker <= 0 {
			panic("timeoutTicker must greater than 0")
		}
		if timeout <= 0 {
			panic("timeoutTicker must greater than 0")
		}

		o.timeoutTicker = timeoutTicker
		o.timeout = timeout
	})
}

func WithKeepAliveInterval(d time.Duration) Option {
	return newFuncServerOption(func(o *options) {
		if d <= 0 {
			panic("keepalive interval must greater than 0")
		}
		o.keepaliveInterval = d
	})
}

func WithListenBacklog(size int) Option {
	return newFuncServerOption(func(o *options) {
		if size <= 0 {
			panic("listen backlog size must greater than 0")
		}
		o.listenBacklog = size
	})
}

func WithIoUringParams(params *gouring.IoUringParams) Option {
	return newFuncServerOption(func(o *options) {
		o.ioUringParams = params
	})
}

func WithIoUringEntries(n uint32) Option {
	return newFuncServerOption(func(o *options) {
		o.ioUringEntries = n
	})
}

func WithIoMode(mode IOMode) Option {
	return newFuncServerOption(func(o *options) {
		o.ioMode = mode
	})
}

func getOptions(opts ...Option) *options {
	cpuNum := runtime.NumCPU()
	options := &options{
		readBufferLen:   1024,
		acceptGNum:      cpuNum,
		ioGNum:          cpuNum,
		ioEventQueueLen: 1024,
		listenBacklog:   1024,
		ioUringEntries:  1024,
	}

	for _, o := range opts {
		o.apply(options)
	}
	return options
}
