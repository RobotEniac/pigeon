package main

import (
	"flag"
	"fmt"
	"google.golang.org/protobuf/proto"
	"log"
	"math/rand"
	"net"
	"udp_server/rpc_message"
	"udp_server/rpc_message/datacloak/server"
)

func main() {
	ip := flag.String("h", "127.0.0.1", "server ip")
	port := flag.Int("p", 9300, "server port")
	flag.Parse()
	if !flag.Parsed() {
		log.Fatalf("flag parse failed")
	}
	remote := net.UDPAddr{
		IP:   net.ParseIP(*ip),
		Port: *port,
	}
	conn, err := net.DialUDP("udp", nil, &remote)
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	defer conn.Close()
	req := server.HelloRequest{
		Name: "wuhaibo",
		Age:  18,
	}
	reqBuf, err := proto.Marshal(&req)
	if err != nil {
		log.Fatalf("marshal failed: %s", err)
	}
	bufs, err := rpc_message.Serialize(
		reqBuf, string(proto.MessageName(&req)), uint32(rand.Int31()), 0, server.RpcMeta_REQUEST)
	if err != nil {
		log.Fatalf("serialize failed: %s", err)
	}

	for _, buf := range bufs {
		wn, err := conn.Write(buf)
		if err != nil {
			log.Fatalf("write failed: %s\n", err)
		}
		log.Printf("write %d bytes\n", wn)
	}
	rdBuf := make([]byte, 65535)
	rn, addr, err := conn.ReadFromUDP(rdBuf)
	if err != nil {
		log.Fatalf("ReadFromUdp failed: %s", err)
	}
	log.Println("from:", addr.String())
	header, err := rpc_message.ParseRpcMessageHeader(rdBuf[:rn])
	if err != nil {
		log.Fatalf("ParseRpcMessageHeader failed: %s", err)
	}
	log.Println("header:", header)
	meta, err := rpc_message.ParseRpcMeta(rdBuf[rpc_message.HeaderLength:rn])
	if err != nil {
		log.Fatalf("ParseRpcMeta failed")
	}
	log.Println("meta:", meta.String())
	resp := server.HelloResponse{}
	err = rpc_message.ParseRpcMessage(meta, &resp)
	if err != nil {
		log.Fatalf("ParseRpcMessage failed: %s", err)
	}
	log.Println("response:", resp.String())
}
