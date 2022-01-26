package innerData

type InnerDataBackward struct {
	StreamID uint32
	Channel chan *InnerDataTransfer
}
