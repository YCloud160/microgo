package xlog

import (
	"fmt"
	"testing"
)

type Data struct {
	Id   int
	Name string
}

func TestEncodeJson(t *testing.T) {
	msg := EncodeJson()
	t.Log(msg)
	msg = EncodeJson(
		Field("error", fmt.Errorf("this is error message")),
		Field("int", 12),
		Field("float", 3.14),
		Field("bool", true),
		Field("string", "hello"),
		Field("[]int", []int64{1, 2, 3}),
		Field("struct", Data{Id: 12, Name: "Jack"}),
	)
	t.Log(msg)
}
