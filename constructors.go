package srpc

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"sync"
)

// run a bridge, can be communicated with another bridge over the write and reader of r and w respectively
func RunOneWayBridge(r io.Reader, w io.Writer, handler Handler) (b *Bridge, errCh chan error) {
	b = &Bridge{
		r:             bufio.NewReader(r),
		w:             bufio.NewWriter(w),
		handler:       handler,
		sendReqCh:     make(chan *iReq),
		errCh:         make(chan error),
		sendRespChs:   map[uint64]chan *rcvResp{},
		sendRespChsMu: &sync.Mutex{},
		rMu:           &sync.Mutex{},
		wMu:           &sync.Mutex{},
	}
	return b, b.loop()
}

// run a bridge with methods to be called by another bridge over stdio
func RunOneWayBridgeOnStdio(handler Handler) (b *Bridge, errCh chan error) {
	return RunOneWayBridge(io.Reader(os.Stdin), io.Writer(os.Stdout), handler)
}

// connects your bridge to another in a cmd,
// allows to call methods on another bridge,
// allows the other bridge to call method on yours
func TwoWayBridgeWithACmd(cmd *exec.Cmd, thisHandler Handler) (b *Bridge, errCh chan error) {
	r1, w1, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	r2, w2, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	cmd.Stdin = r1
	cmd.Stdout = w2
	return RunOneWayBridge(r2, w1, thisHandler)
}
