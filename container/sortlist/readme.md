##### 介绍

1. 使用[skiplist](https://github.com/huandu/skiplist)(支持单向排序)封装成 sortedlist([]byte), 支持并发场景 
2. 对container/list进行修改，加入score([]byte,可以改成Comparable接口来支持不同类型排序, 支持双向排序), 支持排序

##### references

1. [wiki: Skip list](https://en.wikipedia.org/wiki/Skip_list)
1. [Skip Lists：A Probabilistic Alternative to Balanced Trees](https://15721.courses.cs.cmu.edu/spring2018/papers/08-oltpindexes1/pugh-skiplists-cacm1990.pdf)

