package kafka

import (
	"context"

	"github.com/arten331/observability/tracer"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/trace"
)

const (
	spanIDHeader  = "span-id"
	traceIDHeader = "trace-id"
)

func GetSpanFromKafkaMessage(ctx context.Context, msg kafka.Message) (context.Context, trace.Span) {
	var spanID trace.SpanID
	var traceID trace.TraceID

	for i := range msg.Headers {
		switch msg.Headers[i].Key {
		case spanIDHeader:
			parsedSpanID, err := trace.SpanIDFromHex(string(msg.Headers[i].Value))
			if err == nil {
				spanID = parsedSpanID
			}
		case traceIDHeader:
			parsedTraceID, err := trace.TraceIDFromHex(string(msg.Headers[i].Value))
			if err == nil {
				traceID = parsedTraceID
			}
		}
	}

	if spanID.IsValid() && traceID.IsValid() {
		spanContext := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
		})
		ctx = trace.ContextWithSpanContext(ctx, spanContext)
	}

	return ctx, tracer.SpanFromContext(ctx)
}

func InjectSpanToKafkaMessages(ctx context.Context, queueMessages []kafka.Message) {
	if tracer.SpanFromContext(ctx).IsRecording() {
		spanHeader := tracer.SpanFromContext(ctx).SpanContext()
		for i := range queueMessages {
			queueMessages[i].Headers = []kafka.Header{
				{
					Key:   spanIDHeader,
					Value: []byte(spanHeader.SpanID().String()),
				},
				{
					Key:   traceIDHeader,
					Value: []byte(spanHeader.TraceID().String()),
				},
			}
		}
	}
}
