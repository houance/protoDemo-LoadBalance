package netcommon

import "errors"

type ChannelCounter struct {
	counter chan uint32
}

func NewChannelCounter(size int) *ChannelCounter {
	tmp := make(chan uint32, size)
	for i := 1; i < size; i++ {
		tmp <- uint32(i)
	}
	return &ChannelCounter{counter: tmp}
}

func (c *ChannelCounter) Get() (uint32, error) {
	if len(c.counter) == 0 {
		return 0, errors.New("shortfor")
	} else {
		return <-c.counter, nil
	}
}

func (c *ChannelCounter) Release(size uint32) {
	c.counter <- uint32(size)
}

func (c *ChannelCounter) Size() int {
	return len(c.counter)
}
