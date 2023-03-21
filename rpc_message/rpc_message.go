package rpc_message

import (
	"encoding/binary"
	"errors"
	"google.golang.org/protobuf/proto"
	"udp_server/rpc_message/datacloak/server"
	"udp_server/udp_error"
)

type RpcMessageHeader struct {
	Magic     uint32
	Length    uint32
	SessionId uint32
	Seq       uint16
}

const (
	HeaderLength int    = 14
	MagicNumber  uint32 = 0x00424857 // WHB\0
	UDPMTU       int    = 1000
)
const (
	FLAG_MORE uint16 = 1
)

func (h *RpcMessageHeader) ToBytes() []byte {
	buf := make([]byte, HeaderLength)
	binary.LittleEndian.PutUint32(buf[0:4], MagicNumber)
	binary.LittleEndian.PutUint32(buf[4:8], h.Length)
	binary.LittleEndian.PutUint32(buf[8:12], h.SessionId)
	binary.LittleEndian.PutUint16(buf[12:14], h.Seq)
	return buf
}

func ParseRpcMessageHeader(buf []byte) (RpcMessageHeader, error) {
	var ret RpcMessageHeader
	if len(buf) < HeaderLength {
		return ret, udp_error.InvalidLength{}
	}
	magic := binary.LittleEndian.Uint32(buf[0:4])
	if magic != MagicNumber {
		return ret, udp_error.InvalidMagicNumber{}
	}
	ret.Magic = magic
	ret.Length = binary.LittleEndian.Uint32(buf[4:8])
	ret.SessionId = binary.LittleEndian.Uint32(buf[8:12])
	ret.Seq = binary.LittleEndian.Uint16(buf[12:14])
	return ret, nil
}

func Serialize(msg []byte, name string, sid, seq uint32, tp server.RpcMeta_Type) ([]byte, error) {
	meta := server.RpcMeta{
		Type:       tp,
		Method:     name,
		SequenceId: int64(seq),
		Length:     int32(len(msg)),
	}
	meta.Payload = msg
	return SerializeRpcMeta(sid, &meta)
}

func SerializeRpcMeta(sid uint32, meta *server.RpcMeta) ([]byte, error) {
	var ret []byte
	metaByte, err := proto.Marshal(meta)
	if err != nil {
		return nil, errors.New("marshal rpc meta failed")
	}
	var seq uint16 = 0
	header := RpcMessageHeader{
		Magic:     MagicNumber,
		Length:    uint32(HeaderLength + len(metaByte)),
		SessionId: sid,
		Seq:       seq,
	}
	buf := header.ToBytes()
	buf = append(buf, metaByte...)
	return ret, nil
}

func ParseRpcMeta(buf []byte) (*server.RpcMeta, error) {
	var ret server.RpcMeta
	err := proto.Unmarshal(buf, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func ParseRpcMessage(meta *server.RpcMeta, msg proto.Message) error {
	return proto.Unmarshal(meta.Payload, msg)
}
