package main

import (
	"context"
	"fmt"
	"github.com/YCloud160/microgo/example/rpcserver/model"
	"github.com/YCloud160/microgo/utils/tracer"
	"github.com/YCloud160/microgo/utils/xlog"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"sync/atomic"
)

func main() {
	client := model.NewGreetObjClient("demo.rpcServer")
	wg := sync.WaitGroup{}
	var count int32
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := model.SayHelloReq{Name: "MicroGo " + strconv.FormatInt(int64(i), 10)}
			ctx, _ := tracer.WithNewTracer(context.TODO(), req.Name)
			resp, err := client.SayHello(ctx, &req)
			if err != nil {
				xlog.Error(ctx, "请求失败", zap.Error(err))
				return
			}
			atomic.AddInt32(&count, 1)
			xlog.Info(ctx, "请求成功", zap.Any("resp", resp.Message))
		}(i)
	}
	wg.Wait()
	fmt.Println(count)
}
