package kafka

import (
	"context"

	"github.com/Arten331/observability/tracer"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/trace"
)

func GetSpanFromKafkaMessage(ctx context.Context, msg kafka.Message) (context.Context, trace.Span) {
	// get span context from kafka message headers
	var spanID trace.SpanID
	var traceID trace.TraceID
	var counter int
	var err error

	for i := range msg.Headers {
		switch msg.Headers[i].Key {
		case "span-id":
			spanID, err = trace.SpanIDFromHex(string(msg.Headers[i].Value))
			counter++
			if err != nil {
				break
			}
		case "trace-id":
			traceID, err = trace.TraceIDFromHex(string(msg.Headers[i].Value))
			counter++
			if err != nil {
				break
			}
		}
	}

	if counter == 2 {
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
					Key:   "span-id",
					Value: []byte(spanHeader.SpanID().String()),
				},
				{
					Key:   "trace-id",
					Value: []byte(spanHeader.TraceID().String()),
				},
			}
		}
	}
}
