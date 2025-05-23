package main

import (
	"os"
	"os/exec"
	"time"

	"github.com/ondbyte/srpc"
)

func main() {
	r1, w1, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	r2, w2, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	//plugin
	cmd := exec.Command("go", "run", "./plugin")
	cmd.Stdin = r1
	cmd.Stdout = w2

	//bridge
	bridge, errCh := srpc.RunOneWayBridge(r2, w1, nil)
	go func() {
		err := <-errCh
		if err != nil {
			panic(err)
		}
	}()
	go func() {
		time.Sleep(10 * time.Second)
		os.Exit(0)
	}()
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	for range 1000 {
		resp, err := bridge.Cross(srpc.Req{Method: "echo", Body: []byte("body")})
		if err != nil {
			panic(err)
		}
		if string(resp.Body) != "body" {
			panic("body not equal")
		} else {
			println("success")
		}
	}
}
