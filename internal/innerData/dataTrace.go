package innerData

import (
	"houance/protoDemo-LoadBalance/external"

	"google.golang.org/protobuf/proto"
)

type DataTrace struct {
	Span         *external.SpanInfo
	PrefixLength *external.PrefixLength
	lengthData   []byte
	spanByteData []byte
	err          error
	allLength    int
	allByteData  []byte
}

func (dt *DataTrace) Serilize() ([]byte, error) {
	dt.spanByteData, dt.err = proto.Marshal(dt.Span)
	if dt.err != nil {
		return nil, dt.err
	}
	dt.PrefixLength.Length = uint32(len(dt.spanByteData))
	dt.lengthData, dt.err = proto.Marshal(dt.PrefixLength)
	if dt.err != nil {
		return nil, dt.err
	}

	dt.allLength = len(dt.lengthData) + len(dt.spanByteData)

	if dt.allLength > len(dt.allByteData) {
		dt.allByteData = make([]byte, dt.allLength)
	}

	copy(dt.allByteData, dt.lengthData)
	copy(dt.allByteData[len(dt.lengthData):], dt.spanByteData)

	return dt.allByteData[:dt.allLength], nil

}
