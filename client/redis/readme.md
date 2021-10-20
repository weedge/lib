#### 介绍

使用redis client库 [go-redis/redis](github.com/go-redis/redis) 提供分布式锁，分布式缓存，限流等功能

#### 功能

- [x] 分布式锁dlock, 支持unblock TryLock, block Lock, UnLock, watch key to lease util unlock;
- [ ] 分布式限流，
- [ ] 分布式缓存，支持分布式缓存加载至本地缓存，hit 等监控



#### reference

1. [Monitoring using OpenTelemetry Metrics](https://blog.uptrace.dev/posts/opentelemetry-metrics-cache-stats/)
2. [go local cache algorithms benchmark](https://github.com/vmihailenco/go-cache-benchmark)
3. [go-redis/redis_rate](https://github.com/go-redis/redis_rate)

