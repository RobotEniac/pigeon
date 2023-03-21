package rpc_message

import (
	"google.golang.org/protobuf/proto"
	"testing"
	"udp_server/rpc_message/datacloak/server"
)

func TestSerialize(t *testing.T) {
	resp := server.HelloResponse{
		Name: "wuhaibo",
		Ppt:  123456,
	}
	b, err := proto.Marshal(&resp)
	if err != nil {
		t.Fatalf("marshal %s error: %s", resp.String(), err)
	}
	res, err := Serialize(b, string(resp.ProtoReflect().Descriptor().FullName()), 0, 0, server.RpcMeta_RESPONSE)
	if err != nil {
		t.Fatalf("serialize failed: %s", err)
	}
	header, err := ParseRpcMessageHeader(res)
	if err != nil {
		t.Fatalf("parse header failed: %s", err)
	}
	t.Log(header)
	t.Log(len(res))
	t.Log(header.Length - uint32(HeaderLength))
	meta, err := ParseRpcMeta(res[HeaderLength:header.Length])
	if err != nil {
		t.Fatalf("parse meta failed:%s", err)
	}
	t.Log(meta.String())
}
