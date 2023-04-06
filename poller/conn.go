package poller

import (
	"sync/atomic"
	"syscall"
	"time"

	"github.com/weedge/lib/log"
)

// Conn keepalive connection
type Conn struct {
	server       *Server     // server reference
	pollerFD     int         // event poller File descriptor
	fd           int         // socket connect File descriptor
	addr         string      // peer address
	buffer       *Buffer     // Read the buffer
	lastReadTime time.Time   // Time of last read
	data         interface{} // Business custom data, used as an extension
	ioMode       IOMode      // io mode
}

// newConn create tcp connection
func newConn(pollerFD, fd int, addr string, server *Server, ioMode IOMode) *Conn {
	return &Conn{
		server:       server,
		pollerFD:     pollerFD,
		fd:           fd,
		addr:         addr,
		buffer:       NewBuffer(server.readBufferPool.Get().([]byte)),
		lastReadTime: time.Now(),
		ioMode:       ioMode,
	}
}

// GetFd gets the file descriptor
func (c *Conn) GetFd() int {
	return c.fd
}

// GetAddr gets the client address
func (c *Conn) GetAddr() string {
	return c.addr
}

// GetAddr gets the conn buff
func (c *Conn) GetBuff() *Buffer {
	return c.buffer
}

// Read
// block read bytes until read readBufferLen bytes from connect fd
func (c *Conn) Read() error {
	c.lastReadTime = time.Now()
	fd := c.GetFd()
	for {
		err := c.buffer.ReadFromFD(fd)
		if err != nil {
			// There is no data to read in the buffer
			if err == syscall.EAGAIN {
				return nil
			}
			return err
		}

		if c.server.options.decoder == nil {
			c.server.handler.OnMessage(c, c.buffer.ReadAll())
			continue
		}

		err = c.server.options.decoder.Decode(c)
		if err != nil {
			return err
		}
	}
}

// AsyncBlockRead  trigger a async kernerl block read from connect fd to buff
func (c *Conn) AsyncBlockRead() {
	c.lastReadTime = time.Now()
	fd := c.GetFd()
	c.buffer.AsyncReadFromFD(fd, c.server.iouring, c.getReadCallback())
}

func (c *Conn) getReadCallback() EventCallBack {
	return func(e *eventInfo) (err error) {
		if c.server.options.decoder == nil {
			c.server.handler.OnMessage(c, c.buffer.ReadAll())
			return
		}

		err = c.server.options.decoder.Decode(c)
		if err != nil {
			return
		}

		return
	}
}

// processReadEvent
// process connect read complete event
// add async block read bytes event until read readBufferLen bytes from connect fd
func (c *Conn) processReadEvent(e *eventInfo) (err error) {
	for {
		// if un use poll in ready, need add read event op again
		c.AsyncBlockRead()

		err = e.cb(e)
		if err != nil {
			// There is no data to read in the buffer
			if err == syscall.EAGAIN {
				return nil
			}
			return err
		}
	}
}

// AsyncBlockWrite
// async block write bytes
func (c *Conn) AsyncBlockWrite(bytes []byte) {
	c.server.iouring.addSendSqe(func(info *eventInfo) error { return nil }, c.fd, bytes, len(bytes), 0)
	return
}
func (c *Conn) processWirteEvent(e *eventInfo) (err error) {
	if e.cqe.Res < 0 {
		err = ErrIOUringWriteFail
		return
	}

	err = e.cb(e)
	if err != nil {
		// There is no data to write to the buffer
		if err == syscall.EAGAIN {
			return nil
		}
		return err
	}

	return
}

// Write
// Writer impl, block write
func (c *Conn) Write(bytes []byte) (int, error) {
	return syscall.Write(c.fd, bytes)
}

// WriteWithEncoder
// write with encoder, encode bytes to writer
func (c *Conn) WriteWithEncoder(bytes []byte) error {
	return c.server.options.encoder.EncodeToWriter(c, bytes)
}

// Close Closes the connection
func (c *Conn) Close() {
	// Remove from the file descriptor that epoll is listening for
	err := closeFD(c.pollerFD, c.fd)
	if err != nil {
		log.Error(err)
	}

	// Remove conn from conns
	c.server.conns.Delete(c.fd)
	// Return the cache
	c.server.readBufferPool.Put(c.buffer.buf)
	// Subtract one from the number of connections
	atomic.AddInt64(&c.server.connsNum, -1)
}

// CloseRead closes connection
func (c *Conn) CloseRead() error {
	err := closeFDRead(int(c.fd))
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// GetData gets the data
func (c *Conn) GetData() interface{} {
	return c.data
}

// SetData sets the data
func (c *Conn) SetData(data interface{}) {
	c.data = data
}
