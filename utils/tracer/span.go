package tracer

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"time"
)

type Span struct {
	Id       string    `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	CreateAt time.Time `json:"createAt,omitempty"`
	Parent   *Span     `json:"parent,omitempty"`
}

func (s *Span) clone() *Span {
	if s == nil {
		return nil
	}
	ns := &Span{
		Id:       s.Id,
		Name:     s.Name,
		CreateAt: s.CreateAt,
	}
	if s.Parent != nil {
		ns.Parent = s.Parent.clone()
	}
	return ns
}

func (s *Span) SpanID() string {
	if s == nil {
		return ""
	}
	var spanId string
	if s.Parent != nil {
		spanId = s.Parent.SpanID() + "."
	}
	spanId = spanId + s.Id
	return spanId
}

func NewRootSpan(name string) *Span {
	return NewSpan(nil, name)
}

func NewSpan(parent *Span, name string) *Span {
	sp := &Span{
		Id:       genSpanId(name),
		Name:     name,
		CreateAt: time.Now(),
		Parent:   parent,
	}
	return sp
}

func genSpanId(name string) string {
	hash := md5.New()
	hash.Write([]byte(strconv.FormatInt(rd.Int63n(10000000000), 10)))
	hash.Write([]byte(name))
	hash.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	traceId := hash.Sum(nil)
	return hex.EncodeToString(traceId)[:8]
}
