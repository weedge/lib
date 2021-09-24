package workerpool

import (
	"github.com/weedge/lib/strings"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/weedge/lib/log"
)

type Worker struct {
	timer                 *time.Timer
	myWorkerPool          *WorkerPool //所属任务工作池
	hasGoroutineRunning   int32       //是否有对应的goroutine在运行
	chTaskDoRes           chan bool   //任务执行结果 -> check
	chTaskIsTimeOut       chan<- bool //任务是否超时，true 超时,false 正常　由添加任务时创建
	chExecuteGoroutineOut chan int    //执行任务的goroutine结束 -> watch check timeout goroutine 结束
	chWatchGoroutineOut   chan int    //watch check timeout goroutine 结束 ->执行任务的goroutine结束
}

func newWorker(wp *WorkerPool) *Worker {
	return &Worker{
		timer:                 time.NewTimer(DefaultTimeOut),
		myWorkerPool:          wp,
		chTaskDoRes:           make(chan bool, 1),
		chExecuteGoroutineOut: make(chan int, 1),
		chWatchGoroutineOut:   make(chan int, 1),
	}
}

func (worker *Worker) execute() {
	if nil == worker.myWorkerPool {
		log.Error(" myWorkerPool is nil ")
		return
	}

	debug.SetPanicOnFault(true)
	defer worker.myWorkerPool.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			if worker.myWorkerPool.isWorking() {
				worker.chExecuteGoroutineOut <- 1
			}
			atomic.SwapInt32(&worker.hasGoroutineRunning, 0)
			log.Error("worker execute goroutine crash, panic", r, strings.BytesToString(debug.Stack()))
		}
	}()

	for {
		select {
		case task, ok := <-worker.myWorkerPool.chWorkTask:
			if ok {
				worker.chTaskIsTimeOut = task.ChIsTimeOut
				worker.timer.Reset(task.TimeOut)
				worker.chTaskDoRes <- task.Do(task.InParam, task.OutParam)
				worker.timer.Stop()

				worker.destroyAfterTaskDone()
			} else {
				worker.myWorkerPool.close()
				log.Info("no task, myWorkerPool close and exit worker goroutine, the current goroutine number:", atomic.LoadInt32(&worker.myWorkerPool.curWorkerNum))
				runtime.Goexit()
			}
		case <-worker.chWatchGoroutineOut:
			runtime.Goexit()
		}
	}
}

// worker destroy after task done cond:
// 1. worker pool is working
// 2. work task ch is empty
// 3. cur worker num of pool > min worker num
// 4. (now - add worker last time) > worker life time
func (worker *Worker) destroyAfterTaskDone() {
	chWorkTaskLen := len(worker.myWorkerPool.chWorkTask)
	if worker.myWorkerPool.isWorking() && chWorkTaskLen <= 0 &&
		atomic.LoadInt32(&worker.myWorkerPool.curWorkerNum) > worker.myWorkerPool.minWorkerNum &&
		time.Now().Unix()-worker.myWorkerPool.addWorkerLastTime >= workerGoroutineLifeTime {
		worker.chExecuteGoroutineOut <- 1
		log.Info("exit more worker goroutine, the current goroutine number:", atomic.LoadInt32(&worker.myWorkerPool.curWorkerNum))
		runtime.Goexit()
	}
}

func (worker *Worker) executeAndWatch() {
	if nil == worker.myWorkerPool {
		log.Error("myWorkerPool ptr is nil ")
		return
	}

	debug.SetPanicOnFault(true)
	defer worker.myWorkerPool.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			if worker.myWorkerPool.isWorking() {
				worker.chWatchGoroutineOut <- 1
				worker.myWorkerPool.chAddWorker <- 1
			}

			atomic.AddInt32(&(worker.myWorkerPool.curWorkerNum), -1)
			atomic.SwapInt32(&(worker.hasGoroutineRunning), 0)
			log.Error("check timeout worker goroutine is crash, panic", r, strings.BytesToString(debug.Stack()))
		}
	}()

	worker.myWorkerPool.wg.Add(1)
	go worker.execute()

	worker.watch()
}

// watch:
// 1. from worker do task res ch; is ok to task is timeout ch false, or no ret exit this execute worker goroutine
// 2. from worker timeout ch; to task is timeout ch true
// 3. from worker execute goroutine out; is ok && workerPool is working && worker goroutine is running to add new worker by pool, and exit this execute worker goroutine
func (worker *Worker) watch() {
	for {
		select {
		case _, ok := <-worker.chTaskDoRes:
			if ok {
				worker.chTaskIsTimeOut <- false
			} else {
				atomic.AddInt32(&(worker.myWorkerPool.curWorkerNum), -1)
				log.Info("no ret exit this execute watch worker goroutine, the current goroutine number:", atomic.LoadInt32(&worker.myWorkerPool.curWorkerNum))
				runtime.Goexit()
			}
		case <-worker.timer.C:
			worker.chTaskIsTimeOut <- true
		case _, ok := <-worker.chExecuteGoroutineOut:
			if worker.myWorkerPool.isWorking() && atomic.LoadInt32(&worker.hasGoroutineRunning) <= 0 && ok {
				worker.myWorkerPool.chAddWorker <- 1
			}
			atomic.AddInt32(&(worker.myWorkerPool.curWorkerNum), -1)
			log.Info(" exit this execute watch  worker goroutine, the current goroutine number:", atomic.LoadInt32(&worker.myWorkerPool.curWorkerNum))
			runtime.Goexit()
		}
	}
}
