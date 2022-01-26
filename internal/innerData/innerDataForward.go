package innerData

type InnerDataForward struct {
	Address string
	Channel chan *InnerDataTransfer
}