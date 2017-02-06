package lightstep

import (
	"encoding/json"

	cpb "github.com/lightstep/lightstep-tracer-go/collectorpb"
	"github.com/opentracing/opentracing-go/log"
)

const (
	ellipsis = "…"
)

// An implementation of the log.Encoder interface
type logFieldEncoder struct {
	recorder        *Recorder
	buffer          *reportBuffer
	currentKeyValue *cpb.KeyValue
}

func marshalFields(
	recorder *Recorder,
	protoLog *cpb.Log,
	fields []log.Field,
	buffer *reportBuffer,
) {
	lfe := logFieldEncoder{recorder, buffer, nil}
	protoLog.Keyvalues = make([]*cpb.KeyValue, len(fields))
	for i, f := range fields {
		lfe.currentKeyValue = &cpb.KeyValue{}
		f.Marshal(&lfe)
		protoLog.Keyvalues[i] = lfe.currentKeyValue
	}
}

func (lfe *logFieldEncoder) EmitString(key, value string) {
	lfe.emitSafeKey(key)
	lfe.emitSafeString(value)
}
func (lfe *logFieldEncoder) EmitBool(key string, value bool) {
	lfe.emitSafeKey(key)
	lfe.currentKeyValue.Value = &cpb.KeyValue_BoolValue{value}
}
func (lfe *logFieldEncoder) EmitInt(key string, value int) {
	lfe.emitSafeKey(key)
	lfe.currentKeyValue.Value = &cpb.KeyValue_IntValue{int64(value)}
}
func (lfe *logFieldEncoder) EmitInt32(key string, value int32) {
	lfe.emitSafeKey(key)
	lfe.currentKeyValue.Value = &cpb.KeyValue_IntValue{int64(value)}
}
func (lfe *logFieldEncoder) EmitInt64(key string, value int64) {
	lfe.emitSafeKey(key)
	lfe.currentKeyValue.Value = &cpb.KeyValue_IntValue{int64(value)}
}
func (lfe *logFieldEncoder) EmitUint32(key string, value uint32) {
	lfe.emitSafeKey(key)
	lfe.currentKeyValue.Value = &cpb.KeyValue_IntValue{int64(value)}
}
func (lfe *logFieldEncoder) EmitUint64(key string, value uint64) {
	lfe.emitSafeKey(key)
	lfe.currentKeyValue.Value = &cpb.KeyValue_IntValue{int64(value)}
}
func (lfe *logFieldEncoder) EmitFloat32(key string, value float32) {
	lfe.emitSafeKey(key)
	lfe.currentKeyValue.Value = &cpb.KeyValue_DoubleValue{float64(value)}
}
func (lfe *logFieldEncoder) EmitFloat64(key string, value float64) {
	lfe.emitSafeKey(key)
	lfe.currentKeyValue.Value = &cpb.KeyValue_DoubleValue{float64(value)}
}
func (lfe *logFieldEncoder) EmitObject(key string, value interface{}) {
	lfe.emitSafeKey(key)
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		lfe.buffer.logEncoderErrorCount++
		lfe.emitSafeString("<json.Marshal error>")
		return
	}
	lfe.emitSafeJSON(string(jsonBytes))
}
func (lfe *logFieldEncoder) EmitLazyLogger(value log.LazyLogger) {
	// Delegate to `value` to do the late-bound encoding.
	value(lfe)
}

func (lfe *logFieldEncoder) emitSafeKey(key string) {
	if len(key) > lfe.recorder.maxLogKeyLen {
		key = key[:(lfe.recorder.maxLogKeyLen-1)] + ellipsis
	}
	lfe.currentKeyValue.Key = key
}
func (lfe *logFieldEncoder) emitSafeString(str string) {
	if len(str) > lfe.recorder.maxLogValueLen {
		str = str[:(lfe.recorder.maxLogValueLen-1)] + ellipsis
	}
	lfe.currentKeyValue.Value = &cpb.KeyValue_StringValue{str}
}
func (lfe *logFieldEncoder) emitSafeJSON(json string) {
	if len(json) > lfe.recorder.maxLogValueLen {
		str := json[:(lfe.recorder.maxLogValueLen-1)] + ellipsis
		lfe.currentKeyValue.Value = &cpb.KeyValue_StringValue{str}
		return
	}
	lfe.currentKeyValue.Value = &cpb.KeyValue_JsonValue{json}
}
