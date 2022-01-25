package innerData

type InnerDataBackward struct {
	StreamID uint32
	Channel chan *InnerDataTransfer
}

func (idbw *InnerDataBackward) Register(idChannelMap map[uint32]chan *InnerDataTransfer)  {
	idChannelMap[idbw.StreamID] = idbw.Channel
}

func (idbw *InnerDataBackward) DeRegister(idChannelMap map[uint32]chan *InnerDataTransfer) {
	delete(idChannelMap, idbw.StreamID)
	close(idbw.Channel)
}