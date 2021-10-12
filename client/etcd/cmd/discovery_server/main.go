package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/weedge/lib/client/etcd/discovery"
	"google.golang.org/grpc/keepalive"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/weedge/lib/client/etcd/cmd/pb"
	"github.com/weedge/lib/log"
)

const defaultService = "test-service"

type helloService struct {
	pb.UnimplementedHelloServer
}

func (h *helloService) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoRequest, error) {
	fmt.Println(req.Message)
	return req, nil
}

func (h *helloService) Stream(req *pb.EchoRequest, srv pb.Hello_StreamServer) error {
	fmt.Println("stream", req.Message)
	i := 1
	for i < 5 {
		if srv.Context().Err() != nil {
			return srv.Context().Err()
		}

		err := srv.Send(req)
		if err != nil {
			return err
		}
		i++
		time.Sleep(time.Millisecond * 500)
	}
	return nil
}

var kaep = keepalive.EnforcementPolicy{
	MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
	PermitWithoutStream: true,            // Allow pings even when there are no active streams
}

var kasp = keepalive.ServerParameters{
	MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
	MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
	MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
	Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
	Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
}

func main() {
	port := flag.String("port", ":8080", "listen port")
	service := flag.String("service", defaultService, "service name")
	flag.Parse()

	r := discovery.NewRegister([]string{"0.0.0.0:2379"}, 5*time.Second, *service, "localhost"+*port)
	defer r.Close()

	err := r.Do()
	if err != nil {
		log.Error(err.Error())
	}

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp))

	pb.RegisterHelloServer(s, &helloService{})
	if err := s.Serve(lis); err != nil {
		log.Errorf("failed to serve: %v", err)
	}

}
