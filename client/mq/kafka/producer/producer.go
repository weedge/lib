package producer

import (
	"fmt"
	"github.com/weedge/lib/client/mq/kafka/auth"
	"github.com/weedge/lib/runtimer"
	"strconv"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/weedge/lib/log"
)

// @todo: add metrics write to log; add tapper interceptor;

type Producer struct {
	pType         string //sync(default),async
	topic         string
	partition     int32
	config        *sarama.Config
	asyncProducer sarama.AsyncProducer
	syncProducer  sarama.SyncProducer
	wg            *sync.WaitGroup
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

	partitioning = map[string]sarama.PartitionerConstructor{
		"manual":     sarama.NewManualPartitioner,
		"hash":       sarama.NewHashPartitioner,
		"random":     sarama.NewRandomPartitioner,
		"roundrobin": sarama.NewRoundRobinPartitioner,
		//"referencehash": sarama.NewReferenceHashPartitioner,
	}
)

// new sync/async producer to topic with option(requiredAcks,retryMaxCn,partitioning,compressions,TLS...etc)
func NewProducer(topic string, pType string, authOpts []auth.Option, options ...Option) (p *Producer) {
	p = &Producer{
		topic: topic,
		pType: pType,
		//wg: &sync.WaitGroup{},
	}
	p.config = sarama.NewConfig()

	opts := getProducerOptions(authOpts, options...)
	log.Info(fmt.Sprintf("producer options:%+v", opts))

	var err error
	p.config.Version, err = sarama.ParseKafkaVersion(opts.version)
	if err != nil {
		return
	}

	p.config.ClientID = opts.clientID
	p.config.Producer.RequiredAcks = requiredAcks[opts.requiredAcks]
	p.config.Producer.Retry.Max = opts.retryMaxCn
	p.config.Producer.Compression = compressions[opts.compression]                              // Compress messages
	p.config.Producer.Flush.Frequency = time.Duration(opts.flushFrequencyMs) * time.Millisecond // Flush batches every
	p.config.Producer.Return.Successes = true

	partition, err := strconv.ParseInt(opts.partitioning, 10, 64)
	if err != nil {
		p.config.Producer.Partitioner = partitioning[opts.partitioning]
	} else {
		p.config.Producer.Partitioner = partitioning[opts.partitioning]
		p.config.Producer.Partitioner = partitioning["manual"]
		p.partition = int32(partition)
	}

	opts.AuthOptions.InitSSL(p.config)

	opts.AuthOptions.InitSASLSCRAM(p.config)

	switch p.pType {
	case "sync":
		err = p.initSyncProducer(opts)
	case "async":
		err = p.initAsyncProducer(opts)
	default:
		err = p.initSyncProducer(opts)
	}
	if err != nil {
		log.Error("init producer err:", err.Error())
	}

	return
}

// send string msg no key
func (p *Producer) Send(val string) {
	p.send("", val)
}

// send string msg by string key
func (p *Producer) SendByKey(key, val string) {
	p.send(key, val)
}

func (p *Producer) send(key string, val string) {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(val),
	}

	if p.partition > 0 {
		msg.Partition = p.partition
	}

	if len(key) > 0 {
		msg.Key = sarama.StringEncoder(key)
	}

	if p.syncProducer != nil {
		partition, offset, err := p.syncProducer.SendMessage(msg)
		if err != nil {
			log.Error("syncProducer.SendMessage msg err:", err.Error())
		}
		log.Info(fmt.Sprintf("syncProducer.SendMessage success,topic:%s,partition:%d,offset:%d,val:%s", p.topic, partition, offset, msg.Value))
	}

	if p.asyncProducer != nil {
		p.asyncProducer.Input() <- msg
	}
}

// close sync/async producer
func (p *Producer) Close() {
	if p.wg != nil {
		p.wg.Wait()
	}
	if p.syncProducer != nil {
		if err := p.syncProducer.Close(); err != nil {
			log.Error("Failed to close sync producer cleanly:", err)
			return
		}
		log.Info("Success to close sync producer cleanly")
	}
	if p.asyncProducer != nil {
		if err := p.asyncProducer.Close(); err != nil {
			log.Error("Failed to close async producer cleanly:", err)
			return
		}
		log.Info("Success to close async producer cleanly")
	}
}

// On the broker side, you may want to change the following settings to get
// stronger consistency guarantees:
// - For your broker, set `unclean.leader.election.enable` to false
// - For the topic, you could increase `min.insync.replicas`.
func (p *Producer) initSyncProducer(opts *ProducerOptions) (err error) {
	p.syncProducer, err = sarama.NewSyncProducer(opts.brokerList, p.config)
	if err != nil {
		log.Error("Failed to start Sarama sync producer:", err)
		return
	}

	return
}

func (p *Producer) initAsyncProducer(opts *ProducerOptions) (err error) {
	p.asyncProducer, err = sarama.NewAsyncProducer(opts.brokerList, p.config)
	if err != nil {
		log.Error("Failed to start Sarama async producer:", err)
	}

	// We will just log to STDOUT if we're not able to produce messages.
	// Note: messages will only be returned here after all retry attempts are exhausted.
	runtimer.GoSafely(p.wg, false, func() {
		//go func() {
		for err := range p.asyncProducer.Errors() {
			log.Error("Failed to async produce msg error:", err.Error())
		}
		//}()
	}, nil, nil)

	runtimer.GoSafely(p.wg, false, func() {
		//go func() {
		for msg := range p.asyncProducer.Successes() {
			log.Info(fmt.Sprintf("asyncProducer.SendMessage success,topic:%s,partition:%d,offset:%d,val:%s", msg.Topic, msg.Partition, msg.Offset, msg.Value))
		}
		//}()
	}, nil, nil)

	return
}
