package netcommon

import (
	"context"
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
)


func Sender(
	framer *binaryframer.BinaryFramer,
	inChannel chan *innerData.InnerDataTransfer,
	ctx context.Context,
) error  {

	var (
		err error
		idtf *innerData.InnerDataTransfer = &innerData.InnerDataTransfer{}
		bes *BasicErrorMessage = &BasicErrorMessage{}
	)

	defer close(inChannel)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case idtf = <- inChannel:
			err = framer.SendInnerData(idtf)
			if err != nil {
				bes.Wrap(framer, err)
				return bes
			}
		}
	}
}