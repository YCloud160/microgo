package tracer

import (
	"encoding/json"
	"testing"
)

func TestNewSpan(t *testing.T) {
	span := NewRootSpan("test1")
	span = NewSpan(span, "test2")
	span = NewSpan(span, "test3")
	t.Log(span.SpanID())
	bs, _ := json.Marshal(span)
	t.Log(string(bs))
}
