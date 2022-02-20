package innerData

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type TraceAggregation struct {
	StreamID uint32
	complete bool
	done     chan bool
	timeout  bool
	allLog   []*DataTrace
	ticker   *time.Ticker
	latency  []int64
}

func NewTraceAggregation() *TraceAggregation {
	return &TraceAggregation{
		latency: make([]int64, 5),
		done:    make(chan bool),
	}
}

func (ta *TraceAggregation) CaculateLatency() []int64 {

	sort.Sort(DataTraces(ta.allLog))

	// client all latency
	ta.latency[0] = ta.allLog[0].Span.EndTime - ta.allLog[0].Span.StartTime
	// client send to lb latency
	ta.latency[1] = ta.allLog[1].Span.StartTime - ta.allLog[0].Span.StartTime
	// lb send to server latency
	ta.latency[2] = ta.allLog[2].Span.StartTime - ta.allLog[1].Span.EndTime
	// server send to lb latency
	ta.latency[3] = ta.allLog[3].Span.StartTime - ta.allLog[2].Span.EndTime
	// lb send to client latency
	ta.latency[4] = ta.allLog[0].Span.EndTime - ta.allLog[3].Span.EndTime

	return ta.latency
}

func (ta *TraceAggregation) Append(iddt *DataTrace) (bool, error) {

	if ta.timeout {
		return false, errors.New("timeout")
	}

	ta.allLog = append(ta.allLog, iddt)
	if len(ta.allLog) == 4 {
		ta.complete = true
		ta.done <- true
		ta.timeout = false
		ta.StreamID = iddt.Span.StreamID
	}
	return ta.complete, nil
}

func (ta *TraceAggregation) SetTicker(duration time.Duration) {
	ta.ticker = time.NewTicker(duration)
	go func() {
		select {
		case <-ta.ticker.C:
			ta.timeout = true
			break
		case <-ta.done:
			ta.ticker.Stop()
			break
		}
	}()
}

func (ta *TraceAggregation) ResetAllLog(
	spanInfoPool *sync.Pool,
	prefixLengthPool *sync.Pool,
	dataTracePool *sync.Pool,
	traceArregationPool *sync.Pool,
) {

	for _, v := range ta.allLog {
		spanInfoPool.Put(v.Span)
		prefixLengthPool.Put(v.PrefixLength)
		dataTracePool.Put(v)
	}
	traceArregationPool.Put(ta)
}
