package serverside

import (
	"context"
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
	"houance/protoDemo-LoadBalance/internal/netCommon"
	"net"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Start ServerSide Goroutines and Register to The Load-Balancer
// Also Keep Tracking of Server Side Goroutines
// When failed, deRegister from Load-Balancer
func HealthCheck(
	logger *zap.Logger,
	ctx context.Context,
	addressChannel chan string,
	backwardChannel chan *innerData.InnerDataTransfer,
	serverRegisterChannel chan *innerData.InnerDataForward) error {


	var (
		errsChannel chan error
		idfw *innerData.InnerDataForward = &innerData.InnerDataForward{}
		ssem *ServerSideErrorMessage = &ServerSideErrorMessage{}
	)

	defer close(errsChannel)

	for {
		select {
		case <- ctx.Done():
			logger.Error("Outside Distrupt")
			return ctx.Err()

		case address := <- addressChannel:
			con, err := net.Dial("tcp", address)
			if err != nil {
				logger.Warn("Server Dial Failed",zap.Error(err))
				continue
			}
			framer := binaryframer.NewBinaryFramer(con, 5)

			tmpGroup := startServerSideGoroutine(
				address,
				framer,
				backwardChannel,
				serverRegisterChannel,
			)

			go func() {
				errsChannel <- tmpGroup.Wait()
			}()
			logger.Info("Start Server Side Goroutine")

		case err := <- errsChannel:
			bes,ok := err.(*netcommon.BasicErrorMessage)
			logger.Warn("Server Encounter Error",zap.Error(bes))
			ssem.Bes = bes
			if ok {
				// idfw with address only
				// for deregistation
				idfw.Address = ssem.Bes.Framer.GetRemoteAddress()
				idfw.Channel = nil
				serverRegisterChannel <- idfw

				logger.Warn("Server Disconnect",
				zap.String("Address", ssem.Bes.Framer.GetRemoteAddress()),
				zap.Error(bes),
				)

				ssem.Bes.Framer.Close()
			}
		}
	}
}

func startServerSideGoroutine(
	address string, 
	framer *binaryframer.BinaryFramer,
	backwardChannel chan *innerData.InnerDataTransfer,
	registerChannel chan *innerData.InnerDataForward) (
		*errgroup.Group) {

			sendChannel := make(chan *innerData.InnerDataTransfer, 50)
			
			tmpGroup, tmpctx := errgroup.WithContext(context.Background())
			tmpGroup.Go(func() error {
				return netcommon.Receiver(framer, backwardChannel, tmpctx)
			})
			tmpGroup.Go(func() error {
				return netcommon.Sender(framer, sendChannel, tmpctx)
			})

			// idfw with address and channel
			// used for registation
			idlb := &innerData.InnerDataForward{Address: address, Channel: sendChannel}
			registerChannel <- idlb

			return tmpGroup
}