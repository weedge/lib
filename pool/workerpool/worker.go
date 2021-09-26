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
	myWorkerPool          *WorkerPool //所属任务工作池
	hasGoroutineRunning   int32       //是否有对应的goroutine在运行
	chTaskIsTimeOut       chan<- bool //任务是否超时，true 超时,false 正常　由添加任务时创建
	chExecuteGoroutineOut chan int    //执行任务的goroutine结束 -> worker goroutine 结束
}

func newWorker(wp *WorkerPool) *Worker {
	return &Worker{
		myWorkerPool:          wp,
		chExecuteGoroutineOut: make(chan int, 1),
	}
}

func (worker *Worker) safelyDo() {
	if nil == worker.myWorkerPool {
		log.Error("myWorkerPool ptr is nil ")
		return
	}

	debug.SetPanicOnFault(true)
	defer worker.myWorkerPool.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			if worker.myWorkerPool.isWorking() {
				worker.myWorkerPool.chAddWorker <- 1
			}

			atomic.AddInt32(&(worker.myWorkerPool.curWorkerNum), -1)
			atomic.SwapInt32(&(worker.hasGoroutineRunning), 0)
			log.Error("check timeout worker goroutine is crash, panic", r, strings.BytesToString(debug.Stack()))
		}
	}()

	worker.execute()
}

func (worker *Worker) execute() {
	for {
		select {
		case task, ok := <-worker.myWorkerPool.chWorkTask:
			if ok {
				taskTimer := time.NewTimer(task.TimeOut)
				taskDoResCh := make(chan bool, 1)
				go func() {
					taskDoResCh <- task.Do(task.InParam, task.OutParam)
				}()

				select {
				case <-taskTimer.C:
					if task.ChIsTimeOut != nil {
						task.ChIsTimeOut <- true
					}
					if task.OnTimeOut != nil {
						task.OnTimeOut(task.InParam, task.OutParam)
					}
				case _, ok := <-taskDoResCh:
					if ok && task.ChIsTimeOut != nil {
						task.ChIsTimeOut <- false
					}
				}

				worker.destroyAfterTaskDone()
			} else {
				worker.myWorkerPool.close()
				log.Info("close chWorkTask, myWorkerPool close and exit worker goroutine, the current goroutine number:", atomic.LoadInt32(&worker.myWorkerPool.curWorkerNum))
				runtime.Goexit()
			}
		case _, ok := <-worker.chExecuteGoroutineOut:
			if worker.myWorkerPool.isWorking() && atomic.LoadInt32(&worker.hasGoroutineRunning) <= 0 && ok {
				log.Info("send to add worker when ExecuteGoroutineOut")
				worker.myWorkerPool.chAddWorker <- 1
			}
			atomic.AddInt32(&(worker.myWorkerPool.curWorkerNum), -1)
			log.Info(" exit this execute worker goroutine, the current goroutine number:", atomic.LoadInt32(&worker.myWorkerPool.curWorkerNum))
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
