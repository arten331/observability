package logger

import (
	"context"
	"sync"

	"github.com/Arten331/observability/tracer"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	KeyScopePersist = "persist"
	KeyScopeTech    = "tech"

	KeyContextAttributes = "x_trace_attrs"
)

type LContext struct {
	Context
	mu                sync.Mutex
	PersistAttributes []zap.Field
}

type Context interface {
	Value(key any) any
}

func WithCtxValue(ctx Context, key, value interface{}) Context {
	switch ctx.(type) {
	case context.Context:
		return context.WithValue(ctx.(context.Context), key, value)
	case workflow.Context:
		return workflow.WithValue(ctx.(workflow.Context), key, value)
	default:
		L().Error("!IMPORTANT: wrong context for logger", zap.Reflect("obj", ctx))

		return context.WithValue(context.Background(), key, value)
	}
}

func appendContextFields(ctx Context, fields []zap.Field) []zap.Field {
	if lCtx, ok := ctx.(*LContext); ok {
		return append(fields, lCtx.PersistAttributes...)
	}

	lCtx := getContext(ctx)
	if len(lCtx.PersistAttributes) > 0 {
		return append(fields, lCtx.PersistAttributes...)
	}

	return fields
}

func getContext(ctx Context) *LContext {
	var lCtx *LContext
	var ok bool

	lCtx, ok = ctx.(*LContext)

	if !ok {
		var persistAttributes []zap.Field

		//try to find attributes in value and restore
		persistAttributesRaw := ctx.Value(KeyContextAttributes)

		if persistAttributesRaw == nil {
			persistAttributes = make([]zap.Field, 0)
		} else {
			persistAttributes = persistAttributesRaw.([]zapcore.Field)
		}

		lCtx = &LContext{
			Context:           WithCtxValue(ctx, KeyContextAttributes, persistAttributes),
			PersistAttributes: persistAttributes,
		}
	}

	return lCtx
}

func TraceWithID(ctx context.Context, key string, value string) context.Context {
	fields := []zap.Field{zap.String(key, value)}

	span := tracer.SpanFromContext(ctx)
	span.SetAttributes(otelAttributes(fields...)...)

	lCtx := checkAndAppendFieldToCtx(ctx, fields)
	lCtx.Context = WithCtxValue(ctx, KeyContextAttributes, lCtx.PersistAttributes)

	return lCtx.Context.(context.Context)
}

// TraceWFWithID adds a persistent structured field to a workflow context and
// to its active OpenTelemetry span.
func TraceWFWithID(ctx workflow.Context, key string, value string) workflow.Context {
	fields := []zap.Field{zap.String(key, value)}

	tracer.SpanFromWorkflowContext(ctx).SetAttributes(otelAttributes(fields...)...)

	lCtx := checkAndAppendFieldToCtx(ctx, fields)
	return lCtx.Context.(workflow.Context)
}

func checkAndAppendFieldToCtx(ctx Context, fields []zap.Field) *LContext {
	lCtx := getContext(ctx)

	for _, field := range fields {
		exists := false
		for _, existingField := range lCtx.PersistAttributes {
			if existingField.Key == field.Key {
				//lCtx.PersistAttributes[i] = field
				exists = true
				break
			}
		}

		if !exists {
			lCtx.PersistAttributes = append(lCtx.PersistAttributes, field)
		}
	}

	lCtx.Context = WithCtxValue(lCtx.Context, KeyContextAttributes, lCtx.PersistAttributes)

	return lCtx
}

func Scope(scope string) zap.Field {
	return zap.String("scope", scope)
}

func ScopeTech() zap.Field {
	return zap.String("scope", KeyScopeTech)
}

func ScopePersist() zap.Field {
	return zap.String("scope", KeyScopePersist)
}
