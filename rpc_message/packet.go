package rpc_message

import "net"

type Packet struct {
	Addr   *net.UDPAddr
	Msg    []byte
	Length int
	Total  int
	Sid    uint32
	Seq    uint16
}

func ParsePacket(addr *net.UDPAddr, buf []byte) (Packet, error) {
	pkt := Packet{}
	header, err := ParseRpcMessageHeader(buf)
	if err != nil {
		return pkt, err
	}
	pkt.Addr = addr
	pkt.Msg = buf
	pkt.Length = int(header.Length)
	pkt.Total = int(header.Total)
	pkt.Sid = header.SessionId
	pkt.Seq = header.Seq
	return pkt, nil
}
