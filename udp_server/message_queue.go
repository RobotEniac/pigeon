package udp_server

import (
	"log"
	"net"
	"udp_server/rpc_message"
)

type messageQueue chan rpc_message.Packet

func (mq messageQueue) enqueue(addr *net.UDPAddr, msg []byte) {
	pkt, err := rpc_message.ParsePacket(addr, msg)
	if err != nil {
		log.Println("parse packet failed:", err)
		return
	}
	mq <- pkt
}
