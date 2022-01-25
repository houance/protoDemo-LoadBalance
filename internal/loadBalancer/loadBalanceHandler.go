package loadbalancer

import (
	"context"
	"errors"
	"houance/protoDemo-LoadBalance/internal/innerData"

	"go.uber.org/zap"
)

// Start Load-Balancer
// Handle Client Registation, Server Registation, Data Forward and Backward
func LBHandler(
	logger *zap.Logger,
	ctx context.Context,
	serverRegisterChannel chan *innerData.InnerDataForward,
	clientRegisterChannel chan *innerData.InnerDataBackward,
	addressChannelMap map[string]chan *innerData.InnerDataTransfer,
	idChannelMap map[uint32]chan *innerData.InnerDataTransfer,
	forwardChannel chan *innerData.InnerDataTransfer,
	backwardChannel chan *innerData.InnerDataTransfer,
) error {

	var (
		idfw *innerData.InnerDataForward = &innerData.InnerDataForward{}
		idbw *innerData.InnerDataBackward = &innerData.InnerDataBackward{}
		idtfFromClient *innerData.InnerDataTransfer = &innerData.InnerDataTransfer{}
		idtfFromServer *innerData.InnerDataTransfer = &innerData.InnerDataTransfer{}
		address string
		err error
	)

	logger.Info("Load Balance Start")

	for {
		select {
		case <- ctx.Done():
			logger.Error("Outside Distrupt")
			return ctx.Err()

		case idfw = <- serverRegisterChannel:
			if idfw.Channel == nil {
				logger.Warn("Server DeRegist", zap.String("Address", idfw.Address))
				idfw.DeRegister(addressChannelMap)
			} else {
				logger.Info("Server Regist", zap.String("Address", idfw.Address))
				idfw.Register(addressChannelMap)
			}

		case idbw = <- clientRegisterChannel:
			if idbw.Channel == nil {
				logger.Info("Client DeRegist", zap.Uint32("StreamID", idbw.StreamID))
				idbw.DeRegister(idChannelMap)
			} else {
				logger.Info("Client Regist", zap.Uint32("StreamID", idbw.StreamID))
				idbw.Register(idChannelMap)
			}

		case idtfFromClient = <- forwardChannel:
			address, err = lbAlgorithm(addressChannelMap)
			if err != nil {
				return err
			}
			addressChannelMap[address] <- idtfFromClient
			logger.Debug("Recv From Client",
			zap.Uint32("StreamID", idtfFromClient.InnerHeader.StreamID))

		case idtfFromServer = <- backwardChannel:
			idChannelMap[idtfFromServer.InnerHeader.StreamID] <- idtfFromServer
			logger.Debug("Recv From Server",
			zap.Uint32("StreamID", idtfFromServer.InnerHeader.StreamID))
		}
	}

}

func lbAlgorithm(addressChannelMap map[string]chan *innerData.InnerDataTransfer) (address string, err error) {

	if len(addressChannelMap) == 0 {
		return "nil", errors.New("no server available")
	}

	// TODO
	// implement smarter LB Algorithm
	for address = range addressChannelMap {	return }
	return
}