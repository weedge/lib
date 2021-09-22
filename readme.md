### 介绍
累积平时常用的golang开发组件库

### 组件列表

- [x] [poller](https://github.com/weedge/lib/tree/main/poller):  网络event poll库,支持tcp协议

- [x] [asyncbuffer](https://github.com/weedge/lib/tree/main/asyncbuffer): 累积数据至buffer中，异步批量处理

- [x] [concurrent_map](https://github.com/weedge/lib/tree/main/container/concurrent_map):  分片并发map

- [x] [queue](https://github.com/weedge/lib/tree/main/container/queue): 优先队列，延迟队列

- [x] [timingwheel](https://github.com/weedge/lib/tree/main/timingwheel): 层级时间轮

- [x] [buffer pool](https://github.com/weedge/lib/tree/main/pool/bufferpool): 临时对象池, 复用对象，减少gc, 优化逻辑, 

- [x] [worker pool](https://github.com/weedge/lib/tree/main/pool/workerpool)(并发处理的工作任务池，支持超时任务，自动扩缩worker goroutine) 

- [ ] [runtimer](https://github.com/weedge/lib/tree/main/runtimer): 对goroutine运行异常的封装，以及获取goroutine运行时调用的堆栈信息

- [ ] [balance](https://github.com/weedge/lib/tree/main/balance): 负载均衡算法，一致性hash, rr, wrr 等

- [ ] cache: [本地缓存patrickmn/go-cache](https://github.com/patrickmn/go-cache)，[hashicorp/golang-lru](https://github.com/hashicorp/golang-lru)

- [ ] net: 网络工具库

- [ ] zerocopy: 利用底层系统的零拷贝技术，mmap: [edsrzf/mmap-go](https://github.com/edsrzf/mmap-go) sendFile, splice,三方库封装使用

- [ ] [limiter](https://github.com/weedge/lib/tree/main/limiter): 服务提供方限流算法，防止服务过载策略，单机 固定/滑动时间窗口限流算法，漏桶([uber-go/ratelimit]( https://github.com/uber-go/ratelimit) )/( [juju/ratelimit](https://github.com/juju/ratelimit) )令牌桶算法，分布式限流算法(redis 计数，通常在流量入口网关层处理，nginx+lua, golang) 对三方开源服务在业务的基础上进行封装；

- [ ] breaker: 服务消费方调用服务熔断机制，开源实现：[afex/hystrix-go](http://github.com/afex/hystrix-go)  [sony/gobreaker](github.com/sony/gobreaker)  对三方开源服务在业务的基础上进行封装；

- [ ] [log](https://github.com/weedge/lib/tree/main/log): 日志库, 基于uber [zap](https://github.com/uber-go/zap) 库，满足基础日志，访问日志，请求三方日志，panic日志，启动日志，

- [ ] metric： 监控统计方法，比如计算MAU,DAU，精度要求不高可以使用redis HyperLogLog (只需要12K内存，在标准误差0.81%的前提下，能够统计2^64个数据！HyperLogLog是一种基数估计算法)

- [ ] opentrace: 服务链路跟踪，比如生成requestId, traceId, spanId 等, [jaeger](https://github.com/jaegertracing/jaeger)  主要用来学习借鉴开源的服务框架思路

- [ ] consensus: 分布式一致性协议算法，[raft](https://raft.github.io/)  开源方案 [hashicorp/raft](https://github.com/hashicorp/raft)  主要用来学习借鉴开源的服务框架思路

- [ ] xdb: 单数据存储实例，B+tree [etcd-io/bbolt](https://github.com/etcd-io/bbolt)  LMS-tree  [syndtr/goleveldb](https://github.com/syndtr/goleveldb) 主要用来学习借鉴开源的服务框架思路

- [ ] rpc：三方rpc协议框架封装, [grpc-go](https://github.com/grpc/grpc-go), [thrift](https://github.com/apache/thrift), [rpcx](https://github.com/smallnest/rpcx),  [dubbo-go](https://github.com/apache/dubbo-go) 主要用来深入学习框架思路

  

