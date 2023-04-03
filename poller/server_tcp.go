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

const (
	EventIn      = 1 // 数据流入
	EventClose   = 2 // 断开连接
	EventTimeout = 3 // 检测到超时
)

type event struct {
	FD   int32 // 文件描述符
	Type int32 // 时间类型
}

// Server TCP服务
type Server struct {
	options        *options     // 服务参数
	readBufferPool *sync.Pool   // 读缓存区内存池
	handler        Handler      // 注册的处理
	decoder        Decoder      // 解码器
	ioEventQueues  []chan event // IO事件队列集合
	ioQueueNum     int32        // IO事件队列集合数量
	conns          sync.Map     // TCP长连接管理
	connsNum       int64        // 当前建立的长连接数量
	stop           chan int     // 服务器关闭信号
}

// NewServer 创建server服务器
func NewServer(address string, handler Handler, decoder Decoder, opts ...Option) (*Server, error) {
	options := getOptions(opts...)

	// 初始化读缓存区内存池
	readBufferPool := &sync.Pool{
		New: func() interface{} {
			b := make([]byte, options.readBufferLen)
			return b
		},
	}

	// listen and create poller
	err := listen(address, options.listenBacklog)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// 初始化io事件队列
	ioEventQueues := make([]chan event, options.ioGNum)
	for i := range ioEventQueues {
		ioEventQueues[i] = make(chan event, options.ioEventQueueLen)
	}

	return &Server{
		options:        options,
		readBufferPool: readBufferPool,
		handler:        handler,
		decoder:        decoder,
		ioEventQueues:  ioEventQueues,
		ioQueueNum:     int32(options.ioGNum),
		conns:          sync.Map{},
		connsNum:       0,
		stop:           make(chan int),
	}, nil
}

// GetConn 获取Conn
func (s *Server) GetConn(fd int32) (*Conn, bool) {
	value, ok := s.conns.Load(fd)
	if !ok {
		return nil, false
	}
	return value.(*Conn), true
}

// Run 启动服务
func (s *Server) Run() {
	log.Info("start server run")
	s.startAcceptor()
	s.startIOConsumeHandler()
	s.checkTimeout()
	s.startIOEventLooper()
}

// GetConnsNum 获取当前长连接的数量
func (s *Server) GetConnsNum() int64 {
	return atomic.LoadInt64(&s.connsNum)
}

// Stop 启动服务
func (s *Server) Stop() {
	close(s.stop)
	for _, queue := range s.ioEventQueues {
		close(queue)
	}
}

// handleEvent 处理事件
func (s *Server) handleEvent(event event) {
	index := event.FD % s.ioQueueNum
	s.ioEventQueues[index] <- event
}

// startIOEventLooper 启动生产者 EvenLoop
func (s *Server) startIOEventLooper() {
	log.Info("start io producer")
	for {
		select {
		case <-s.stop:
			log.Error("stop producer")
			return
		default:
			events, err := getEvents()
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

// startAcceptor 开始接收连接请求
func (s *Server) startAcceptor() {
	for i := 0; i < s.options.acceptGNum; i++ {
		go s.accept()
	}
	log.Info(fmt.Sprintf("start accept by %d goroutine", s.options.acceptGNum))
}

// accept 接收连接请求
func (s *Server) accept() {
	for {
		select {
		case <-s.stop:
			return
		default:
			nfd, addr, err := accept(s.options.keepaliveInterval)
			if err != nil {
				log.Error(err)
				continue
			}

			fd := int32(nfd)
			conn := newConn(fd, addr, s)
			s.conns.Store(fd, conn)
			atomic.AddInt64(&s.connsNum, 1)
			s.handler.OnConnect(conn)
		}
	}
}

// startIOConsumeHandler 启动消费者
func (s *Server) startIOConsumeHandler() {
	for _, queue := range s.ioEventQueues {
		queue := queue
		go s.consumeIOEvent(queue)
	}
	log.Info(fmt.Sprintf("start io event consumer by %d goroutine handler", len(s.ioEventQueues)))
}

// consumeIOEvent 消费IO事件
func (s *Server) consumeIOEvent(queue chan event) {
	for event := range queue {
		v, ok := s.conns.Load(event.FD)
		if !ok {
			log.Error("not found in conns,", event.FD)
			continue
		}
		c := v.(*Conn)

		if event.Type == EventClose {
			c.Close()
			s.handler.OnClose(c, io.EOF)
			continue
		}
		if event.Type == EventTimeout {
			c.Close()
			s.handler.OnClose(c, ErrReadTimeout)
			continue
		}

		err := c.Read()
		if err != nil {
			// 服务端关闭连接
			if err == syscall.EBADF {
				continue
			}
			c.Close()
			s.handler.OnClose(c, err)

			log.Debug(err)
		}
	}
}

// checkTimeout 定时检查超时的TCP长连接
func (s *Server) checkTimeout() {
	if s.options.timeout == 0 || s.options.timeoutTicker == 0 {
		return
	}
	log.Info(fmt.Sprintf("check timeout goroutine run,check_time:%v,timeout:%v", s.options.timeoutTicker, s.options.timeout))
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
						s.handleEvent(event{FD: c.fd, Type: EventTimeout})
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
