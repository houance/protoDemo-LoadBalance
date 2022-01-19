package clientside

import (
	"houance/protoDemo-LoadBalance/internal/innerData"
	"net"
)

func SocketServer(
	listenAddress string,
	lbChannel chan *innerData.InnerDataTransfer,
	clientRegisterChannel chan *InnerDataMuxer)  {

	ls, err := net.Listen("tcp", listenAddress)
	if err != nil {
		panic(err)
	}

	for {
		con, err := ls.Accept()
		if err != nil {
			panic(err)
		}
		go func() {
			ClientRecvHandler(
				con,
				lbChannel,
				clientRegisterChannel,
			)
		}()
	}


}