## 介绍

该workerpool 通过channel 存放任务，多个 Worker 共享同一个任务 Channel，通过多个协程来消费池中的任务执行，协程根据提交的任务数动态扩缩协程；任务可以定义输入，输出，超时时间；通过channel 返回是否超时。



## 场景

接口请求中经常会有<u>批量任务</u>执行(可以是不同任务)，将这些任务放入任务工作池中并发处理，提高接口吞吐率。

**Tips:** 

ants 通过sync.Pool管理worker，动态扩缩管理worker池；每提交一个任务之前会从池中获取worker，初始化一个协程来执行，这个worker有单独的任务通道channel；然后将任务提交至channel中，对应协程异步执行。如果池子满了，是否非阻塞模式，是直接nil返回，否则继续等待可用worker执行，执行完放入池中；

ants 是运行时从池中获取管道初始协程，然后往管道提交任务协程异步处理；

而这里实现的workerpool是启动时初始化缓冲任务管道大小，运行时根据提交任务的数量/速度，动态扩缩处理任务协程数目；

一个是突增式处理，一个是扩展式处理，如果是潮汐🌊流量耗时短任务可以用第一种方式，如果是大量批量耗时相对比较高的任务可以采用第二种方式；

## 框架

![workerpool](https://raw.githubusercontent.com/weedge/lib/main/pool/workerpool/workerpool.png)

##### reference

1. [ants](https://github.com/panjf2000/ants) 
2. [Concurrency in Golang And WorkerPool](https://hackernoon.com/concurrency-in-golang-and-workerpool-part-1-e9n31ao) [Go语言的并发与WorkerPool](https://mp.weixin.qq.com/s?__biz=MzI2MDA1MTcxMg==&mid=2648468373&idx=1&sn=dc9c6e56cbd20c79a2593481100c69da) Github:[goworkerpool](https://github.com/Joker666/goworkerpool.git)
3. [The Case For A Go Worker Pool](https://brandur.org/go-worker-pool) GitHub: [worker-pool](https://github.com/vardius/worker-pool)
4. [一文搞懂如何实现 Go 超时控制](https://segmentfault.com/a/1190000039731121)
5. [使用 Golang Timer 的正确方式](http://russellluo.com/2018/09/the-correct-way-to-use-timer-in-golang.html)
5. [Pool：性能提升大杀器](https://time.geekbang.org/column/article/301716)
5. [Visually Understanding Worker Pool](https://medium.com/coinmonks/visually-understanding-worker-pool-48a83b7fc1f5)
5. [Java线程池实现原理及其在美团业务中的实践](https://tech.meituan.com/2020/04/02/java-pooling-pratice-in-meituan.html)



##### 修复问题：

1. 新增任务task定义超时时间，以及处理超时时间回调函数，woker获取任务执行的时候，进行超时任务处理，去掉worker对应的watch, 去掉冗余逻辑； 2021/9/26

