package producer

import (
	"regexp"
	"strings"
	"time"

	"github.com/weedge/lib/client/mq/kafka/auth"
)

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

type Option interface {
	apply(o *ProducerOptions)
}

type funcServerOption struct {
	f func(o *ProducerOptions)
}

func (fdo *funcServerOption) apply(o *ProducerOptions) {
	fdo.f(o)
}

func newFuncServerOption(f func(o *ProducerOptions)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

// kafka version support kafka min version 0.8.2.0
func WithVersion(version string) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		if len(version) == 0 {
			version = "0.8.2.0"
		}
		o.version = version
	})
}

// The client ID sent with every request to the brokers.
func WithClientID(clientId string) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		o.clientID = clientId
	})
}

// The duration the producer will wait to receive -required-acks
func WithTimeOut(timeOut time.Duration) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		o.timeOut = timeOut
	})
}

func WithChannelBufferSize(channelBufferSize int) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		o.channelBufferSize = channelBufferSize
	})
}

func WithMaxOpenRequests(maxOpenRequests int) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		o.maxOpenRequests = maxOpenRequests
	})
}

func WithMaxMessageBytes(maxMessageBytes int) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		o.maxMessageBytes = maxMessageBytes
	})
}

// kafka brokers ip:port,ip:port
func WithBrokerList(brokers string) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		if len(brokers) == 0 {
			panic("empty brokers")
		}
		o.brokerList = strings.Split(brokers, ",")
	})
}

// key partition: partition(manual),hash,random
// - manual partitioning if a partition number is provided
// - hash partitioning by msg key
// - random partitioning otherwise.
func WithPartitioning(partitioning string) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		o.partitioning = "hash"
		match, _ := regexp.MatchString(`^[0-9]+$|^hash$|^random$|^roundrobin$`, partitioning)
		if match {
			o.partitioning = partitioning
		}
	})
}

// RequiredAcks is used in Produce Requests to tell the broker how many replica acknowledgements
// it must see before responding. Any of the constants defined here are valid. On broker versions
// prior to 0.8.2.0 any other positive int16 is also valid (the broker will wait for that many
// acknowledgements) but in 0.8.2.0 and later this will raise an exception (it has been replaced
// by setting the `min.isr` value in the brokers configuration).
// 0: NoResponse doesn't send any response, the TCP ACK is all you get.
// 1: WaitForLocal waits for only the local commit to succeed before responding.
// -1: WaitForAll waits for all in-sync replicas to commit before responding.
// The minimum number of in-sync replicas is configured on the broker via
// the `min.insync.replicas` configuration key.
func WithRequiredAcks(requiredAcks int) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		if requiredAcks < -1 {
			requiredAcks = -1
		}
		if requiredAcks > 1 {
			requiredAcks = 1
		}
		o.requiredAcks = requiredAcks
	})
}

func WithRetryMaxCn(retryMaxCn int) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		if retryMaxCn <= 0 {
			retryMaxCn = 3
		}
		o.retryMaxCn = retryMaxCn
	})
}

func WithCompression(compression string) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		if _, ok := compressions[compression]; !ok {
			panic("un support this compression" + compression)
		}
		o.compression = compression
	})
}

func WithFlushFrequencyMs(flushFrequencyMs int) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		if flushFrequencyMs < 0 {
			flushFrequencyMs = 0
		}

		o.flushFrequencyMs = flushFrequencyMs
	})
}

func WithFlushBytes(flushBytes int) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		if flushBytes < 0 {
			flushBytes = 0
		}

		o.flushBytes = flushBytes
	})
}

func WithFlushMessages(flushMessages int) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		if flushMessages < 0 {
			flushMessages = 0
		}

		o.flushMessages = flushMessages
	})
}

func WithFlushMaxMessages(flushMaxMessages int) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		if flushMaxMessages < 0 {
			flushMaxMessages = 0
		}

		o.flushMaxMessages = flushMaxMessages
	})
}

func getProducerOptions(authOpts []auth.Option, opts ...Option) *ProducerOptions {
	producerOptions := &ProducerOptions{
		version:           "0.8.2.0",
		clientID:          "",
		brokerList:        nil,
		partitioning:      "",
		requiredAcks:      -1,
		timeOut:           5 * time.Second,
		retryMaxCn:        3,
		compression:       "",
		maxOpenRequests:   5,
		maxMessageBytes:   1000000,
		channelBufferSize: 256,
		flushFrequencyMs:  0,
		flushBytes:        0,
		flushMessages:     0,
		flushMaxMessages:  0,
		AuthOptions:       auth.GetAuthOptions(authOpts...),
	}

	for _, o := range opts {
		o.apply(producerOptions)
	}

	return producerOptions
}
