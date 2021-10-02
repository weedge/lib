package consumer

import (
	"github.com/weedge/lib/client/mq/kafka/auth"
	"strings"
)

type ConsumerGroupOptions struct {
	version           string   // kafka version
	brokerList        []string // kafka broker ip:port list
	topicList         []string // subscribe topic list
	groupId           string   // consumer groupId
	reBalanceStrategy string   // consumer group partition assignment strategy (range, roundrobin, sticky)
	initialOffset     string   // initial offset to consumer (oldest, newest)

	*auth.AuthOptions
}

type Option interface {
	apply(ConsumerGroupOptions *ConsumerGroupOptions)
}

type funcServerOption struct {
	f func(*ConsumerGroupOptions)
}

func (fdo *funcServerOption) apply(do *ConsumerGroupOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(ConsumerGroupOptions *ConsumerGroupOptions)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

// kafka brokers ip:port,ip:port
func WithBrokerList(brokers string) Option {
	return newFuncServerOption(func(o *ConsumerGroupOptions) {
		if len(brokers) == 0 {
			panic("empty brokers")
		}
		o.brokerList = strings.Split(brokers, ",")
	})
}

// consumer groupId
func WithGroupId(groupId string) Option {
	return newFuncServerOption(func(o *ConsumerGroupOptions) {
		if len(groupId) == 0 {
			panic("no groupId to consumer")
		}
		o.groupId = groupId
	})
}

// kafka version
func WithVersion(version string) Option {
	return newFuncServerOption(func(o *ConsumerGroupOptions) {
		if len(version) == 0 {
			return
		}
		o.version = version
	})
}

// subscribe topics "test1,test2"
func WithTopicList(topics string) Option {
	return newFuncServerOption(func(o *ConsumerGroupOptions) {
		if len(topics) == 0 {
			panic("empty topics")
		}
		o.topicList = strings.Split(topics, ",")
	})
}

// consumer group partition assignment strategy (range, roundrobin, sticky)
func WithReBalanceStrategy(reBalanceStrategy string) Option {
	return newFuncServerOption(func(o *ConsumerGroupOptions) {
		if len(reBalanceStrategy) == 0 {
			return
		}
		o.reBalanceStrategy = reBalanceStrategy
	})
}

// initial offset to consumer (oldest, newest)
func WithInitialOffset(initialOffset string) Option {
	return newFuncServerOption(func(o *ConsumerGroupOptions) {
		if len(initialOffset) == 0 {
			return
		}
		o.initialOffset = initialOffset
	})
}

func getConsumerOptions(authOpts []auth.Option, opts ...Option) *ConsumerGroupOptions {
	ConsumerGroupOptions := &ConsumerGroupOptions{
		version:           "0.8.2.0",
		groupId:           "",
		brokerList:        nil,
		topicList:         nil,
		reBalanceStrategy: "sticky",
		initialOffset:     "newest",
		AuthOptions:       auth.GetAuthOptions(authOpts...),
	}

	for _, o := range opts {
		o.apply(ConsumerGroupOptions)
	}
	return ConsumerGroupOptions
}
