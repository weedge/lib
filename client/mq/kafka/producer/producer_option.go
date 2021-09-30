package producer

import (
	"strings"
)

type ProducerOptions struct {
	brokerList       []string
	requiredAcks     int    //required ack
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
		if flushFrequencyMs <= 0 {
			flushFrequencyMs = 300
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
		flushFrequencyMs: 300,
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
