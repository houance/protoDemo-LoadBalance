package internal

import (
	"context"
	clientside "houance/protoDemo-LoadBalance/internal/clientSide"
	consultool "houance/protoDemo-LoadBalance/internal/consulTool"
	"houance/protoDemo-LoadBalance/internal/helper"
	"houance/protoDemo-LoadBalance/internal/innerData"
	lb "houance/protoDemo-LoadBalance/internal/loadBalancer"
	netcommon "houance/protoDemo-LoadBalance/internal/netCommon"
	serverside "houance/protoDemo-LoadBalance/internal/serverSide"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// LB Middleward cost way less than 1ms
// newest change, use aliyun inter connection setting(like localhost),
// latency down to 50 ms
func StartAllComponent() {

	// config
	config, err := helper.ReadConf("../config.yaml")
	if err != nil {
		panic(err)
	}

	// all Channels
	serverRegisterChannel := make(chan *innerData.InnerDataForward, config.S.InfoChannelSize)
	clientRegisterChannel := make(chan *innerData.InnerDataBackward, config.S.InfoChannelSize)
	addressChannel := make(chan string, config.S.InfoChannelSize)
	forwardChannel := make(chan *innerData.InnerDataTransfer, config.S.LBChannelSize)
	backwardChannel := make(chan *innerData.InnerDataTransfer, config.S.LBChannelSize)
	channelCounter := netcommon.NewChannelCounter(config.S.ConnectionChannelSize)

	// all Map
	addressChannelMap := make(map[string]chan *innerData.InnerDataTransfer)
	idChannelMap := make(map[uint32]chan *innerData.InnerDataTransfer)

	// for managing all components
	g, ctx := errgroup.WithContext(context.Background())

	// logger
	logger := zap.NewExample()

	// consul init
	err = consultool.NewConsulClient(config.C.Server.IP, config.C.Server.Port)
	if err != nil {
		panic(err)
	}
	err = consultool.Register(
		"golangLB",
		config.N.IP,
		config.N.LBPort,
		config.C.Tags,
		config.C.Check.Interval,
		config.C.Check.Timeout,
		config.C.Check.HttpHealthCheckUrl)
	if err != nil {
		panic(err)
	}

	g.Go(func() error {
		return consultool.HealthCheckHttpServer(
			config.N.HealthPort,
			config.N.HealthCheckPath,
		)
	})

	g.Go(func() error {
		return consultool.ScheduleHealthCheck(
			ctx,
			addressChannelMap,
			addressChannel,
			config.C.Filter,
		)
	})

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
			config.N.LBPort,
			forwardChannel,
			clientRegisterChannel,
			channelCounter,
			config.S.InfoChannelSize,
			config.S.DataChannelSize,
		)
	})

	g.Go(func() error {
		return serverside.ServersManager(
			logger,
			ctx,
			addressChannel,
			backwardChannel,
			serverRegisterChannel,
			config.S.InfoChannelSize,
			config.S.DataChannelSize,
		)
	})

	panic(g.Wait())
}
