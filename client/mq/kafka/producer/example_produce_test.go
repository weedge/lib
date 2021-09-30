package producer

func ExampleProducer_Ops() {
	p := NewProducer("test", "sync",
		WithBrokerList("127.0.0.1:9092,127.0.0.1:9093,127.0.0.1:9094"),
		WithRequiredAcks(-1),
		WithRetryMaxCn(3),
		WithCompression(""),
		WithPartitioning(""),
	)
	p.Send("hi")

	p.Close()
	// Output:
	//
}
