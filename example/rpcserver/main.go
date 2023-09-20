package main

import (
	"github.com/YCloud160/microgo"
	"github.com/YCloud160/microgo/example/rpcserver/handle"
	"github.com/YCloud160/microgo/example/rpcserver/model"
)

func main() {
	server := microgo.NewTCPServer("demo.rpcServer", &handle.HelloServer{}, model.GreetObjCall)
	microgo.RegisterServer(server)
	microgo.Run()
}
