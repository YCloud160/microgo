package microgo

import (
	"context"
	"github.com/YCloud160/microgo/config"
	"github.com/YCloud160/microgo/meta"
	"github.com/YCloud160/microgo/utils/header"
	"github.com/YCloud160/microgo/utils/xlog"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	startWaitGroup sync.WaitGroup

	serverMap = make(map[string]Server)
	stopCh    = make(chan struct{})

	isClosed atomic.Bool
)

func init() {
	conf := config.GetConfig()
	xlog.InitXlog(
		xlog.WithField(
			xlog.Field("pid", os.Getpid()),
			xlog.Field("service", conf.Service),
		),
		xlog.WithLevel(conf.LogLevel),
		xlog.WithContextWrite(func(ctx context.Context) []*xlog.Entry {
			data, ok := meta.FromOutContext(ctx)
			if !ok {
				return nil
			}
			var fields []*xlog.Entry
			traceId := data[header.TraceID]
			if len(traceId) > 0 {
				fields = append(fields, xlog.Field("traceId", traceId))
			}
			spanId := data[header.SpanID]
			if len(spanId) > 0 {
				fields = append(fields, xlog.Field("spanId", spanId))
			}
			return fields
		}))
}

func RegisterServer(servers ...Server) {
	for _, s := range servers {
		serverMap[s.Name()] = s
	}
}

func Run() error {
	initAdminF()

	for _, server := range serverMap {
		startWaitGroup.Add(1)
		go func(server Server) {
			if err := server.Start(); err != nil {
				xlog.Error(context.TODO(), "server start failed", xlog.Field("server", server.Name()), xlog.Field("error", err))
			}
		}(server)
	}
	startWaitGroup.Wait()

	return loop()
}

func loop() error {
	conf := config.GetConfig()
	tick := time.NewTicker(time.Duration(conf.KeepAlive) * time.Millisecond)
	for {
		select {
		case <-tick.C:
			for _, srv := range serverMap {
				xlog.Info(context.TODO(), "keepAlive", xlog.Field("server", srv.Name()))
			}
		case <-stopCh:
			for _, srv := range serverMap {
				srv.Stop()
			}
			xlog.Info(context.TODO(), "stop service success")
			stopCh <- struct{}{}
			<-stopCh
			return nil
		}
	}
}
