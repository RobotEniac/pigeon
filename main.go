package main

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"os/signal"
	"udp_server/rpc_message/datacloak/server"
	"udp_server/udp_server"
)

type HelloMethod struct{}

func (m *HelloMethod) RequestMessage() proto.Message {
	ret := server.HelloRequest{}
	return &ret
}

func (m *HelloMethod) ResponseMessage() proto.Message {
	ret := server.HelloResponse{}
	return &ret
}

func (m *HelloMethod) CallMethod(name string, req proto.Message) (proto.Message, error) {
	hReq := req.(*server.HelloRequest)
	fmt.Printf("name = %s, age = %d\n", hReq.Name, hReq.Age)
	resp := server.HelloResponse{
		Name: "whb",
		Ppt:  12354,
	}
	return &resp, nil
}

func main() {
	a := HelloMethod{}
	req := server.HelloRequest{
		Name: "whb",
		Age:  18,
	}
	ss := udp_server.NewUdpServer("", 9300)
	if ss == nil {
		log.Fatalf("invalid udp server")
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ss.Register(string(req.ProtoReflect().Descriptor().FullName()), &a)
	go ss.Serve()
	<-c
	ss.Stop()
	log.Printf("quiting...")
}
