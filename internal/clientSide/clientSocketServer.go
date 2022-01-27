package clientside

import (
	"context"
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
	netcommon "houance/protoDemo-LoadBalance/internal/netCommon"
	"net"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func SocketServer(
	logger *zap.Logger,
	ctx context.Context,
	listenAddress string,
	forwardChannel chan *innerData.InnerDataTransfer,
	clientRegisterChannel chan *innerData.InnerDataBackward) error {

	var (
		errsChannel chan error
		f *binaryframer.BinaryFramer
		err error
		idbw *innerData.InnerDataBackward = &innerData.InnerDataBackward{}
	)

	defer close(errsChannel)

	ls, err := net.Listen("tcp", listenAddress)
	if err != nil {
		logger.Error("Client Socket Server Listen Failed",
		zap.Error(err),
		)
		return err
	}
	logger.Info("Client Server Start Listening",
		zap.String("Address", listenAddress),
	)

	for {
		select {
		case <-ctx.Done():
			logger.Warn("Outside Distrupt")
			return ctx.Err()

		case err = <- errsChannel:
			re,ok := err.(*ClientSideErrorMessage)
			if ok {
				// idbw with StreamID only
				// for deregistation
				idbw.StreamID = re.StreamID
				idbw.Channel = nil
				clientRegisterChannel <- idbw

				logger.Info("Client Disconnect",
				zap.String("Address", re.Bes.Framer.GetRemoteAddress()),
				zap.Error(re),
				)
				re.Bes.Framer.Close()
			}

		default:
			con, err := ls.Accept()
			if err != nil {
				logger.Error("Client Socket Server Accept() Method Error",
				zap.Error(err),
				)
				return err
			}
			

			f = binaryframer.NewBinaryFramer(con, 5)

			tmpGroup := startClientSideGoroutine(
				f,
				forwardChannel,
				clientRegisterChannel,
			)
			go func() {
				errsChannel <- tmpGroup.Wait()
			}()
			logger.Info("Start Client Goroutine")
		}
	}

}

func startClientSideGoroutine(
	framer *binaryframer.BinaryFramer,
	forwardChannel chan *innerData.InnerDataTransfer,
	registerChannel chan *innerData.InnerDataBackward) (
		*errgroup.Group) {

			sendChannel := make(chan *innerData.InnerDataTransfer, 50)

			tmpGroup, tmpctx := errgroup.WithContext(context.Background())
			tmpGroup.Go(func() error {
				return Receiver(framer, forwardChannel, tmpctx, sendChannel, registerChannel)
			})
			tmpGroup.Go(func() error {
				return netcommon.Sender(framer, sendChannel, tmpctx)
			})

			return tmpGroup
}