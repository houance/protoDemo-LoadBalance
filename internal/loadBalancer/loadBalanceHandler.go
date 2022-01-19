package loadbalancer

import (
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/clientSide"
	"houance/protoDemo-LoadBalance/internal/innerData"

)

// Start Load-Balancer
// Handle Client Registation, Server Registation, Data Forward and Backward
func LBHandler(
	serverRegisterChannel chan *innerData.InnerDataLB,
	clientRegisterChannel chan *clientside.InnerDataMuxer,
	addressChannelMap map[string]chan *innerData.InnerDataTransfer,
	idFramerMap map[uint32]*binaryframer.BinaryFramer,
	lbChannel chan *innerData.InnerDataTransfer,
	muxerChannel chan *innerData.InnerDataTransfer,
)  {

	var (
		idlb *innerData.InnerDataLB = &innerData.InnerDataLB{}
		idmx *clientside.InnerDataMuxer = &clientside.InnerDataMuxer{}
		idtfFromClient *innerData.InnerDataTransfer = &innerData.InnerDataTransfer{}
		idtfFromServer *innerData.InnerDataTransfer = &innerData.InnerDataTransfer{}
		address string
		f *binaryframer.BinaryFramer
	)
	
	for {
		select {
		case idlb = <- serverRegisterChannel:
			idlb.Handle(addressChannelMap)
		case idmx = <- clientRegisterChannel:
			idmx.Handle(idFramerMap)
		case idtfFromClient = <- lbChannel:
			address = lbAlgorithm(addressChannelMap)
			addressChannelMap[address] <- idtfFromClient
		case idtfFromServer = <- muxerChannel:
			f = idFramerMap[idtfFromServer.InnerHeader.StreamID]
			go func() {
				f.SendInnerData(idtfFromServer)
			}()
		}
	}

}

func lbAlgorithm(addressChannelMap map[string]chan *innerData.InnerDataTransfer) (address string) {

	// TODO
	// implement smarter LB Algorithm
	for address = range addressChannelMap {	return }
	return
}