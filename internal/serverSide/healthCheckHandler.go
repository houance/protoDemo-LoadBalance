package serverside

import (
	"context"
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
	"net"

	"golang.org/x/sync/errgroup"
)

// Start ServerSide Goroutines and Register to The Load-Balancer
// Also Keep Tracking of Server Side Goroutines
// When failed, deRegister from Load-Balancer
func HealthCheck(
	addressChannel chan string,
	muxerChannel chan *innerData.InnerDataTransfer,
	serverRegisterChannel chan *innerData.InnerDataLB)  {

	var errsChannel chan error
	idlb := &innerData.InnerDataLB{}

	defer close(errsChannel)

	for {
		select {
		case address := <- addressChannel:
			con, err := net.Dial("tcp", address)
			if err != nil {
				continue
			}
			framer := binaryframer.NewBinaryFramer(con, 5)

			tmpGroup := startServerSideGoroutine(
				address,
				framer,
				muxerChannel,
				serverRegisterChannel)
	
			go func() {
				errsChannel <- tmpGroup.Wait()
			}()
		case err := <- errsChannel:
			re,ok := err.(*ServerSideErrorMessage)
			if ok {
				// idlb with address only
				// for deregistation
				idlb.Address = re.GetAddress()
				serverRegisterChannel <- idlb
			}
		}
	}
}

func startServerSideGoroutine(
	address string,
	framer *binaryframer.BinaryFramer,
	muxerChannel chan *innerData.InnerDataTransfer,
	registerChannel chan *innerData.InnerDataLB) (
		*errgroup.Group) {

			sendChannel := make(chan *innerData.InnerDataTransfer, 100)
			
			tmpGroup, tmpctx := errgroup.WithContext(context.Background())
			tmpGroup.Go(func() error {
				return ServerRecvHandler(address, tmpctx, framer, muxerChannel)
			})
			tmpGroup.Go(func() error {
				return ServerSendHandler(address, tmpctx, framer, sendChannel)
			})

			// idlb with address and channel
			// used for registation
			idlb := &innerData.InnerDataLB{Address: address, Channel: sendChannel}
			registerChannel <- idlb

			return tmpGroup
}