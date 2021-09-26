package workerpool

import (
	"github.com/weedge/lib/strings"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/weedge/lib/log"
)

type WorkerPool struct {
	minWorkerNum      int32      //worker goroutine 最小数目
	maxWorkerNum      int32      //worker goroutine 最大数目
	curWorkerNum      int32      //当前worker goroutine数量
	stat              int32      //任务工作池状态
	addTaskStat       int32      //添加任务状态
	chAddWorker       chan int32 //添加worker ch -> watchAddWorker
	addWorkerLastTime int64      //最新添加worker时间
	wg                *sync.WaitGroup
	chWorkTask        chan Task
	workers           []*Worker
	indexWorkers      int // watch workers 循环下标
	lock              *sync.Mutex
}

func NewWorkerPool(minWorkerNum, maxWorkerNum, taskChSize int) *WorkerPool {
	if minWorkerNum <= 0 || minWorkerNum > maxWorkerNum {
		log.Error("worker pool init error , the min number: ", minWorkerNum, " the max number: ", maxWorkerNum)
		return nil
	}

	if taskChSize <= 0 {
		log.Error("worker pool init error , the nChan : ", taskChSize)
		return nil
	}

	wp := &WorkerPool{
		minWorkerNum: int32(minWorkerNum),
		maxWorkerNum: int32(maxWorkerNum),
		wg:           &sync.WaitGroup{},
		lock:         &sync.Mutex{},
	}

	wp.init(taskChSize)

	return wp
}

func (wp *WorkerPool) init(taskChSize int) {
	if atomic.LoadInt32(&(wp.stat)) > 0 {
		log.Warn("worker pool is already initialized ! the stat is ", wp.stat)
		return
	}

	atomic.SwapInt32(&(wp.stat), WorkerPool_Stat_Start)

	wp.chWorkTask = make(chan Task, taskChSize)
	wp.chAddWorker = make(chan int32, 1)
	wp.workers = make([]*Worker, 0, wp.maxWorkerNum)
	for i := int32(0); i < wp.maxWorkerNum; i++ {
		wp.workers = append(wp.workers, newWorker(wp))
	}

	log.Info(" init worker pool ok , the min number:  ", wp.minWorkerNum, " the max number : ", wp.maxWorkerNum)

	return
}

func (wp *WorkerPool) Run() {
	for i := int32(0); i < wp.minWorkerNum; i++ {
		if nil == wp.workers[i] {
			log.Error(" workers is  nil ,the index is", i)
			return
		}

		atomic.AddInt32(&(wp.curWorkerNum), 1)
		atomic.SwapInt32(&(wp.workers[i].hasGoroutineRunning), 1)
		wp.wg.Add(1)
		go wp.workers[i].safelyDo()
	}

	wp.indexWorkers = int(wp.minWorkerNum)

	wp.wg.Add(1)
	go wp.watchAddWorker()
	atomic.SwapInt32(&(wp.stat), WorkerPool_Stat_Running)

	log.Info(" run worker pool ok , run the min number:  ", wp.minWorkerNum, " timeout workers and one watch add worker")
}

func (wp *WorkerPool) watchAddWorker() {
	debug.SetPanicOnFault(true)
	defer wp.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Error("watch add worker goroutine crash, panic", r, strings.BytesToString(debug.Stack()))
		}
	}()

	for ch := range wp.chAddWorker {
		for nTimes := int32(0); wp.isWorking() && nTimes < wp.maxWorkerNum; {
			nTimes++
			if nil == wp.workers[wp.indexWorkers] {
				log.Warn(" the workers is nil the index is ", wp.indexWorkers)
				continue
			}

			if atomic.LoadInt32(&wp.workers[wp.indexWorkers].hasGoroutineRunning) <= 0 {
				wp.addWorker()
				log.Info("create a new worker, current worker number is : ", atomic.LoadInt32(&(wp.curWorkerNum)), " the flag ", ch)
				break
			}
		}
	}
}

func (wp *WorkerPool) addWorker() {
	atomic.SwapInt32(&wp.workers[wp.indexWorkers].hasGoroutineRunning, 1)
	wp.wg.Add(1)
	go wp.workers[wp.indexWorkers].safelyDo()
	atomic.AddInt32(&(wp.curWorkerNum), 1)
	wp.indexWorkers++
	if int32(wp.indexWorkers) >= wp.maxWorkerNum {
		wp.indexWorkers = 0
	}

	return
}

func (wp *WorkerPool) AddTask(task *Task) {
	if nil == task {
		log.Warn("AddTask task is nil")
		return
	}

	if nil == task.Do {
		log.Warn("add task Do func is nil")
		return
	}
	if nil == task.InParam {
		log.Warn("add task Do func inParam is nil")
		return
	}
	if nil == task.OutParam {
		log.Warn("add task Do func outParam is nil")
		return
	}

	if atomic.LoadInt32(&(wp.stat)) != WorkerPool_Stat_Running {
		log.Warn("this worker pool is not a running stat! the stat is ", wp.stat)
		return
	}

	atomic.AddInt32(&wp.addTaskStat, 1)
	wp.chWorkTask <- *task
	atomic.AddInt32(&wp.addTaskStat, -1)

	wp.addWorkerWhenAddTask()

	return
}

// add worker when add task cond:
// 1. 1 < work task num < (min worker num)/2
// 2. cur worker num < max worker num
func (wp *WorkerPool) addWorkerWhenAddTask() {
	chWorkTaskNum := len(wp.chWorkTask)
	if chWorkTaskNum > 1 && int32(chWorkTaskNum) > wp.minWorkerNum/2 && atomic.LoadInt32(&wp.curWorkerNum) < wp.maxWorkerNum {
		wp.chAddWorker <- 1
		wp.addWorkerLastTime = time.Now().Unix()
	}

	return
}

func (wp *WorkerPool) GetStat() int32 {
	return atomic.LoadInt32(&wp.stat)
}

func (wp *WorkerPool) GetCurrentGoNumber() int32 {
	return atomic.LoadInt32(&wp.curWorkerNum)
}

func (wp *WorkerPool) isWorking() bool {
	return atomic.LoadInt32(&(wp.stat)) == WorkerPool_Stat_Running ||
		atomic.LoadInt32(&(wp.stat)) == WorkerPool_Stat_Stoping
}

func (wp *WorkerPool) close() {
	wp.lock.Lock()
	if wp.isWorking() {
		atomic.SwapInt32(&wp.stat, WorkerPool_Stat_Stop)
		for index, _ := range wp.workers {
			atomic.SwapInt32(&wp.workers[index].hasGoroutineRunning, 1) // 不让自启
			//close(wp.workers[index].chWatchGoroutineOut)
			close(wp.workers[index].chExecuteGoroutineOut)
			//close(wp.workers[index].chTaskDoRes)
		}
		close(wp.chAddWorker)
		//close(wp.chWorkTask)
	}
	wp.lock.Unlock()
}

func (wp *WorkerPool) Stop() {
	if atomic.LoadInt32(&(wp.stat)) != WorkerPool_Stat_Running {
		log.Warn("this worker pool is not a running stat,can't stop ! the stat is ", wp.stat)
		return
	}
	atomic.SwapInt32(&(wp.stat), WorkerPool_Stat_Stoping)

	//wait add task product over
	stopTime := time.Now().Unix()
	for atomic.LoadInt32(&wp.addTaskStat) > 0 && time.Now().Unix()-stopTime < addWaitingTimeWhenStopPool {
		time.Sleep(10 * time.Millisecond)
	}

	log.Info("stop worker pool, stop to close chWorkTask -> close worker pool")
	close(wp.chWorkTask) // stop to close chWorkTask -> close worker pool

	wp.wg.Wait()
}
