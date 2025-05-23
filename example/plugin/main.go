package main

import (
	"os"

	"github.com/ondbyte/srpc"
)

func main() {
	_, errCh := srpc.RunOneWayBridge(os.Stdin, os.Stdout, func(msg srpc.Req) srpc.Resp {
		switch msg.Method {
		case "echo":
			return srpc.Resp{Body: msg.Body}
		default:
			return srpc.Resp{Err: srpc.ErrMethodNotFound}
		}
	})
	err := <-errCh
	if err != nil {
		panic(err)
	}
}
