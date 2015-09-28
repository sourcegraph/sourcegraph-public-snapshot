package client

import (
	"bytes"
	"flag"
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/base"
	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument/crouton_thrift"
	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/thrift_0_9_2/lib/go/thrift"
)

var (
	flagMaxLogMessageLen     = flag.Int("max_log_message_len_bytes", 1024, "the maximum number of bytes used by a single log message")
	flagMaxPayloadFieldBytes = flag.Int("max_log_payload_field_bytes", 1024, "the maximum number of bytes exported in a single payload field")
	flagMaxPayloadTotalBytes = flag.Int("max_log_payload_max_total_bytes", 4096, "the maximum number of bytes exported in an entire payload")
)

const kLogInMemoryThreshold = base.Micros(10 * 1000000)

func (r *Runtime) Log(arg interface{}) {
	rec := &logRecord{}
	switch arg := arg.(type) {
	case *instrument.LogBuilder:
		rec.LogRecord = arg.LogRecord()
	case *instrument.LogRecord:
		rec.LogRecord = arg
	default:
		rec.LogRecord = &instrument.LogRecord{Message: fmt.Sprint(arg)}
	}
	// Set the runtime guid below.
	r.log(rec)
}

// logRecord is a representation of log records before they are sent
// over the wire.
type logRecord struct {
	*instrument.LogRecord
	SpanGuid    *instrument.SpanGuid // may be nil
	RuntimeGuid instrument.RuntimeGuid
}

func (l *logRecord) String() string {
	var buf bytes.Buffer
	buf.WriteString(l.LogRecord.String())
	buf.WriteString(":{span:")
	if l.SpanGuid == nil {
		buf.WriteString("nil")
	} else {
		buf.WriteString(string(*l.SpanGuid))
	}
	buf.WriteString(" runtime:")
	buf.WriteString(string(l.RuntimeGuid))
	buf.WriteString("}")
	return buf.String()
}

func (r *Runtime) log(rec *logRecord) {
	// Ensure that every record written has a timestamp and a runtime
	// guid. We allow logs to be passed in with a timestamp to support
	// the proxy use case.
	if rec.TimestampMicros == 0 {
		rec.TimestampMicros = base.NowMicros()
	}
	rec.RuntimeGuid = r.guid

	r.reporter.AddRecords([]*crouton_thrift.LogRecord{rec.toThrift()}, nil)
}

func (r *logRecord) toThrift() *crouton_thrift.LogRecord {
	var msg *string
	if len(r.Message) > 0 {
		// Don't allow for arbitrarily long log messages.
		if len(r.Message) > *flagMaxLogMessageLen {
			msg = thrift.StringPtr(r.Message[:(*flagMaxLogMessageLen-1)] + kEllipsis)
		} else {
			msg = thrift.StringPtr(r.Message)
		}
	}

	var thriftPayload *string
	if r.Payload != nil {
		// This converts values to strings to avoid lossy encoding. I.e.
		// not the same as a call to json.Marshal().
		jsonString, err := ValueToSanitizedJSONString(
			r.Payload, *flagMaxPayloadFieldBytes, *flagMaxPayloadTotalBytes)
		if err != nil {
			log.Printf("Error encoding payload object: %v", err)
		} else {
			thriftPayload = &jsonString
		}
	}
	return &crouton_thrift.LogRecord{
		TimestampMicros: thrift.Int64Ptr(r.TimestampMicros.Int64()),
		RuntimeGuid:     thrift.StringPtr(string(r.RuntimeGuid)),
		SpanGuid:        (*string)(r.SpanGuid),
		StableName:      base.StringPtr(r.EventName),
		Message:         msg,
		Level:           thrift.StringPtr(r.Level),
		Filename:        thrift.StringPtr(r.FileName),
		LineNumber:      thrift.Int64Ptr(int64(r.LineNumber)),
		StackFrames:     r.StackFrames,
		PayloadJson:     thriftPayload,
		ErrorFlag:       base.BoolPtr(r.IsError),
	}
}
