package xlog

import (
	"context"
	"testing"
)

func TestInitXlog(t *testing.T) {
	const v = 1
	ctx := context.WithValue(context.TODO(), v, "123")
	InitXlog(WithContextWrite(func(ctx context.Context) []*Entry {
		fields := make([]*Entry, 0)
		s := ctx.Value(v)
		fields = append(fields, Field("trace", s))
		return fields
	}))
	Debug(ctx, "test debug message", Field("string", "hello"))
	Info(ctx, "test info message", Field("string", "hello"))
	SetLevel("error")
	Error(ctx, "test error message", Field("string", "hello"))
	Warn(ctx, "test warn message", Field("string", "hello"))
	Fatal(ctx, "test fatal message", Field("string", "hello"))
}

func TestRecover(t *testing.T) {
	defer Recover(context.TODO())

	var a, b int
	a = 10
	t.Log(a / b)
}
