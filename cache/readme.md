#### 介绍

​	并发读流量高的场景，将远端交互的不易变的数据加载入本地缓存中使用，降低网络交互，减少网络io和存储读io； 比如，策略数据，配置数据，有序追加数据等。



##### 本地缓存 zset  热key range 有序追加数据流程

![lru-cache-sortedlist](https://raw.githubusercontent.com/weedge/lib/main/cache/lru-cache-sortedlist.png)

