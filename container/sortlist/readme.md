#### 介绍

1. 使用[skiplist](https://github.com/huandu/skiplist) 封装成 sortedlist(MemberScore), 支持并发场景,Range操作O(log(n)+m)
2. 对container/list进行修改，加入score([]byte,可以改成Comparable接口来支持不同类型排序)，支持并发场景，Range操作O(n+m)

#### 使用场景

两者可用于从 redis zset 通过 `ZRANGE ** start stop WITHSCORES` (O(log(n)+m))或 `ZRANGEBYSCORE ** min max WITHSCORES`(O(log(n)+m)) 获取的数据放入本地进程SortedList结构中使用，减少网络io，并发请求大时，缓解出现热key的情况

tips: 写入redis zset的数据是时序append加入到有序集合中的，不能出现更新历史数据的情况，以防顺序变化，导致本地缓存不一致 

#### references

1. [wiki: Skip list](https://en.wikipedia.org/wiki/Skip_list)
1. [Skip Lists：A Probabilistic Alternative to Balanced Trees](https://15721.courses.cs.cmu.edu/spring2018/papers/08-oltpindexes1/pugh-skiplists-cacm1990.pdf)

