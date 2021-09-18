package queue

import (
	"fmt"
	"time"
)

type Task struct {
	Id  int
	Exp int64
}

func (t *Task) Do() {
	fmt.Println(t.Id, t.Exp)
}

func ExampleDelayQueue_Ops() {
	dq := NewDelayQueue(10)

	pollExitCh, doExitCh := make(chan struct{}), make(chan struct{})
	go dq.Poll(pollExitCh, func() int64 { return 1000 })
	go dq.Do(doExitCh, func(item interface{}) {
		task := item.(*Task)
		task.Do()
	})

	tasks := []*Task{
		{Id: 2, Exp: 2000},
		{Id: 3, Exp: 3000},
		{Id: 1, Exp: 1000},
		{Id: 4, Exp: 4000},
	}

	for _, task := range tasks {
		dq.Offer(task, task.Exp)
	}

	time.Sleep(3 * time.Second)

	pollExitCh <- struct{}{}
	doExitCh <- struct{}{}

	// Output:
	// 1 1000
}
