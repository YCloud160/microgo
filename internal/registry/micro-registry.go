package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/YCloud160/microgo/utils/xlog"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type RouteResp struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

type Route struct {
	Addr string `json:"addr"`
	Name string `json:"name"`
}

type MicroRegistry struct {
	Host string
}

func NewMicroRegistry(host string) *MicroRegistry {
	return &MicroRegistry{Host: host}
}

func (mr *MicroRegistry) Register(name string, addr string) error {
	data := map[string]string{
		"name": name,
		"addr": addr,
	}
	url := fmt.Sprintf("http://%s/micro/route/register", mr.Host)
	_, err := mr.request(url, data)
	return err
}

func (mr *MicroRegistry) UnRegister(name string, addr string) error {
	data := map[string]string{
		"name": name,
		"addr": addr,
	}
	url := fmt.Sprintf("http://%s/micro/route/unregister", mr.Host)
	_, err := mr.request(url, data)
	return err
}

func (mr *MicroRegistry) KeepAlive(name string, addr string) error {
	data := map[string]string{
		"name": name,
		"addr": addr,
	}
	url := fmt.Sprintf("http://%s/micro/route/keepalive", mr.Host)
	_, err := mr.request(url, data)
	return err
}

func (mr *MicroRegistry) request(url string, data map[string]string) (*RouteResp, error) {
	bs, _ := json.Marshal(data)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bs))
	if err != nil {
		xlog.Error(context.TODO(), "请求失败", zap.String("url", url), zap.Error(err))
		return nil, err
	}
	req.Header.Set("content-type", "application/json;charset=utf-8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		xlog.Error(context.TODO(), "请求失败", zap.String("url", url), zap.Error(err))
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		xlog.Error(context.TODO(), "解析数据失败", zap.String("url", url), zap.Error(err))
		return nil, err
	}
	res := &RouteResp{}
	if err := json.Unmarshal(body, res); err != nil {
		xlog.Error(context.TODO(), "解析数据失败", zap.String("res", string(body)), zap.Error(err))
		return nil, err
	}
	if res.Code != 200 {
		xlog.Error(context.TODO(), "获取数据失败", zap.String("res", string(body)))
		return nil, fmt.Errorf("%s", res.Msg)
	}
	return res, nil
}
