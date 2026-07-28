package logger

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.uber.org/zap/zapcore"
)

func otelAttributes(fields ...zapcore.Field) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(fields))

	for _, f := range fields {
		attrs = appendField(attrs, f)
	}

	return attrs
}

func appendField(attrs []attribute.KeyValue, f zapcore.Field) []attribute.KeyValue {
	switch f.Type {
	case zapcore.BoolType:
		attr := attribute.Bool(f.Key, f.Integer == 1)
		return append(attrs, attr)

	case zapcore.Int8Type, zapcore.Int16Type, zapcore.Int32Type, zapcore.Int64Type:
		attr := attribute.Int64(f.Key, f.Integer)
		return append(attrs, attr)
	case zapcore.Uint8Type, zapcore.Uint16Type, zapcore.Uint32Type, zapcore.Uint64Type, zapcore.UintptrType:
		value := strconv.FormatUint(uint64(f.Integer), 10) //nolint:gosec // Zap stores unsigned field values in Integer's bit pattern.
		attr := attribute.String(f.Key, value)
		return append(attrs, attr)

	case zapcore.Float32Type, zapcore.Float64Type:
		value := math.Float64frombits(uint64(f.Integer)) //nolint:gosec // Zap stores float64 values in Integer's bit pattern.
		attr := attribute.Float64(f.Key, value)
		return append(attrs, attr)

	case zapcore.Complex64Type:
		s := strconv.FormatComplex(complex128(f.Interface.(complex64)), 'E', -1, 64)
		attr := attribute.String(f.Key, s)
		return append(attrs, attr)
	case zapcore.Complex128Type:
		s := strconv.FormatComplex(f.Interface.(complex128), 'E', -1, 128)
		attr := attribute.String(f.Key, s)
		return append(attrs, attr)

	case zapcore.StringType:
		attr := attribute.String(f.Key, f.String)
		return append(attrs, attr)
	case zapcore.BinaryType, zapcore.ByteStringType:
		attr := attribute.String(f.Key, string(f.Interface.([]byte)))
		return append(attrs, attr)
	case zapcore.StringerType:
		attr := attribute.String(f.Key, f.Interface.(fmt.Stringer).String())
		return append(attrs, attr)

	case zapcore.DurationType, zapcore.TimeType:
		attr := attribute.Int64(f.Key, f.Integer)
		return append(attrs, attr)
	case zapcore.TimeFullType:
		attr := attribute.Int64(f.Key, f.Interface.(time.Time).UnixNano())
		return append(attrs, attr)
	case zapcore.ErrorType:
		err := f.Interface.(error)
		typ := reflect.TypeOf(err).String()
		attrs = append(attrs, semconv.ExceptionTypeKey.String(typ))
		attrs = append(attrs, semconv.ExceptionMessageKey.String(err.Error()))
		return attrs
	case zapcore.ReflectType:
		attr := Attribute(f.Key, f.Interface)
		return append(attrs, attr)
	case zapcore.SkipType:
		return attrs

	case zapcore.ArrayMarshalerType:
		var attr attribute.KeyValue
		arrayEncoder := &bufferArrayEncoder{
			stringsSlice: []string{},
		}
		err := f.Interface.(zapcore.ArrayMarshaler).MarshalLogArray(arrayEncoder)
		if err != nil {
			attr = attribute.String(f.Key+"_error", fmt.Sprintf("otelzap: unable to marshal array: %v", err))
		} else {
			attr = attribute.StringSlice(f.Key, arrayEncoder.stringsSlice)
		}
		return append(attrs, attr)

	case zapcore.ObjectMarshalerType:
		attr := attribute.String(f.Key+"_error", "otelzap: zapcore.ObjectMarshalerType is not implemented")

		attrs = append(attrs, attr)

		attr = Attribute(f.Key, f.Interface)
		attrs = append(attrs, attr)

		return attrs
	default:
		attr := attribute.String(f.Key+"_error", fmt.Sprintf("otelzap: unknown field type: %v", f))
		return append(attrs, attr)
	}
}

func Attribute(key string, value any) attribute.KeyValue {
	switch value := value.(type) {
	case nil:
		return attribute.String(key, "<nil>")
	case string:
		return attribute.String(key, value)
	case int:
		return attribute.Int(key, value)
	case int64:
		return attribute.Int64(key, value)
	case uint64:
		return attribute.String(key, strconv.FormatUint(value, 10))
	case float64:
		return attribute.Float64(key, value)
	case bool:
		return attribute.Bool(key, value)
	case fmt.Stringer:
		return attribute.String(key, value.String())
	}

	rv := reflect.ValueOf(value)

	switch rv.Kind() {
	case reflect.Array:
		rv = rv.Slice(0, rv.Len())
		fallthrough
	case reflect.Slice:
		switch reflect.TypeOf(value).Elem().Kind() {
		case reflect.Bool:
			return attribute.BoolSlice(key, rv.Interface().([]bool))
		case reflect.Int:
			return attribute.IntSlice(key, rv.Interface().([]int))
		case reflect.Int64:
			return attribute.Int64Slice(key, rv.Interface().([]int64))
		case reflect.Float64:
			return attribute.Float64Slice(key, rv.Interface().([]float64))
		case reflect.String:
			return attribute.StringSlice(key, rv.Interface().([]string))
		default:
			return attribute.KeyValue{Key: attribute.Key(key)}
		}
	case reflect.Bool:
		return attribute.Bool(key, rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return attribute.Int64(key, rv.Int())
	case reflect.Float64:
		return attribute.Float64(key, rv.Float())
	case reflect.String:
		return attribute.String(key, rv.String())
	}
	if b, err := json.Marshal(value); b != nil && err == nil {
		return attribute.String(key, string(b))
	}
	return attribute.String(key, fmt.Sprint(value))
}
