package srpc

import (
	"encoding/json"
	"fmt"
)

type iReq struct {
	Id uint64 `json:"id"`
	Req
}
type Req struct {
	Method string `json:"method"`
	Args   []any  `json:"args"`
	Body   []byte `json:"body"`
}

var ErrMethodNotFound = "method not found"

type iResp struct {
	Id uint64 `json:"id"`
	Resp
}
type Resp struct {
	Body []byte `json:"body"`
	Err  string `json:"err"`
}

func (rw *Bridge) readMsg() (*iReq, *iResp, error) {
	rw.rMu.Lock()
	line, _, err := rw.r.ReadLine()
	rw.rMu.Unlock()
	if err != nil {
		return nil, nil, fmt.Errorf("read line: %w", err)
	}
	if line[0] == '0' {
		// req
		msg := &iReq{}
		err = json.Unmarshal(line[1:], msg)
		if err != nil {
			return nil, nil, fmt.Errorf("unmarshal msg: %w", err)
		}
		return msg, nil, nil
	} else if line[0] == '1' {
		// resp
		msg := &iResp{}
		err = json.Unmarshal(line[1:], msg)
		if err != nil {
			return nil, nil, fmt.Errorf("unmarshal msg: %w", err)
		}
		return nil, msg, nil
	}
	return nil, nil, fmt.Errorf("unknown msg type")
}

func (rw *Bridge) writeMsg(prefix byte, req any) error {
	if line, err := json.Marshal(req); err != nil {
		return fmt.Errorf("marshal msg: %w", err)
	} else {
		line = append([]byte{prefix}, line...)
		rw.wMu.Lock()
		defer rw.wMu.Unlock()
		if _, err := rw.w.Write(append(line, '\n')); err != nil {
			return fmt.Errorf("write msg: %w", err)
		}
		err := rw.w.Flush()
		if err != nil {
			return fmt.Errorf("flush msg: %w", err)
		}
	}
	return nil
}
