package internal

import (
	"context"
	clientside "houance/protoDemo-LoadBalance/internal/clientSide"
	"houance/protoDemo-LoadBalance/internal/innerData"
	lb "houance/protoDemo-LoadBalance/internal/loadBalancer"
	serverside "houance/protoDemo-LoadBalance/internal/serverSide"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// LB Middleward cost way less than 1ms
// python client send to golang (~80k), has latency ~18 ms,
// with NAGLE algorithm disable, timeout = None, sendbuf = 128k, recvbuf twice

// golang send to python server(~80k), has latency ~40 ms,
// with NAGLE algorithm disable, sendbuf = 128k, recvbuf twice

// python server send to golang(~1k), has latency ~6ms, with NAGLE algorithm enable

// golang send to python client(~1k), has latency ~12 ms, need enhancement
func StartAllComponent(listenAddress string,
	testServerAddress string) {

	// all Channels
	serverRegisterChannel := make(chan *innerData.InnerDataForward, 10)
	clientRegisterChannel := make(chan *innerData.InnerDataBackward, 10)
	forwardChannel := make(chan *innerData.InnerDataTransfer, 100)
	backwardChannel := make(chan *innerData.InnerDataTransfer, 100)
	addressChannel := make(chan string, 10)

	// all Map
	addressChannelMap := make(map[string]chan *innerData.InnerDataTransfer)
	idChannelMap := make(map[uint32]chan *innerData.InnerDataTransfer)

	// for managing all components
	g, ctx := errgroup.WithContext(context.Background())

	// logger
	logger := zap.NewExample()

	g.Go(func() error {
		return lb.LBHandler(
			logger,
			ctx,
			serverRegisterChannel,
			clientRegisterChannel,
			addressChannelMap,
			idChannelMap,
			forwardChannel,
			backwardChannel,
		)
	})

	g.Go(func() error {
		return clientside.SocketServer(
			logger,
			ctx,
			listenAddress,
			forwardChannel,
			clientRegisterChannel,
		)
	})

	g.Go(func() error {
		return serverside.HealthCheck(
			logger,
			ctx,
			addressChannel,
			backwardChannel,
			serverRegisterChannel,
		)
	})

	addressChannel <- testServerAddress

	panic(g.Wait())
}
