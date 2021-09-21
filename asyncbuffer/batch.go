package asyncbuffer

//@todo: add send obj sequence id

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"unsafe"

	"time"

	"github.com/weedge/lib/log"
)

func (sd *SendData) AddBufferItem(item *InputBufferItem) (err error) {
	ch, ok := sd.MapDataCh[item.ChName]
	if ok == false {
		err = fmt.Errorf("bufferName: %s chName: %s don't exist", sd.BufferName, item.ChName)
		return
	}
	if item == nil {
		err = fmt.Errorf("bufferName: %s AddBufferItem InputBufferItem is nil", sd.BufferName)
		return
	}
	sd.ISendObj = item.Data
	err, res := sd.ISendObj.FormatInput()
	if err != nil {
		return
	}
	ch <- res

	return
}

func (sd *SendData) InitAsyncSub(chName string, workerNum int) {
	for i := 1; i <= workerNum; i++ {
		sd.asyncSubFromOneCh(chName)
	}
}

func (sd *SendData) initFlushTicker(buffName string) {
	go func() {
		for {
			now := time.Now()
			beginTs := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
			endTs := beginTs + 86400 - 1
			//0:0:10
			afterTs := endTs - now.Unix() + 10
			afterTs = afterTs % 86400
			if afterTs <= 0 {
				afterTs = 86400
			}

			select {
			case <-time.After(3 * time.Second):
				//println(buffName, "flushTicker")
				log.Debugf("%s flushTicker", buffName)
				sd.BufferSend([]byte{})
			case <-time.After(time.Duration(afterTs) * time.Second):
				//println(buffName, "counterTicker dayBufferCounter", sd.BufferDayCounter, "clear")
				log.Infof("%s dayBufferCounter:%d clear~!", buffName, sd.BufferDayCounter)
				sd.counterClear()
			}
		}
	}()
}

func (sd *SendData) asyncSubFromOneCh(chName string) {
	go func(name string) {
		defer func() {
			if err := recover(); err != nil {
				panicBuffer := ""
				if sd != nil && sd.BufferData != nil {
					for _, item := range sd.BufferData {
						panicBuffer += *(*string)(unsafe.Pointer(&item))
					}
				}
				if sd != nil {
					sd.BufferIndex = 0
				}

				log.Errorf("bufferName: %s AsyncSubFromOneCh chName[%s] panicBuffer[%s] panic recovered err[%s] stack[%s]", sd.BufferName, chName, panicBuffer, err, string(debug.Stack()))
				//println("panicBuffer-->:", panicBuffer)
				//println("stack:", string(debug.Stack()))
				time.Sleep(200 * time.Millisecond)
				sd.asyncSubFromOneCh(name)
			}
		}()
		sd.subFromOneCh(name)
	}(chName)
}

func (sd *SendData) counter() {
	atomic.AddUint64(&sd.BufferDayCounter, 1)
}
func (sd *SendData) counterClear() {
	atomic.StoreUint64(&sd.BufferDayCounter, 0)
}

func (sd *SendData) subFromOneCh(chName string) {
	log.Infof("bufferName: %s subFromOneCh select ch chName:%s bufferWindowSize: %d delaySendTimeMS: %d", sd.BufferName, chName, sd.BufferWindowSize, sd.DelaySendTime)
	ch, ok := sd.MapDataCh[chName]
	if ok == false {
		return
	}
	for {
		select {
		case data, ok := <-ch:
			if ok == false { //close
				log.Infof("chName: %s close continue", chName)
				time.Sleep(3 * time.Second)
				continue
			}

			//println("getSendDataFrom data:", *(*string)(unsafe.Pointer(&data)))
			log.Infof("getSendDataFrom ch:%s data: %s", chName, *(*string)(unsafe.Pointer(&data)))
			sd.counter()
			sd.BufferSend(data)
		case isFlush, ok := <-sd.IsFlushCh:
			if ok == false { //close
				log.Infof("IsFlushCh close continue")
				time.Sleep(3 * time.Second)
				continue
			}
			log.Debugf("bufferName: %s IsFlushCh send isFlush: %v start flush~!", sd.BufferName, isFlush)
			if isFlush {
				sd.BufferSend([]byte{})
			}
		}
	}
}

// one sendData -> multi pub and batch sub
func (sd *SendData) batchSubFromCh() {
	log.Debugf("batchSubFromCh")
	for chName := range sd.MapDataCh {
		sd.subFromOneCh(chName)
	}
}

func (sd *SendData) BufferSend(data []byte) {
	sd.OpLock.Lock()
	defer sd.OpLock.Unlock()

	if len(data) > 0 {
		sd.BufferData[atomic.LoadInt64(&sd.BufferIndex)] = data
		atomic.AddInt64(&sd.BufferIndex, 1)
	}

	if atomic.LoadInt64(&sd.BufferIndex) == int64(sd.BufferWindowSize) || len(data) == 0 {
		sd.flush()
	}

	return
}

//notice: if close ch panic
func (sd *SendData) FlushBuffer() {
	sd.IsFlushCh <- true
}

func (sd *SendData) flush() {
	if atomic.LoadInt64(&sd.BufferIndex) <= 0 {
		//log.Infof("bufferName: %s un flush~!", sd.BufferName)
		return
	}
	if sd.ISendObj == nil {
		log.Errorf("bufferName: %s sd.ISendObj is nil, flush fail~!", sd.BufferName)
		return
	}

	time.Sleep(time.Duration(sd.DelaySendTime) * time.Millisecond)
	sd.ISendObj.BatchDo(sd.BufferData[0:atomic.LoadInt64(&sd.BufferIndex)])
	atomic.StoreInt64(&sd.BufferIndex, 0)
}
