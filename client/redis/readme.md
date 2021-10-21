#### 介绍

使用redis client库 [go-redis/redis](github.com/go-redis/redis) 提供分布式锁，分布式缓存，限流等功能

#### 功能

- [x] 分布式锁dlock, 支持unblock TryLock, block Lock, UnLock, watch key to lease util unlock;(注意依赖支持redis协议部署的集群满足CP还是AP, 满足AP的锁是不可靠的，比如redis主从哨兵模式，为了提高锁的可靠性可以部署至少5个实例的redis实现redlock, 成本会高很多)
- [ ] 分布式限流，
- [ ] 分布式缓存，支持分布式缓存加载至本地缓存，hit 等监控



#### reference

1. [Monitoring using OpenTelemetry Metrics](https://blog.uptrace.dev/posts/opentelemetry-metrics-cache-stats/)
2. [go local cache algorithms benchmark](https://github.com/vmihailenco/go-cache-benchmark)
3. [go-redis/redis_rate](https://github.com/go-redis/redis_rate)

