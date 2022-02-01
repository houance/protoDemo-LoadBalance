package loadbalancer

import (
	"context"
	"errors"
	"houance/protoDemo-LoadBalance/internal/innerData"
	"time"


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
				delete(addressChannelMap, idfw.Address)
				logger.Warn("Server DeRegist", zap.String("Address", idfw.Address))
			} else {
				addressChannelMap[idfw.Address] = idfw.Channel
				logger.Info("Server Regist", zap.String("Address", idfw.Address))
			}

		case idbw = <- clientRegisterChannel:
			if idbw.Channel == nil {
				delete(idChannelMap, idbw.StreamID)
				logger.Info("Client DeRegist", zap.Uint32("StreamID", idbw.StreamID))
			} else {
				idChannelMap[idbw.StreamID] = idbw.Channel
				logger.Info("Client Regist", zap.Uint32("StreamID", idbw.StreamID))
			}

		case idtfFromClient = <- forwardChannel:
			address, err = lbAlgorithm(addressChannelMap)
			if err != nil {
				return err
			}
			addressChannelMap[address] <- idtfFromClient
			logger.Info(
				"Recv From PyClient, Send To PyServer",
				zap.String("time", time.Now().String()),
			)


		case idtfFromServer = <- backwardChannel:
			idChannelMap[idtfFromServer.InnerHeader.StreamID] <- idtfFromServer
			logger.Info(
				"Recv From PyServer, Send To PyClient",
				zap.String("time", time.Now().String()),
			)
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