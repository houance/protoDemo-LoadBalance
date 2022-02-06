package serverside

import (
	"context"
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
	netcommon "houance/protoDemo-LoadBalance/internal/netCommon"
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
		errsChannel chan error                  = make(chan error, 10)
		idfw        *innerData.InnerDataForward = &innerData.InnerDataForward{}
		ssem        *ServerSideErrorMessage     = &ServerSideErrorMessage{}
	)

	defer close(errsChannel)

	for {
		select {
		case <-ctx.Done():
			logger.Error("Outside Distrupt, Return From healthCheckHandler")
			return ctx.Err()

		case address := <-addressChannel:
			con, err := net.Dial("tcp", address)
			if err != nil {
				logger.Warn("Server Dial Failed", zap.Error(err))
				continue
			}

			framer, err := binaryframer.NewBinaryFramer(con, 5, logger)
			if err != nil {
				logger.Warn("Binary Framer Init Failed", zap.Error(err))
				continue
			}

			tmpGroup := startServerSideGoroutine(
				address,
				framer,
				backwardChannel,
				serverRegisterChannel,
				logger,
			)

			go func() {
				errsChannel <- tmpGroup.Wait()
			}()
			logger.Info("Connect Server Sucessful, Start Server Side Goroutines")

		case err := <-errsChannel:
			bes, ok := err.(*netcommon.BasicErrorMessage)
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
	registerChannel chan *innerData.InnerDataForward,
	logger *zap.Logger) *errgroup.Group {

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
	idfw := &innerData.InnerDataForward{Address: address, Channel: sendChannel}
	registerChannel <- idfw

	return tmpGroup
}
