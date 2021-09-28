#### 介绍

对Kafka client 开源库 [Shopify/sarama](https://github.com/Shopify/sarama) 进行单独封装，提供单一功能接口，KISS

#### 功能接口

```go
// user just defined open consumer group option, init consumer conf to new ConsumerGroup
func NewConsumerGroup(name string, msg IConsumerMsg, options ...Option) (consumer *ConsumerGroup, err error) {

// start with ctx to cancel
func (consumer *ConsumerGroup) Start() 
func (consumer *ConsumerGroup) StartWithTimeOut(timeout time.Duration) 
func (consumer *ConsumerGroup) StartWithDeadline(time time.Time)

// cancel to close consumer group client 
func (consumer *ConsumerGroup) Close()

// user intance interface to do（ConsumerMessage）  
type IConsumerMsg interface {
	Do(msg *sarama.ConsumerMessage) error
}


```

具体操作见example test



##### mac安装测试

```shell
# 启动zookeeper，单个服务，
/usr/local/bin/zookeeper-server-start /usr/local/etc/kafka/zookeeper.properties
# 启动kafka broker 0
/usr/local/bin/kafka-server-start /usr/local/etc/kafka/server.properties
# 启动kafka broker 1 修改broker.id=1 listeners=PLAINTEXT://:9093 log.dirs=/usr/local/var/lib/kafka-1-logs
/usr/local/bin/kafka-server-start /usr/local/etc/kafka/server-1.properties
# 启动kafka broker 2 修改broker.id=2 listeners=PLAINTEXT://:9094 log.dirs=/usr/local/var/lib/kafka-2-logs
/usr/local/bin/kafka-server-start /usr/local/etc/kafka/server-2.properties
# 创建topic, topic 分片数，副本数：为1只有一个主副本，没有备份
/usr/local/bin/kafka-topics  --create --zookeeper localhost:2181 --replication-factor 3 --partitions 2 --topic sarama
# 对sarama topic生产数据
/usr/local/bin/kafka-console-producer --broker-list localhost:9092 --topic sarama
# 从开始offset 消费 sarama topic
/usr/local/bin/kafka-console-consumer --bootstrap-server localhost:9092 --topic sarama --from-beginning
```

 

