package innerData

type InnerDataForward struct {
	Address string
	Channel chan *InnerDataTransfer
}

func (idfw *InnerDataForward) Register(addressChannelMap map[string]chan *InnerDataTransfer)  {
	addressChannelMap[idfw.Address] = idfw.Channel
}

func (idfw *InnerDataForward) DeRegister(addressChannelMap map[string]chan *InnerDataTransfer) {
	delete(addressChannelMap, idfw.Address)
	close(idfw.Channel)
}