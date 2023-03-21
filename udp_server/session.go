package udp_server

import (
	"net"
	"udp_server/rpc_message"
	"udp_server/udp_error"
)

type Session struct {
	Endpoint   net.UDPAddr
	CurrentSid uint32
	conn       *net.UDPConn
	mq         messageQueue
	recvPkts   []rpc_message.Packet
}

func (s *Session) Recv(pkt rpc_message.Packet) error {
	if s.recvPkts == nil {
		s.CurrentSid = pkt.Sid
		s.recvPkts = append(s.recvPkts, pkt)
		return nil
	}
	if pkt.Sid != s.CurrentSid {
		return udp_error.SessionNotMatch{}
	}
	var i = 0
	for i < len(s.recvPkts) && s.recvPkts[i].Seq <= pkt.Seq {
		if s.recvPkts[i].Seq == pkt.Seq {
			return nil
		}
		i++
	}
	s.recvPkts = append(s.recvPkts, pkt)
	for j := len(s.recvPkts) - 2; j >= i; j-- {
		s.recvPkts[j+1] = s.recvPkts[j]
	}
	s.recvPkts[i] = pkt
	return nil
}

func (s *Session) RecvDone() bool {
	total := 0
	expectTotal := 0
	for i, p := range s.recvPkts {
		if p.Seq != uint16(i+1) {
			return false
		}
		total += p.Length - rpc_message.HeaderLength
		expectTotal = p.Total
	}
	return total == expectTotal
}

func (s *Session) RecvBuf() []byte {
	var buf []byte
	for i, p := range s.recvPkts {
		if p.Seq != uint16(i+1) {
			return nil
		}
		buf = append(buf, p.Msg[rpc_message.HeaderLength:]...)
	}
	return buf
}
