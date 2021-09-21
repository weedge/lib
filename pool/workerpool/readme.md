##### 介绍

自动扩缩的任务工作池，任务可以定义超时时间，通过channel 返回是否超时



##### 场景

接口请求中经常会有批量任务执行，将这些任务放入任务工作池中并发处理，提高接口吞吐率，减少相应时间(RT)



##### 框架

![workerpool](https://raw.githubusercontent.com/weedge/lib/main/pool/workerpool/workerpool.png)

