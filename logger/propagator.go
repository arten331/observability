package logger

import (
	"context"
	"fmt"
	"strings"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const temporalHeaderPrefix = "x_oculus_"

// ValuesPropagator carries persistent Oculus fields through Temporal headers.
type ValuesPropagator struct{}

// NewPersistenceValuesPropagator returns a Temporal context propagator for
// fields added with TraceWithID or TraceWFWithID.
func NewPersistenceValuesPropagator() workflow.ContextPropagator {
	return &ValuesPropagator{}
}

func (p *ValuesPropagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	return injectFields(getContext(ctx).PersistAttributes, writer)
}

func (p *ValuesPropagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	return injectFields(getContext(ctx).PersistAttributes, writer)
}

func (p *ValuesPropagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	fields, err := extractFields(reader)
	if err != nil || len(fields) == 0 {
		return ctx, err
	}

	return context.WithValue(ctx, contextAttributesKey, fields), nil
}

func (p *ValuesPropagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	fields, err := extractFields(reader)
	if err != nil || len(fields) == 0 {
		return ctx, err
	}

	return workflow.WithValue(ctx, contextAttributesKey, fields), nil
}

func injectFields(fields []zap.Field, writer workflow.HeaderWriter) error {
	for _, field := range fields {
		payload, err := converter.GetDefaultDataConverter().ToPayload(field.String)
		if err != nil {
			return err
		}
		writer.Set(fmt.Sprintf("%s%s", temporalHeaderPrefix, field.Key), payload)
	}

	return nil
}

func extractFields(reader workflow.HeaderReader) ([]zap.Field, error) {
	fields := make([]zap.Field, 0)
	err := reader.ForEachKey(func(key string, value *commonpb.Payload) error {
		if !strings.HasPrefix(key, temporalHeaderPrefix) {
			return nil
		}

		var decodedValue string
		if err := converter.GetDefaultDataConverter().FromPayload(value, &decodedValue); err != nil {
			return err
		}

		fields = append(fields, zap.String(strings.TrimPrefix(key, temporalHeaderPrefix), decodedValue))
		return nil
	})

	return fields, err
}
