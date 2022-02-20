package trace

import (
	"context"
	"houance/protoDemo-LoadBalance/internal/innerData"
	"sync"
	"time"

	"go.uber.org/zap"
)

func ProcessHandler(
	ctx context.Context,
	logger *zap.Logger,
	inChannel chan *innerData.DataTrace,
	traceIDAggregationMap map[uint32]*innerData.TraceAggregation,
	steamIDStreamLatencyMap map[uint32]*innerData.StreamLatency,
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

	logger.Info("Start Trace Process")

	for {

		select {
		case <-ctx.Done():
			logger.Error("Out Side Disturpt, Return Trace Process")
			return ctx.Err()

		case iddt := <-inChannel:
			// Arregation
			if _, ok := traceIDAggregationMap[iddt.Span.TraceID]; !ok {
				logger.Info(
					"first time get trace",
					zap.Uint32("Trace ID", iddt.Span.TraceID),
				)
				tmpTraceArregation := traceArregationPool.Get().(*innerData.TraceAggregation)
				traceIDAggregationMap[iddt.Span.TraceID] = tmpTraceArregation
				tmpTraceArregation.SetTicker(10 * time.Second)
			}
			ta = traceIDAggregationMap[iddt.Span.TraceID]
			complete, err = ta.Append(iddt)

			// timeout
			// doing recycle
			if err != nil {
				ta.ResetAllLog(
					spanInfoPool,
					prefixLengthPool,
					dataTracePool,
					traceArregationPool,
				)
				traceIDAggregationMap[iddt.Span.SpanID] = nil
				delete(traceIDAggregationMap, iddt.Span.TraceID)
			}

			// recv full trace
			if complete {
				logger.Info(
					"Recv Full Trace",
					zap.Uint32("Stream ID", ta.StreamID),
				)

				// Analyse
				// doing recycle
				if _, ok := steamIDStreamLatencyMap[ta.StreamID]; !ok {
					logger.Info("first time got full stream trace")
					steamIDStreamLatencyMap[ta.StreamID] = innerData.NewStreamLatency()
				}
				sl = steamIDStreamLatencyMap[ta.StreamID]

				sl.AcumulateLatency(ta.CaculateLatency())
				ta.ResetAllLog(
					spanInfoPool,
					prefixLengthPool,
					dataTracePool,
					traceArregationPool,
				)
				traceIDAggregationMap[iddt.Span.TraceID] = nil
				delete(traceIDAggregationMap, iddt.Span.TraceID)
			}
		}
	}
}
