package microgo

import (
	"context"
	"github.com/YCloud160/microgo/config"
	"github.com/YCloud160/microgo/errors"
	"github.com/YCloud160/microgo/internal/generator"
	"github.com/YCloud160/microgo/meta"
	"github.com/YCloud160/microgo/utils/header"
	"github.com/YCloud160/microgo/utils/tracer"
	"github.com/YCloud160/microgo/utils/xlog"
	"go.uber.org/zap"
	"sync"
	"time"
)

type ClientOption func(client *Client)

func WithClientOptionHosts(hosts ...string) ClientOption {
	return func(client *Client) {
		client.mu.Lock()
		defer client.mu.Unlock()

		existHost := make(map[string]struct{}, len(client.hosts))
		for _, host := range client.hosts {
			existHost[host] = struct{}{}
		}
		for _, host := range hosts {
			if _, ok := existHost[host]; ok {
				continue
			}
			client.hosts = append(client.hosts, host)
		}
	}
}

type Client struct {
	name  string
	mu    sync.Mutex
	conf  *config.ClientConfig
	idx   int
	hosts []string
	pool  map[string]*clientConnPool
	reqCh sync.Map
}

func NewClient(name string, options ...ClientOption) *Client {
	client := &Client{
		name:  name,
		conf:  config.GetClientConfig(),
		hosts: make([]string, 0),
		pool:  make(map[string]*clientConnPool),
	}

	for _, option := range options {
		option(client)
	}

	if discovery != nil {
		hosts, err := discovery.QueryRoute(name)
		xlog.Info(context.TODO(), "节点", zap.Strings("hosts", hosts))
		if err == nil && len(hosts) > 0 {
			WithClientOptionHosts(hosts...)(client)
		}
		go client.updateNode()
	}

	return client
}

func (client *Client) updateNode() {
	client._updateNode()
	tick := time.NewTicker(time.Second * time.Duration(client.conf.RefreshEndpointInterval))
	for {
		select {
		case <-tick.C:
			client._updateNode()
		}
	}
}

func (client *Client) _updateNode() {
	defer xlog.Recover(context.TODO())

	hosts, err := discovery.QueryRoute(client.name)
	if err != nil {
		//xlog.Error(ctx, "更新节点失败", zap.Error(err))
		return
	}
	var (
		oldHost     = make(map[string]struct{})
		delHost     = make(map[string]struct{})
		newHost     = make(map[string]struct{})
		newHostList []string
	)
	for _, h := range hosts {
		newHost[h] = struct{}{}
	}

	client.mu.Lock()
	for _, h := range client.hosts {
		oldHost[h] = struct{}{}
	}

	for h := range oldHost {
		if _, ok := newHost[h]; !ok {
			delHost[h] = struct{}{}
		} else {
			newHostList = append(newHostList, h)
		}
	}
	for h := range newHost {
		if _, ok := oldHost[h]; ok {
			continue
		}
		newHostList = append(newHostList, h)
	}

	var delPool []*clientConnPool

	client.hosts = newHostList
	for h := range delHost {
		if p, ok := client.pool[h]; ok {
			delPool = append(delPool, p)
			delete(client.pool, h)
		}
	}
	client.mu.Unlock()

	for _, p := range delPool {
		p.close()
	}
}

func (client *Client) handle(msg *Message) {
	val, ok := client.reqCh.Load(msg.Data.RequestId)
	if ok {
		if reqCh, reqOk := val.(chan *Message); reqOk {
			reqCh <- msg
		}
	}
}

func (client *Client) Call(ctx context.Context, host, contentType, method string, input []byte) (out []byte, err error) {
	return client.call(ctx, host, contentType, method, input)
}

func (client *Client) BroadcastCall(ctx context.Context, contentType, method string, input []byte) (map[string][]byte, map[string]error) {
	hosts := client.getActiveHosts()
	var (
		outs = make(map[string][]byte)
		errs = make(map[string]error)
	)
	for _, host := range hosts {
		out, err := client.call(ctx, host, contentType, method, input)
		outs[host] = out
		errs[host] = err
	}
	return outs, errs
}

func (client *Client) call(ctx context.Context, host, contentType, method string, input []byte) (out []byte, err error) {
	trace, _ := tracer.OutTracer(ctx)
	meta := meta.FromOutContext(ctx)
	meta[header.ContentType] = contentType
	if trace != nil {
		meta[header.Tracer] = trace.String()
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(client.conf.RequestTimeout))
	defer cancel()

	req := getMessage()
	reqId := generator.NextRequestId()
	req.Type = MessageType_Data
	req.ContentType = defaultContentType
	req.Data.RequestId = reqId
	req.Data.Obj = client.name
	req.Data.Method = method
	req.Data.Meta = meta
	req.Data.Body = input

	var respChan = make(chan *Message, 1)

	rw, err := client.getConn(host)
	if err != nil {
		return nil, err
	}

	client.reqCh.Store(reqId, respChan)
	defer client.reqCh.Delete(reqId)

	if err := rw.sendMessage(req); err != nil {
		return nil, err
	}
	putMessage(req)

	var (
		resp *Message
		ok   bool
	)

	select {
	case <-ctx.Done():
	case resp, ok = <-respChan:
		if ok {
			out = resp.Data.Body
			if resp.Data.Code != 0 {
				err = errors.New("", resp.Data.Desc, resp.Data.Code)
			}
			putMessage(resp)
			return out, err
		}
	}

	return nil, errors.New("", "request timeout", 9999)
}

func (client *Client) getActiveHosts() []string {
	var hosts []string
	client.mu.Lock()
	for host := range client.pool {
		hosts = append(hosts, host)
	}
	client.mu.Unlock()
	return hosts
}

func (client *Client) getConn(host string) (*conn, error) {
	if len(host) > 0 {
		return client.getTargetConn(host)
	}
	return client.getRingConn()
}

func (client *Client) getTargetConn(host string) (*conn, error) {
	client.mu.Lock()

	exist := false
	for _, domain := range client.hosts {
		if domain == host {
			exist = true
			break
		}
	}
	client.mu.Unlock()

	if exist {
		return client._getConn(host)
	}

	return nil, ErrNotFoundConnection
}

func (client *Client) getRingConn() (*conn, error) {
	if len(client.hosts) == 0 {
		return nil, ErrNotFoundConnection
	}

	client.mu.Lock()
	if client.idx >= len(client.hosts) {
		client.idx = 0
	}
	host := client.hosts[client.idx]
	client.idx++
	client.mu.Unlock()

	return client._getConn(host)
}

func (client *Client) _getConn(host string) (*conn, error) {
	client.mu.Lock()
	pool, ok := client.pool[host]
	if !ok {
		pool = newClientConnPool(client, host, 0)
		client.pool[host] = pool
	}
	client.mu.Unlock()
	return pool.getConn()
}
