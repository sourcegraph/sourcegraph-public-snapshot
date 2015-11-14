package instrument

import "fmt"

type nopRuntime struct{}

// NewNopRuntime returns a Runtime with a minimal implementation.  It should
// incur little if any overhead.
func NewNopRuntime() Runtime {
	return &nopRuntime{}
}

func (r *nopRuntime) RunInSpan(f func(s ActiveSpan) error, options ...SpanOption) error {
	return f(&nopActiveSpan{r, false, make(map[string]string)})
}
func (r *nopRuntime) StartSpan() ActiveSpan {
	return &nopActiveSpan{r, true, make(map[string]string)}
}
func (r *nopRuntime) AddTraceJoinIdToSpansInStack(key string, value interface{}) error {
	return nil
}

func (r *nopRuntime) Log(arg interface{}) {}

func (r *nopRuntime) RecordTraceJoin(joinIds ...interface{}) {}

func (r *nopRuntime) MergeAttributes(attrs map[string]interface{}) {}
func (r *nopRuntime) Flush()                                       {}
func (r *nopRuntime) Disable()                                     {}

type nopActiveSpan struct {
	runtime       *nopRuntime
	fromStartSpan bool
	joinIds       map[string]string
}

func (s *nopActiveSpan) Finish() {
	// Should only be called for Spans created with StartSpan().
	if !s.fromStartSpan {
		fmt.Print("Error: ActiveSpan.Finish() called without StartSpan()")
	}
}

func (s *nopActiveSpan) SetOperation(op string) ActiveSpan {
	return s
}

func (s *nopActiveSpan) SetName(name string) ActiveSpan {
	return s
}
func (s *nopActiveSpan) AddTraceJoinId(key string, value interface{}) ActiveSpan {
	s.joinIds[key] = fmt.Sprint(value)
	return s
}
func (s *nopActiveSpan) SetEndUserId(id interface{}) ActiveSpan {
	s.joinIds[TraceJoinKeyEndUserId] = fmt.Sprint(id)
	return s
}
func (s *nopActiveSpan) AddAttribute(key, val string) ActiveSpan {
	return s
}

func (s *nopActiveSpan) SetParent(parentSpan ActiveSpan) ActiveSpan {
	return s
}

func (s *nopActiveSpan) MergeTraceJoinIdsFromStack() error {
	return nil
}

func (s *nopActiveSpan) SetParentFromStack() error {
	return nil
}

func (s *nopActiveSpan) Guid() SpanGuid {
	return "not-implemented"
}

func (s *nopActiveSpan) TraceJoinIds() map[string]string {
	rval := make(map[string]string)
	for k, v := range s.joinIds {
		rval[k] = v
	}
	return rval
}

func (s *nopActiveSpan) Log(arg interface{}) {
	s.runtime.Log(arg)
}
