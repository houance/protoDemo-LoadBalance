package trace

import (
	"context"
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
	"net"
	"strconv"
	"sync"

	"go.uber.org/zap"
)

func TraceServer(
	logger *zap.Logger,
	ctx context.Context,
	listenPort int,
	outChannel chan *innerData.DataTrace,
	infoChannelSize int,
	spanInfoPool *sync.Pool,
	prefixLengthPool *sync.Pool,
	dataTracePool *sync.Pool) (err error) {

	var (
		errsChannel         chan error    = make(chan error, infoChannelSize)
		acceptErrorsChannel chan error    = make(chan error, infoChannelSize)
		connectionsChannel  chan net.Conn = make(chan net.Conn, infoChannelSize)
		f                   *binaryframer.BinaryFramer
		con                 net.Conn
		ls                  net.Listener
		listenAddress       string = "0.0.0.0:" + strconv.Itoa(listenPort)
	)

	defer close(errsChannel)
	defer close(acceptErrorsChannel)
	defer close(connectionsChannel)

	// listen
	ls, err = net.Listen("tcp", listenAddress)
	if err != nil {
		logger.Error("Trace Server Listen Failed",
			zap.Error(err),
		)
		return err
	} else {
		logger.Info("Trace Server Start Listening",
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
			logger.Error("Outside Distrupt, Return From TraceServer")
			return ctx.Err()

		case err = <-errsChannel:
			logger.Info("MiddleWare Client Exit", zap.Error(err))

		case con = <-connectionsChannel:
			f, err = binaryframer.NewBinaryFramer(con, 1, logger)
			if err != nil {
				logger.Warn("Binary Framer Init Failed", zap.Error(err))
				break
			}

			go func() {
				errsChannel <- Receiver(
					f,
					outChannel,
					spanInfoPool,
					prefixLengthPool,
					dataTracePool,
				)
			}()

			logger.Info("MiddleWare Connect, Start Trace Recv Goroutine")

		case err = <-acceptErrorsChannel:
			logger.Error("Trace Server Accept() Method Error",
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
