package microgo

import (
	"context"
	"github.com/YCloud160/microgo/utils/xlog"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

var adminFServer *http.Server

func initAdminF() {
	mux := http.NewServeMux()
	mux.HandleFunc("/microgo/stop", stopApplication)
	addr := ":0"
	//conf := config.GetConfig()
	//if len(conf.AppListen) > 0 {
	//	addr = conf.AppListen
	//}
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	adminFServer = &http.Server{
		Handler: mux,
	}
	go func() {
		xlog.Info(context.TODO(), "adminF start", zap.Any("addr", listen.Addr()))
		if err := adminFServer.Serve(listen); err != nil {
			xlog.Error(context.TODO(), "adminF stop", zap.Any("addr", listen.Addr()), zap.Error(err))
		}
	}()
}

func stopApplication(writer http.ResponseWriter, request *http.Request) {
	isClosed.Store(true)
	if registry != nil {
		for _, srv := range serverMap {
			registry.UnRegister(srv.Name(), srv.Addr())
			xlog.Info(context.TODO(), "unregister server", zap.String("server", srv.Name()))
		}
	}

	time.Sleep(time.Second * 15)
	stopCh <- struct{}{}
	<-stopCh
	writer.Write([]byte("stop service success"))
	go func() {
		time.Sleep(time.Second)
		adminFServer.Close()
		stopCh <- struct{}{}
	}()
}
