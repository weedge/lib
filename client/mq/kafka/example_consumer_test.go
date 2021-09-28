package kafka

import (
	"fmt"

	"github.com/weedge/lib/strings"

	"github.com/Shopify/sarama"
)

type testMsg struct {
}

func (m *testMsg) Do(msg *sarama.ConsumerMessage) (err error) {
	fmt.Println("msg", msg)

	if strings.BytesToString(msg.Value) == "error" {
		err = fmt.Errorf("msg:%v err:%s", msg, err.Error())
		return
	}

	return
}

func ExampleConsumerGroup_Ops() {
	cg, err := NewConsumerGroup("consumer.group.test", &testMsg{},
		WithVersion("2.8.0"), //kafka version
		WithBrokerList("127.0.0.1:9092"),
		WithGroupId("consumer.group.test"),
		WithTopicList("sarama"),
		WithInitialOffset("newest"),
		WithReBalanceStrategy("sticky"),
	)
	if err != nil {
		println(err)
	}
	cg.Start()

	cg.Close()

	// output:
	//
}
