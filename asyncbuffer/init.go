package asyncbuffer

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/weedge/lib/log"
)

var gBufferSendDataInstances map[string]*SendData

func init() {
	gBufferSendDataInstances = map[string]*SendData{}
}

// init instances by conf
func InitInstancesByConf() {
	asyncBuffer := viper.GetStringMap(`async_buffer`)
	for bufferName := range asyncBuffer {
		bufferWindowSize := viper.GetInt(fmt.Sprintf(`async_buffer.%s.buffer_win_size`, bufferName))
		delaySendTime := viper.GetInt(fmt.Sprintf(`async_buffer.%s.delay_do_ms`, bufferName))
		chs := viper.GetStringMap(fmt.Sprintf(`async_buffer.%s.chs`, bufferName))
		for chName := range chs {
			chLen := viper.GetInt(fmt.Sprintf(`async_buffer.%s.chs.%s.ch_len`, bufferName, chName))
			subWorkNum := viper.GetInt(fmt.Sprintf(`async_buffer.%s.chs.%s.sub_worker_num`, bufferName, chName))
			InitInstance(&Conf{
				BufferName: bufferName,
				BufferSendChannel: map[string]*SendChannel{
					chName: {
						ChName:       chName,
						ChLen:        chLen,
						SubWorkerNum: subWorkNum,
					}},
				BufferWindowSize: bufferWindowSize,
				DelaySendTime:    delaySendTime,
			})
		} //end for
	} //end for
}

// get instance
func GetInstance(bufferName string) (err error, sd *SendData) {
	if _, ok := gBufferSendDataInstances[bufferName]; !ok {
		err = fmt.Errorf("bufferName: %s un init instance", bufferName)
		return
	}
	sd = gBufferSendDataInstances[bufferName]

	return
}

// init instance
func InitInstance(conf *Conf) {
	if sd, ok := gBufferSendDataInstances[conf.BufferName]; !ok {
		if conf.BufferWindowSize <= 0 {
			conf.BufferWindowSize = DefaultBufferWindowSize
		}

		if conf.DelaySendTime <= 0 {
			conf.DelaySendTime = DefaultDelaySendTimeMs
		}

		mapDataCh := map[string]chan []byte{}
		for chName, sendCh := range conf.BufferSendChannel {
			if sendCh.ChLen <= 0 {
				sendCh.ChLen = 0
			}
			mapDataCh[chName] = make(chan []byte, sendCh.ChLen)
		}

		sd = &SendData{
			BufferName:       conf.BufferName,
			MapDataCh:        mapDataCh,
			BufferData:       make([][]byte, conf.BufferWindowSize),
			BufferIndex:      0,
			BufferWindowSize: conf.BufferWindowSize,
			IsFlushCh:        make(chan bool),
			DelaySendTime:    conf.DelaySendTime,
			BufferDayCounter: 0,
		}
		for chName, sendCh := range conf.BufferSendChannel {
			if sendCh.SubWorkerNum <= 0 {
				sendCh.SubWorkerNum = 1
			}
			sd.InitAsyncSub(chName, sendCh.SubWorkerNum)
		}

		sd.initFlushTicker(conf.BufferName)
		gBufferSendDataInstances[conf.BufferName] = sd
		log.Infof("GetInstance: %s new instance: %v", conf.BufferName, sd)
		return
	}
	log.Debugf("GetInstance: %s instance: %v", conf.BufferName, gBufferSendDataInstances[conf.BufferName])
	return
}

func FlushAll() {
	for bufferName, instance := range gBufferSendDataInstances {
		log.Infof("flush bufferName:%s nmq send buffer~!", bufferName)
		instance.FlushBuffer()
	}
}

func FlushOne(bufferName string) {
	log.Infof("flushOne bufferName:%s nmq send buffer~!", bufferName)
	if _, ok := gBufferSendDataInstances[bufferName]; !ok {
		log.Errorf("flush bufferName: %s error~!", bufferName)
		return
	}
	gBufferSendDataInstances[bufferName].FlushBuffer()
}
