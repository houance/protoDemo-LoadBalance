package innerData

import (
	message "houance/protoDemo-LoadBalance/external"
	"google.golang.org/protobuf/proto"
)

type DataTransfer struct {
	InnerHeader *message.Header
	Data []byte
	serilizeData []byte
	length int
	headerData []byte
	err error
}

func (innerData *DataTransfer) Serilize() ([]byte , error) {
	innerData.InnerHeader.Length = uint32(len(innerData.Data))
	innerData.headerData, innerData.err= proto.Marshal(innerData.InnerHeader)
	if innerData.err != nil {
		return nil, innerData.err
	}

	innerData.length = len(innerData.headerData) + len(innerData.Data)
	
	if len(innerData.serilizeData) < innerData.length {
		innerData.serilizeData = make([]byte, innerData.length)
	}
	copy(innerData.serilizeData, innerData.headerData)
	copy(innerData.serilizeData[len(innerData.headerData):], innerData.Data)

	return innerData.serilizeData[:innerData.length], innerData.err
}