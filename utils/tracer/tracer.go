package tracer

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"strconv"
	"time"
)

var rd = rand.New(rand.NewSource(time.Now().UnixNano()))

type tracerKey struct{}

var _tracerKey = tracerKey{}

type Tracer struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	CreateAt time.Time `json:"createAt"`
	spanId   string
	Span     *Span `json:"span"`
}

func (t *Tracer) clone() *Tracer {
	nt := &Tracer{
		Id:       t.Id,
		Name:     t.Name,
		CreateAt: t.CreateAt,
	}
	if t.Span != nil {
		nt.spanId = t.spanId
		nt.Span = t.Span.clone()
	}
	return nt
}

func (t *Tracer) TraceID() string {
	if t == nil {
		return ""
	}
	return t.Id
}

func (t *Tracer) SpanID() string {
	if t == nil {
		return ""
	}
	return t.spanId
}

func (t *Tracer) String() string {
	if t == nil {
		return ""
	}
	bs, _ := json.Marshal(t)
	return string(bs)
}

func WithTracer(ctx context.Context, trace *Tracer, name string) context.Context {
	trace.Span = NewSpan(trace.Span, name)
	trace.spanId = trace.Span.SpanID()
	return context.WithValue(ctx, _tracerKey, trace)
}

func WithNewTracer(ctx context.Context, name string) (context.Context, *Tracer) {
	val := ctx.Value(_tracerKey)
	trace, ok := val.(*Tracer)
	if !ok {
		trace = newTracer(name)
	} else {
		trace = trace.clone()
		trace.Span = NewSpan(trace.Span, name)
	}
	trace.spanId = trace.Span.SpanID()
	return context.WithValue(ctx, _tracerKey, trace), trace
}

func ParseTrace(s string) *Tracer {
	trace := &Tracer{}
	if err := json.Unmarshal([]byte(s), trace); err != nil {
		return nil
	}
	return trace
}

func newTracer(name string) *Tracer {
	trace := &Tracer{
		Id:       genTraceId(name),
		Name:     name,
		CreateAt: time.Now(),
		Span:     NewRootSpan(name),
	}
	return trace
}

func genTraceId(name string) string {
	hash := md5.New()
	hash.Write([]byte(strconv.FormatInt(rd.Int63n(10000000000), 10)))
	hash.Write([]byte(name))
	hash.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	traceId := hash.Sum(nil)
	return hex.EncodeToString(traceId)
}
