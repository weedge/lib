#### 介绍

使用etcd提供分布式锁，服务发现，服务元数据配置等

#### 功能

- [x] 服务发现, service provider register(keepalive 健康检查,使用etcd续租的特性); server cusumer resolver by grpc (etcdv3:///{服务单元名称}); 

  ps: 服务发现之后负载均衡分3种模式：

  	1. 客户端解析服务单元的ip列表，然后进程内lb，需要不同语言支持；(**采用这个模式**，通过etcd事件通知实时感知获取ip列表，ral/[brpc](https://github.com/apache/incubator-brpc/blob/master/docs/cn/load_balancing.md#%E8%B4%9F%E8%BD%BD%E5%9D%87%E8%A1%A1)轮训bns获取ip列表)
   	2. 客户端解析服务单元的ip列表，然后发给本机单独lb进程，不需要多语言支持，但是不便排查问题;
   	3. 客户端解析服务单元的ip列表，发给lb服务，存在单点问题，会成为整体服务的性能瓶颈；

- [x] 分布式锁，etcd clientv3 本身提供，直接使用就行;

- [ ] 服务元数据配置管理;



#### reference

1. [etcd: 从应用场景到实现原理的全方位解读](https://www.infoq.cn/article/etcd-interpretation-application-scenario-implement-principle/)

