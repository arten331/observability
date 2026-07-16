package logger

import (
	"context"

	"github.com/Arten331/observability/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (l *Logger) WithError(msg string, err error, fields ...zap.Field) error {
	fields = append(fields, zap.Error(err))
	l.ErrorLogger.Error(msg, fields...)

	return err
}

func (l *Logger) WithCtxError(
	ctx context.Context,
	msg string,
	err error,
	fields ...zap.Field,
) error {
	span := tracer.SpanFromContext(ctx)
	span.RecordError(err, trace.WithAttributes(otelAttributes(fields...)...))
	span.SetStatus(codes.Error, "")

	fields = appendTraceFields(span, append(appendContextFields(ctx, fields), zap.Error(err)))
	l.ErrorLogger.Error(msg, fields...)

	return err
}

func (l *Logger) WithCtxInfo(
	ctx context.Context,
	msg string,
	fields ...zap.Field,
) {
	span := tracer.SpanFromContext(ctx)
	span.AddEvent(msg, trace.WithAttributes(otelAttributes(fields...)...))

	fields = appendTraceFields(span, appendContextFields(ctx, fields))

	l.CtxLogger.Info(msg, fields...)
}

func (l *Logger) WithCtxDebug(
	ctx context.Context,
	msg string,
	fields ...zap.Field,
) {
	span := tracer.SpanFromContext(ctx)
	span.AddEvent(msg, trace.WithAttributes(otelAttributes(fields...)...))

	fields = appendTraceFields(span, appendContextFields(ctx, fields))

	l.CtxLogger.Debug(msg, fields...)
}

// WithWFCtxError records and logs an error against the active workflow span.
func (l *Logger) WithWFCtxError(
	ctx workflow.Context,
	msg string,
	err error,
	fields ...zap.Field,
) error {
	span := tracer.SpanFromWorkflowContext(ctx)
	span.RecordError(err, trace.WithAttributes(otelAttributes(fields...)...))
	span.SetStatus(codes.Error, "")

	fields = appendTraceFields(span, append(appendContextFields(ctx, fields), zap.Error(err)))
	l.ErrorLogger.Error(msg, fields...)

	return err
}

// WithWFCtxInfo records and logs an event against the active workflow span.
func (l *Logger) WithWFCtxInfo(ctx workflow.Context, msg string, fields ...zap.Field) {
	span := tracer.SpanFromWorkflowContext(ctx)
	span.AddEvent(msg, trace.WithAttributes(otelAttributes(fields...)...))

	l.CtxLogger.Info(msg, appendTraceFields(span, appendContextFields(ctx, fields))...)
}

// WithWFCtxDebug records and logs a debug event against the active workflow span.
func (l *Logger) WithWFCtxDebug(ctx workflow.Context, msg string, fields ...zap.Field) {
	span := tracer.SpanFromWorkflowContext(ctx)
	span.AddEvent(msg, trace.WithAttributes(otelAttributes(fields...)...))

	l.CtxLogger.Debug(msg, appendTraceFields(span, appendContextFields(ctx, fields))...)
}

func appendTraceFields(span trace.Span, fields []zap.Field) []zap.Field {
	spanContext := span.SpanContext()
	if !spanContext.IsValid() {
		return fields
	}

	return append(fields,
		zap.String("trace_id", spanContext.TraceID().String()),
		zap.String("span_id", spanContext.SpanID().String()),
	)
}
