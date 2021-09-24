package asynctask

import (
	"fmt"
	"runtime/debug"

	"github.com/weedge/lib/container/set"
	"github.com/weedge/lib/log"
	"github.com/weedge/lib/strings"
)

var asyncTaskNames *set.HashSet

func init() {
	asyncTaskNames = set.NewSet()
}

type IAsyncTask interface {
	Run() error
}

type AsyncTask struct {
	name string
	ch   chan IAsyncTask
}

func (this *AsyncTask) Close() {
	close(this.ch)
}

func NewAsyncTask(name string, taskChanNumber int64, onError func(err error)) (*AsyncTask, error) {
	if ok := asyncTaskNames.Contains(name); ok {
		return nil, fmt.Errorf("asynctask name duplicated: %v", name)
	}
	asyncTaskNames.Add(name)

	asyncTask := new(AsyncTask)
	asyncTask.name = name
	asyncTask.ch = make(chan IAsyncTask, taskChanNumber)
	asyncTask.run(onError)
	return asyncTask, nil
}

func Recover() {
	if e := recover(); e != nil {
		log.Errorf("panic: %v, stack: %v", e, strings.BytesToString(debug.Stack()))
	}
}

func (this *AsyncTask) run(onError func(err error)) {
	go func(name string, ch chan IAsyncTask) {
		defer Recover()

		log.Infof("AsyncTask [%v] created", name)
		for {
			asyncTask, ok := <-ch
			if !ok {
				break
			}
			if err := asyncTask.Run(); nil != err {
				if nil != onError {
					onError(err)
				}
			}
		}
		log.Infof("AsyncTask [%v] quit", name)
	}(this.name, this.ch)
}

func (this *AsyncTask) Post(task IAsyncTask) {
	this.ch <- task
}
