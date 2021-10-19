package dlock

import "time"

func ExampleMockTest() {
	go MockTest("A")
	go MockTest("B")
	go MockTest("C")
	go MockTest("D")
	go MockTest("E")

	// 用于测试goroutine接收到ctx.Done()信号后的打印
	time.Sleep(time.Second * 2)

	//output:
	//
}
