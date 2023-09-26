package main

import (
	"context"
	"github.com/YCloud160/microgo"
	"github.com/YCloud160/microgo/example/rpcserver/model"
	"github.com/YCloud160/microgo/utils/xlog"
	"go.uber.org/zap"
	"strconv"
	"sync"
)

func main() {
	client := model.NewGreetObjClient("demo.rpcServer", microgo.WithClientOptionHosts("127.0.0.1:8080"))
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := model.SayHelloReq{Name: "MicroGo " + strconv.FormatInt(int64(i), 10)}
			resp, err := client.SayHello(context.TODO(), &req)
			if err != nil {
				xlog.Error(context.TODO(), "请求失败", zap.Error(err))
				return
			}
			xlog.Info(context.TODO(), "请求成功", zap.Any("resp", resp.Message))
		}(i)
	}
	wg.Wait()
}
