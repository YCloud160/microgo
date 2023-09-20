package main

import (
	"github.com/YCloud160/microgo"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello microgo"))
	})
	server := microgo.NewHttpServer("demo.httpServer", mux)
	microgo.RegisterServer(server)
	microgo.Run()
}
