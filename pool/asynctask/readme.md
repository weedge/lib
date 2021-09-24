#### 介绍

异步执行任务，任务发至channel中，协程异步处理，无序关注超时，执行任务错误回调处理

#### 功能

```go
// 初始异步任务
func NewAsyncTask(name string, taskChanNumber int64, onError func(err error)) (*AsyncTask, error) 

// 提交任务
func (this *AsyncTask) Post(task IAsyncTask) 
```

