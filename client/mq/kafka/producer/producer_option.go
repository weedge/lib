package producer

import (
	"regexp"
	"strings"
)

type ProducerOptions struct {
	brokerList       []string
	partitioning     string // key {partition}(manual),hash,random
	requiredAcks     int    // required ack
	retryMaxCn       int    // retry max cn
	compression      string // msg compression(gzip,snappy,lz4,zstd)
	flushFrequencyMs int    // flush batches frequency
	certFile         string // the optional certificate file for client authentication
	keyFile          string // the optional key file for client authentication
	caFile           string // the optional certificate authority file for TLS client authentication
	verifySSL        bool   // the optional verify ssl certificates chain
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

func WithCertFile(certFile string) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		o.certFile = certFile
	})
}

func WithKeyFile(keyFile string) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		o.keyFile = keyFile
	})
}

func WithCaFile(caFile string) Option {
	return newFuncServerOption(func(o *ProducerOptions) {
		o.caFile = caFile
	})
}

func getProducerOptions(opts ...Option) *ProducerOptions {
	ConsumerGroupOptions := &ProducerOptions{
		brokerList:       nil,
		requiredAcks:     -1,
		retryMaxCn:       3,
		compression:      "",
		flushFrequencyMs: 0,
		certFile:         "",
		keyFile:          "",
		caFile:           "",
		verifySSL:        false,
	}

	for _, o := range opts {
		o.apply(ConsumerGroupOptions)
	}

	return ConsumerGroupOptions
}
