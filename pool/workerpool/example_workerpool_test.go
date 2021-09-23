package workerpool

import (
	"fmt"
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

	time.Sleep(5 * time.Second)

	outParam.(*OutParam).err = nil
	outParam.(*OutParam).res = 1111

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

	for i, ch := range timeoutChs {
		go func(index int, ch chan bool) {
			for {
				select {
				case isTimeout, ok := <-ch:
					fmt.Println("timeOutChs--->", index, isTimeout, ok)
				}
			}
		}(i, ch)
	}
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

		/*
		isTimeout, ok := <-timeoutChs[i]
		fmt.Println("timeOutChs--->", i, isTimeout, ok)
		 */
	}

	for i, ch := range timeoutChs {
		go func(index int, ch chan bool) {
			for {
				select {
				case isTimeout, ok := <-ch:
					if isTimeout == true && ok {
						fmt.Println("timeOutChs--->", index, isTimeout, ok)
					}
				}
			}
		}(i, ch)
	}
}
