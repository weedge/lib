package producer

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/weedge/lib/log"

	"github.com/Shopify/sarama"
)

type Producer struct {
	config        *sarama.Config
	msg           *sarama.ProducerMessage
	asyncProducer sarama.AsyncProducer
	syncProducer  sarama.SyncProducer
}

var (
	compressions = map[string]sarama.CompressionCodec{
		"":       sarama.CompressionNone,
		"gzip":   sarama.CompressionGZIP,
		"snappy": sarama.CompressionSnappy,
		"lz4":    sarama.CompressionLZ4,
		"zstd":   sarama.CompressionZSTD,
	}
	requiredAcks = map[int]sarama.RequiredAcks{
		0:  sarama.NoResponse,
		1:  sarama.WaitForLocal,
		-1: sarama.WaitForAll,
	}
)

func NewProducer(options ...Option) (p *Producer) {
	p = &Producer{}

	opts := getProducerOptions(options...)
	log.Info(fmt.Sprintf("producer options:%+v", opts))

	p.config.Producer.RequiredAcks = requiredAcks[opts.requiredAcks]
	p.config.Producer.Retry.Max = opts.retryMaxCn
	p.config.Producer.Compression = compressions[opts.compression] // Compress messages

	tlsConfig := createTlsConfiguration(opts)
	if tlsConfig != nil {
		p.config.Net.TLS.Config = tlsConfig
		p.config.Net.TLS.Enable = true
	}

	return
}

// On the broker side, you may want to change the following settings to get
// stronger consistency guarantees:
// - For your broker, set `unclean.leader.election.enable` to false
// - For the topic, you could increase `min.insync.replicas`.
func (p *Producer) initSyncProducer(opts *ProducerOptions) (err error) {
	p.config.Producer.Return.Successes = true
	p.syncProducer, err = sarama.NewSyncProducer(opts.brokerList, p.config)
	if err != nil {
		log.Error("Failed to start Sarama sync producer:", err)
		return
	}

	return
}

func (p *Producer) initAsyncProducer(opts *ProducerOptions) (err error) {
	p.config.Producer.Flush.Frequency = time.Duration(opts.flushFrequencyMs) * time.Millisecond // Flush batches every

	producer, err := sarama.NewAsyncProducer(opts.brokerList, p.config)
	if err != nil {
		log.Error("Failed to start Sarama async producer:", err)
	}

	// We will just log to STDOUT if we're not able to produce messages.
	// Note: messages will only be returned here after all retry attempts are exhausted.
	go func() {
		for err := range producer.Errors() {
			log.Error("Failed to write access log entry:", err)
		}
	}()

	return
}

func createTlsConfiguration(opts *ProducerOptions) (t *tls.Config) {
	if opts.certFile == "" || opts.keyFile == "" || opts.caFile == "" {
		return
	}

	cert, err := tls.LoadX509KeyPair(opts.certFile, opts.keyFile)
	if err != nil {
		log.Error("tls.LoadX509KeyPair err", err)
		return
	}

	caCert, err := ioutil.ReadFile(opts.caFile)
	if err != nil {
		log.Error("ioutil.ReadFile err", err)
		return
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	t = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: opts.verifySSL,
	}

	return
}
