package clientside

import binaryframer "houance/protoDemo-LoadBalance/internal/binaryFramer"

type InnerDataMuxer struct {
	StreamID uint32
	Framer *binaryframer.BinaryFramer
}

func (idmx *InnerDataMuxer) regist(idFramerMap map[uint32]*binaryframer.BinaryFramer)  {
	idFramerMap[idmx.StreamID] = idmx.Framer
}

func (idmx *InnerDataMuxer) deRegist(idFramerMap map[uint32]*binaryframer.BinaryFramer)  {
	delete(idFramerMap, idmx.StreamID)
	idmx.Framer.Close()
}

func (idmx *InnerDataMuxer) Handle(idFramerMap map[uint32]*binaryframer.BinaryFramer) {
	if idmx.Framer == nil {
		idmx.deRegist(idFramerMap)
	}else {
		idmx.regist(idFramerMap)
	}
}