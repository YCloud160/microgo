package handle

import (
	"context"
	"github.com/YCloud160/microgo/example/rpcserver/model"
	"github.com/YCloud160/microgo/utils/xlog"
	"go.uber.org/zap"
)

type HelloServer struct{}

func (obj *HelloServer) SayHello(ctx context.Context, req *model.SayHelloReq) (*model.SayHelloResp, error) {
	xlog.Info(ctx, "收到请求", zap.Any("req", req))
	return &model.SayHelloResp{
		Message: "hello " + req.Name,
	}, nil
}
