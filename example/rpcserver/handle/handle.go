package handle

import (
	"context"
	"github.com/YCloud160/microgo/example/rpcserver/model"
	"github.com/YCloud160/microgo/utils/xlog"
	"time"
)

type HelloServer struct{}

func (obj *HelloServer) SayHello(ctx context.Context, req *model.SayHelloReq) (*model.SayHelloResp, error) {
	xlog.Info(ctx, "收到请求", xlog.Field("req", req))
	time.Sleep(time.Second * 10)
	return &model.SayHelloResp{
		Message: "hello " + req.Name,
	}, nil
}
