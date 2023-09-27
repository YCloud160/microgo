package discovery

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
	Code   int32    `json:"code"`
	Msg    string   `json:"msg"`
	Routes []*Route `json:"routes"`
}

type Route struct {
	Addr string `json:"addr"`
	Name string `json:"name"`
}

type MicroDiscovery struct {
	Host string
}

func NewMicroDiscovery(host string) *MicroDiscovery {
	return &MicroDiscovery{Host: host}
}

func (md *MicroDiscovery) QueryRoute(name string) ([]string, error) {
	url := fmt.Sprintf("http://%s/micro/route/query", md.Host)
	bs, _ := json.Marshal(map[string]string{"name": name})
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
	routes := make([]string, 0, len(res.Routes))
	for _, r := range res.Routes {
		routes = append(routes, r.Addr)
	}
	return routes, nil
}
