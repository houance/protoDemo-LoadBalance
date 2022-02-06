package clientside

import (
	"context"
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
	netcommon "houance/protoDemo-LoadBalance/internal/netCommon"
)

func ClientSender(
	framer *binaryframer.BinaryFramer,
	inChannel chan *innerData.InnerDataTransfer,
	ctx context.Context,
) error {

	var (
		err  error
		idtf *innerData.InnerDataTransfer = &innerData.InnerDataTransfer{}
		bes  *netcommon.BasicErrorMessage = &netcommon.BasicErrorMessage{}
		csem *ClientSideErrorMessage      = &ClientSideErrorMessage{}
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case idtf = <-inChannel:
			err = framer.SendInnerData(idtf)
			if err != nil {
				bes.Wrap(framer, err)
				csem.Wrap(bes, idtf.InnerHeader.StreamID)
				return csem
			}
		}
	}
}
