package microgo

import "github.com/YCloud160/microgo/utils/encoder"

type Encoder interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(bs []byte, v any) error
	Name() string
}

var encodes = make(map[string]Encoder)

func init() {
	RegisterEncoder(encoder.NewJsonEncoder())
	RegisterEncoder(encoder.NewProtoEncoder())
}

func RegisterEncoder(enc Encoder) {
	if enc == nil {
		return
	}
	encodes[enc.Name()] = enc
}

func GetEncoder(name string) Encoder {
	if enc, ok := encodes[name]; ok {
		return enc
	}
	return encodes[encoder.JsonEncoder]
}
