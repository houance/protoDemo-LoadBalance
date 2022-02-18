package clientside

import (
	"context"
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
	netcommon "houance/protoDemo-LoadBalance/internal/netCommon"
	"net"
	"strconv"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func SocketServer(
	logger *zap.Logger,
	ctx context.Context,
	listenPort int,
	forwardChannel chan *innerData.DataTransfer,
	clientRegisterChannel chan *innerData.DataBackward,
	counter *netcommon.ChannelCounter,
	infoChannelSize int,
	dataChannelSize int) error {

	var (
		errsChannel         chan error    = make(chan error, infoChannelSize)
		acceptErrorsChannel chan error    = make(chan error, infoChannelSize)
		connectionsChannel  chan net.Conn = make(chan net.Conn, infoChannelSize)
		f                   *binaryframer.BinaryFramer
		err                 error
		idbw                *innerData.DataBackward = &innerData.DataBackward{}
		con                 net.Conn
		ls                  net.Listener
		listenAddress       string = "0.0.0.0:" + strconv.Itoa(listenPort)
	)

	defer close(errsChannel)
	defer close(acceptErrorsChannel)
	defer close(connectionsChannel)

	ls, err = net.Listen("tcp", listenAddress)
	if err != nil {
		logger.Error("Client Socket Server Listen Failed",
			zap.Error(err),
		)
		return err
	} else {
		logger.Info("Client Server Start Listening",
			zap.String("Address", listenAddress),
		)
		go func() {
			AcceptGoroutine(
				ls,
				connectionsChannel,
				acceptErrorsChannel,
			)
		}()
	}
	defer ls.Close()

	for {
		select {
		case <-ctx.Done():
			logger.Error("Outside Distrupt, Return From ClientSocketServer")
			return ctx.Err()

		case err = <-errsChannel:
			csem, ok := err.(*ClientSideErrorMessage)
			if ok {
				// idbw with StreamID only
				// for deregistation
				idbw.StreamID = csem.StreamID
				clientRegisterChannel <- idbw

				logger.Info("Client Disconnect",
					zap.Uint32("Client ID", csem.StreamID),
					zap.Error(csem),
				)

				csem.Bes.Framer.Close()
			}

		case con = <-connectionsChannel:
			f, err = binaryframer.NewBinaryFramer(con, 5, logger)
			if err != nil {
				logger.Warn("Binary Framer Init Failed", zap.Error(err))
				break
			}

			id, err := counter.Get()
			if err != nil {
				logger.Warn("Reach Max Connection", zap.Int("Size", counter.Size()))
				break
			}

			tmpGroup := startClientSideGoroutine(
				f,
				forwardChannel,
				clientRegisterChannel,
				id,
				dataChannelSize,
			)

			go func() {
				errsChannel <- tmpGroup.Wait()
			}()
			logger.Info("Client Connect, Start Client Goroutine")

		case err = <-acceptErrorsChannel:
			logger.Error("Client Socket Server Accept() Method Error",
				zap.Error(err),
			)
		}
	}
}

func AcceptGoroutine(listener net.Listener,
	connectionsChannel chan net.Conn,
	acceptErrorsChannel chan error) {

	for {
		con, err := listener.Accept()
		if err != nil {
			acceptErrorsChannel <- err
			continue
		} else {
			connectionsChannel <- con
		}
	}
}

func startClientSideGoroutine(
	framer *binaryframer.BinaryFramer,
	forwardChannel chan *innerData.DataTransfer,
	registerChannel chan *innerData.DataBackward,
	id uint32,
	dataChannelSize int) *errgroup.Group {

	sendChannel := make(chan *innerData.DataTransfer, dataChannelSize)

	tmpGroup, tmpctx := errgroup.WithContext(context.Background())
	tmpGroup.Go(func() error {
		return ClientReceiver(framer, forwardChannel, tmpctx)
	})
	tmpGroup.Go(func() error {
		return ClientSender(framer, sendChannel, tmpctx, id)
	})

	idbw := &innerData.DataBackward{StreamID: id, Channel: sendChannel}
	registerChannel <- idbw

	return tmpGroup
}
