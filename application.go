package microgo

import (
	"context"
	"github.com/YCloud160/microgo/config"
	"github.com/YCloud160/microgo/utils/xlog"
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
