package innerData

type StreamLatency struct {
	latency []int64
	counter int64
}

func NewStreamLatency() *StreamLatency {
	return &StreamLatency{latency: make([]int64, 5)}
}

func (sl *StreamLatency) AcumulateLatency(latency []int64) {
	sl.counter++
	for k := range sl.latency {
		sl.latency[k] += latency[k]
	}
}

func (sl *StreamLatency) GetAverageLatency() []int64 {
	for k := range sl.latency {
		sl.latency[k] = sl.latency[k] / sl.counter
	}
	return sl.latency
}

func (sl *StreamLatency) SetLatency(tmp []int64) {
	sl.latency = tmp
}

func (sl *StreamLatency) SetCounter(v int64) {
	sl.counter = v
}
