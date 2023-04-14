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

// Server TCP server
type Server struct {
	options        *options                    // Service parameters
	readBufferPool *sync.Pool                  // Read cache memory pool
	handler        Handler                     // Indicates the processing of registration
	ioEventQueues  []chan *eventInfo           // IO A collection of event queues
	ioQueueNum     int                         // Number of I/O event queues
	conns          sync.Map                    // TCP long connection management
	connsNum       int64                       // Indicates the number of established long connections
	stop           chan int                    // Indicates the server shutdown signal
	listenFD       int                         // listen fd
	pollerFD       int                         // event poller fd (epoll/kqueue, poll, select)
	iouring        *ioUring                    // iouring async event
	asyncEventCb   map[EventType]EventCallBack // async event call back register
}

// NewServer
// init server to start
func NewServer(address string, handler Handler, opts ...Option) (*Server, error) {
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

	// init poller(epoll/kqueue)
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
		ioEventQueues:  ioEventQueues,
		ioQueueNum:     options.ioGNum,
		conns:          sync.Map{},
		connsNum:       0,
		stop:           make(chan int),
		listenFD:       lfd,
		pollerFD:       pollerFD,
		iouring:        ring,
		asyncEventCb:   map[EventType]EventCallBack{},
	}, nil
}

func (s *Server) registerEventCb() {

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
	if s.iouring != nil {
		go s.asyncBlockAccept()
		log.Info("start trigger async block accept")
		return
	}

	for i := 0; i < s.options.acceptGNum; i++ {
		go s.accept()
	}
	log.Infof("start accept by %d goroutine", s.options.acceptGNum)
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
	var rsa syscall.RawSockaddrAny
	var len uint32 = syscall.SizeofSockaddrAny
	s.iouring.addAcceptSqe(s.getAcceptCallback(&rsa), s.listenFD, &rsa, len, 0)
}

func (s *Server) getAcceptCallback(rsa *syscall.RawSockaddrAny) EventCallBack {
	return func(e *eventInfo) (err error) {
		if e.cqe.Res < 0 {
			err = fmt.Errorf("accept err res %d", e.cqe.Res)
			return
		}

		cfd := int(e.cqe.Res)
		err = setConnectOption(cfd, s.options.keepaliveInterval)
		if err != nil {
			return
		}

		socketAddr, err := anyToSockaddr(rsa)
		if err != nil {
			return
		}
		addr := getAddr(socketAddr)

		conn := newConn(s.pollerFD, cfd, addr, s)
		s.conns.Store(cfd, conn)
		atomic.AddInt64(&s.connsNum, 1)
		s.handler.OnConnect(conn)

		// new connected client, async read data from socket
		conn.AsyncBlockRead()

		// re-add accept to monitor for new connections
		s.asyncBlockAccept()

		return
	}
}

// startIOEventLooper
// from poller events or io_uring cqe event entries
func (s *Server) startIOEventLooper() {
	//runtime.LockOSThread()
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
			log.Info("stop io_uring event op poll dispatcher")
			return
		default:
			event, err := s.iouring.getEventInfo()
			if err != nil {
				log.Warnf("iouring get events error:%s continue", err.Error())
				continue
			}
			if event == nil {
				continue
			}

			// dispatch
			s.handleEvent(event)
			// commit cqe is seen
			s.iouring.cqeDone(event.cqe)
		}
	} // end for
}

// handleEvent
// use hash dispatch event to channel(queue)
// need balance(hash fd, same connect have orderly event process)
// golang scheduler have a good way to schedule thread in bound cpu affinity
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
	for event := range queue {
		// process async accept connect complete event
		if event.etype == ETypeAccept {
			err := event.cb(event)
			if err != nil {
				log.Errorf("accept event %s cb error:%s, continue next event", event, err.Error())
			}
			continue
		}

		// get connect from fd
		v, ok := s.conns.Load(event.fd)
		if !ok {
			log.Errorf("fd %d not found in conns, event:%s , continue next event", event.fd, event)
			continue
		}
		c := v.(*Conn)

		// process async read complete event
		if event.etype == ETypeRead {
			err := c.processReadEvent(event)
			if err != nil {
				// notice: if next connect use closed cfd (TIME_WAIT stat between 2MSL eg:4m),
				// read from closed cfd return EBADF
				if err == syscall.EBADF {
					log.Errorf("read closed connect fd %d EBADF, continue next event", event.fd)
					continue
				}

				// no bytes available on socket, client must be disconnected
				log.Warnf("process read event %s err:%s , client connect must be disconnected", event, err.Error())
				// close and free connect
				c.CloseConnect()
				s.handler.OnClose(c, err)
			}
		}

		// async write complete event
		if event.etype == ETypeWrite {
			err := c.processWirteEvent(event)
			if err != nil {
				log.Errorf("process write event %s err:%s, continue next event", event, err.Error())
				continue
			}
		}
	} // end for
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

			// no bytes available on socket, client must be disconnected
			log.Warnf("process sync read err:%s , client connect must be disconnected", err.Error())
			// close and free connect
			c.Close()
			s.handler.OnClose(c, err)

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
