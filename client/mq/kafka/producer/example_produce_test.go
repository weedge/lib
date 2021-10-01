package producer

import (
	"fmt"
	"github.com/Shopify/sarama"
	"strings"
	"sync"
	"time"
)

func ExampleSend() {
	cf := sarama.NewConfig()
	cf.Producer.RequiredAcks = sarama.WaitForAll
	cf.Producer.Return.Successes = true
	//cf.Producer.Flush.Frequency = 300 * time.Millisecond

	blist := strings.Split("127.0.0.1:9092,127.0.0.1:9093,127.0.0.1:9094", ",")
	p, err := sarama.NewSyncProducer(blist, cf)
	if err != nil {
		println("new err", err.Error())
		return
	}

	for i := 0; i < 500; i++ {
		partition, offset, err := p.SendMessage(&sarama.ProducerMessage{
			Topic: "sarama",
			Value: sarama.StringEncoder(fmt.Sprintf("hi:%d", i)),
		})
		if err != nil {
			println("send err", err.Error())
			return
		}
		println(partition, offset)
	}

	err = p.Close()
	if err != nil {
		println("close err", err.Error())
		return
	}

	// Output:
	//
}

func ExampleProducer_SyncProducerOps() {
	p := NewProducer("sarama", "sync",
		WithBrokerList("127.0.0.1:9092,127.0.0.1:9093,127.0.0.1:9094"),
		WithRequiredAcks(-1),
		WithRetryMaxCn(3),
		WithCompression(""),
		WithPartitioning(""),
	)

	t := time.NewTimer(3 * time.Second)
	cn := 0
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-t.C:
				println("send over, send totalCn:", cn)
				return
			default:
				p.Send(fmt.Sprintf("hi:%d", cn))
				cn++
				//time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	wg.Wait()

	p.Close()
	// Output:
	//
}

func ExampleProducer_ASyncProducerOps() {
	p := NewProducer("test", "async",
		WithBrokerList("127.0.0.1:9092,127.0.0.1:9093,127.0.0.1:9094"),
		WithRequiredAcks(-1),
		WithRetryMaxCn(3),
		WithCompression(""),
		WithPartitioning(""),
	)

	t := time.NewTimer(3 * time.Second)
	cn := 0
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-t.C:
				println("send over, send totalCn:", cn)
				return
			default:
				p.Send(fmt.Sprintf("hi:%d", cn+100))
				cn++
				//time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	wg.Wait()

	p.Close()
	// Output:
	//
}
