package poller

import (
	"io"
	"syscall"
)

// Buffer Read buffer, one read buffer for each tcp long connection
type Buffer struct {
	// todo: use bytes.Buffer ->  buffer pool
	// buff   bytes.Buffer
	buf   []byte // In-application cache
	start int    // The start position of a valid byte
	end   int    // End position of a valid byte
}

// NewBuffer Creates a buffer
func NewBuffer(bytes []byte) *Buffer {
	return &Buffer{buf: bytes, start: 0, end: 0}
}

func (b *Buffer) Len() int {
	return b.end - b.start
}

// reset Reset cache (moves useful bytes forward)
func (b *Buffer) reset() {
	if b.start == 0 {
		return
	}
	copy(b.buf, b.buf[b.start:b.end])
	b.end -= b.start
	b.start = 0
}

// ReadFromFD reads data from the file descriptor
func (b *Buffer) ReadFromFD(fd int) error {
	b.reset()
	n, err := syscall.Read(fd, b.buf[b.end:])
	if err != nil {
		return err
	}
	if n == 0 {
		return syscall.EAGAIN
	}
	b.end += n
	return nil
}

// AsyncReadFromFD
// async block read event from fd to buff
func (b *Buffer) AsyncReadFromFD(fd int, uring *ioUring, cb EventCallBack) error {
	b.reset()
	uring.addRecvSqe(func(info *eventInfo) error {
		n := info.cqe.Res
		if n < 0 {
			return ErrIOUringReadFail
		}
		if n == 0 {
			return syscall.EAGAIN
		}
		b.end += int(n)
		return cb(info)
	}, len(b.buf[b.end:]), b.buf[b.end:], 0)

	return nil
}

// ReadFromReader reads data from the reader. If the reader blocks, it will block
func (b *Buffer) ReadFromReader(reader io.Reader) (int, error) {
	b.reset()
	n, err := reader.Read(b.buf[b.end:])
	if err != nil {
		return n, err
	}
	b.end += n
	return n, nil
}

// Seek returns n bytes without a shift, or an error if there are not enough bytes
func (b *Buffer) Seek(len int) ([]byte, error) {
	if b.end-b.start <= len {
		buf := b.buf[b.start : b.start+len]
		return buf, nil
	}
	return nil, ErrBufferNotEnough
}

// Read Discard offset fields and read n fields. If there are not enough bytes, an error is returned
func (b *Buffer) Read(offset, limit int) ([]byte, error) {
	if b.Len() < offset+limit {
		return nil, ErrBufferNotEnough
	}
	b.start += offset
	buf := b.buf[b.start : b.start+limit]
	b.start += limit
	return buf, nil
}

// ReadAll Reads all bytes
func (b *Buffer) ReadAll() []byte {
	buf, _ := b.Read(b.start, b.end)
	return buf
}
