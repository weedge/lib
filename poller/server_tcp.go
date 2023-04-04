package poller

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/weedge/lib/log"
)

// Server TCP服务
type Server struct {
	options        *options          // 服务参数
	readBufferPool *sync.Pool        // 读缓存区内存池
	handler        Handler           // 注册的处理
	decoder        Decoder           // 解码器
	ioEventQueues  []chan *eventInfo // IO事件队列集合
	ioQueueNum     int               // IO事件队列集合数量
	conns          sync.Map          // TCP长连接管理
	connsNum       int64             // 当前建立的长连接数量
	stop           chan int          // 服务器关闭信号
	listenFD       int               // listen fd
	pollerFD       int               // event poller fd (epoll/kqueue, poll, select)
	iouring        *ioUring          // iouring async event
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
	var ring *ioUring
	if options.ioMode == IOModeUring {
		ring, err = newIoUring(options.ioUringEntries, options.ioUringParams)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	// init io event channel(queue)
	ioEventQueues := make([]chan *eventInfo, options.ioGNum)
	for i := range ioEventQueues {
		ioEventQueues[i] = make(chan *eventInfo, options.ioEventQueueLen)
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
// free io uring
func (s *Server) Stop() {
	close(s.stop)
	for _, queue := range s.ioEventQueues {
		close(queue)
	}

	if s.iouring != nil {
		s.iouring.CloseRing()
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

// nonBlockPollAccept
// non block accept, when return EAGAIN, add/produce event poll op to sqe
func (s *Server) nonBlockPollAccept() {
}

// asyncBlockAccept
// async add/produce block accept op to sqe
func (s *Server) asyncBlockAccept() {
}

// startIOEventLooper
// from poller events or io_uring cqe event entries
func (s *Server) startIOEventLooper() {
	if s.iouring != nil {
		s.startIOUringPollDispatcher()
	} else {
		s.startIOEventPollDispatcher()
	}
}

// startIOEventPollDispatcher
// get ready events from poller, distpatch to event channel(queue)
func (s *Server) startIOEventPollDispatcher() {
	log.Info("start io event poll dispatcher")
	for {
		select {
		case <-s.stop:
			log.Error("stop io event poll dispatcher")
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
				s.handleEvent(&events[i])
			}
		}
	} // end for
}

// startIOUringPollDispatcher
// get completed event ops from io_uring cqe event entries, distpatch to event channel(queue)
func (s *Server) startIOUringPollDispatcher() {
	log.Info("start io_uring event op poll dispatcher")
	for {
		select {
		case <-s.stop:
			log.Error("stop io_uring event op poll dispatcher")
			return
		default:
			event, err := s.iouring.getEventInfo()
			if err != nil {
				log.Error(err)
			}

			// dispatch
			s.handleEvent(event)
		}
	} // end for
}

// handleEvent
// use hash dispatch event to channel(queue)
// need balance
func (s *Server) handleEvent(event *eventInfo) {
	index := event.fd % s.ioQueueNum
	s.ioEventQueues[index] <- event
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

func (s *Server) consumeIOEvent(queue chan *eventInfo) {
	if s.iouring != nil {
		s.consumeIOCompletionEvent(queue)
	} else {
		s.consumeIOReadyEvent(queue)
	}
}

// consumeIOCompletionEvent
func (s *Server) consumeIOCompletionEvent(queue chan *eventInfo) {
}

// consumeIOReadyEvent
// handle ready r/w, close, connect timeout etc event
func (s *Server) consumeIOReadyEvent(queue chan *eventInfo) {
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
	} // end for
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
						s.handleEvent(&eventInfo{fd: int(c.fd), etype: ETypeTimeout})
					}
					return true
				})
			}
		} // end for
	}()
}

// roundDurationUp rounds d to the next multiple of to.
func roundDurationUp(d time.Duration, to time.Duration) time.Duration {
	return (d + to - 1) / to
}
