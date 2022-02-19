package innerData

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type TraceAggregation struct {
	StreamID    uint32
	complete    bool
	killChannel chan bool
	timeout     bool
	allLog      []*DataTrace
	ticker      *time.Ticker
	Latency     []int64
}

func (ta *TraceAggregation) CaculateLatency() {

	sort.Sort(DataTraces(ta.allLog))

	// client all latency
	ta.Latency[0] = ta.allLog[0].Span.EndTime - ta.allLog[0].Span.StartTime
	// client send to lb latency
	ta.Latency[1] = ta.allLog[1].Span.StartTime - ta.allLog[0].Span.StartTime
	// lb send to server latency
	ta.Latency[2] = ta.allLog[2].Span.StartTime - ta.allLog[1].Span.EndTime
	// server send to lb latency
	ta.Latency[3] = ta.allLog[3].Span.EndTime - ta.allLog[4].Span.StartTime
	// lb send to client latency
	ta.Latency[4] = ta.allLog[4].Span.EndTime - ta.allLog[0].Span.EndTime
}

func (ta *TraceAggregation) Append(iddt *DataTrace) (bool, error) {

	if ta.timeout {
		return false, errors.New("timeout")
	}

	ta.allLog = append(ta.allLog, iddt)
	if iddt.Span.ParentID == 1 && len(ta.allLog) == 4 {
		ta.complete = true
		ta.killChannel <- true
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
		case <-ta.killChannel:
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
