package encoder

import (
	"fmt"
	"google.golang.org/protobuf/proto"
)

const ProtoEncoder = "proto"

var ErrNotProtoMessage = fmt.Errorf("need proto message")

type protoEncoder struct{}

func NewProtoEncoder() *protoEncoder {
	return &protoEncoder{}
}

func (*protoEncoder) Marshal(v any) ([]byte, error) {
	m, ok := v.(proto.Message)
	if !ok {
		return nil, ErrNotProtoMessage
	}
	return proto.Marshal(m)
}

func (*protoEncoder) Unmarshal(bs []byte, v any) error {
	m, ok := v.(proto.Message)
	if !ok {
		return ErrNotProtoMessage
	}
	return proto.Unmarshal(bs, m)
}

func (*protoEncoder) Name() string {
	return ProtoEncoder
}
