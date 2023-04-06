package poller

import (
	"encoding/binary"
	"io"
	"sync"
)

type headerLenDecoder struct {
	// headerlen
	// Length of the TCP packet header, which is used to describe the byte length of the packet
	headerLen int
}

// NewHeaderLenDecoder Creates a decoder based on header length
// headerLen Indicates the header of a TCP packet, which describes the byte length of the packet
// readMaxLen Specifies the maximum length of the client packet read. The length of the packet sent by the client cannot exceed this length
func NewHeaderLenDecoder(headerLen int) Decoder {
	if headerLen <= 0 {
		panic("headerLen or readMaxLen must must greater than 0")
	}

	return &headerLenDecoder{
		headerLen: headerLen,
	}
}

// Decode
func (d *headerLenDecoder) DecodeBuffer(buffer *Buffer) (value []byte, err error) {
	value = []byte{}
	header, err := buffer.Seek(d.headerLen)
	if err != nil {
		return
	}

	valueLen := int(binary.BigEndian.Uint16(header))
	value, err = buffer.Read(d.headerLen, valueLen)
	if err != nil {
		return
	}

	return
}

func (d *headerLenDecoder) Decode(c *Conn) error {
	for {
		value, err := d.DecodeBuffer(c.buffer)
		if err == ErrBufferNotEnough {
			return nil
		}
		if err != nil {
			return err
		}

		c.server.handler.OnMessage(c, value)
	}
}

type headerLenEncoder struct {
	// headerLen
	// Length of the TCP packet header, which is used to describe the byte length of the packet
	headerLen int
	// writeBufferLen
	// The recommended length of packets sent by the server to the client. Writebufferlen int takes advantage of pool optimization when packets are sent less than this value
	writeBufferLen  int
	writeBufferPool *sync.Pool
}

// NewHeaderLenEncoder
// Creates an encoder based on header length
// headerLen Indicates the header of a TCP packet, which describes the byte length of the packet
// writeBufferLen Indicates the recommended length of a packet sent by the server to the client. When a packet is sent less than this value, it takes advantage of the pool optimization
func NewHeaderLenEncoder(headerLen, writeBufferLen int) *headerLenEncoder {
	if headerLen <= 0 || writeBufferLen <= 0 {
		panic("headerLen or writeBufferLen must must greater than 0")
	}

	return &headerLenEncoder{
		headerLen:      headerLen,
		writeBufferLen: writeBufferLen,
		writeBufferPool: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, writeBufferLen)
				return b
			},
		},
	}
}

// EncodeToWriter
// Encodes data and writes it to Writer
func (e headerLenEncoder) EncodeToWriter(w io.Writer, bytes []byte) error {
	l := len(bytes)
	var buffer []byte
	if l <= e.writeBufferLen-e.headerLen {
		obj := e.writeBufferPool.Get()
		defer e.writeBufferPool.Put(obj)
		buffer = obj.([]byte)[0 : l+e.headerLen]
	} else {
		buffer = make([]byte, l+e.headerLen)
	}

	// Write the message length to buffer
	binary.BigEndian.PutUint16(buffer[0:e.headerLen], uint16(l))
	// Writes the message content to the buffer
	copy(buffer[e.headerLen:], bytes)

	_, err := w.Write(buffer)
	return err
}
