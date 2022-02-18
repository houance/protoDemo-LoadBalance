package innerData

type DataBackward struct {
	StreamID uint32
	Channel  chan *DataTransfer
}
