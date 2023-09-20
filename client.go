package microgo

import (
	"context"
	"github.com/YCloud160/microgo/config"
	"github.com/YCloud160/microgo/errors"
	"github.com/YCloud160/microgo/internal/generator"
	"github.com/YCloud160/microgo/meta"
	"github.com/YCloud160/microgo/utils/header"
	"sync"
	"time"
)

const (
	defaultRequestTimeout = time.Second * 5
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
	msgCh chan *ReadData
	reqCh sync.Map
}

func NewClient(name string, options ...ClientOption) *Client {
	client := &Client{
		name:  name,
		conf:  config.GetClientConfig(),
		hosts: make([]string, 0),
		pool:  make(map[string]*clientConnPool),
		msgCh: make(chan *ReadData, defaultReadCh),
	}

	for _, option := range options {
		option(client)
	}

	go client.handle()
	return client
}

func (client *Client) handle() {
	for {
		select {
		case msg := <-client.msgCh:
			val, ok := client.reqCh.Load(msg.msg.Data.RequestId)
			if ok {
				if reqCh, reqOk := val.(chan *Message); reqOk {
					reqCh <- msg.msg
				}
			}
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
	meta, _ := meta.FromOutContext(ctx)
	meta[header.ContentType] = contentType

	ctx, cancel := context.WithTimeout(ctx, time.Duration(client.conf.RequestTimeout))
	defer cancel()

	conn, err := client.getConn(host)
	if err != nil {
		return nil, err
	}
	req := getMessage()
	req.Type = MessageType_Data
	req.ContentType = defaultContentType
	req.Data.RequestId = generator.NextRequestId()
	req.Data.Obj = client.name
	req.Data.Method = method
	req.Data.Meta = meta
	req.Data.Body = input
	if err := conn.sendMessage(req); err != nil {
		return nil, err
	}

	var (
		respChan = make(chan *Message, 1)
		resp     *Message
		ok       bool
	)
	client.reqCh.Store(req.Data.RequestId, respChan)
	defer client.reqCh.Delete(req.Data.RequestId)

	select {
	case <-ctx.Done():
	case resp, ok = <-respChan:
		if ok {
			if resp.Data.Code == 0 {
				return resp.Data.Body, nil
			} else {
				return resp.Data.Body, errors.New("", resp.Data.Desc, resp.Data.Code)
			}
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
	return client.getRandConn()
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

func (client *Client) getRandConn() (*conn, error) {
	client.mu.Lock()

	if client.idx >= len(client.hosts) {
		client.idx = 0
	}
	host := client.hosts[client.idx]
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