package client

import (
	"fmt"
	"sync"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/base"
	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument/crouton_thrift"
	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/thrift_0_9_2/lib/go/thrift"

	"github.com/golang/glog"
)

const (
	// See the comment for shouldFlush() for more about these tuning
	// parameters.
	kDefaultMaxReportingPeriod = 5 * time.Second
	kMinReportingPeriod        = 500 * time.Millisecond

	kMaxBufferedLogs  = 1000
	kMaxBufferedSpans = 1000
)

// ReporterFuncs is a global registry of Reporter implementations. Runtime
// users can set the reporter using the ReporterImpl option.
var ReporterFuncs = make(map[string]func(options *Options, guid instrument.RuntimeGuid) Reporter)

const BufferingReporterImpl = "buffering"

func init() {
	ReporterFuncs[BufferingReporterImpl] = NewBufferingReporter
}

// Reporter is a low-level interface to various backends that process
// log and span records. It corresponds to a single Runtime instance
// and forwards records asynchronously or when Flush() is called.
type Reporter interface {
	AddRecords([]*crouton_thrift.LogRecord, []*crouton_thrift.SpanRecord)
	MergeAttributes(map[string]string)
	Flush()
	// Disable causes this reporter to discard any buffered data. Once a
	// Reported is disabled, all methods behave as no-ops. A Reporter
	// cannot be re-enabled.
	Disable()
}

// BufferingReporter is a Reporter that simply buffers records and
// forwards them to a ReportingService.
type BufferingReporter struct {
	lock sync.Mutex

	// auth and runtime information

	auth      *crouton_thrift.Auth
	guid      instrument.RuntimeGuid
	groupName string
	attrs     map[string]string
	start     base.Micros

	// buffered logs and spans

	// will retain up to kMaxBufferedLogs
	// XXX: this is unsafe... need some additional constraints on log
	// size, for example, on the length of messages or payloads
	logRecords []*crouton_thrift.LogRecord
	// will retain up to kMaxBufferedSpans
	// XXX: same issues as recentLogs.
	spanRecords []*crouton_thrift.SpanRecord

	// the backend and related info
	killChan chan bool
	// the last time an outgoing report was made
	lastOutgoing       base.Micros
	maxReportingPeriod base.Micros
	reportInFlight     bool
	// Remote service that will receive log and span reports
	backend crouton_thrift.ReportingService

	// We allow our remote peer to disable this instrumentation at any
	// time, turning all potentially costly runtime operations into
	// no-ops.
	disabled bool

	debugString string
}

func NewBufferingReporter(options *Options, guid instrument.RuntimeGuid) Reporter {
	attrs := make(map[string]string)
	for k, v := range options.Attributes {
		attrs[k] = fmt.Sprint(v)
	}
	attrs["cruntime_platform"] = "golang"
	attrs["cruntime_version"] = "0.1.2"
	attrs["cruntime_impl"] = BufferingReporterImpl

	var backend crouton_thrift.ReportingService
	var debugString string
	if options.Backend != nil {
		backend = options.Backend
		debugString = fmt.Sprintf("buffering, forwarding reports to %#v", backend)
	} else {
		serviceHost := DefaultServiceHost
		if len(options.ServiceHost) > 0 {
			serviceHost = options.ServiceHost
		}
		servicePort := ReportingServiceThriftSecurePort
		if options.ServicePort > 0 {
			servicePort = options.ServicePort
		}

		transport, err := thrift.NewTHttpPostClient(
			fmt.Sprintf("https://%s:%d%s", serviceHost, servicePort,
				ReportingServiceThriftPathBinary))
		if err != nil {
			glog.Warningf("Unable to initialize Traceguide client: %v", err)
		} else {
			backend = crouton_thrift.NewReportingServiceClientFactory(
				transport, thrift.NewTBinaryProtocolFactoryDefault())
		}
		debugString = fmt.Sprintf("buffering, forwarding reports to %s:%d", serviceHost, servicePort)
	}

	rval := &BufferingReporter{
		auth: &crouton_thrift.Auth{
			AccessToken: thrift.StringPtr(options.AccessToken),
		},
		guid:      guid,
		groupName: options.GroupName,
		attrs:     attrs,
		start:     base.NowMicros(),

		killChan: make(chan bool),
		backend:  backend,
		// Disable if we weren't able to initialize a backend.
		disabled:    backend == nil,
		debugString: debugString,

		maxReportingPeriod: base.DurationInMicros(kDefaultMaxReportingPeriod),
	}

	go rval.reportLoop()
	return rval
}

func (r *BufferingReporter) String() string {
	return r.debugString
}

func (r *BufferingReporter) MergeAttributes(attrs map[string]string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	// Early-out for disabled runtimes.
	if r.disabled {
		return
	}

	for k, v := range attrs {
		r.attrs[k] = v
	}
}

func (r *BufferingReporter) AddRecords(logs []*crouton_thrift.LogRecord, spans []*crouton_thrift.SpanRecord) {
	r.lock.Lock()
	defer r.lock.Unlock()

	// Early-out for disabled runtimes.
	if r.disabled {
		return
	}

	// Don't append to logRecords if the log buffer is past our max.
	//
	// (Note that this drops data to protect the client)
	if l := kMaxBufferedLogs - len(r.logRecords); l > 0 {
		if l > len(logs) {
			l = len(logs)
		}
		if l > 0 {
			r.logRecords = append(r.logRecords, logs[:l]...)
		}
	}

	// Don't append to spanRecords if the buffer is past our max.
	//
	// (Note that this drops data to protect the client)
	if l := kMaxBufferedSpans - len(r.spanRecords); l > 0 {
		if l > len(spans) {
			l = len(spans)
		}
		if l > 0 {
			r.spanRecords = append(r.spanRecords, spans[:l]...)
		}
	}
}

func (r *BufferingReporter) Flush() {
	r.lock.Lock()

	if r.disabled {
		r.lock.Unlock()
		return
	}

	if r.reportInFlight == true {
		fmt.Printf("A previous Report is still in flight; aborting Flush().")
		r.lock.Unlock()
		return
	}

	r.lastOutgoing = base.NowMicros()
	var req *crouton_thrift.ReportRequest
	if len(r.logRecords) != 0 || len(r.spanRecords) != 0 {
		req = &crouton_thrift.ReportRequest{
			Runtime:     r.thriftRuntime(),
			LogRecords:  r.logRecords,
			SpanRecords: r.spanRecords,
		}
	}
	if req == nil {
		// Nothing happened since the last tick; no need to send a report.
		r.lock.Unlock()
		return
	}
	r.reportInFlight = true
	r.lock.Unlock() // unlock before making the RPC itself

	// Send the [non-empty] report along with our accessToken.
	resp, err := r.backend.Report(r.auth, req)
	glog.V(1).Infof("Report: resp=%v, err=%v", resp, err)

	r.lock.Lock()
	r.reportInFlight = false
	if err != nil {
		r.lock.Unlock()
		return
	}

	// Clear the buffers
	r.logRecords = nil
	r.spanRecords = nil
	// TODO something about timing
	r.lock.Unlock()

	for _, c := range resp.Commands {
		if c.Disable != nil && *c.Disable {
			r.Disable()
		}
	}
}

// caller must hold r.lock
func (r *BufferingReporter) thriftRuntime() *crouton_thrift.Runtime {
	runtimeAttrs := []*crouton_thrift.KeyValue{}
	for k, v := range r.attrs {
		runtimeAttrs = append(runtimeAttrs, &crouton_thrift.KeyValue{k, v})
	}
	return &crouton_thrift.Runtime{
		Guid:        thrift.StringPtr(string(r.guid)),
		StartMicros: thrift.Int64Ptr(r.start.Int64()),
		GroupName:   thrift.StringPtr(r.groupName),
		Attrs:       runtimeAttrs,
	}
}

func (r *BufferingReporter) Disable() {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.disabled {
		return
	}

	fmt.Printf("Disabling Runtime instance: %p", r)

	r.logRecords = nil
	r.spanRecords = nil
	r.disabled = true

	r.killChan <- true
}

// Every kMinReportingPeriod the reporting loop wakes up and checks to see if
// either (a) the Runtime's max reporting period is about to expire (see
// maxReportingPeriod()), (b) the number of buffered log records is
// approaching kMaxBufferedLogs, or if (c) the number of buffered span records
// is approaching kMaxBufferedSpans. If any of those conditions are true,
// pending data is flushed to the remote peer. If not, the reporting loop waits
// until the next cycle. See Runtime.maybeFlush() for details.
//
// This could alternatively be implemented using flush channels and so forth,
// but that would introduce opportunities for client code to block on the
// runtime library, and we want to avoid that at all costs (even dropping data,
// which can certainly happen with high data rates and/or unresponsive remote
// peers).
func (r *BufferingReporter) shouldFlush() bool {
	r.lock.Lock()
	defer r.lock.Unlock()

	if (base.NowMicros()+base.DurationInMicros(kMinReportingPeriod))-r.lastOutgoing > r.maxReportingPeriod {
		// Flush timeout.
		glog.V(2).Infof("--> timeout")
		return true
	} else if len(r.logRecords) > kMaxBufferedLogs/2 {
		// Too many queued log records.
		glog.V(2).Infof("--> log queue")
		return true
	} else if len(r.spanRecords) > kMaxBufferedSpans/2 {
		// Too many queued span records.
		glog.V(2).Infof("--> span queue")
		return true
	}
	return false
}

func (r *BufferingReporter) reportLoop() {
	// (Thrift really should do this internally, but we saw some too-many-fd's
	// errors and thrift is the most likely culprit.)
	switch b := r.backend.(type) {
	case *crouton_thrift.ReportingServiceClient:
		// TODO This is a bit racy with other calls to Flush, but we're
		// currently assuming that no one calls Flush after Disable.
		defer b.Transport.Close()
	}

	tickerChan := time.Tick(kMinReportingPeriod)
	for {
		select {
		// Rare case:
		case <-r.killChan:
			glog.Infof("Exiting reporting loop")
			return

		// Common case:
		case <-tickerChan:
			glog.V(1).Infof("reporting alarm fired")
			if r.shouldFlush() {
				r.Flush()
			}
		}
	}
}
