package trace

import (
	"fmt"

	otlog "github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"
)

// Copied from opentelemetry-go/bridge/opentracing
// https://sourcegraph.com/github.com/open-telemetry/opentelemetry-go/-/tree/bridge/opentracing

type bridgeFieldEncoder struct {
	pairs []attribute.KeyValue
}

var _ otlog.Encoder = &bridgeFieldEncoder{}

func (e *bridgeFieldEncoder) EmitString(key, value string) {
	e.emitCommon(key, value)
}

func (e *bridgeFieldEncoder) EmitBool(key string, value bool) {
	e.emitCommon(key, value)
}

func (e *bridgeFieldEncoder) EmitInt(key string, value int) {
	e.emitCommon(key, value)
}

func (e *bridgeFieldEncoder) EmitInt32(key string, value int32) {
	e.emitCommon(key, value)
}

func (e *bridgeFieldEncoder) EmitInt64(key string, value int64) {
	e.emitCommon(key, value)
}

func (e *bridgeFieldEncoder) EmitUint32(key string, value uint32) {
	e.emitCommon(key, value)
}

func (e *bridgeFieldEncoder) EmitUint64(key string, value uint64) {
	e.emitCommon(key, value)
}

func (e *bridgeFieldEncoder) EmitFloat32(key string, value float32) {
	e.emitCommon(key, value)
}

func (e *bridgeFieldEncoder) EmitFloat64(key string, value float64) {
	e.emitCommon(key, value)
}

func (e *bridgeFieldEncoder) EmitObject(key string, value interface{}) {
	e.emitCommon(key, value)
}

func (e *bridgeFieldEncoder) EmitLazyLogger(value otlog.LazyLogger) {
	value(e)
}

func (e *bridgeFieldEncoder) emitCommon(key string, value interface{}) {
	e.pairs = append(e.pairs, otTagToOTelAttr(key, value))
}

func otLogFieldsToOTelAttrs(fields []otlog.Field) []attribute.KeyValue {
	encoder := &bridgeFieldEncoder{}
	for _, field := range fields {
		field.Marshal(encoder)
	}
	return encoder.pairs
}

// otTagToOTelAttr converts given key-value into attribute.KeyValue.
// Note that some conversions are not obvious:
// - int -> int64
// - uint -> string
// - int32 -> int64
// - uint32 -> int64
// - uint64 -> string
// - float32 -> float64
func otTagToOTelAttr(k string, v interface{}) attribute.KeyValue {
	key := otTagToOTelAttrKey(k)
	switch val := v.(type) {
	case bool:
		return key.Bool(val)
	case int64:
		return key.Int64(val)
	case uint64:
		return key.String(fmt.Sprintf("%d", val))
	case float64:
		return key.Float64(val)
	case int32:
		return key.Int64(int64(val))
	case uint32:
		return key.Int64(int64(val))
	case float32:
		return key.Float64(float64(val))
	case int:
		return key.Int(val)
	case uint:
		return key.String(fmt.Sprintf("%d", val))
	case string:
		return key.String(val)
	default:
		return key.String(fmt.Sprint(v))
	}
}

func otTagToOTelAttrKey(k string) attribute.Key {
	return attribute.Key(k)
}
