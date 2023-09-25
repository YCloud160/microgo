package microgo

import (
	"context"
	"github.com/YCloud160/microgo/config"
	"github.com/YCloud160/microgo/utils/xlog"
	"go.uber.org/zap"
	"net"
	"net/http"
)

type ServerHTTP struct {
	name       string
	conf       *config.ServerConfig
	httpServer *http.Server

	mux http.Handler
}

func NewHttpServer(name string, mux http.Handler) Server {
	srv := &ServerHTTP{
		name: name,
		mux:  mux,
		conf: config.GetServerConfig(name),
	}
	srv.httpServer = &http.Server{
		Handler: srv,
	}
	return srv
}

func (srv *ServerHTTP) Start() error {
	listenAddr := ":" + srv.conf.Port
	listen, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	startWaitGroup.Done()
	xlog.Info(context.TODO(), "start http server", zap.String("server", srv.Name()), zap.String("listen", srv.conf.Port))
	return srv.httpServer.Serve(listen)
}

func (srv *ServerHTTP) Stop() error {
	err := srv.httpServer.Close()
	xlog.Info(context.TODO(), "stop http server", zap.String("server", srv.Name()))
	return err
}

func (srv *ServerHTTP) Name() string {
	return srv.name
}

func (srv *ServerHTTP) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := context.TODO()
	req = req.WithContext(ctx)

	defer xlog.Recover(ctx)

	srv.mux.ServeHTTP(rw, req)
}
