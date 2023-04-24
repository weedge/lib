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
	stop           chan struct{}               // Indicates the server shutdown signal
	listenFD       int                         // listen fd
	pollerFD       int                         // event poller fd (epoll/kqueue, poll, select)
	iourings       []*ioUring                  // iouring async event rings
	asyncEventCb   map[EventType]EventCallBack // async event call back register
	looperWg       sync.WaitGroup              // main looper group wait
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
	var rings []*ioUring
	if options.ioMode == IOModeUring || options.ioMode == IOModeEpollUring {
		rings = make([]*ioUring, options.ioUringNum)
		for i := 0; i < options.ioUringNum; i++ {
			ring, err := newIoUring(options.ioUringEntries, options.ioUringParams)
			if err != nil {
				log.Errorf("newIoUring %d err %s", i, err.Error())
				return nil, err
			}

			// register eventfd
			if options.ioMode == IOModeEpollUring {
				err = ring.RegisterEventFd()
				if err != nil {
					log.Errorf("ring.RegisterEventFd %d err %s", i, err.Error())
					ring.CloseRing()
					return nil, err
				}
			}

			rings[i] = ring
		} // end for
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
		stop:           make(chan struct{}),
		listenFD:       lfd,
		pollerFD:       pollerFD,
		iourings:       rings,
		asyncEventCb:   map[EventType]EventCallBack{},
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
	log.Info("start server runing...")
	// rigister event
	s.rigisterEpollIouringEvent()

	// monitor
	s.report()
	s.checkTimeout()

	// start server
	s.startAcceptor()
	s.startIOConsumeHandler()
	s.startIOEventLooper()
}

// rigisterEpollIouringEvent
// rigister epoll iouring eventfd, wait eventfd iouring cq read ready event,
// notify iouring to get cqe
func (s *Server) rigisterEpollIouringEvent() {
	if s.options.ioMode != IOModeEpollUring {
		return
	}
	log.Info("start notify iouring cq event by epoll rigistered iouring eventfd")

	err := s.rigisterIoUringEvent()
	if err != nil {
		log.Errorf("rigisterIoUringEvent err %s", err.Error())
		return
	}

	go s.startNotifyIoUringCQEvent()
}

// rigisterIoUringEvent
// add read event to epoll item list for registered iouring eventfd (just read)
func (s *Server) rigisterIoUringEvent() (err error) {
	for i := 0; i < s.options.ioUringNum; i++ {
		log.Debugf("pollerFD %d eventfd %d add epoll readable event", s.pollerFD, s.iourings[i].eventfd)
		err = addReadEvent(s.pollerFD, s.iourings[i].eventfd)
		if err != nil {
			return
		}
	}
	return
}

// CloseIoUring
// remove rigistered  eventfd iouring event, free iouring mmap
func (s *Server) CloseIoUring() {
	for i := 0; i < len(s.iourings); i++ {
		delEventFD(s.pollerFD, s.iourings[i].eventfd)
		s.iourings[i].CloseRing()
	}
}

// Stop
// stop server, close communication channel(queue)
// free io uring
func (s *Server) Stop() {
	close(s.stop)
	for _, queue := range s.ioEventQueues {
		close(queue)
	}

	s.CloseIoUring()
}

// GetConnsNum
func (s *Server) GetConnsNum() int64 {
	return atomic.LoadInt64(&s.connsNum)
}

// startAcceptor
// setup accept connect goroutine
func (s *Server) startAcceptor() {
	if len(s.iourings) != 0 {
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

			conn := newConn(s.pollerFD, cfd, addr, s)
			s.conns.Store(cfd, conn)
			atomic.AddInt64(&s.connsNum, 1)

			err = addReadEvent(s.pollerFD, cfd)
			if err != nil {
				log.Error(err)
				conn.Close()
				continue
			}

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
	s.GetIoUring(s.listenFD).addAcceptSqe(s.getAcceptCallback(&rsa), s.listenFD, &rsa, len, 0)
}

func (s *Server) GetIoUring(fd int) *ioUring {
	return s.iourings[fd%s.options.ioUringNum]
}

func (s *Server) GetEventIoUring(efd int) *ioUring {
	if len(s.iourings) == 0 {
		return nil
	}
	for i := 0; i < s.options.ioUringNum; i++ {
		if efd == s.iourings[i].eventfd {
			return s.iourings[i]
		}
	}

	return nil
}

func (s *Server) startNotifyIoUringCQEvent() {
	for {
		select {
		case <-s.stop:
			return
		default:
			events, err := getEvents(s.pollerFD)
			if err != nil {
				if err != syscall.EINTR {
					log.Errorf("getEvents err %s", err.Error())
				}
				continue
			}
			for _, event := range events {
				ring := s.GetEventIoUring(event.fd)
				ring.cqeSignCh <- struct{}{}
			}
		}
	}
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

// startIOEventLooper main looper
// from poller events or io_uring cqe event entries
func (s *Server) startIOEventLooper() {
	//runtime.LockOSThread()
	if s.iourings == nil {
		s.startIOEventPollDispatcher()
		return
	}

	s.looperWg.Add(len(s.iourings))
	for i := 0; i < len(s.iourings); i++ {
		go s.startIOUringPollDispatcher(i)
	}
	s.looperWg.Wait()
}

// startIOEventPollDispatcher
// get ready events from poller, distpatch to event channel(queue)
func (s *Server) startIOEventPollDispatcher() {
	log.Info("start io event poll dispatcher")
	for {
		select {
		case <-s.stop:
			log.Infof("stop io event poll dispatcher")
			return
		default:
			var err error
			var events []eventInfo
			events, err = getEvents(s.pollerFD)
			if err != nil {
				if err != syscall.EINTR {
					log.Error(err)
				}
				continue
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
func (s *Server) startIOUringPollDispatcher(id int) {
	defer s.looperWg.Done()
	if s.iourings[id] == nil {
		return
	}
	log.Infof("start io_uring event op poll dispatcher id %d", id)
	for {
		select {
		case <-s.stop:
			log.Infof("stop io_uring event op poll dispatcher id %d", id)
			return
		default:
			event, err := s.iourings[id].getEventInfo()
			if err != nil {
				log.Warnf("id %d iouring get events error:%s continue", id, err.Error())
				continue
			}
			if event == nil {
				continue
			}

			// dispatch
			s.handleEvent(event)
			// commit cqe is seen
			s.iourings[id].cqeDone(event.cqe)
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
		go s.consumeIOEvent(queue)
	}
	log.Info(fmt.Sprintf("start io event consumer by %d goroutine handler", len(s.ioEventQueues)))
}

func (s *Server) consumeIOEvent(queue chan *eventInfo) {
	if s.iourings != nil {
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
			log.Warnf("fd %d not found in conns, event:%s , continue next event", event.fd, event)
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
			log.Warn("not found in conns,", event.fd, event)
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
			log.Warnf("process sync read connect fd %d err:%s , client connect must be disconnected", c.fd, err.Error())
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

	log.Infof("check timeout goroutine run,check_time:%v, timeout:%v", s.options.timeoutTicker, s.options.timeout)
	go func() {
		ticker := time.NewTicker(s.options.timeoutTicker)
		for {
			select {
			case <-s.stop:
				return
			case <-ticker.C:
				s.conns.Range(func(key, value interface{}) bool {
					c := value.(*Conn)
					//log.Infof("check connect %+v", c)
					if time.Since(c.lastReadTime) > s.options.timeout {
						s.handleEvent(&eventInfo{fd: int(c.fd), etype: ETypeTimeout})
					}
					return true
				})
			}
		} // end for
	}()
}

func (s *Server) report() {
	if s.options.reportTicker == 0 {
		return
	}

	log.Infof("start report server info, report tick time %v", s.options.reportTicker)
	go func() {
		ticker := time.NewTicker(s.options.reportTicker)
		for {
			select {
			case <-s.stop:
				return
			case <-ticker.C:
				n := s.GetConnsNum()
				if n > 0 {
					log.Infof("current active connect num %d", n)
				}
			}
		}
	}()
}
