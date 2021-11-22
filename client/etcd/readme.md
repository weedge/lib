#### 介绍

使用etcd提供分布式锁，服务发现，服务元数据配置等

#### 功能

- [x] 服务发现, service provider register(keepalive 健康检查,使用etcd续租的特性); server cusumer resolver by grpc (etcdv3:///{服务单元名称});

  ps: 服务发现之后负载均衡分3种模式：(gRPC client  dial 和server建立的长连接,gRPC 负载均衡是基于每次调用)

  1. 客户端解析服务单元的ip列表，然后进程内lb，需要不同语言支持；(**采用这个模式**，通过etcd事件通知实时感知获取ip列表，ral/[brpc](https://github.com/apache/incubator-brpc/blob/master/docs/cn/load_balancing.md)轮训bns获取ip列表，client  dial 和server建立短链接的方式)
  2. 客户端解析服务单元的ip列表，然后发给本机单独lb进程，不需要多语言支持，但是不便排查问题;
  3. 客户端解析服务单元的ip列表，发给lb服务，存在单点问题，会成为整体服务的性能瓶颈；
  4. 容器化通过边车代理的负载均衡，和客户端服务部署在同一个pod, 通过虚拟化网络进行数据交互(unix domian socket)，网络消耗小， istio负载均衡模块通过envoy+polit实现，服务注册通过coreDNS, envoy 通过k8s部署在sidecar中，用于pod中服务的注册和发现, 负载均衡，限流，分流，熔断，监控 等服务治理策略方便升级，无需改动部署的app服务(画面: nice~)；

- [x] 分布式锁，etcd clientv3 本身提供，直接使用就行（服务侧对应server/etcdserver/api/v3lock中的代码）  适用于严格可靠锁的场景，etcd满足CP系统，通过raft一致性协议保证。

   Etcd 通过以下机制：Watch 机制、Lease 机制、Revision 机制和 Prefix 机制，赋予了 Etcd 实现分布式锁的能力。

  - **Lease 机制**：即租约机制（TTL，Time To Live），Etcd 可以为存储的 Key-Value 对设置租约，当租约到期，Key-Value 将失效删除；同时也支持续约，通过客户端可以在租约到期之前续约，以避免 Key-Value 对过期失效。Lease 机制可以保证分布式锁的安全性，为锁对应的 Key 配置租约，即使锁的持有者因故障而不能主动释放锁，锁也会因租约到期而自动释放。
  - **Revision 机制**：每个 Key 带有一个 Revision 号，每进行一次事务便加一，因此它是全局唯一的，如初始值为 0，进行一次 `put(key, value)`，Key 的 Revision 变为 1，同样的操作，再进行一次，Revision 变为 2；换成 key1 进行 put(key1, value) 操作，Revision 将变为 3；这种机制有一个作用：通过 Revision 的大小就可以知道写操作的顺序。在实现分布式锁时，多个客户端同时抢锁，根据 Revision 号大小依次获得锁，可以避免 “羊群效应” （也称“惊群效应”），实现公平锁。
  - **Prefix 机制**：即前缀机制，也称目录机制，例如，一个名为 `/mylock` 的锁，两个争抢它的客户端进行写操作，实际写入的 Key 分别为：`key1="/mylock/UUID1",key2="/mylock/UUID2"`，其中，UUID 表示全局唯一的 ID，确保两个 Key 的唯一性。很显然，写操作都会成功，但返回的 Revision 不一样，那么，如何判断谁获得了锁呢？通过前缀“/mylock” 查询，返回包含两个 Key-Value 对的 Key-Value 列表，同时也包含它们的 Revision，通过 Revision 大小，客户端可以判断自己是否获得锁，如果抢锁失败，则等待锁释放（对应的 Key 被删除或者租约过期），然后再判断自己是否可以获得锁。
  - **Watch 机制**：即监听机制(客户端和服务端通过grpc stream进行交互)，Watch 机制支持监听某个固定的 Key，也支持监听一个范围（前缀机制），当被监听的 Key 或范围发生变化，客户端将收到通知；在实现分布式锁时，如果抢锁失败，可通过 Prefix 机制返回的 Key-Value 列表获得 Revision 比自己小且相差最小的 Key（称为 Pre-Key），对 Pre-Key 进行监听，因为只有它释放锁，自己才能获得锁，如果监听到 Pre-Key 的 DELETE 事件，则说明 Pre-Key 已经释放，自己已经持有锁。



#### reference

1. [etcd: 从应用场景到实现原理的全方位解读](https://www.infoq.cn/article/etcd-interpretation-application-scenario-implement-principle/)
2. [grpc-example-features](https://github.com/grpc/grpc-go/tree/master/examples/features)
3. [grpc-core-concepts](https://grpc.io/docs/what-is-grpc/core-concepts/)
4. [gRPC 长连接在微服务业务系统中的实践](https://www.infoq.cn/article/cpxr35bwjttgncltyekz)
5. [客户端负载均衡](http://icyfenix.cn/distribution/connect/load-balancing.html)
6. [k8s-service-discovery-and-loadbalancing](https://jimmysong.io/kubernetes-handbook/practice/service-discovery-and-loadbalancing.html)
7. [istio集成服务注册中心](https://www.servicemesher.com/istio-handbook/practice/integration-registry.html)
7. [分布式锁：为什么基于etcd实现分布式锁比Redis锁更安全？](https://time.geekbang.org/column/article/350285)

