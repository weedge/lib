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
	fd           int32       // File descriptor
	addr         string      // peer address
	buffer       *Buffer     // Read the buffer
	lastReadTime time.Time   // Time of last read
	data         interface{} // Business custom data, used as an extension
}

// newConn create tcp connection
func newConn(fd int32, addr string, server *Server) *Conn {
	return &Conn{
		server:       server,
		fd:           fd,
		addr:         addr,
		buffer:       NewBuffer(server.readBufferPool.Get().([]byte)),
		lastReadTime: time.Now(),
	}
}

// GetFd gets the file descriptor
func (c *Conn) GetFd() int32 {
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

// Read reads the data
func (c *Conn) Read() error {
	c.lastReadTime = time.Now()
	fd := int(c.GetFd())
	for {
		err := c.buffer.ReadFromFD(fd)
		if err != nil {
			// There is no data to read in the buffer
			if err == syscall.EAGAIN {
				return nil
			}
			return err
		}

		err = c.server.decoder.Decode(c)
		if err != nil {
			return err
		}
	}
}

// Write writes the data
func (c *Conn) Write(bytes []byte) (int, error) {
	return syscall.Write(int(c.fd), bytes)
}

// Close Closes the connection
func (c *Conn) Close() error {
	// Remove from the file descriptor that epoll is listening for
	err := closeFD(int(c.fd))
	if err != nil {
		log.Error(err)
		return err
	}

	// Remove conn from conns
	c.server.conns.Delete(c.fd)
	// Return the cache
	c.server.readBufferPool.Put(c.buffer.Buf)
	// Subtract one from the number of connections
	atomic.AddInt64(&c.server.connsNum, -1)
	return nil
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
