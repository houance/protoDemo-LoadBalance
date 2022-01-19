package internal

import (
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	clientside "houance/protoDemo-LoadBalance/internal/clientSide"
	"houance/protoDemo-LoadBalance/internal/innerData"
	serverside "houance/protoDemo-LoadBalance/internal/serverSide"
	lb "houance/protoDemo-LoadBalance/internal/loadBalancer"
)

func StartAllComponent(listenAddress string,
	testServerAddress string) {
	
	// all Channels
	serverRegisterChannel := make(chan *innerData.InnerDataLB, 10)
	clientRegisterChannel := make(chan *clientside.InnerDataMuxer, 10)
	lbChannel := make(chan *innerData.InnerDataTransfer, 100)
	muxerChannel := make(chan *innerData.InnerDataTransfer, 100)
	addressChannel := make(chan string, 10)


	// all Map
	addressChannelMap := make(map[string]chan *innerData.InnerDataTransfer)
	idFramerMap := make(map[uint32]*binaryframer.BinaryFramer)

	// for blocking
	blockingChannel := make(chan int)

	go func() {
		clientside.SocketServer(
			listenAddress,
			lbChannel,
			clientRegisterChannel,
		)
	}()

	go func() {
		serverside.HealthCheck(
			addressChannel,
			muxerChannel,
			serverRegisterChannel,
		)
	}()

	go func() {
		lb.LBHandler(
			serverRegisterChannel,
			clientRegisterChannel,
			addressChannelMap,
			idFramerMap,
			lbChannel,
			muxerChannel,
		)
	}()

	addressChannel <- testServerAddress

	<- blockingChannel
}