package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/weedge/lib/client/etcd/cmd/pb"
	"github.com/weedge/lib/client/etcd/discovery"
	"github.com/weedge/lib/log"

	"google.golang.org/grpc"
)

const defaultService = "test-service"

func testService(service string, stream bool) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:///%s", discovery.Scheme, service), grpc.WithBlock(), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`), grpc.WithInsecure())
	if err != nil {
		log.Errorf("did not connect: %v", err)
		return
	}
	defer conn.Close()
	c := pb.NewHelloClient(conn)

	i := 1

	for {
		if !stream {
			resp, err := c.Echo(context.Background(), &pb.EchoRequest{Message: fmt.Sprintf("test-%d", i)})
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(resp.Message)
			}
		} else {
			s, err := c.Stream(context.Background(), &pb.EchoRequest{Message: fmt.Sprintf("test-%d", i)})
			if err != nil {
				fmt.Println(err)
			} else {
				for {
					msg, err := s.Recv()
					if err == io.EOF {
						break
					}

					if err != nil {
						fmt.Println(err)
						break
					}
					fmt.Println(msg.Message)
				}
			}
		}
		i++
		time.Sleep(time.Second)
	}
}

func main() {
	stream := flag.Bool("stream", false, "If test stream.")
	service := flag.String("service", defaultService, "service name")
	flag.Parse()

	r := discovery.NewGrpcResolver([]string{"0.0.0.0:2379"}, time.Second*5)
	err := r.InitGrpcResolver()
	if err != nil {
		log.Error(err.Error())
		return
	}

	tmpArr := strings.Split(*service, ",")
	for _, s := range tmpArr {
		s := s
		go testService(s, *stream)
	}

	for {
		time.Sleep(time.Second * 5)
		r.DebugStore()
	}
}
