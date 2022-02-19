package trace

import (
	"houance/protoDemo-LoadBalance/external"
	binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"
	"houance/protoDemo-LoadBalance/internal/innerData"
	"sync"
)

func Receiver(
	f *binaryframer.BinaryFramer,
	outChannel chan *innerData.DataTrace,
	spanPool *sync.Pool,
	prefixLengthPool *sync.Pool,
	iddtPool *sync.Pool,
) (err error) {

	for {
		iddt := iddtPool.Get().(*innerData.DataTrace)
		spanInfo := spanPool.Get().(*external.SpanInfo)
		prefixLength := prefixLengthPool.Get().(*external.PrefixLength)

		err = f.RecvSpan(prefixLength, spanInfo)
		if err != nil {
			return
		}
		iddt.PrefixLength = prefixLength
		iddt.Span = spanInfo
		outChannel <- iddt
	}
}
