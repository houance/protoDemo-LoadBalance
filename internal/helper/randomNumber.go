package helper

import (
	"math/rand"
)

type RandomGenerator struct {
	sourceMap map[uint32]*rand.Rand
}

func NewRandomGenerator(streamIDStart uint32, streamIDEnd uint32) *RandomGenerator {
	tmp := make(map[uint32]*rand.Rand)
	for i := streamIDStart; i < streamIDEnd; i++ {
		tmp[i] = rand.New(rand.NewSource(int64(i)))
	}
	return &RandomGenerator{
		sourceMap: tmp,
	}
}

func (r *RandomGenerator) RandomNumber(streamID uint32) uint32 {
	return r.sourceMap[streamID].Uint32()
}
