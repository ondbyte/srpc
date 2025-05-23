package sconn

import (
	"encoding/json"
	"fmt"
)

type Req struct {
	Id     uint64 `json:"id"`
	Method string `json:"method"`
	Args   []any  `json:"args"`
	Body   []byte `json:"body"`
}

type Resp struct {
	Id   uint64 `json:"id"`
	Body []byte `json:"body"`
	Err  string `json:"err"`
}

func (rw *RW) readMsg() (*Req, *Resp, error) {
	line, _, err := rw.r.ReadLine()
	if err != nil {
		return nil, nil, fmt.Errorf("read line: %w", err)
	}
	if line[0] == '0' {
		// req
		msg := &Req{}
		err = json.Unmarshal(line, msg)
		if err != nil {
			return nil, nil, fmt.Errorf("unmarshal msg: %w", err)
		}
	}
	if line[0] == '1' {
		// resp
		msg := &Resp{}
		err = json.Unmarshal(line, msg)
		if err != nil {
			return nil, nil, fmt.Errorf("unmarshal msg: %w", err)
		}
	}
	return nil, nil, fmt.Errorf("unknown msg type")
}

func (rw *RW) writeReq(req *Req) error {
	// 0 for req
	return rw.writeMsg('0', req)
}

func (rw *RW) writeResp(req *Req) error {
	// 0 for req
	return rw.writeMsg('1', req)
}

func (rw *RW) writeMsg(prefix byte, req any) error {
	if line, err := json.Marshal(req); err != nil {
		return fmt.Errorf("marshal msg: %w", err)
	} else {
		if _, err := rw.w.Write(append(line, '\n')); err != nil {
			return fmt.Errorf("write msg: %w", err)
		}
	}
	return nil
}
