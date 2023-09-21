package microgo

import (
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"net"
	"runtime"
	"sync"
)

const (
	defaultReadBufSize = 4096
	defaultHeadSize    = 5

	messageType = 0xF0
	contentType = 0x0F

	maxBodyLen = 1<<31 - 1 - defaultHeadSize
)

var (
	ErrNotSupportContentType = fmt.Errorf("not support content type")
	ErrFullBodyLen           = fmt.Errorf("full body lenght")
	ErrBadConnection         = fmt.Errorf("bad connection")
	ErrNotFoundConnection    = fmt.Errorf("not found connection")
)

type conn struct {
	mu         sync.Mutex
	rw         net.Conn
	getReadBuf func(size int32) []byte
	readBuf    []byte
	headBuf    []byte
	lastErr    error
	ip         string
}

func newConn(rw net.Conn, msgCh chan *ReadData, removeCh chan *conn) *conn {
	c := &conn{
		rw:      rw,
		readBuf: make([]byte, defaultReadBufSize),
		headBuf: make([]byte, defaultHeadSize),
	}
	ip := net.ParseIP(rw.RemoteAddr().String())
	c.ip = ip.String()

	c.getReadBuf = func(size int32) []byte {
		if size <= int32(len(c.readBuf)) {
			return c.readBuf[:size]
		}
		c.readBuf = make([]byte, size)
		return c.readBuf[:]
	}
	go c.read(msgCh, removeCh)
	return c
}

func (conn *conn) Close() error {
	conn.rw.Close()
	return nil
}

func (conn *conn) read(msgCh chan *ReadData, removeCh chan *conn) {
	defer func() {
		conn.Close()
		removeCh <- conn
	}()

	for {
		msg, err := conn.readMessage()
		if err != nil {
			return
		}
		msgCh <- &ReadData{
			msg:  msg,
			conn: conn,
		}
	}
}

func (conn *conn) readMessage() (*Message, error) {
	headBuf := conn.headBuf[:]
	_, err := io.ReadFull(conn.rw, headBuf)
	if err != nil {
		return nil, err
	}
	msg := getMessage()
	msg.BodyLen = int32(headBuf[0]) | int32(headBuf[1])<<8 | int32(headBuf[2])<<16 | int32(headBuf[3])<<24
	msg.Type = MessageType(headBuf[4]&messageType) >> 4
	msg.ContentType = MessageContentType(headBuf[4]) & contentType

	bodyBuf := conn.getReadBuf(msg.BodyLen)
	_, err = io.ReadFull(conn.rw, bodyBuf)
	if err != nil {
		putMessage(msg)
		return nil, err
	}

	if msg.BodyLen > 0 {
		switch msg.ContentType {
		case MessageContentType_Json:
			err = json.Unmarshal(bodyBuf, msg.Data)
		case MessageContentType_Proto:
			err = proto.Unmarshal(bodyBuf, msg.Data)
		default:
			putMessage(msg)
			return nil, ErrNotSupportContentType
		}
	}

	return msg, nil
}

func (conn *conn) sendMessage(msg *Message) (err error) {
	var dataBytes []byte
	switch msg.ContentType {
	case MessageContentType_Json:
		dataBytes, err = json.Marshal(msg.Data)
	case MessageContentType_Proto:
		dataBytes, err = proto.Marshal(msg.Data)
	default:
		return ErrNotSupportContentType
	}
	if err != nil {
		return err
	}
	bodyLen := int64(len(dataBytes))
	if bodyLen > maxBodyLen {
		return ErrFullBodyLen
	}

	body := make([]byte, bodyLen+defaultHeadSize)
	body[0] = byte(bodyLen)
	body[1] = byte(bodyLen >> 8)
	body[2] = byte(bodyLen >> 16)
	body[3] = byte(bodyLen >> 24)
	body[4] = (byte(msg.Type)<<4)&messageType | byte(msg.ContentType)&contentType
	copy(body[5:], dataBytes)

	_, err = conn.rw.Write(body)

	return err
}

type clientConnPool struct {
	client    *Client
	mu        sync.Mutex
	addr      string
	dial      func(addr string) (net.Conn, error)
	poolSize  int
	index     int
	idleConns []*conn
	removeCh  chan *conn
}

func newClientConnPool(client *Client, addr string, size int) *clientConnPool {
	p := &clientConnPool{
		client:   client,
		addr:     addr,
		poolSize: size,
		removeCh: make(chan *conn, 10),
	}
	p.dial = func(addr string) (net.Conn, error) {
		return net.Dial("tcp", addr)
	}
	if p.poolSize <= 0 {
		p.poolSize = runtime.NumCPU()
	}
	p.idleConns = make([]*conn, 0, p.poolSize)
	go p.listenRemove()
	return p
}

const tryGetConnTimes = 3

func (p *clientConnPool) getConn() (*conn, error) {
	for i := 0; i < tryGetConnTimes; i++ {
		c, err := p.tryGetConn()
		if err != nil {
			if err == ErrBadConnection {
				continue
			}
			return nil, err
		}
		return c, nil
	}
	return nil, ErrNotFoundConnection
}

func (p *clientConnPool) tryGetConn() (*conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.idleConns) >= p.poolSize {
		if p.index >= p.poolSize {
			p.index = 0
		}
		for {
			c := p.idleConns[p.index]
			p.index++
			return c, nil
		}
	}

	rw, err := p.dial(p.addr)
	if err != nil {
		return nil, err
	}
	c := newConn(rw, p.client.msgCh, p.removeCh)
	p.index++
	p.idleConns = append(p.idleConns, c)
	return c, nil
}

func (p *clientConnPool) listenRemove() {
	for range p.removeCh {

	}
}
