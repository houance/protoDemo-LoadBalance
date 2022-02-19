package trace

import (
	"context"
	"houance/protoDemo-LoadBalance/internal/innerData"
	"sync"
	"time"
)

func ProcessHandler(
	ctx context.Context,
	inChannel chan *innerData.DataTrace,
	traceIDMap map[uint32]*innerData.TraceAggregation,
	steamIDMap map[uint32]*innerData.StreamLatency,
	spanInfoPool *sync.Pool,
	prefixLengthPool *sync.Pool,
	dataTracePool *sync.Pool,
	traceArregationPool *sync.Pool,
) (err error) {

	var (
		complete bool
		sl       *innerData.StreamLatency    = &innerData.StreamLatency{}
		ta       *innerData.TraceAggregation = &innerData.TraceAggregation{}
	)

	for {

		select {
		case <-ctx.Done():
			return ctx.Err()

		case iddt := <-inChannel:
			// Arregation
			if _, ok := traceIDMap[iddt.Span.TraceID]; !ok {
				tmpTraceArregation := traceArregationPool.Get().(*innerData.TraceAggregation)
				traceIDMap[iddt.Span.TraceID] = tmpTraceArregation
				tmpTraceArregation.SetTicker(5 * time.Second)
			}
			ta = traceIDMap[iddt.Span.TraceID]
			complete, err = ta.Append(iddt)
			if err != nil {
				traceArregationPool.Put(ta)
				// traceIDMap[iddt.Span.SpanID] = nil
				delete(traceIDMap, iddt.Span.TraceID)
			}
			if complete {
				// Analyse
				if _, ok := steamIDMap[ta.StreamID]; !ok {
					steamIDMap[ta.StreamID] = &innerData.StreamLatency{}
				}
				sl = steamIDMap[ta.StreamID]
				sl.AcumulateLatency(ta.Latency)
				ta.ResetAllLog(
					spanInfoPool,
					prefixLengthPool,
					dataTracePool,
					traceArregationPool,
				)

			}
		}
	}
}
