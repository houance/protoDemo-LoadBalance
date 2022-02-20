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
	traceIDMap map[uint32]*innerData.TraceAggregation,
	steamIDMap map[uint32]*innerData.StreamLatency,
	spanInfoPool *sync.Pool,
	prefixLengthPool *sync.Pool,
	dataTracePool *sync.Pool,
	traceArregationPool *sync.Pool,
) (err error) {

	var (
		// complete bool
		sl *innerData.StreamLatency    = &innerData.StreamLatency{}
		ta *innerData.TraceAggregation = &innerData.TraceAggregation{}
	)

	logger.Info("Start Trace Process")

	for {

		select {
		case <-ctx.Done():
			logger.Error("Out Side Disturpt, Return Trace Process")
			return ctx.Err()

		case iddt := <-inChannel:
			// Arregation
			if _, ok := traceIDMap[iddt.Span.TraceID]; !ok {
				logger.Info(
					"first time get trace",
					zap.Uint32("Trace ID", iddt.Span.TraceID),
				)
				tmpTraceArregation := traceArregationPool.Get().(*innerData.TraceAggregation)
				traceIDMap[iddt.Span.TraceID] = tmpTraceArregation
				tmpTraceArregation.SetTicker(10 * time.Second)
			}
			ta = traceIDMap[iddt.Span.TraceID]
			complete, length, err := ta.Append(iddt)
			logger.Info(
				"status",
				zap.Bool("complete", complete),
				zap.Error(err),
				zap.Int("length", length),
			)

			// timeout
			if err != nil {
				traceArregationPool.Put(ta)
				// traceIDMap[iddt.Span.SpanID] = nil
				delete(traceIDMap, iddt.Span.TraceID)
			}

			// recv full trace
			if complete {
				logger.Info(
					"Recv Full Trace",
					zap.Uint32("Stream ID", ta.StreamID),
				)

				// Analyse
				if _, ok := steamIDMap[ta.StreamID]; !ok {
					logger.Info("first time got full stream trace")
					steamIDMap[ta.StreamID] = innerData.NewStreamLatency()
				}
				sl = steamIDMap[ta.StreamID]

				// logger.Info(
				// 	"Latency",
				// 	zap.Int64("overall latency", ta.Latency[0]),
				// 	zap.Int64("client send to lb", ta.Latency[1]),
				// 	zap.Int64("lb send to server", ta.Latency[2]),
				// 	zap.Int64("server send to lb", ta.Latency[3]),
				// 	zap.Int64("lb send to client", ta.Latency[4]),
				// )

				sl.AcumulateLatency(ta.CaculateLatency())
				lat := sl.GetAverageLatency()
				logger.Info(
					"Latency",
					zap.Int64("overall latency", lat[0]),
					zap.Int64("client send to lb", lat[1]),
					zap.Int64("lb send to server", lat[2]),
					zap.Int64("server send to lb", lat[3]),
					zap.Int64("lb send to client", lat[4]),
				)
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
