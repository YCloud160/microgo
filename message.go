package microgo

import (
	"github.com/YCloud160/microgo/pb"
	"sync"
)

type MessageType uint8

const (
	MessageType_Ping MessageType = 0x10
	MessageType_Data MessageType = 0x20
)

type MessageContentType uint8

const (
	MessageContentType_Json  MessageContentType = 0x01
	MessageContentType_Proto MessageContentType = 0x02

	defaultContentType = MessageContentType_Json
)

type CompressType uint8

const (
	CompressType_Gzip CompressType = 0x10
)

type ReadData struct {
	msg  *Message
	conn *conn
}

type Message struct {
	BodyLen      int32
	Type         MessageType
	ContentType  MessageContentType
	CompressType CompressType
	Data         *pb.Message
}

func (msg *Message) reset() {
	msg.BodyLen = 0
	msg.Type = 0
	msg.ContentType = 0
	msg.CompressType = 0
	msg.Data.RequestId = 0
	msg.Data.Obj = ""
	msg.Data.Method = ""
	msg.Data.Meta = nil
	msg.Data.Body = nil
	msg.Data.Code = 0
	msg.Data.Desc = ""
}

var messagePool = sync.Pool{
	New: func() any {
		return newMessage()
	},
}

func newMessage() *Message {
	return &Message{
		BodyLen:      0,
		Type:         0,
		ContentType:  0,
		CompressType: 0,
		Data:         new(pb.Message),
	}
}

func getMessage() *Message {
	val := messagePool.Get()
	if msg, ok := val.(*Message); ok {
		return msg
	}
	return newMessage()
}

func putMessage(msg *Message) {
	msg.reset()
	messagePool.Put(msg)
}
