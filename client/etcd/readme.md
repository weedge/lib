#### 介绍

使用etcd提供分布式锁，服务发现，服务元数据配置等

#### 功能

- [x] 服务发现, service provider register(keepalive 健康检查,使用etcd续租的特性); server cusumer resolver by grpc (etcdv3:///{服务单元名称});

  ps: 服务发现之后负载均衡分3种模式：(gRPC client  dial 和server建立的长连接,gRPC 负载均衡是基于每次调用)

  1. 客户端解析服务单元的ip列表，然后进程内lb，需要不同语言支持；(**采用这个模式**，通过etcd事件通知实时感知获取ip列表，ral/[brpc](https://github.com/apache/incubator-brpc/blob/master/docs/cn/load_balancing.md)轮训bns获取ip列表，client  dial 和server建立短链接的方式)
  2. 客户端解析服务单元的ip列表，然后发给本机单独lb进程，不需要多语言支持，但是不便排查问题;
  3. 客户端解析服务单元的ip列表，发给lb服务，存在单点问题，会成为整体服务的性能瓶颈；
  4. 容器化通过边车代理的负载均衡，和客户端服务部署在同一个pod, 通过虚拟化网络进行数据交互(unix domian socket)，网络消耗小， istio负载均衡模块通过envoy+polit实现，服务注册通过coreDNS, envoy 通过k8s部署在ingress 和 egress 中，用于pod中服务的注册和发现；

- [x] 分布式锁，etcd clientv3 本身提供，直接使用就行;

- [ ] 服务元数据配置管理;



#### reference

1. [etcd: 从应用场景到实现原理的全方位解读](https://www.infoq.cn/article/etcd-interpretation-application-scenario-implement-principle/)
2. [grpc-example-features](https://github.com/grpc/grpc-go/tree/master/examples/features)
3. [grpc-core-concepts](https://grpc.io/docs/what-is-grpc/core-concepts/)
4. [gRPC 长连接在微服务业务系统中的实践](https://www.infoq.cn/article/cpxr35bwjttgncltyekz)
5. [客户端负载均衡](http://icyfenix.cn/distribution/connect/load-balancing.html)

