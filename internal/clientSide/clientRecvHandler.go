package clientside

import (
	"fmt"
	message "houance/protoDemo-LoadBalance/external"
	framer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
	"net"
	"sync"
)

func ClientRecvHandler(
	con net.Conn, 
	lbChannel chan *innerData.InnerDataTransfer,
	clientRegisterChannel chan *InnerDataMuxer) {
	
	var (
		f *framer.BinaryFramer = framer.NewBinaryFramer(con, 1)
		header *message.Header = &message.Header{}
		innerD *innerData.InnerDataTransfer = &innerData.InnerDataTransfer{}
		data []byte
		once sync.Once
		err error
		idmx *InnerDataMuxer = &InnerDataMuxer{}
	)

	for {

		err = f.RecvHeader(header)
		if err != nil{
			fmt.Println(err)
			return
		}

		once.Do(func() {
			idmx.StreamID = header.StreamID
			idmx.Framer = f
			clientRegisterChannel <- idmx
		})

		data, err = f.RecvBytes(header)
		if err != nil {
			fmt.Println(err)
			idmx.Framer = nil
			clientRegisterChannel <- idmx
			return
		}

		innerD.InnerHeader = header
		innerD.Data = data
		lbChannel <- innerD
	}
}