package asyncbuffer

import (
	"bytes"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	_ "syscall"
	"testing"
	"time"
	"unsafe"
)

var itemMap map[string]int = map[string]int{}

//var gItemMap sync.Map
var gCounter uint64

func (m *DemoBuffer) BatchDo(data [][]byte) {
	//println("bufferLen:", len(data))
	bufferData := ""
	for _, item := range data {
		//println("getSendDataFrom item:", *(*string)(unsafe.Pointer(&item)))
		//time.Sleep(300 * time.Millisecond)

		//_, ok := gItemMap.Load(*(*string)(unsafe.Pointer(&item)))
		_, ok := itemMap[*(*string)(unsafe.Pointer(&item))]
		if !ok {
			atomic.AddUint64(&gCounter, 1)
			//gItemMap.Store(*(*string)(unsafe.Pointer(&item)), 1)
			itemMap[*(*string)(unsafe.Pointer(&item))] = 1
			//println("lenItemMap:", len(itemMap))
		} else {
			println("repeat item:", *(*string)(unsafe.Pointer(&item)))
		}

		bufferData += *(*string)(unsafe.Pointer(&item)) + "\t"
		if *(*string)(unsafe.Pointer(&item)) == "10" {
			//panic("panic:" + string(item))
		}
	}

	//println("bufferData:", bufferData)
}

func (m *DemoBuffer) FormatInput() (err error, bytes []byte) {
	return nil, []byte(m.Name)
}

func TestMain(m *testing.M) {
	m.Run()
}

func TestFlushOne(t *testing.T) {
	InitInstancesByConf()
	time.Sleep(3 * time.Second)

	type InputCase struct {
		BufferName string
	}
	inputCases := []InputCase{{BufferName: "default"}, {BufferName: "default"}, {BufferName: "nmq_live_user"}}

	for _, inputCase := range inputCases {
		FlushOne(inputCase.BufferName)
		time.Sleep(3 * time.Second)
	}
}

func TestSendOneCh(t *testing.T) {
	InitInstancesByConf()
	time.Sleep(3 * time.Second)

	type InputCase struct {
		BufferName string
		ChName     string
		Data       IBuffer
	}
	inputCases := []InputCase{
		{
			BufferName: "default",
			ChName:     "default",
			Data:       &DemoBuffer{Name: "testDefault1"},
		},
		{
			BufferName: "default",
			ChName:     "default",
			Data:       &DemoBuffer{Name: "testDefault2"},
		},
		{
			BufferName: "nmq_live_user",
			ChName:     "nmq",
			Data:       &DemoBuffer{Name: "testUser1"},
		},
		{
			BufferName: "nmq_live_user",
			ChName:     "nmq",
			Data:       &DemoBuffer{Name: "testUser2"},
		},
	}

	for _, inputCase := range inputCases {
		err := SendOneCh(inputCase.BufferName, inputCase.ChName, inputCase.Data)
		if err != nil {
			t.Errorf("SendOneCh bufferName: %s chName: %s data: %v err: %s", inputCase.BufferName, inputCase.ChName, inputCase.Data, err.Error())
		}
	}

	FlushAll()

	ch := make(chan int)
	ch <- 1

}

func TestMultiWorkerSendOneCh(t *testing.T) {
	InitInstancesByConf()
	//time.Sleep(3 * time.Second)

	var start time.Time
	var end time.Time

	type InputCase struct {
		BufferName string
		ChName     string
		Data       IBuffer
	}
	start = time.Now()
	//exitCh := make(chan int)

run:
	// N+1 goroutine  send D (N+1 * D)
	var wg sync.WaitGroup
	N := 9
	D := 303
	for i := 1; i <= N*D+1; i += D {
		wg.Add(1)
		go func(uid int) {
			opCn := 0
			for {
				if opCn >= 3 {
					end = time.Now()
					// 执行时间 单位:微秒
					_ = end.Sub(start).Nanoseconds() / 1e3
					//println("gid:", getGID())
					//println("pid", syscall.Getpid())
					break
				}
				opCn += 1
				testCases := []InputCase{}
				for i := 1; i <= 101; i++ {
					strUid := strconv.Itoa(uid)
					testCase := InputCase{
						BufferName: "default",
						ChName:     "default",
						Data:       &DemoBuffer{Name: "" + strUid},
					}
					testCases = append(testCases, testCase)
					uid++
				}
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				randVal := r.Intn(10)
				//println("sleep", randVal, "flush")
				time.Sleep(time.Duration(randVal) * time.Second)
				//FlushOne("default")

				for _, testCase := range testCases {
					err := SendOneCh(testCase.BufferName, testCase.ChName, testCase.Data)
					if err != nil {
						t.Errorf("SendOneCh bufferName: %s chName: %s data: %v err: %s", testCase.BufferName, testCase.ChName, testCase.Data, err.Error())
					}
				}
				//time.Sleep(200 * time.Millisecond)
			}
			//println("flushOne")
			FlushOne("default")
			time.Sleep(200 * time.Millisecond)
			wg.Done()
		}(i)
	}
	wg.Wait()

	//println("flushAll")
	//FlushAll()
	//time.Sleep(200 * time.Millisecond)

	cn := 0
	/*
		for i := 1; i <= (N+1) * D; i++ {
			if _, ok := gItemMap.Load(strconv.Itoa(i)); !ok {
				println("un send---->", i)
			} else {
				cn++
			}
		}
	*/

	for i := 1; i <= (N+1)*D; i++ {
		if _, ok := itemMap[strconv.Itoa(i)]; !ok {
			println("un send---->", i)
		}
	}
	cn = len(itemMap)

	println("cn:", cn)

	println("counter:", gCounter)
	println("buffCounter:", gBufferSendDataInstances["default"].BufferDayCounter)
	println("bufferIndex:", gBufferSendDataInstances["default"].BufferIndex)
	for j := int64(0); j < gBufferSendDataInstances["default"].BufferIndex; j++ {
		println("unSendBufferData:", string(gBufferSendDataInstances["default"].BufferData[j]))
	}

	time.Sleep(5000 * time.Millisecond)
	println("bufferIndex:", gBufferSendDataInstances["default"].BufferIndex)

	atomic.StoreUint64(&gCounter, 0)
	itemMap = map[string]int{}
	goto run

	ch := make(chan int)
	ch <- 1
	//time.Sleep(10000 * time.Millisecond)

}

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
