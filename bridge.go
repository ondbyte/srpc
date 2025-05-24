package srpc

import (
	"bufio"
	"fmt"
	"sync"
	"sync/atomic"
)

type Handler func(msg Req) Resp

type rcvResp struct {
	iResp iResp
	err   error
}

type Bridge struct {
	rMu *sync.Mutex
	r   *bufio.Reader
	wMu *sync.Mutex
	w   *bufio.Writer

	handler Handler

	sendRespChsMu *sync.Mutex

	errCh       chan error
	sendReqCh   chan *iReq
	sendRespChs map[uint64]chan *rcvResp
}

// do a req to the other side
// and get a response
func (rw *Bridge) Do(r Req) (Resp, error) {
	req := &iReq{Id: genId(), Req: r}
	responseCh := make(chan *rcvResp)
	defer close(responseCh)
	rw.sendRespChsMu.Lock()
	rw.sendRespChs[req.Id] = responseCh
	rw.sendRespChsMu.Unlock()
	rw.sendReqCh <- req
	rr, ok := <-responseCh
	if !ok {
		return Resp{}, fmt.Errorf("response channel closed")
	}
	return rr.iResp.Resp, rr.err
}

var id atomic.Uint64

func genId() uint64 {
	return id.Add(1)
}

// shutdown
func (rw *Bridge) Burn() {
	close(rw.sendReqCh)
	rw.sendRespChsMu.Lock()
	defer rw.sendRespChsMu.Unlock()
	for _, v := range rw.sendRespChs {
		close(v)
	}
	clear(rw.sendRespChs)
}

// maintains a actual communication bw this and the other party
func (rw *Bridge) loop() chan error {
	go func() {
		for {
			req := <-rw.sendReqCh
			rw.sendRespChsMu.Lock()
			respondCh, ok := rw.sendRespChs[req.Id]
			rw.sendRespChsMu.Unlock()
			if !ok {
				continue
			}
			if err := rw.writeMsg('0', req); err != nil {
				respondCh <- &rcvResp{err: err}
				continue
			}
		}
	}()
	go func() {
		for {
			req, resp, err := rw.readMsg()
			if err != nil {
				rw.errCh <- fmt.Errorf("err while reading message %v", err)
				break
			}
			if req != nil {
				if rw.handler == nil {
					rw.errCh <- fmt.Errorf("handler is nil")
					break
				}
				go func(msg *iReq) {
					resp := rw.handler(msg.Req)
					if err := rw.writeMsg('1', &iResp{msg.Id, resp}); err != nil {
						rw.errCh <- fmt.Errorf("err while writing msg %v", err)
					}
				}(req)
			}
			if resp != nil {
				rw.sendRespChsMu.Lock()
				respnseCh, ok := rw.sendRespChs[resp.Id]
				delete(rw.sendRespChs, resp.Id)
				rw.sendRespChsMu.Unlock()
				if !ok {
					continue
				}
				respnseCh <- &rcvResp{iResp: *resp}
			}
		}
	}()
	return rw.errCh
}
