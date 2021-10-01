package producer

import (
	"fmt"
	"sync"
	"time"
)

func ExampleProducer_SyncProducerOps() {
	p := NewProducer("test", "sync",
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
