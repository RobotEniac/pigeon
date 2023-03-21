package udp_server

import (
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net"
	"sync"
	"time"
	"udp_server/rpc_message"
	"udp_server/rpc_message/datacloak/server"
	"udp_server/udp_error"
)

const (
	MessageQueueSize = 10240
	UDPPacketSize    = 65535
	UDPServerTimeout = 70 * time.Second
)

type UdpServer struct {
	conn        *net.UDPConn
	localAddr   net.UDPAddr
	sessionLock sync.Mutex
	sessions    map[string]*Session
	methods     map[string]RpcMethod
}

type RpcMethod interface {
	RequestMessage() proto.Message
	ResponseMessage() proto.Message
	CallMethod(name string, req proto.Message) (proto.Message, error)
}

func writeResponse(se *Session, sid, seq uint32, meta *server.RpcMeta) error {
	bufs, err := rpc_message.SerializeRpcMeta(sid, meta)
	if err != nil {
		log.Printf("serialize response failed: %s", err)
		return err
	}
	total := 0
	for _, buf := range bufs {
		for total < len(buf) {
			n, err := se.conn.WriteToUDP(buf[total:], &se.Endpoint)
			if err != nil {
				log.Printf("WriteToUDP failed: %s", err)
				return err
			}
			total += n
		}
	}
	return nil
}

func processRequest(meta *server.RpcMeta, mp map[string]RpcMethod) (proto.Message, error) {
	method := mp[meta.GetMethod()]
	req := method.RequestMessage().ProtoReflect().Interface()
	err := proto.Unmarshal(meta.Payload, req)
	if err != nil {
		return nil, err
	}
	return method.CallMethod(meta.Method, req)
}

func (s *UdpServer) handleRpcMeta(se *Session) error {
	var header rpc_message.RpcMessageHeader
	var buf []byte
	var err error
	buf = se.RecvBuf()
	if buf != nil {
		log.Printf("RecvBuf not complete")
		return udp_error.RecvNotComplete{}
	}
	meta, err := rpc_message.ParseRpcMeta(buf)
	if err != nil {
		log.Printf("ParseRpcMeta failed: %s", err)
		meta = &server.RpcMeta{
			Type:   server.RpcMeta_RESPONSE,
			Method: "",
			Error:  server.RpcMeta_InvalidMessage,
			Length: 0,
		}
		err = writeResponse(se, header.SessionId, 0, meta)
		if err != nil {
			log.Printf("write response failed: %s", err)
			return err
		}
	} else {
		resp, err := processRequest(meta, s.methods)
		if err != nil {
			log.Printf("processRequest failed: %s", err)
			meta = &server.RpcMeta{
				Type:   server.RpcMeta_RESPONSE,
				Method: "",
				Error:  server.RpcMeta_Unknown,
				Length: 0,
			}
			err = writeResponse(se, se.CurrentSid, 0, meta)
			if err != nil {
				log.Printf("write response failed: %s", err)
				return err
			}
		}
		respBuf, err := proto.Marshal(resp)
		if err != nil {
			log.Printf("marsh proto failed: %s", err)
			meta = &server.RpcMeta{
				Type:   server.RpcMeta_RESPONSE,
				Method: "",
				Error:  server.RpcMeta_Unknown,
				Length: 0,
			}
			err = writeResponse(se, se.CurrentSid, 0, meta)
			if err != nil {
				log.Printf("write response failed: %s", err)
				return err
			}
		}
		respMeta := &server.RpcMeta{
			Type:       server.RpcMeta_RESPONSE,
			Method:     meta.Method,
			Error:      0,
			Timeout:    meta.Timeout,
			SequenceId: meta.SequenceId,
			Length:     int32(len(respBuf)),
			Payload:    respBuf,
		}
		err = writeResponse(se, se.CurrentSid, 0, respMeta)
		if err != nil {
			log.Printf("write response error: %s", err)
			return err
		}
	}
	return nil
}

func (s *UdpServer) handlePacket(se *Session) {
	tk := time.NewTimer(UDPServerTimeout)
loop:
	for {
		select {
		case pkt, ok := <-se.mq:
			if !ok {
				continue
			}
			err := se.Recv(pkt)
			if err != nil {
				log.Printf("Recv failed: %s\n", err)
				continue
			}
			if se.RecvDone() {
				s.handlePacket(se)
			}
			tk.Reset(UDPServerTimeout)
		case <-tk.C:
			break loop
		}
	}
	s.sessionLock.Lock()
	delete(s.sessions, se.Endpoint.String())
}

func (s *UdpServer) receive() {
	for {
		msg := make([]byte, UDPPacketSize)
		n, addr, err := s.conn.ReadFromUDP(msg)
		if err != nil {
			log.Printf("read from udp failed: %s", err)
			if err == io.EOF {
				return
			}
			return
		}
		log.Printf("recv from: %s, %d bytes", addr.String(), n)
		s.sessionLock.Lock()
		session, ok := s.sessions[addr.String()]
		log.Printf("session size: %d", len(s.sessions))
		if !ok {
			se := &Session{
				Endpoint:   *addr,
				CurrentSid: 0,
				conn:       s.conn,
				mq:         make(messageQueue, MessageQueueSize),
			}
			s.sessions[addr.String()] = se
			s.sessionLock.Unlock()
			go s.handlePacket(se)
			se.mq.enqueue(addr, msg[:n])
		} else {
			s.sessionLock.Unlock()
			session.mq.enqueue(addr, msg[:n])
		}
	}
}

func NewUdpServer(ip string, port int) *UdpServer {
	if ip == "" {
		ip = "0.0.0.0"
	}
	return &UdpServer{
		conn: nil,
		localAddr: net.UDPAddr{
			IP:   net.ParseIP(ip),
			Port: port,
		},
		sessions: make(map[string]*Session),
		methods:  make(map[string]RpcMethod),
	}
}

func (s *UdpServer) Register(name string, method RpcMethod) {
	s.methods[name] = method
}

func (s *UdpServer) callMethod(name string, req proto.Message) (proto.Message, error) {
	return s.methods[name].CallMethod(name, req)
}

func (s *UdpServer) Serve() {
	conn, err := net.ListenUDP("udp4", &s.localAddr)
	if err != nil {
		panic(err)
	}
	s.conn = conn
	log.Printf("listening at %s\n", conn.LocalAddr().String())
	s.receive()
	log.Printf("Udp server[%s] stopped", s.localAddr.String())
}

func (s *UdpServer) Stop() {
	s.conn.Close()
}
