package poller

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ii64/gouring"
	"github.com/weedge/lib/log"
)

// Server TCP服务
type Server struct {
	options        *options         // 服务参数
	readBufferPool *sync.Pool       // 读缓存区内存池
	handler        Handler          // 注册的处理
	decoder        Decoder          // 解码器
	ioEventQueues  []chan eventInfo // IO事件队列集合
	ioQueueNum     int              // IO事件队列集合数量
	conns          sync.Map         // TCP长连接管理
	connsNum       int64            // 当前建立的长连接数量
	stop           chan int         // 服务器关闭信号
	listenFD       int              // listen fd
	pollerFD       int              // event poller fd (epoll/kqueue, poll, select)
	iouring        *gouring.IoUring // iouring async event
}

// NewServer
// init server to start
func NewServer(address string, handler Handler, decoder Decoder, opts ...Option) (*Server, error) {
	options := getOptions(opts...)

	// init read buffer pool
	readBufferPool := &sync.Pool{
		New: func() interface{} {
			b := make([]byte, options.readBufferLen)
			return b
		},
	}

	// listen
	lfd, err := listen(address, options.listenBacklog)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// init poller(io_uring,epoll)
	var pollerFD int
	pollerFD, err = createPoller()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// init io_uring setup
	var ring *gouring.IoUring
	if options.ioMode == IOModeUring {
		ring, err = newRing(options.ioUringEntries, options.ioUringParams)
	}

	// init io event channel(queue)
	ioEventQueues := make([]chan eventInfo, options.ioGNum)
	for i := range ioEventQueues {
		ioEventQueues[i] = make(chan eventInfo, options.ioEventQueueLen)
	}

	return &Server{
		options:        options,
		readBufferPool: readBufferPool,
		handler:        handler,
		decoder:        decoder,
		ioEventQueues:  ioEventQueues,
		ioQueueNum:     options.ioGNum,
		conns:          sync.Map{},
		connsNum:       0,
		stop:           make(chan int),
		listenFD:       lfd,
		pollerFD:       pollerFD,
		iouring:        ring,
	}, nil
}

// GetConn
// get connect by connect fd from session connect sync Map
func (s *Server) GetConn(fd int32) (*Conn, bool) {
	value, ok := s.conns.Load(fd)
	if !ok {
		return nil, false
	}
	return value.(*Conn), true
}

// Run run server
// acceptor accept connet fron listen fd,
// eventLooper dispatch event,
// ioConsumeHandler hanle event for biz logic
// check time out connenct session,
func (s *Server) Run() {
	log.Info("start server run")
	s.startAcceptor()
	s.startIOConsumeHandler()
	s.checkTimeout()
	s.startIOEventLooper()
}

// GetConnsNum
func (s *Server) GetConnsNum() int64 {
	return atomic.LoadInt64(&s.connsNum)
}

// Stop
// stop server, close communication channel(queue)
func (s *Server) Stop() {
	close(s.stop)
	for _, queue := range s.ioEventQueues {
		close(queue)
	}
}

// handleEvent
// use hash dispatch event to channel(queue)
// need balance
func (s *Server) handleEvent(event eventInfo) {
	index := event.fd % s.ioQueueNum
	s.ioEventQueues[index] <- event
}

// startIOEventLooper
// from poller events or io_uring cqe event entries
func (s *Server) startIOEventLooper() {
	log.Info("start io producer")
	for {
		select {
		case <-s.stop:
			log.Error("stop producer")
			return
		default:
			var err error
			var events []eventInfo
			events, err = getEvents(s.pollerFD)
			if err != nil {
				log.Error(err)
			}

			// dispatch
			for i := range events {
				s.handleEvent(events[i])
			}
		}
	}
}

// startAcceptor
// setup accept connect goroutine
func (s *Server) startAcceptor() {
	for i := 0; i < s.options.acceptGNum; i++ {
		go s.accept()
	}
	log.Info(fmt.Sprintf("start accept by %d goroutine", s.options.acceptGNum))
}

// accept
// block accept connect from listen fd
// save non block connect fd session and OnConnect logic handle
func (s *Server) accept() {
	for {
		select {
		case <-s.stop:
			return
		default:
			cfd, socketAddr, err := accept(s.listenFD, s.options.keepaliveInterval)
			if err != nil {
				log.Error(err)
				continue
			}
			addr := getAddr(socketAddr)

			err = addReadEvent(s.pollerFD, cfd)
			if err != nil {
				log.Error(err)
				continue
			}

			conn := newConn(s.pollerFD, cfd, addr, s)
			s.conns.Store(cfd, conn)
			atomic.AddInt64(&s.connsNum, 1)
			s.handler.OnConnect(conn)
		}
	}
}

func (s *Server) asyncAccept() {
}

// startIOConsumeHandler
// setup io event consume goroutine
func (s *Server) startIOConsumeHandler() {
	for _, queue := range s.ioEventQueues {
		queue := queue
		go s.consumeIOEvent(queue)
	}
	log.Info(fmt.Sprintf("start io event consumer by %d goroutine handler", len(s.ioEventQueues)))
}

// consumeIOEvent
// handle r/w, close, connect timeout etc event
func (s *Server) consumeIOEvent(queue chan eventInfo) {
	for event := range queue {
		v, ok := s.conns.Load(event.fd)
		if !ok {
			log.Error("not found in conns,", event.fd)
			continue
		}
		c := v.(*Conn)

		if event.etype == ETypeClose {
			c.Close()
			s.handler.OnClose(c, io.EOF)
			continue
		}
		if event.etype == ETypeTimeout {
			c.Close()
			s.handler.OnClose(c, ErrReadTimeout)
			continue
		}

		err := c.Read()
		if err != nil {
			// notice: if next connect use closed cfd (TIME_WAIT stat between 2MSL eg:4m),
			// read from closed cfd return EBADF
			if err == syscall.EBADF {
				continue
			}

			// close and free connect
			c.Close()
			s.handler.OnClose(c, err)

			log.Debug(err)
		}
	}
}

// checkTimeout
// tick to check connect time out
func (s *Server) checkTimeout() {
	if s.options.timeout == 0 || s.options.timeoutTicker == 0 {
		return
	}

	log.Infof("check timeout goroutine run,check_time:%v,timeout:%v", s.options.timeoutTicker, s.options.timeout)
	go func() {
		ticker := time.NewTicker(s.options.timeoutTicker)
		for {
			select {
			case <-s.stop:
				return
			case <-ticker.C:
				s.conns.Range(func(key, value interface{}) bool {
					c := value.(*Conn)

					if time.Since(c.lastReadTime) > s.options.timeout {
						s.handleEvent(eventInfo{fd: int(c.fd), etype: ETypeTimeout})
					}
					return true
				})
			}
		}
	}()
}

// roundDurationUp rounds d to the next multiple of to.
func roundDurationUp(d time.Duration, to time.Duration) time.Duration {
	return (d + to - 1) / to
}
