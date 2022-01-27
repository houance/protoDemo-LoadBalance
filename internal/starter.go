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
// python client send to golang, has latency ~50 ms, with NAGLE algorithm disable
// golang send to python server, has latency ~64 ms, need enhancement
// python server send to golang, has latency only ~6ms, with NAGLE algorithm enable
// golang send to python client, has latency ~20 ms, need enhancement
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