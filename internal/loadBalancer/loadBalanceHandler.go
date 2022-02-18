package loadbalancer

import (
	"context"
	"errors"
	"houance/protoDemo-LoadBalance/external"
	"houance/protoDemo-LoadBalance/internal/helper"
	"houance/protoDemo-LoadBalance/internal/innerData"
	"houance/protoDemo-LoadBalance/internal/trace"
	"time"

	"go.uber.org/zap"
)

var (
	traceByte []byte
)

// Start Load-Balancer
// Handle Client Registation, Server Registation, Data Forward and Backward
func LBHandler(
	logger *zap.Logger,
	ctx context.Context,
	serverRegisterChannel chan *innerData.DataForward,
	clientRegisterChannel chan *innerData.DataBackward,
	addressChannelMap map[string]chan *innerData.DataTransfer,
	idChannelMap map[uint32]chan *innerData.DataTransfer,
	forwardChannel chan *innerData.DataTransfer,
	backwardChannel chan *innerData.DataTransfer,
	randomGenerate *helper.RandomGenerator,
	batchProcess *trace.BatchProcess,
) (err error) {

	var (
		idtfFromClient *innerData.DataTransfer = &innerData.DataTransfer{}
		idtfFromServer *innerData.DataTransfer = &innerData.DataTransfer{}
		spanInfo       *external.SpanInfo      = &external.SpanInfo{}
		prefixLength   *external.PrefixLength  = &external.PrefixLength{}
		iddt           *innerData.DataTrace    = &innerData.DataTrace{}
		name           string                  = "lb"
		statuOne       string                  = "clbs"
		statuTwo       string                  = "slbc"
		address        string
	)

	logger.Info("Load Balance Start")

	for {
		select {
		case <-ctx.Done():
			logger.Error("Outside Distrupt, Return From LB")
			return ctx.Err()

		case idfw := <-serverRegisterChannel:
			if idfw.Channel == nil {
				close(addressChannelMap[idfw.Address])
				delete(addressChannelMap, idfw.Address)
				logger.Warn("Server DeRegist", zap.String("Address", idfw.Address))
			} else {
				addressChannelMap[idfw.Address] = idfw.Channel
				logger.Info("Server Regist", zap.String("Address", idfw.Address))
			}

		case idbw := <-clientRegisterChannel:
			if idbw.Channel == nil {
				close(idChannelMap[idbw.StreamID])
				delete(idChannelMap, idbw.StreamID)
				logger.Info("Client DeRegist", zap.Uint32("StreamID", idbw.StreamID))
			} else {
				idChannelMap[idbw.StreamID] = idbw.Channel
				logger.Info("Client Regist", zap.Uint32("StreamID", idbw.StreamID))
			}

		default:
			select {
			case idtfFromClient = <-forwardChannel:
				StartSpan(
					idtfFromClient,
					spanInfo,
					randomGenerate,
					name,
					statuOne,
				)

				address, err = lbAlgorithm(addressChannelMap)
				if err != nil {
					return err
				}
				addressChannelMap[address] <- idtfFromClient

				err = EndSpan(
					spanInfo,
					batchProcess,
					iddt,
					prefixLength,
				)
				if err != nil {
					return err
				}

			case idtfFromServer = <-backwardChannel:
				StartSpan(
					idtfFromServer,
					spanInfo,
					randomGenerate,
					name,
					statuTwo,
				)

				idChannelMap[idtfFromServer.InnerHeader.StreamID] <- idtfFromServer

				err = EndSpan(
					spanInfo,
					batchProcess,
					iddt,
					prefixLength,
				)
				if err != nil {
					return err
				}

			default:
				break
			}
		}
	}

}

func lbAlgorithm(addressChannelMap map[string]chan *innerData.DataTransfer) (string, error) {

	var address string

	if len(addressChannelMap) == 0 {
		return "nil", errors.New("no server available")
	}

	// TODO
	// implement smarter LB Algorithm
	for address = range addressChannelMap {
		break
	}
	return address, nil
}

func StartSpan(
	idtf *innerData.DataTransfer,
	spanInfo *external.SpanInfo,
	randomGenerato *helper.RandomGenerator,
	name string,
	status string) {

	spanInfo.StreamID = idtf.InnerHeader.StreamID
	spanInfo.TraceID = idtf.InnerHeader.TraceID
	spanInfo.ParentID = idtf.InnerHeader.SpanID
	spanInfo.SpanID = randomGenerato.RandomNumber(idtf.InnerHeader.StreamID)
	spanInfo.Name = name
	spanInfo.Status = status
	spanInfo.StartTime = time.Now().UnixMicro()
}

func EndSpan(
	spanInfo *external.SpanInfo,
	batchProcess *trace.BatchProcess,
	dataTrace *innerData.DataTrace,
	prefixLength *external.PrefixLength) (err error) {

	spanInfo.EndTime = time.Now().UnixMicro()

	dataTrace.PrefixLength = prefixLength
	dataTrace.Span = spanInfo
	traceByte, err = dataTrace.Serilize()
	if err != nil {
		return err
	}
	batchProcess.SendSpan(traceByte)
	return nil
}
