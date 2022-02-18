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
func ServersManager(
	logger *zap.Logger,
	ctx context.Context,
	addressChannel chan string,
	backwardChannel chan *innerData.DataTransfer,
	serverRegisterChannel chan *innerData.DataForward,
	infoChannelSize int,
	dataChannelSize int,
) error {

	var (
		errsChannel chan error                  = make(chan error, infoChannelSize)
		idfw        *innerData.DataForward = &innerData.DataForward{}
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
				dataChannelSize,
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
	backwardChannel chan *innerData.DataTransfer,
	registerChannel chan *innerData.DataForward,
	logger *zap.Logger,
	dataChannelSize int) *errgroup.Group {

	sendChannel := make(chan *innerData.DataTransfer, dataChannelSize)

	tmpGroup, tmpctx := errgroup.WithContext(context.Background())
	tmpGroup.Go(func() error {
		return netcommon.Receiver(framer, backwardChannel, tmpctx)
	})
	tmpGroup.Go(func() error {
		return netcommon.Sender(framer, sendChannel, tmpctx)
	})

	// idfw with address and channel
	// used for registation
	idfw := &innerData.DataForward{Address: address, Channel: sendChannel}
	registerChannel <- idfw

	return tmpGroup
}
