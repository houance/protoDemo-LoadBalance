package clientside

import (
	"context"
	message "houance/protoDemo-LoadBalance/external"
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
	netcommon "houance/protoDemo-LoadBalance/internal/netCommon"
)

func ClientSender(
	framer *binaryframer.BinaryFramer,
	inChannel chan *innerData.DataTransfer,
	ctx context.Context,
	id uint32,
) error {

	var (
		err    error
		idtf   *innerData.DataTransfer = &innerData.DataTransfer{}
		bes    *netcommon.BasicErrorMessage = &netcommon.BasicErrorMessage{}
		csem   *ClientSideErrorMessage      = &ClientSideErrorMessage{}
		header *message.Header              = &message.Header{StreamID: id, Length: id}
	)

	err = framer.SendHeader(header)
	if err != nil {
		bes.Wrap(framer, err)
		csem.Wrap(bes, header.StreamID)
		return csem
	}

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
