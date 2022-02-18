package netcommon

import (
	"context"
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	message "houance/protoDemo-LoadBalance/external"
	"houance/protoDemo-LoadBalance/internal/innerData"
)

func Receiver(
	framer *binaryframer.BinaryFramer,
	outChannel chan *innerData.DataTransfer,
	ctx context.Context,
) error {

	var (
		err error
		bes *BasicErrorMessage = &BasicErrorMessage{}
		header *message.Header = &message.Header{}
		data []byte
		idata  *innerData.DataTransfer = &innerData.DataTransfer{}
	)
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = framer.RecvHeader(header)
			if err != nil {
				bes.Wrap(framer, err)
				return bes
			}

			data, err = framer.RecvBytes(header)
			if err != nil {
				bes.Wrap(framer, err)
				return bes
			}

			idata.InnerHeader = header
			idata.Data = data
			outChannel <- idata
		}
	}
}