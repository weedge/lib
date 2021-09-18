#### 介绍

队列库，提供优先队列，延迟队列基础操作



#### 功能

- [x] priority_queue: 优先队列，基于container/heap实现，采用min heap结构，提供Push，Pop, Top, PeekAndShift, Update 等操作函数接口
- [ ] delay_queue: 延迟队列, 基于优先队列，提供Offer, Poll, Do 操作函数，Offer（添加 bucket）和 Poll（获取并删除 bucket）的运作方式，



##### tips

heap 使用场景：最小顶堆，最大顶堆；优先级队列；有序小文件合并成大文件；定时任务； golang timer采用最小顶堆实现；

