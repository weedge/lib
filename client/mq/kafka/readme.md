#### 介绍

对Kafka client 开源库 [Shopify/sarama](https://github.com/Shopify/sarama) 进行单独封装，提供单一功能接口，KISS

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

// user intance interface to do（ConsumerMessage）  
type IConsumerMsg interface {
	Do(msg *sarama.ConsumerMessage) error
}
```



具体操作见example test


