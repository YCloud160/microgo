package encoder

import "encoding/json"

const JsonEncoder = "json"

type jsonEncoder struct{}

func NewJsonEncoder() *jsonEncoder {
	return &jsonEncoder{}
}

func (*jsonEncoder) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (*jsonEncoder) Unmarshal(bs []byte, v any) error {
	return json.Unmarshal(bs, v)
}

func (*jsonEncoder) Name() string {
	return JsonEncoder
}
