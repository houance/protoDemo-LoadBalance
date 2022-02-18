package clientside

import (
	"context"
	message "houance/protoDemo-LoadBalance/external"
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
	netcommon "houance/protoDemo-LoadBalance/internal/netCommon"
)

func ClientReceiver(
	framer *binaryframer.BinaryFramer,
	outChannel chan *innerData.DataTransfer,
	ctx context.Context,
) error {

	var (
		err    error
		bes    *netcommon.BasicErrorMessage = &netcommon.BasicErrorMessage{}
		csem   *ClientSideErrorMessage      = &ClientSideErrorMessage{}
		header *message.Header              = &message.Header{}
		data   []byte
		idtf   *innerData.DataTransfer = &innerData.DataTransfer{}
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = framer.RecvHeader(header)
			if err != nil {
				bes.Wrap(framer, err)
				csem.Wrap(bes, header.StreamID)
				return csem
			}

			data, err = framer.RecvBytes(header)
			if err != nil {
				bes.Wrap(framer, err)
				csem.Wrap(bes, header.StreamID)
				return csem
			}

			idtf.InnerHeader = header
			idtf.Data = data
			outChannel <- idtf
		}
	}
}
