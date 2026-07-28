package logger

import (
	"strconv"
	"testing"

	"go.uber.org/zap"
)

func TestOtelAttributesPreserveUint64(t *testing.T) {
	const value = ^uint64(0)

	attrs := otelAttributes(zap.Uint64("value", value))
	if got, want := attrs[0].Value.AsString(), strconv.FormatUint(value, 10); got != want {
		t.Errorf("uint64 attribute = %q, want %q", got, want)
	}
}

func TestAttributePreservesUint64(t *testing.T) {
	const value = ^uint64(0)

	attr := Attribute("value", value)
	if got, want := attr.Value.AsString(), strconv.FormatUint(value, 10); got != want {
		t.Errorf("uint64 attribute = %q, want %q", got, want)
	}
}
