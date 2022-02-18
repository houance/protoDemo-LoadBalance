package binaryframer

import (
	"bufio"
	"errors"
	message "houance/protoDemo-LoadBalance/external"
	"houance/protoDemo-LoadBalance/internal/innerData"
	"io"
	"net"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type BinaryFramer struct {
	reader *bufio.Reader
	writer *net.TCPConn

	recvHeaderBuf []byte
	recvPrefixBuf []byte
	recvDataBuf   []byte
	recvSpanBuf   []byte

	recvError error

	sendAllDataBuf []byte
	sendHeaderBuf  []byte
	sendError      error
	logger         *zap.Logger
}

const MB = 1000000

const headerLength = 25

const prefixLengthSize = 5

func NewBinaryFramer(con net.Conn, size int, logger *zap.Logger) (*BinaryFramer, error) {

	var err error

	if size > 10 {
		return nil, errors.New("size must smaller than 10")
	}
	tcpcon := con.(*net.TCPConn)

	err = tcpcon.SetNoDelay(true)
	if err != nil {
		return nil, err
	}

	err = tcpcon.SetLinger(0)
	if err != nil {
		return nil, err
	}

	return &BinaryFramer{
		reader:        bufio.NewReaderSize(tcpcon, size*MB),
		recvDataBuf:   make([]byte, size*MB),
		recvSpanBuf:   make([]byte, size*MB),
		recvHeaderBuf: make([]byte, headerLength),
		recvPrefixBuf: make([]byte, prefixLengthSize),
		writer:        tcpcon,
		logger:        logger,
	}, nil
}

func (framer *BinaryFramer) RecvHeader(header *message.Header) error {
	_, framer.recvError = io.ReadAtLeast(framer.reader, framer.recvHeaderBuf, headerLength)
	if framer.recvError != nil {
		return framer.recvError
	}
	return proto.Unmarshal(framer.recvHeaderBuf, header)
}

func (framer *BinaryFramer) RecvBytes(header *message.Header) ([]byte, error) {
	_, framer.recvError = io.ReadAtLeast(framer.reader, framer.recvDataBuf, int(header.Length))
	if framer.recvError != nil {
		return nil, framer.recvError
	}
	return framer.recvDataBuf[:int(header.Length)], framer.recvError
}

func (framer *BinaryFramer) RecvSpan(
	prefixLength *message.PrefixLength,
	spanInfo *message.SpanInfo) error {

	_, framer.recvError = io.ReadAtLeast(framer.reader, framer.recvPrefixBuf, prefixLengthSize)
	if framer.recvError != nil {
		return framer.recvError
	}

	framer.recvError = proto.Unmarshal(framer.recvPrefixBuf, prefixLength)
	if framer.recvError != nil {
		return framer.recvError
	}

	_, framer.recvError = io.ReadAtLeast(framer.reader, framer.recvSpanBuf, int(prefixLength.Length))
	if framer.recvError != nil {
		return framer.recvError
	}

	return proto.Unmarshal(framer.recvSpanBuf, spanInfo)
}

func (framer *BinaryFramer) SendHeader(header *message.Header) error {

	framer.sendHeaderBuf, framer.sendError = proto.Marshal(header)
	if framer.sendError != nil {
		return framer.sendError
	}

	_, framer.sendError = framer.writer.Write(framer.sendHeaderBuf)
	return framer.sendError
}

func (framer *BinaryFramer) SendInnerData(innerData *innerData.DataTransfer) error {

	framer.sendAllDataBuf, framer.sendError = innerData.Serilize()
	if framer.sendError != nil {
		return framer.sendError
	}

	_, framer.sendError = framer.writer.Write(framer.sendAllDataBuf)
	return framer.sendError
}

func (framer *BinaryFramer) SendDataTrace(dataTrace *innerData.DataTrace) error {

	framer.sendAllDataBuf, framer.sendError = dataTrace.Serilize()
	if framer.sendError != nil {
		return framer.sendError
	}

	_, framer.sendError = framer.writer.Write(framer.sendAllDataBuf)
	return framer.sendError
}

func (framer *BinaryFramer) GetRemoteAddress() (address string) {
	return framer.writer.RemoteAddr().String()
}

func (framer *BinaryFramer) Close() {
	framer.writer.Close()
}
