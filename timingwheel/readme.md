##### from

https://github.com/RussellLuo/timingwheel

##### 使用场景

延迟一段时间后执行任务；C1000k的场景，每个连接一个Timer，小顶堆的结构查找删除时间复杂度O(logn)，效率会比较低；采用时间轮实现的 Timer,创建和删除的时间复杂度为 O(1); 效仿kafka Purgatory 的实现，层级时间轮(Hierarchical Timing Wheels); 主要场景：

1. 长链接心跳检测，聊天，推送
2. 客户端发起 HTTP 请求后，如果在指定时间内没有收到服务器的响应，则自动断开连接

##### 对比

Go 1.14 对time.Timer进行优化，在go1.15 的测试：

```shell
go test -bench=. -run=none -benchmem -memprofile=mem.pprof -cpuprofile=cpu.pprof -blockprofile=block.pprof
goos: darwin
goarch: amd64
pkg: github.com/weedge/lib/timingwheel
BenchmarkTimingWheel_StartStop/N-1m-8         	 4408240	       238 ns/op	      84 B/op	       2 allocs/op
BenchmarkTimingWheel_StartStop/N-5m-8         	 4696995	       261 ns/op	     104 B/op	       2 allocs/op
BenchmarkTimingWheel_StartStop/N-10m-8        	 2271856	       467 ns/op	      82 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-1m-8       	 6194773	       204 ns/op	      83 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-5m-8       	 6753042	       219 ns/op	      80 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-10m-8      	 1000000	     51426 ns/op	    3859 B/op	       9 allocs/op
#最后一个操作慢有gc导致

#第二次运行结果，使用go1.15
BenchmarkTimingWheel_StartStop/N-1m-8         	 4488100	       237 ns/op	      84 B/op	       2 allocs/op
BenchmarkTimingWheel_StartStop/N-5m-8         	 4677676	       263 ns/op	     103 B/op	       2 allocs/op
BenchmarkTimingWheel_StartStop/N-10m-8        	 4872208	       518 ns/op	      54 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-1m-8       	 6061422	       200 ns/op	      81 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-5m-8       	 6735716	       193 ns/op	      83 B/op	       1 allocs/op
BenchmarkStandardTimer_StartStop/N-10m-8      	 4947601	       259 ns/op	      85 B/op	       1 allocs/op

go tool pprof -http=":8080" cpu.pprof
Serving web UI on http://localhost:8080
#可以通过火焰图查看cpu, 内存，函数运行耗时，整体timingwheel 耗时相对少些
```

总结： time.Timer 经过优化之后，性能有所提升，但是整体小顶堆结构的添加删除操作O(logn)比双向循环链表O(1)的效率要低

##### reference

1. [层级时间轮的 Golang 实现](http://russellluo.com/2018/10/golang-implementation-of-hierarchical-timing-wheels.html) 
2. [Apache Kafka, Purgatory, and Hierarchical Timing Wheels](confluent.io/blog/apache-kafka-purgatory-hierarchical-timing-wheels/)
3. [完全兼容golang定时器的高性能时间轮实现(go-timewheel)](http://xiaorui.cc/archives/6160) 
4. [golang netty timewheel](https://github.com/dubbogo/gost/blob/master/time/timer.go#L158-L172)
5. [golang timer(计时器)](https://golang.design/under-the-hood/zh-cn/part2runtime/ch06sched/timer/)
6. [第 74 期 time.Timer 源码分析 (Go 1.14)](https://github.com/talkgo/night/issues/541)
7. [论golang Timer Reset方法使用的正确姿势](https://tonybai.com/2016/12/21/how-to-use-timer-reset-in-golang-correctly/)
8. [George Varghese , Anthony Lauck, Hashed and hierarchical timing wheels: efficient data structures for implementing a timer facility, IEEE/ACM Transactions on Networking (TON), v.5 n.6, p.824-834, Dec. 1997](http://www.cs.columbia.edu/~nahum/w6998/papers/ton97-timing-wheels.pdf)

