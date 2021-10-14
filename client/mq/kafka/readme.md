#### 介绍

对Kafka client 开源库 [Shopify/sarama](https://github.com/Shopify/sarama) ,本身兼容kraft mode(kafka version 2.8.0+)； 进行单独封装，提供单一功能接口

#### 功能接口

##### Authentication Options:

```go
type AuthOptions struct {
	// security-protocol: SSL (broker service config default PLAINTEXT, need security.inter.broker.protocol=SSL)
	enableSSL bool
	certFile  string // the optional certificate file for client authentication
	keyFile   string // the optional key file for client authentication
	caFile    string // the optional certificate authority file for TLS client authentication
	verifySSL bool   // the optional verify ssl certificates chain

	//SASL: PLAINTEXT(Kafka 0.10.0.0+), SCRAM(kafka 0.10.2.0+ dynamic add user), OAUTHBEARER(Kafka2.0.0+,JWT)
	enableSASL     bool
	authentication string // PLAINTEXT,SCRAM,OAUTHBEARER,if enableSASL true,default SCRAM
	saslUser       string // The SASL username
	saslPassword   string // The SASL password
	scramAlgorithm string // The SASL SCRAM SHA algorithm sha256 or sha512 as mechanism
}
```

##### Consumer Options:

```go
type ConsumerGroupOptions struct {
	version           string   // kafka version
	brokerList        []string // kafka broker ip:port list
	topicList         []string // subscribe topic list
	groupId           string   // consumer groupId
	reBalanceStrategy string   // consumer group partition assignment strategy (range, roundrobin, sticky)
	initialOffset     string   // initial offset to consumer (oldest, newest)

	*auth.AuthOptions
}
```

##### Consumer Group:

```go
// user just defined open consumer group option, init consumer conf to new ConsumerGroup
func NewConsumerGroup(name string, msg IConsumerMsg, authOpts []auth.Option, options ...Option) (consumer *ConsumerGroup, err error) {

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

##### Producer Options:

```go
type ProducerOptions struct {
	version           string        // kafka version
	clientID          string        // The client ID sent with every request to the brokers.
	brokerList        []string      // broker list
	partitioning      string        // key {partition}(manual),hash,random
	requiredAcks      int           // required ack
	timeOut           time.Duration // The duration the producer will wait to receive -required-acks
	retryMaxCn        int           // retry max cn (default: 3)
	compression       string        // msg compression(gzip,snappy,lz4,zstd)
	maxOpenRequests   int           // The maximum number of unacknowledged requests the client will send on a single connection before blocking (default: 5)
	maxMessageBytes   int           // The max permitted size of a message (default: 1000000)
	channelBufferSize int           // The number of events to buffer in internal and external channels.

	// flush batches
	flushFrequencyMs int // The best-effort frequency of flushes
	flushBytes       int // The best-effort number of bytes needed to trigger a flush.
	flushMessages    int // The best-effort number of messages needed to trigger a flush.
	flushMaxMessages int // The maximum number of messages the producer will send in a single request.

	*auth.AuthOptions
}
```

##### Producer:

```go
// new sync/async producer to topic with option(requiredAcks,retryMaxCn,partitioning,compressions,TLS ...etc)
func NewProducer(topic string, pType string, authOpts []auth.Option, options ...Option) (p *Producer)

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
7. [GoLang：你真的了解 HTTPS 吗?](https://mp.weixin.qq.com/s/ibwNtDc2zd2tdhMN7iROJw)
8. [知乎基于Kubernetes的kafka平台的设计和实现](https://zhuanlan.zhihu.com/p/36366473)

