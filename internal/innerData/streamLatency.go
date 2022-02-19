package innerData

type StreamLatency struct {
	latency []int64
	counter int
}

func (sl *StreamLatency) AcumulateLatency(latency []int64) {
	sl.counter++
	for k := range sl.latency {
		sl.latency[k] += latency[k]
	}
}

func (sl *StreamLatency) GetAverageLatency() {
	for k := range sl.latency {
		sl.latency[k] = sl.latency[k] / int64(sl.counter)
	}
}
