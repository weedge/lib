package workerpool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type InParam struct {
	key string
	val int
}

type OutParam struct {
	res int

	err error
}

func MyDo(inParam interface{}, outParam interface{}) bool {
	println("InParam:", inParam.(*InParam).key, inParam.(*InParam).val)

	time.Sleep(2 * time.Second)

	outParam.(*OutParam).err = nil
	outParam.(*OutParam).res = 1111

	return true
}

func MyDoSlow(inParam interface{}, outParam interface{}) bool {
	println("InParam:", inParam.(*InParam).key, inParam.(*InParam).val)

	time.Sleep(4 * time.Second)

	outParam.(*OutParam).err = nil
	outParam.(*OutParam).res = 1111
	//println("MyDoSlow run ok", outParam)

	return true
}

func ExampleWorkerPool_RunOk() {
	wp := NewWorkerPool(3, 5, 3)

	wp.Run()
	defer wp.Stop()

	inParam := &InParam{
		key: "wo qu",
		val: 110,
	}
	outParam := &OutParam{}
	timeoutCh := make(chan bool, 1)
	task := &Task{
		Do:          MyDo,
		InParam:     inParam,
		OutParam:    outParam,
		ChIsTimeOut: timeoutCh,
		TimeOut:     3 * time.Second,
	}
	wp.AddTask(task)
	isTimeout := <-timeoutCh
	fmt.Println(isTimeout, outParam.err, outParam.res)

	// Output:
	// false <nil> 1111
}

func ExampleWorkerPool_RunTimeout() {
	wp := NewWorkerPool(3, 5, 3)

	wp.Run()
	defer wp.Stop()

	inParam1 := &InParam{
		key: "wo qu",
		val: 110,
	}
	outParam1 := &OutParam{}
	timeoutCh1 := make(chan bool, 1)
	task1 := &Task{
		Do:          MyDoSlow,
		InParam:     inParam1,
		OutParam:    outParam1,
		ChIsTimeOut: timeoutCh1,
		TimeOut:     3 * time.Second,
	}
	wp.AddTask(task1)
	isTimeout1 := <-timeoutCh1
	fmt.Println(isTimeout1, outParam1.err, outParam1.res)

	// Output:
	// true <nil> 0
}

func TestWorkerPool_RunManyOk(t *testing.T) {
	wp := NewWorkerPool(3, 5, 3)

	wp.Run()
	defer wp.Stop()

	timeoutChs := make([]chan bool, 10)
	for i := 0; i < 10; i++ {
		inParam := &InParam{
			key: "wo qu",
			val: i + 100000,
		}
		outParam := &OutParam{}
		timeoutChs[i] = make(chan bool, 1)
		task := &Task{
			Do:          MyDo,
			InParam:     inParam,
			OutParam:    outParam,
			ChIsTimeOut: timeoutChs[i],
			TimeOut:     3 * time.Second,
		}
		wp.AddTask(task)
	}

	batchAsyncDoTimeout(timeoutChs)
}

func TestWorkerPool_RunManyTimeout(t *testing.T) {
	wp := NewWorkerPool(3, 5, 3)

	wp.Run()
	defer wp.Stop()

	timeoutChs := make([]chan bool, 10)
	for i := 0; i < 10; i++ {
		inParam := &InParam{
			key: "wo qu",
			val: i + 1000,
		}
		outParam := &OutParam{}
		timeoutChs[i] = make(chan bool, 1)
		task := &Task{
			Do:          MyDoSlow,
			InParam:     inParam,
			OutParam:    outParam,
			ChIsTimeOut: timeoutChs[i],
			TimeOut:     3 * time.Second,
		}
		wp.AddTask(task)
	}

	batchAsyncDoTimeout(timeoutChs)
}

func TestWorkerPool_RunManyOkTimeout(t *testing.T) {
	wp := NewWorkerPool(3, 5, 3)

	wp.Run()
	defer wp.Stop()

	timeoutChs := make([]chan bool, 10)
	for i := 0; i < 10; i++ {
		inParam := &InParam{
			key: "wo qu",
			val: i + 1000,
		}
		outParam := &OutParam{}
		timeoutChs[i] = make(chan bool, 1)
		timeOut := 3 * time.Second
		if i%2 == 0 {
			timeOut = 5 * time.Second
		}
		task := &Task{
			Do:          MyDoSlow,
			InParam:     inParam,
			OutParam:    outParam,
			ChIsTimeOut: timeoutChs[i],
			TimeOut:     timeOut,
		}
		wp.AddTask(task)
	}

	batchAsyncDoTimeout(timeoutChs)
}

func batchAsyncDoTimeout(timeoutChs []chan bool) {
	wg := &sync.WaitGroup{}
	for i, ch := range timeoutChs {
		wg.Add(1)
		go func(wg *sync.WaitGroup, index int, ch chan bool) {
			defer wg.Done()
			for {
				select {
				case isTimeout, ok := <-ch:
					fmt.Println("timeOutChs--->", index, isTimeout, ok)
					return
				}
			}
		}(wg, i, ch)
	}
	wg.Wait()
}

func TestWorkerPool_RunManyOkTimeoutWithTimeOutHandle(t *testing.T) {
	wp := NewWorkerPool(3, 5, 3)

	wp.Run()
	defer wp.Stop()

	//timeoutChs := make([]chan bool, 10)
	for i := 0; i < 10; i++ {
		inParam := &InParam{
			key: "wo qu",
			val: i + 1000,
		}
		outParam := &OutParam{}
		//timeoutChs[i] = make(chan bool, 1)
		timeOut := 3 * time.Second
		if i%2 == 0 {
			timeOut = 5 * time.Second
		}
		task := &Task{
			Do:          MyDoSlow,
			InParam:     inParam,
			OutParam:    outParam,
			//ChIsTimeOut: timeoutChs[i],
			TimeOut:     timeOut,
			OnTimeOut: func(inParam interface{}, outParam interface{}) {
				key := inParam.(*InParam).key
				val := inParam.(*InParam).val
				println("inParam", key, val, "timeout")
			},
		}
		wp.AddTask(task)
	}

}
