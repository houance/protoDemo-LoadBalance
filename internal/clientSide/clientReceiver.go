package clientside

import (
	"context"
	message "houance/protoDemo-LoadBalance/external"
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
	netcommon "houance/protoDemo-LoadBalance/internal/netCommon"
	"sync"
)

func ClientReceiver(
	framer *binaryframer.BinaryFramer,
	outChannel chan *innerData.InnerDataTransfer,
	ctx context.Context,
	sendChannel chan *innerData.InnerDataTransfer,
	registChannel chan *innerData.InnerDataBackward,
) error {

	var (
		err    error
		bes    *netcommon.BasicErrorMessage = &netcommon.BasicErrorMessage{}
		csem   *ClientSideErrorMessage      = &ClientSideErrorMessage{}
		header *message.Header              = &message.Header{}
		data   []byte
		idtf   *innerData.InnerDataTransfer = &innerData.InnerDataTransfer{}
		do     sync.Once
		regist *innerData.InnerDataBackward = &innerData.InnerDataBackward{}
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = framer.RecvHeader(header)
			if err != nil {
				bes.Wrap(framer, err)
				if data == nil {
					csem.Wrap(bes, 0)
					return csem
				} else {
					csem.Wrap(bes, header.StreamID)
					return csem
				}
			}

			do.Do(func() {
				regist.StreamID = header.StreamID
				regist.Channel = sendChannel
				registChannel <- regist
			})

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
