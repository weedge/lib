package asynctask

import (
	"fmt"
)

type testTask struct {
	Name string
}

func newTestTask(Name string) (m *testTask) {
	m = &testTask{
		Name: Name,
	}

	return m
}

func (m *testTask) Run() (err error) {
	//println("name", m.Name)
	if m.Name == "err" {
		return fmt.Errorf("wo qu err le")
	}

	return
}

func ExampleAsyncTask_NewAsyncTask() {
	asyncTask, err := NewAsyncTask("async.task.test", 3, 3, func(err error) {
		return
	})
	if err != nil {
		println("err", err.Error())
		return
	}
	asyncTask, err = NewAsyncTask("async.task.test", 3, 3, func(err error) {
		return
	})
	if err != nil {
		println("err", err.Error())
		return
	}

	asyncTask.Post(newTestTask("hello"))

	// Output:
	//
}

func ExampleAsyncTask_Ops() {
	asyncTask, err := NewAsyncTask("async.task.test", 3, 3, func(err error) {
		fmt.Println("Run err", err.Error())
		return
	})
	if err != nil {
		println("err", err.Error())
		return
	}

	asyncTask.Post(newTestTask("hello"), newTestTask("err"))

	asyncTask.Close()

	// Output:
	// Run err wo qu err le
}
