package sconn

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type Handler func(msg *Req) *Resp
type rcvResp struct {
	resp *Resp
	err  error
}
type RW struct {
	r *bufio.Reader
	w *os.File

	handler Handler

	mu *sync.Mutex

	sendReqCh   chan *Req
	sendRespChs map[uint64]chan *rcvResp
}

func NewRW(r, w *os.File, handler Handler) *RW {
	return &RW{r: bufio.NewReader(r), w: w, handler: handler}
}

func (rw *RW) SendReq(req *Req) (*Resp, error) {
	responseCh := make(chan *rcvResp)
	defer close(responseCh)
	rw.sendRespChs[req.Id] = responseCh
	rw.sendReqCh <- req
	rr := <-responseCh
	return rr.resp, rr.err
}

func (rw *RW) RcvReqLoop() (e chan error) {
	e = make(chan error)
	go func() {
		for {
			req, resp, err := rw.readMsg()
			if err != nil {
				e <- fmt.Errorf("err while reading msg %v", err)
				break
			}
			if req != nil {
				go func(msg *Req) {
					if err := rw.writeMsg('1', rw.handler(msg)); err != nil {
						e <- fmt.Errorf("err while writing msg %v", err)
					}
				}(req)
			}
			if resp != nil {
				rw.mu.Lock()
				responseCh, ok := rw.sendRespChs[resp.Id]
				delete(rw.sendRespChs, resp.Id)
				rw.mu.Unlock()
				if !ok {
					continue
				}
				responseCh <- &rcvResp{resp: resp}
				close(responseCh)
			}
		}
	}()
	return
}

func (rw *RW) SendReqLoop() (e chan error) {
	e = make(chan error)
	go func() {
		for {
			req := <-rw.sendReqCh
			rw.mu.Lock()
			respondCh, ok := rw.sendRespChs[req.Id]
			rw.mu.Unlock()
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
				e <- fmt.Errorf("err while reading message %v", err)
				break
			}
			if req != nil {
				go func(msg *Req) {
					if err := rw.writeMsg('1', rw.handler(msg)); err != nil {
						e <- fmt.Errorf("err while writing msg %v", err)
					}
				}(req)
			}
			if resp != nil {
				rw.mu.Lock()
				respnseCh, ok := rw.sendRespChs[resp.Id]
				delete(rw.sendRespChs, resp.Id)
				rw.mu.Unlock()
				if !ok {
					continue
				}
				respnseCh <- &rcvResp{resp: resp}
				close(respnseCh)
			}
		}
	}()
	return
}
