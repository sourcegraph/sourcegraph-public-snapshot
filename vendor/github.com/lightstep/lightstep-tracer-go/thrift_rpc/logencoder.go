package thrift_rpc

import (
	"encoding/json"
	"fmt"

	"github.com/lightstep/lightstep-tracer-go/lightstep_thrift"
	"github.com/lightstep/lightstep-tracer-go/thrift_0_9_2/lib/go/thrift"
	"github.com/opentracing/opentracing-go/log"
)

const (
	deprecatedFieldKeyEvent   = "event"
	deprecatedFieldKeyPayload = "payload"
	ellipsis                  = "…"
)

// thrift_rpc.logFieldEncoder is an implementation of the log.Encoder interface
// that handles only the deprecated OpenTracing
// Span.LogEvent/LogEventWithPayload calls. (Since the thrift client is being
// phased out anyway)
type logFieldEncoder struct {
	logRecord *lightstep_thrift.LogRecord
	recorder  *Recorder
}

func (lfe *logFieldEncoder) EmitString(key, value string) {
	if key == deprecatedFieldKeyEvent {
		if len(value) > lfe.recorder.maxLogMessageLen {
			value = value[:(lfe.recorder.maxLogMessageLen-1)] + ellipsis
		}
		lfe.logRecord.StableName = thrift.StringPtr(value)
	}
}
func (lfe *logFieldEncoder) EmitObject(key string, value interface{}) {
	if key == deprecatedFieldKeyPayload {
		var thriftPayload string
		jsonString, err := json.Marshal(value)
		if err != nil {
			thriftPayload = fmt.Sprintf("Error encoding payload object: %v", err)
		} else {
			thriftPayload = string(jsonString)
		}
		if len(thriftPayload) > lfe.recorder.maxLogMessageLen {
			thriftPayload = thriftPayload[:(lfe.recorder.maxLogMessageLen-1)] + ellipsis
		}
		lfe.logRecord.PayloadJson = thrift.StringPtr(thriftPayload)
	}
}

// All other log.Encoder methods do nothing in the thrift_rpc implementation.
func (lfe *logFieldEncoder) EmitBool(key string, value bool)       {}
func (lfe *logFieldEncoder) EmitInt(key string, value int)         {}
func (lfe *logFieldEncoder) EmitInt32(key string, value int32)     {}
func (lfe *logFieldEncoder) EmitInt64(key string, value int64)     {}
func (lfe *logFieldEncoder) EmitUint32(key string, value uint32)   {}
func (lfe *logFieldEncoder) EmitUint64(key string, value uint64)   {}
func (lfe *logFieldEncoder) EmitFloat32(key string, value float32) {}
func (lfe *logFieldEncoder) EmitFloat64(key string, value float64) {}
func (lfe *logFieldEncoder) EmitLazyLogger(value log.LazyLogger)   {}
