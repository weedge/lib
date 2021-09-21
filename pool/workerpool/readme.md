##### 介绍

自动扩缩的任务工作池，任务可以定义超时时间，通过channel 返回是否超时



##### 场景

接口请求中经常会有批量任务执行，将这些任务放入任务工作池中并发处理，提高接口吞吐率，减少相应时间(RT)



##### 框架

![workerpool](https://raw.githubusercontent.com/weedge/lib/main/pool/workerpool/workerpool.png)

##### reference

1. [ants](github.com/panjf2000/ants)
2. [Concurrency in Golang And WorkerPool](https://hackernoon.com/concurrency-in-golang-and-workerpool-part-1-e9n31ao) [Go语言的并发与WorkerPool](https://mp.weixin.qq.com/s?__biz=MzI2MDA1MTcxMg==&mid=2648468373&idx=1&sn=dc9c6e56cbd20c79a2593481100c69da) Github:[goworkerpool](https://github.com/Joker666/goworkerpool.git)
3. [The Case For A Go Worker Pool](https://brandur.org/go-worker-pool) GitHub: [worker-pool](https://github.com/vardius/worker-pool)

