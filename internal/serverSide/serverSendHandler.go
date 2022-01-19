package serverside

import (
	"context"
	"houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
)

func ServerSendHandler(
	address string,
	ctx context.Context,
	framer *binaryframer.BinaryFramer,
	sendChannel chan *innerData.InnerDataTransfer,
) error {

	var (
		err error
		idtf *innerData.InnerDataTransfer = &innerData.InnerDataTransfer{}
		ssem *ServerSideErrorMessage = &ServerSideErrorMessage{}
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case idtf = <- sendChannel:
			err = framer.SendInnerData(idtf)
			if err != nil {
				ssem.Wrap(address, err)
				framer.Close()
				return ssem
			}
		}
	}
}