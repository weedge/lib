package workerpool

import (
	"errors"
	"time"
)

const (
	WorkerPool_Stat_Uninitialized int32 = iota
	WorkerPool_Stat_Start
	WorkerPool_Stat_Running
	WorkerPool_Stat_Stoping
	WorkerPool_Stat_Stop
)

const (
	DefaultTimeOut             = 30 * time.Second
	workerGoroutineLifeTime    = 30
	addWaitingTimeWhenStopPool = 30
)

// 一些错误码
var ErrTimeout = errors.New("received timeout")
var ErrInterrupt = errors.New("receive interrupt")
