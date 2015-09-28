package client

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/base/goroutinelocal"
	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument/crouton_thrift"
)

const (
	ReportingServiceThriftPathPrefix     = "/_rpc/v1/reports/"
	ReportingServiceThriftPathBinary     = ReportingServiceThriftPathPrefix + "binary"
	ReportingServiceThriftPathJSON       = ReportingServiceThriftPathPrefix + "json"
	ReportingServiceThriftPathURIEncoded = ReportingServiceThriftPathPrefix + "uri_encoded"

	ReportingServiceThriftPlainPort  = 9998
	ReportingServiceThriftSecurePort = 9997
)

const DefaultServiceHost = "traceguide-api.mttr.to"

type Runtime struct {
	lock sync.Mutex
	guid instrument.RuntimeGuid

	reporter Reporter
}

type Options struct {
	AccessToken string

	// ServiceHost describes the service to which span and log data will be
	// sent.  If empty, the default will be used.  Backend takes precedence
	// over ServiceHost.
	ServiceHost string
	// ServicePort describes the service to which span and log data will be
	// sent.  If zero, the default will be used.  Backend takes precedence
	// over ServicePort.
	ServicePort int

	// Backend overrides the values of ServiceHost and ServicePort.  It is
	// useful for testing.
	Backend crouton_thrift.ReportingService

	GroupName    string
	Attributes   map[string]interface{}
	ReporterImpl string
}

func NewRuntime(options *Options) *Runtime {
	if options.GroupName == "" {
		options.GroupName = path.Base(os.Args[0])
	}
	if options.Attributes == nil {
		options.Attributes = make(map[string]interface{})
	}
	// Set some default attributes if not found in options
	if _, found := options.Attributes["hostname"]; !found {
		hostname, _ := os.Hostname()
		options.Attributes["hostname"] = hostname
	}
	if _, found := options.Attributes["cmdline"]; !found {
		options.Attributes["cmdline"] = strings.Join(os.Args, " ")
	}
	rval := &Runtime{
		guid: instrument.RuntimeGuid(genSeededGuid()),
	}
	reporterImpl := options.ReporterImpl
	if len(reporterImpl) == 0 {
		reporterImpl = BufferingReporterImpl
	}
	rval.reporter = ReporterFuncs[reporterImpl](options, rval.guid)
	logString := fmt.Sprintf("Traceguide client Runtime initialized; %v\n", rval.reporter)
	// Good to go!
	log.Print(logString)
	rval.Log(instrument.FileLine(1).Info().Print(logString))
	return rval
}

func (r *Runtime) MergeAttributes(attrs map[string]interface{}) {
	newAttrs := make(map[string]string)
	for k, v := range attrs {
		newAttrs[k] = fmt.Sprint(v)
	}
	r.reporter.MergeAttributes(newAttrs)
}

func (r *Runtime) StartSpan() instrument.ActiveSpan {
	return r.startSpan()
}

// startSpan behaves like StartSpan but returns a more specific type.
func (r *Runtime) startSpan() *ActiveSpan {
	rval := newActiveSpan(r)
	rval.logPerfStats()
	return rval
}

func (r *Runtime) RunInSpan(f func(s instrument.ActiveSpan) error,
	options ...instrument.SpanOption) error {
	span := r.startSpan()
	defer span.Finish()

	// This method is subtle to understand but powerful in that spans
	// separated by arbitrary numbers of [direct] function calls can
	// connect via the goroutine-local storage.
	var localActiveSpans *activeSpanStack
	onStack := false
	for _, o := range options {
		switch o {
		case instrument.OnStack:
			onStack = true
		default:
			r.Log(instrument.Printf("Unknown SpanOption: %#v", o).Warning())
		}
	}
	if onStack {
		localActiveSpans = goroutinelocal.GetWithDefault(
			kActiveSpansGoroutineLocalKey, &activeSpanStack{}).(*activeSpanStack)
		localActiveSpans.Push(span)
		defer localActiveSpans.PopSpan(span)
	}

	err := f(span)
	if err != nil {
		span.Log(instrument.Print(err).Error().CallStack(3))
	}
	return err
}

func (r *Runtime) AddTraceJoinIdToSpansInStack(key string, value interface{}) error {
	goroutineActiveSpans, ok := goroutinelocal.Get(kActiveSpansGoroutineLocalKey).(*activeSpanStack)
	if !ok || len(goroutineActiveSpans.stack) == 0 {
		return fmt.Errorf("No active Spans found on stack")
	}
	for _, parentSpan := range goroutineActiveSpans.stack {
		parentSpan.AddTraceJoinId(key, value)
	}
	return nil
}

var (
	seededGuidGen     *rand.Rand
	seededGuidGenOnce sync.Once
	seededGuidLock    sync.Mutex
)

// NOTE: we are not happy about these being strings. In a non-prototype
// universe, they should be more like int128s. Or, at least, they should be
// really compact and fast to manipulate.
func genSeededGuid() string {
	// Golang does not seed the rng for us. Make sure it happens.
	seededGuidGenOnce.Do(func() {
		seededGuidGen = rand.New(rand.NewSource(time.Now().UnixNano()))
	})

	// The goland rand generators are *not* intrinsically thread-safe.
	seededGuidLock.Lock()
	defer seededGuidLock.Unlock()
	return strconv.FormatUint(uint64(seededGuidGen.Int63()), 36)
}

func (r *Runtime) Flush() {
	r.reporter.Flush()
}

// Disable the instrumentation for this Runtime instance permanently
// (i.e., this is not toggle-able). Only call this if you know what
// you're doing or worry that the library is doing something harmful.
func (r *Runtime) Disable() {
	r.reporter.Disable()
}

func (r *Runtime) RecordTraceJoin(keyVals ...interface{}) {
	if len(keyVals) < 4 || (len(keyVals)%2 != 0) {
		return
	}
	activeSpan := r.StartSpan()
	defer activeSpan.Finish()
	activeSpan.SetOperation("_trace_join")
	for i := 0; i < len(keyVals); i += 2 {
		key, ok := keyVals[i].(string)
		if !ok {
			continue
		}
		val := fmt.Sprint(keyVals[i+1])
		activeSpan.AddTraceJoinId(key, val)
	}
}

func (r *Runtime) String() string {
	return fmt.Sprintf("Runtime:{guid:%s}", string(r.guid))
}

const kActiveSpansGoroutineLocalKey = "active_spans"

//type activeSpanMap map[instrument.SpanGuid]*ActiveSpan
type activeSpanStack struct {
	stack []*ActiveSpan
}

func (p *activeSpanStack) Push(span *ActiveSpan) {
	p.stack = append(p.stack, span)
}

// Pop this particular span...
func (p *activeSpanStack) PopSpan(span *ActiveSpan) {
	i := len(p.stack) - 1
	if i < 0 {
		// TODO: where do we report an internal error like this?
		return
	}

	topSpan := p.stack[i]
	if topSpan.Guid() != span.Guid() {
		// TODO: where should this unexpected GUID mismatch be reported?
	}

	// Pop the stack even if the GUIDs didn't match with the rationale
	// that a runaway growing active span stack would be worse than the
	// alternative of mismatched spans (we're already in a bad place if
	// the GUIDs don't match!).
	p.stack = p.stack[0:i]
}

func (p *activeSpanStack) Top() *ActiveSpan {
	i := len(p.stack) - 1
	if i < 0 {
		return nil
	}
	return p.stack[i]
}
