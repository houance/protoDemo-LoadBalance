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
	reader         *bufio.Reader
	writer         *net.TCPConn
	headerBuf      []byte
	dataBuf        []byte
	recvError      error
	n              int
	sendAllDataBuf []byte
	sendHeaderBuf  []byte
	sendError      error
	logger         *zap.Logger
}

const MB = 1000000

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

	err = tcpcon.SetWriteBuffer(128000)
	if err != nil {
		return nil, err
	}

	err = tcpcon.SetReadBuffer(256000)
	if err != nil {
		return nil, err
	}

	return &BinaryFramer{
		reader:    bufio.NewReaderSize(tcpcon, size*MB),
		dataBuf:   make([]byte, size*MB),
		headerBuf: make([]byte, 10),
		writer:    tcpcon,
		logger:    logger,
	}, nil
}

func (framer *BinaryFramer) RecvHeader(header *message.Header) error {
	framer.n, framer.recvError = io.ReadAtLeast(framer.reader, framer.headerBuf, 10)
	if framer.recvError != nil {
		return framer.recvError
	}
	return proto.Unmarshal(framer.headerBuf, header)
}

func (framer *BinaryFramer) RecvBytes(header *message.Header) ([]byte, error) {
	framer.n, framer.recvError = io.ReadAtLeast(framer.reader, framer.dataBuf, int(header.Length))
	if framer.recvError != nil {
		return nil, framer.recvError
	}
	return framer.dataBuf[:int(header.Length)], framer.recvError
}

func (framer *BinaryFramer) SendHeader(header *message.Header) error {

	framer.sendHeaderBuf, framer.sendError = proto.Marshal(header)
	if framer.sendError != nil {
		return framer.sendError
	}

	_, framer.sendError = framer.writer.Write(framer.headerBuf)
	return framer.sendError
}

func (framer *BinaryFramer) SendInnerData(innerData *innerData.InnerDataTransfer) error {

	framer.sendAllDataBuf, framer.sendError = innerData.Serilize()
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
