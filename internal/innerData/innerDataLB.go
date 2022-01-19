package innerData

type InnerDataLB struct {
	Address string
	Channel chan *InnerDataTransfer
}

func (idlb *InnerDataLB) register(addressChannelMap map[string]chan *InnerDataTransfer)  {
	addressChannelMap[idlb.Address] = idlb.Channel
}

func (idlb *InnerDataLB) deRegister(addressChannelMap map[string]chan *InnerDataTransfer) {
	delete(addressChannelMap, idlb.Address)
	close(idlb.Channel)
}

func (idlb *InnerDataLB) Handle (addressChannelMap map[string]chan *InnerDataTransfer) {
	if idlb.Channel==nil  {
		idlb.deRegister(addressChannelMap)
	}else {
		idlb.register(addressChannelMap)
	}
}