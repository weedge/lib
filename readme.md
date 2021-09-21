### 介绍
累积平时常用的golang开发组件库

### 组件列表

- [x] poller:  网络event poll库

- [x] asyncbuffer: 累积数据至buffer中，异步批量处理

- [x] concurrent_map:  分片并发map

- [x] queue: 优先队列，延迟队列

- [x] timingwheel: 层级时间轮

- [x] pool: 通过池化方法，复用对象，减少gc, 优化逻辑，buffer pool(临时对象池), worker pool(并发处理的工作任务池，支持超时任务，自动扩缩worker goroutine) 

- [ ] runtimer: 对goroutine运行异常的封装，以及获取goroutine运行时调用的堆栈信息

- [ ] balance: 负载均衡算法，一致性hash, rr, wrr 等

- [ ] cache: [本地缓存](https://github.com/patrickmn/go-cache)，[lru](https://github.com/hashicorp/golang-lru)

- [ ] net: 网络工具库

- [ ] log: 日志库, 基于uber [zap](https://github.com/uber-go/zap) 库

  

