package microgo

import (
	"context"
	"github.com/YCloud160/microgo/config"
	ierrors "github.com/YCloud160/microgo/errors"
	"github.com/YCloud160/microgo/meta"
	"github.com/YCloud160/microgo/utils/header"
	"github.com/YCloud160/microgo/utils/tracer"
	"github.com/YCloud160/microgo/utils/xlog"
	"go.uber.org/zap"
	"io"
	"net"
	"sync"
	"time"
)

const (
	defaultReadCh = 10000
)

type ServerTCP struct {
	name string
	mu   sync.Mutex

	conf   *config.ServerConfig
	listen net.Listener
	conns  map[*conn]struct{}

	tick chan struct{}

	impl any
	call Call
}

func NewTCPServer(name string, impl any, call Call) Server {
	srv := &ServerTCP{
		name:  name,
		conf:  config.GetServerConfig(name),
		conns: make(map[*conn]struct{}),
		impl:  impl,
		call:  call,
	}
	srv.tick = make(chan struct{}, srv.conf.MaxInvoke)
	return srv
}

func (srv *ServerTCP) Start() error {
	listenAddr := ":" + srv.conf.Port
	listen, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	startWaitGroup.Done()
	xlog.Info(context.TODO(), "start tcp server", zap.String("server", srv.Name()), zap.String("listen", srv.conf.Port))
	srv.listen = listen
	return srv.accept()
}

func (srv *ServerTCP) Stop() error {
	var removeConn []*conn
	srv.mu.Lock()
	srv.listen.Close()
	for conn := range srv.conns {
		delete(srv.conns, conn)
		removeConn = append(removeConn, conn)
	}
	srv.mu.Unlock()

	xlog.Info(context.TODO(), "stop tcp server", zap.String("server", srv.Name()))
	return nil
}

func (srv *ServerTCP) Name() string {
	return srv.name
}

func (srv *ServerTCP) accept() error {
	defer srv.Stop()

	for {
		rw, err := srv.listen.Accept()
		if err != nil {
			operr, ok := err.(*net.OpError)
			if ok && operr.Err == net.ErrClosed {
				return nil
			}
			return err
		}
		if tcpConn, ok := rw.(*net.TCPConn); ok {
			tcpConn.SetWriteBuffer(4096)
			tcpConn.SetReadBuffer(4096)
		}
		c := newConn(rw)
		srv.addConn(c)
		go srv.handle(c)
	}
}

func (srv *ServerTCP) handle(c *conn) {
	defer xlog.Recover(context.TODO())
	defer srv.removeConn(c)

	for {
		msg, err := c.readMessage()
		if err != nil {
			if err != io.EOF {
				xlog.Error(context.TODO(), "read message failed", zap.Error(err))
			}
			return
		}
		switch msg.Type {
		case MessageType_Ping:
		case MessageType_Data:
			srv.invoke(c, msg)
		default:
			xlog.Error(context.TODO(), "error message type", zap.Any("data type", msg.Type))
			return
		}
	}
}

func (srv *ServerTCP) addConn(conn *conn) {
	srv.mu.Lock()
	srv.conns[conn] = struct{}{}
	srv.mu.Unlock()
}

func (srv *ServerTCP) removeConn(conn *conn) {
	srv.mu.Lock()
	_, ok := srv.conns[conn]
	if ok {
		delete(srv.conns, conn)
	}
	srv.mu.Unlock()
	conn.Close()
}

func (srv *ServerTCP) ping(message *Message) {

}

type outChan struct {
	data []byte
	err  error
}

func (srv *ServerTCP) invoke(conn *conn, req *Message) {
	ctx := context.TODO()
	var cancel context.CancelFunc
	if srv.conf.InvokeTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(srv.conf.InvokeTimeout))
		defer cancel()
	}

	resp := getMessage()
	resp.Type = MessageType_Data
	resp.ContentType = defaultContentType
	resp.Data.RequestId = req.Data.RequestId

	select {
	case <-ctx.Done():
		resp.Data.Code = 9999
		resp.Data.Desc = "request timeout"
		conn.sendMessage(resp)
		putMessage(resp)
		return
	case srv.tick <- struct{}{}:
	}
	defer func() {
		<-srv.tick
	}()

	ctxData := req.Data.Meta
	if ctxData == nil {
		ctxData = make(map[string]string)
	}
	ctxData[header.RemoteIP] = conn.ip
	ctx, ctxData = setTrace(ctx, ctxData, req.Data.Method)
	ctx = meta.NewOutContext(ctx, ctxData)
	contentType := ctxData[header.ContentType]
	enc := GetEncoder(contentType)

	var (
		respData *outChan
		ok       bool
		respCh   = make(chan *outChan, 1)
	)

	go func() {
		defer xlog.Recover(ctx)
		out, err := srv.call(ctx, srv.impl, enc, req.Data.Method, req.Data.Body)
		respCh <- &outChan{data: out, err: err}
	}()

	select {
	case <-ctx.Done():
		respData, ok = <-respCh
		if !ok {
			resp.Data.Code = 9999
			resp.Data.Desc = "request timeout"
		}
	case respData, ok = <-respCh:
		if !ok {
			resp.Data.Code = 9999
			resp.Data.Desc = "request timeout"
		}
	}
	close(respCh)

	if ok {
		resp.Data.Body = respData.data
		if respData.err != nil {
			var parseErr *ierrors.Error
			if ie, ok := respData.err.(*ierrors.Error); ok {
				parseErr = ie
			} else {
				parseErr = ierrors.ParseError(respData.err.Error())
			}
			resp.Data.Code = parseErr.Code
			resp.Data.Desc = parseErr.Desc
		}
	}
	conn.sendMessage(resp)
	putMessage(resp)
}

func setTrace(ctx context.Context, ctxData map[string]string, name string) (context.Context, map[string]string) {
	isSetTracer := false
	if traceData, ok := ctxData[header.Tracer]; ok {
		trace := tracer.ParseTrace(traceData)
		if trace != nil {
			isSetTracer = true
			ctx = tracer.WithTracer(ctx, trace, name)
			ctxData[header.TraceID] = trace.TraceID()
			ctxData[header.SpanID] = trace.SpanID()
		}
	}
	if isSetTracer == false {
		tracerCtx, trace := tracer.WithNewTracer(ctx, name)
		ctx = tracerCtx
		if trace != nil {
			ctxData[header.TraceID] = trace.TraceID()
			ctxData[header.SpanID] = trace.SpanID()
		}
	}
	return ctx, ctxData
}
