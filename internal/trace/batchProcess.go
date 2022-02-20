package trace

import (
	"bytes"
	"context"
	"net"
)

type BatchProcess struct {
	currentBatch int
	limit        int
	con          net.Conn
	buffer       bytes.Buffer
	fullChannel  chan bool
}

var (
	err error
)

func NewBatchProcess(
	maxSpan int,
	address string,
) (*BatchProcess, error) {

	con, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	return &BatchProcess{
		limit:       maxSpan,
		con:         con,
		buffer:      bytes.Buffer{},
		fullChannel: make(chan bool, 2)}, nil
}

func (bp *BatchProcess) SendSpan(spanByte []byte) {
	bp.buffer.Write(spanByte)
	bp.currentBatch++

	if bp.currentBatch > bp.limit {
		bp.fullChannel <- true
	}
}

func (bp *BatchProcess) SendBatchReachLimit(ctx context.Context) error {
	for {
		select {
		case <-bp.fullChannel:
			_, err = bp.con.Write(bp.buffer.Bytes())
			if err != nil {
				return err
			}
			bp.buffer.Reset()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
