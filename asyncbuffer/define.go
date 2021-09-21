package asyncbuffer

import (
	"sync"
)

const (
	DefaultBufferWindowSize int = 1
	DefaultDelaySendTimeMs  int = 10
)

type IBuffer interface {
	BatchDo([][]byte)
	FormatInput() (error, []byte)
}

type DemoBuffer struct {
	Name string
}

type InputBufferItem struct {
	ChName string
	Data   IBuffer
}

type SendChannel struct {
	ChName       string
	ChLen        int
	SubWorkerNum int
}

type Conf struct {
	BufferName        string
	BufferSendChannel map[string]*SendChannel
	BufferWindowSize  int
	DelaySendTime     int
}

type SendData struct {
	BufferName       string
	MapDataCh        map[string]chan []byte
	BufferData       [][]byte
	BufferIndex      int64
	BufferWindowSize int
	ISendObj         IBuffer
	IsFlushCh        chan bool
	DelaySendTime    int
	OpLock           sync.Mutex
	BufferDayCounter uint64
}
