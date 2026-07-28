package kafka

import (
	"testing"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/trace"
)

func TestGetSpanFromKafkaMessage(t *testing.T) {
	ctx, _ := GetSpanFromKafkaMessage(t.Context(), kafka.Message{
		Headers: []kafka.Header{
			{Key: spanIDHeader, Value: []byte("0123456789abcdef")},
			{Key: traceIDHeader, Value: []byte("0123456789abcdef0123456789abcdef")},
		},
	})

	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		t.Fatal("expected a valid span context")
	}
}

func TestGetSpanFromKafkaMessageIgnoresInvalidHeaders(t *testing.T) {
	ctx, _ := GetSpanFromKafkaMessage(t.Context(), kafka.Message{
		Headers: []kafka.Header{
			{Key: spanIDHeader, Value: []byte("not-a-span-id")},
			{Key: traceIDHeader, Value: []byte("0123456789abcdef0123456789abcdef")},
		},
	})

	if spanContext := trace.SpanContextFromContext(ctx); spanContext.IsValid() {
		t.Fatal("invalid Kafka headers must not create a span context")
	}
}
