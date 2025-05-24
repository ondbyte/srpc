package srpc_test

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/ondbyte/srpc"
)

func TestBridge(t *testing.T) {
	r1, w1, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	r2, w2, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	b1, errCh1 := srpc.RunOneWayBridge(r1, w2, func(msg srpc.Req) srpc.Resp {
		switch msg.Method {
		case "ping1":
			return srpc.Resp{Body: []byte("pong1")}
		}
		return srpc.Resp{Err: srpc.ErrMethodNotFound}
	})
	b2, errCh2 := srpc.RunOneWayBridge(r2, w1, func(msg srpc.Req) srpc.Resp {
		switch msg.Method {
		case "ping2":
			return srpc.Resp{Body: []byte("pong2")}
		}
		return srpc.Resp{Err: srpc.ErrMethodNotFound}
	})

	go func() {
		select {
		case err := <-errCh1:
			t.Errorf("b1 reported error %v", err)
		case err := <-errCh2:
			t.Errorf("b2 reported error %v", err)
		}
	}()
	// multiple calls
	var wg sync.WaitGroup
	n := 50000
	wg.Add(n * 2)
	errCh := make(chan error)
	for range n {
		go func() {
			resp, err := b2.Do(srpc.Req{Method: "ping1"})
			if err != nil {
				errCh <- err
			}
			if string(resp.Body) != "pong1" {
				errCh <- fmt.Errorf("pong1 expected")
			}
			wg.Done()
		}()
		go func() {
			resp, err := b1.Do(srpc.Req{Method: "ping2"})
			if err != nil {
				errCh <- err
			}
			if string(resp.Body) != "pong2" {
				errCh <- fmt.Errorf("pong2 expected")
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
