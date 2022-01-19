package binaryframer

import (
	"bufio"
	message "houance/protoDemo-LoadBalance/external"
	"io"
	"net"
	"houance/protoDemo-LoadBalance/internal/innerData"
	"google.golang.org/protobuf/proto"
)

type BinaryFramer struct {
	reader *bufio.Reader
	writer net.Conn
	headerBuf []byte
	dataBuf []byte
	recvError error
	n int
	sendAllDataBuf []byte
	sendHeaderBuf []byte
	sendError error
}

const MB = 1000000

func NewBinaryFramer(con net.Conn, size int) (*BinaryFramer) {

	if size > 10{
		return nil
	}

	return &BinaryFramer{
		reader: bufio.NewReaderSize(con, size * MB),
		dataBuf: make([]byte, size * MB),
		headerBuf: make([]byte, 10),
		writer: con,
	}
}


func (framer *BinaryFramer) RecvHeader(header *message.Header) (error){
	framer.n, framer.recvError = io.ReadAtLeast(framer.reader, framer.headerBuf, 10)
	if framer.recvError != nil {
		return framer.recvError
	}
	return proto.Unmarshal(framer.headerBuf, header)
}

func (framer *BinaryFramer) RecvBytes(header *message.Header) ([]byte, error){
	framer.n, framer.recvError = io.ReadAtLeast(framer.reader, framer.dataBuf, int(header.Length))
	if framer.recvError != nil {
		return nil, framer.recvError
	}
	return framer.dataBuf, framer.recvError
}

func (framer *BinaryFramer) SendHeader(header *message.Header) (error) {
	
	framer.sendHeaderBuf, framer.sendError = proto.Marshal(header)
	if framer.sendError != nil {
		return framer.sendError
	}

	_,framer.sendError = framer.writer.Write(framer.headerBuf)
	return framer.sendError
}

func (framer *BinaryFramer) SendInnerData(innerData *innerData.InnerDataTransfer) (error) {
	
	framer.sendAllDataBuf, framer.sendError = innerData.Serilize()
	if framer.sendError != nil {
		return framer.sendError
	}

	_, framer.sendError = framer.writer.Write(framer.sendAllDataBuf)
	return framer.sendError
}

func (framer *BinaryFramer) Close()  {
	framer.writer.Close()
}