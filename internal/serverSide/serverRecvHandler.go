package serverside

import (
	"context"
	message "houance/protoDemo-LoadBalance/external"
	"houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
)

func ServerRecvHandler(
	address string,
	ctx context.Context, 
	framer *binaryframer.BinaryFramer,
	muxerChannel chan *innerData.InnerDataTransfer,
	) error {
	
	var (
		err error
		header *message.Header = &message.Header{}
		data []byte
		idata  *innerData.InnerDataTransfer = &innerData.InnerDataTransfer{}
		ems *ServerSideErrorMessage = &ServerSideErrorMessage{}
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = framer.RecvHeader(header)
			if err != nil {
				ems.Wrap(address, err)
				framer.Close()
				return ems
			}

			data, err = framer.RecvBytes(header)
			if err != nil {
				ems.Wrap(address, err)
				framer.Close()
				return ems
			}

			idata.InnerHeader = header
			idata.Data = data
			muxerChannel <- idata
		}
	}
}