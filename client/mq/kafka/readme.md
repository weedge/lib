#### 介绍

对Kafka client 开源库 [Shopify/sarama](https://github.com/Shopify/sarama) ,本身兼容kraft mode； 进行单独封装，提供单一功能接口

#### 功能接口

##### Consumer Group:

```go
// user just defined open consumer group option, init consumer conf to new ConsumerGroup
func NewConsumerGroup(name string, msg IConsumerMsg, options ...Option) (consumer *ConsumerGroup, err error) {

// start with ctx to cancel
func (consumer *ConsumerGroup) Start() 
func (consumer *ConsumerGroup) StartWithTimeOut(timeout time.Duration) 
func (consumer *ConsumerGroup) StartWithDeadline(time time.Time)

// cancel to close consumer group client 
func (consumer *ConsumerGroup) Close()

// user instance interface to do（ConsumerMessage）  
type IConsumerMsg interface {
	Consumer(msg *sarama.ConsumerMessage) error
}
```

##### Producer:

```go
// new sync/async producer to topic with option(requiredAcks,retryMaxCn,partitioning,compressions,TLS ...etc)
func NewProducer(topic string, pType string, options ...Option) (p *Producer)

// send string msg no key 
func (p *Producer) Send(val string) 

// send string msg by string key
func (p *Producer) SendByKey(key, val string)

// close sync/async producer
func (p *Producer) Close()
```

具体操作见example test

#### Kafka 拓扑结构

![kafka-zk](https://raw.githubusercontent.com/weedge/lib/main/client/mq/kafka/kafka-zk.png)



#### reference

1. [Kafka 0.10.0 doc](https://kafka.apache.org/0100/documentation.html)
2. [Kafka doc](https://kafka.apache.org/documentation.html) 最新版文档(2021/9/21 3.0版本)
3. [Apache Kafka 3.0 发布，离彻底去掉 ZooKeeper 更进一步](https://www.infoq.cn/article/RTTzLOMBPOx2TsL7dM9T)
4. [KIP-500: Replace ZooKeeper with a Self-Managed Metadata Quorum](https://cwiki.apache.org/confluence/display/KAFKA/KIP-500%3A+Replace+ZooKeeper+with+a+Self-Managed+Metadata+Quorum)
5. [KRaft (aka KIP-500) mode Early Access Release](https://github.com/apache/kafka/blob/6d1d68617ecd023b787f54aafc24a4232663428d/config/kraft/README.md)
6. [2.8 版本去掉zk简单操作视频](https://asciinema.org/a/403794/embed)

