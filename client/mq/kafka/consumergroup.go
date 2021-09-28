package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/weedge/lib/container/set"
	"github.com/weedge/lib/log"
	"github.com/weedge/lib/runtimer"

	"github.com/Shopify/sarama"
)

var consumerGroupNames *set.HashSet

func init() {
	consumerGroupNames = set.NewSet()
}

type IConsumerMsg interface {
	Do(msg *sarama.ConsumerMessage) error
}

// Consumer represents a Sarama consumer group consumer
type ConsumerGroup struct {
	name      string // the same to groupId
	ready     chan bool
	config    *sarama.Config
	topicList []string
	client    sarama.ConsumerGroup
	cancel    context.CancelFunc
	wg        *sync.WaitGroup
	msg       IConsumerMsg
	msgMeta   string
}

// user just defined open consumer group option
func NewConsumerGroup(name string, msg IConsumerMsg, options ...Option) (consumer *ConsumerGroup, err error) {
	consumer = &ConsumerGroup{name: name, wg: &sync.WaitGroup{}, msg: msg}

	consumerOpts := getConsumerOptions(options...)
	log.Info(fmt.Sprintf("consumer options:%+v", consumerOpts))

	consumer.ready = make(chan bool)
	consumer.config = sarama.NewConfig()
	consumer.topicList = consumerOpts.topicList
	consumer.config.Version, err = sarama.ParseKafkaVersion(consumerOpts.version)
	if err != nil {
		return
	}

	switch consumerOpts.reBalanceStrategy {
	case "sticky":
		consumer.config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	case "roundrobin":
		consumer.config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "range":
		consumer.config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	default:
		err = fmt.Errorf("un define consumer group rebalance strategy")
		return
	}

	switch consumerOpts.initialOffset {
	case "newest":
		consumer.config.Consumer.Offsets.Initial = sarama.OffsetNewest
	case "oldest":
		consumer.config.Consumer.Offsets.Initial = sarama.OffsetOldest
	default:
		err = fmt.Errorf("un define consumer group rebalance strategy")
		return
	}

	consumer.client, err = sarama.NewConsumerGroup(consumerOpts.brokerList, consumerOpts.groupId, consumer.config)
	if err != nil {
		err = fmt.Errorf("error creating consumer group client: %v", err)
		return
	}
	log.Info("init consumer group ok!")

	return
}

func (consumer *ConsumerGroup) Start() {
	var ctx context.Context
	ctx, consumer.cancel = context.WithCancel(context.Background())
	consumer.startWithContext(ctx)
}

func (consumer *ConsumerGroup) StartWithTimeOut(timeout time.Duration) {
	var ctx context.Context
	ctx, consumer.cancel = context.WithTimeout(context.Background(), timeout)
	consumer.startWithContext(ctx)
}

func (consumer *ConsumerGroup) StartWithDeadline(time time.Time) {
	var ctx context.Context
	ctx, consumer.cancel = context.WithDeadline(context.Background(), time)
	consumer.startWithContext(ctx)
}

func (consumer *ConsumerGroup) startWithContext(ctx context.Context) {
	if consumerGroupNames.Contains(consumer.name) {
		log.Warn("have the same consumer to start name", consumer.name)
		return
	}
	runtimer.GoSafely(nil, false, func() {
		// Track errors
		for err := range consumer.client.Errors() {
			log.Error(err)
		}
	}, nil)

	runtimer.GoSafely(consumer.wg, false, func() {
		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims
		for {
			if err := consumer.client.Consume(ctx, consumer.topicList, consumer); err != nil {
				log.Error("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				log.Info("Sarama consumer stop!...")
				return
			}
			consumer.ready = make(chan bool)
		}
	}, nil)

	<-consumer.ready // Await till the consumer has been set up
	log.Info("Sarama consumer up and running!...")
	consumerGroupNames.Add(consumer.name)
}

func (consumer *ConsumerGroup) Close() {
	if consumer.cancel != nil {
		consumer.cancel()
	}

	consumerGroupNames.Remove(consumer.name)
	err := consumer.client.Close()
	if err != nil {
		log.Error("consumer.client.Close err", err.Error())
	}

	consumer.wg.Wait()
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *ConsumerGroup) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *ConsumerGroup) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *ConsumerGroup) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		log.Info(fmt.Sprintf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic))
		err := consumer.msg.Do(message)
		if err != nil {
			log.Error(fmt.Sprintf("consumer.msg.Do error:%s", err.Error()))
			continue
		}

		//commit msg ack to consumer ok
		session.MarkMessage(message, consumer.msgMeta)
	}

	return nil
}
