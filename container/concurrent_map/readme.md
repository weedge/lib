#### 介绍

分片并发map

#### 功能

- [x] 提供基础map操作，Get, Set, Del, Count, Iter(Range), Snapshot, Clear 等功能；
- [x]  通过自定义hash算法生成key进行partition; 比如：google [cityhash](https://github.com/zentures/cityhash) 为了满足大数据量的需求，减少碰撞

####  参考

1. [concurrent-map](https://github.com/orcaman/concurrent-map)
2. [allegro/bigcache](https://github.com/allegro/bigcache)

