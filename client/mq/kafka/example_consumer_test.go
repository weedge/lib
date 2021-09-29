package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/weedge/lib/log"
	"github.com/weedge/lib/strings"
	"os"
	"os/signal"
	"syscall"
)

type testMsg struct {
}

func (m *testMsg) Do(msg *sarama.ConsumerMessage) (err error) {
	log.Info(fmt.Sprintf("msg: %+v", msg))

	if strings.BytesToString(msg.Value) == "error" {
		err = fmt.Errorf("msg:%v is err", msg)
		return
	}

	return
}

func ExampleConsumerGroup_Ops() {
	cg, err := NewConsumerGroup("consumer.group.test", &testMsg{},
		WithVersion("2.8.0"), //kafka version
		WithBrokerList("127.0.0.1:9092,127.0.0.1:9093,127.0.0.1:9094"),//兼容kraft mode
		WithGroupId("consumer.group.test"),
		WithTopicList("foo"),
		WithInitialOffset("oldest"),
		WithReBalanceStrategy("sticky"),
	)
	if err != nil {
		println(err)
	}
	cg.Start()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigterm:
		println("terminating: via signal")
	}

	cg.Close()

	// output:
	//
}
