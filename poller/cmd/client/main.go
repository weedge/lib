package main

import (
	"log"
	"net"
	"strconv"

	"github.com/weedge/lib/poller/cmd/common"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	go start()
	select {}
}

func start() {
	log.Println("start")
	conn, err := net.Dial("tcp", "127.0.0.1:8085")
	if err != nil {
		log.Println("Error dialing", err.Error())
		return // 终止程序
	}

	codec := common.NewCodec(conn)

	go func() {
		for {
			_, err = codec.Read()
			if err != nil {
				log.Println(err)
				return
			}
			for {
				bytes, ok, err := codec.Decode()
				// 解码出错，需要中断连接
				if err != nil {
					log.Println(err)
					return
				}
				if ok {
					log.Println("receive:", string(bytes))
					continue
				}
				break
			}
		}
	}()

	for i := 0; i < 1000; i++ {
		msg := []byte("hello" + strconv.Itoa(i))
		_, err := conn.Write(common.Encode(msg))
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("send:", string(msg))
	}
}
