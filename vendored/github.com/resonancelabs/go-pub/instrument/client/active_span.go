package client

import (
	"fmt"

	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/base"
	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/base/goroutinelocal"
	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument/crouton_thrift"
	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/thrift_0_9_2/lib/go/thrift"
)

// TODO This is public so that code outside this package can call ToThrift.

type ActiveSpan struct {
	guid        instrument.SpanGuid
	runtime     *Runtime
	operation   string
	joinIds     map[string]string
	startMicros base.Micros
	endMicros   base.Micros
	attributes  map[string]string
}

func newActiveSpan(r *Runtime) *ActiveSpan {
	return &ActiveSpan{
		guid:        newSpanGuid(),
		runtime:     r,
		joinIds:     map[string]string{},
		startMicros: base.NowMicros(),
		endMicros:   base.Micros(-1), // not yet valid!
		attributes:  map[string]string{},
	}
}

func (s *ActiveSpan) Finish() {
	s.endMicros = base.NowMicros()
	if len(s.operation) > 0 {
		s.runtime.reporter.AddRecords(nil, []*crouton_thrift.SpanRecord{s.ToThrift()})
	} else {
		// TODO: should we panic() here? this is a serious API misuse.
		// TODO: ...or do something reasonable with anonymous spans
	}
}

func (s *ActiveSpan) SetOperation(op string) instrument.ActiveSpan {
	s.operation = op
	return s
}

func (s *ActiveSpan) SetName(name string) instrument.ActiveSpan {
	s.AddAttribute(instrument.SpanAttributeDeprecatedName, "")
	return s.SetOperation(name)
}

func (s *ActiveSpan) AddTraceJoinId(key string, value interface{}) instrument.ActiveSpan {
	s.joinIds[key] = fmt.Sprint(value)
	return s
}

func (s *ActiveSpan) SetEndUserId(id interface{}) instrument.ActiveSpan {
	s.joinIds[instrument.TraceJoinKeyEndUserId] = fmt.Sprint(id)
	return s
}

func (s *ActiveSpan) AddAttribute(key, val string) instrument.ActiveSpan {
	s.attributes[key] = val
	return s
}

func (s *ActiveSpan) SetParent(parentSpan instrument.ActiveSpan) instrument.ActiveSpan {
	if parentSpan == nil {
		return s
	}

	parentGuid := string(parentSpan.Guid())
	s.AddAttribute("parent_span_guid", parentGuid)

	// Merge all the parent join IDs
	for key, val := range parentSpan.TraceJoinIds() {
		s.AddTraceJoinId(key, val)
	}
	return s
}

func (s *ActiveSpan) TraceJoinIds() map[string]string {
	rval := make(map[string]string, len(s.joinIds))
	for k, v := range s.joinIds {
		rval[k] = v
	}
	return rval
}

func (s *ActiveSpan) Guid() instrument.SpanGuid {
	return s.guid
}

func (s *ActiveSpan) MergeTraceJoinIdsFromStack() error {
	goroutineActiveSpans, ok := goroutinelocal.Get(kActiveSpansGoroutineLocalKey).(*activeSpanStack)
	if !ok || len(goroutineActiveSpans.stack) == 0 {
		return fmt.Errorf("No active Spans found on stack")
	}
	for _, parentSpan := range goroutineActiveSpans.stack {
		for key, val := range parentSpan.TraceJoinIds() {
			s.AddTraceJoinId(key, val)
		}
	}
	return nil
}

func (s *ActiveSpan) SetParentFromStack() error {
	goroutineActiveSpans, ok := goroutinelocal.Get(kActiveSpansGoroutineLocalKey).(*activeSpanStack)
	if !ok {
		return fmt.Errorf("No active Spans found on stack")
	}
	parentSpan := goroutineActiveSpans.Top()
	if parentSpan == nil {
		return fmt.Errorf("No active Spans found on stack")
	}
	s.SetParent(parentSpan)
	return nil
}

func (s *ActiveSpan) Log(arg interface{}) {
	rec := &logRecord{}
	switch arg := arg.(type) {
	case *instrument.LogBuilder:
		rec.LogRecord = arg.LogRecord()
	case *instrument.LogRecord:
		rec.LogRecord = arg
	default:
		rec.LogRecord = &instrument.LogRecord{Message: fmt.Sprint(arg)}
	}
	rec.SpanGuid = &s.guid
	s.runtime.log(rec)
}

func newSpanGuid() instrument.SpanGuid {
	return instrument.SpanGuid(genSeededGuid())
}

func (s *ActiveSpan) logPerfStats() {
	maybeRefreshPerfStats()

	gPerfLock.RLock()
	defer gPerfLock.RUnlock()
	s.Log(instrument.Printf("perf snapshot (%v ago)",
		(base.NowMicros() - gPerfStats.PerfSampleMicros).ToDuration()).
		Payload(gPerfStats))
}

func (s *ActiveSpan) ToThrift() *crouton_thrift.SpanRecord {
	joinIds := []*crouton_thrift.TraceJoinId{}
	for k, v := range s.joinIds {
		joinIds = append(joinIds, &crouton_thrift.TraceJoinId{
			TraceKey: k,
			Value:    v,
		})
	}

	var attributes []*crouton_thrift.KeyValue
	if len(s.attributes) > 0 {
		attributes = make([]*crouton_thrift.KeyValue, 0, len(s.attributes))
		for k, v := range s.attributes {
			attributes = append(attributes, &crouton_thrift.KeyValue{
				Key:   k,
				Value: v,
			})
		}
	}

	return &crouton_thrift.SpanRecord{
		SpanGuid:       thrift.StringPtr(string(s.guid)),
		RuntimeGuid:    thrift.StringPtr(string(s.runtime.guid)),
		SpanName:       thrift.StringPtr(s.operation),
		JoinIds:        joinIds,
		OldestMicros:   thrift.Int64Ptr(s.startMicros.Int64()),
		YoungestMicros: thrift.Int64Ptr(s.endMicros.Int64()),
		Attributes:     attributes,
	}
}
