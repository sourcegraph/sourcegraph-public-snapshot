package client

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument"
	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/instrument/crouton_thrift"
)

// ---------
// CONSTANTS
// ---------
const (
	// Number of threads to launch. For the purposes of the below tests kNumThreads must be 0 - 10.
	kNumThreads = 10
	// Number of requests (log & spans) per thread (go routine).
	kNumRequests = 50
	// Number of total requests made.
	kTotalRequests = kNumThreads * kNumRequests

	// Number of automatic logs
	kNumAutomaticLogs = 1
	// Due to automatic logs, after n >= kMaxLogsRetained log requests, logs are randomly thrown out
	kMaxLogsRetained = kMaxBufferedLogs / (kNumAutomaticLogs + 1)
)

// ---------------
// TYPES & METHODS
// ---------------

// Implements crouton_thrift.ReportingService such that it can intercept
// report requests for local testing and debugging of the go runtime.
type testingReportingService struct {
	Testing *testing.T
}

// Do not send reports to the server, instead inspect & validate the reports.
func (r *testingReportingService) Report(auth *crouton_thrift.Auth, request *crouton_thrift.ReportRequest) (*crouton_thrift.ReportResponse, error) {
	checkSpans(r.Testing, request)
	checkLogs(r.Testing, request)

	return &crouton_thrift.ReportResponse{}, nil
}

// -----
// TESTS
// -----

// This check makes sure that it is possible to correctly establish a Runtime
// that sends reports via the tsetReportingService rather than sending reports
// to the service.
// This might fail if the Runtime fails to initialize properly.
func TestNew(t *testing.T) {
	initializeRuntime(t)
}

// Launch kNumThreads go routines, each with kNumRequests logs & spans.
// This may fail if all the spans and logs that are sent, are not received
// by the testReportingService in a report request.
func TestCreateLogsAndSpans(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	initializeRuntime(t)

	var wg sync.WaitGroup
	wg.Add(kNumThreads)

	for i := 0; i < kNumThreads; i++ {
		go func(i int) {
			defer wg.Done()
			createLogsAndSpans(t, kNumRequests, i*kNumRequests, i)
		}(i)
	}

	wg.Wait()
	instrument.Flush()
}

// --------------
// HELPER METHODS
// --------------

// Construct testReportingService and initialize a new runtime with this service
// for debugging purposes.
func initializeRuntime(t *testing.T) {
	rs := &testingReportingService{
		Testing: t,
	}

	instrument.SetDefaultRuntime(NewRuntime(&Options{
		AccessToken: "fake_access_token",
		GroupName:   "testing_threads",
		Backend:     rs,
	}))
}

// Send numRequests spans and logs
func createLogsAndSpans(t *testing.T, numRequests int, startIndex int, threadId int) {
	for i := 0; i < numRequests; i++ {
		id := i + startIndex
		span := instrument.StartSpan().SetOperation(fmt.Sprintf("%v", id))
		log := instrument.Payload(i).Printf("Log for span #: %v", id)
		span.Log(log)
		span.Finish()
	}
}

// Up to kMaxLogsRetained logs supported, typically 1 + kTotalRequests * 2 logs presented
func checkLogs(t *testing.T, request *crouton_thrift.ReportRequest) {
	// TODO: Make test more robust

	logs := request.GetLogRecords()

	if kTotalRequests < kMaxLogsRetained && len(logs) != (1+(kNumAutomaticLogs+1)*kTotalRequests) {
		t.Error("Incorrect number of logs.")
	} else if kTotalRequests >= kMaxLogsRetained && len(logs) != kMaxBufferedLogs {
		t.Error("Incorrect number of logs.")
	}

	// If # of logs < kMaxLogsRetained, check to make sure all submitted logs exist
	if kTotalRequests < kMaxLogsRetained {
		// Accumulate all log messages
		logMessages := make([]string, 0, len(logs))
		for _, log := range logs {
			logMessages = append(logMessages, log.GetMessage())
		}
		sort.Strings(logMessages)

		// Search for every desired log
		for i := 0; i < kTotalRequests; i++ {
			message := fmt.Sprintf("Log for span #: %v", i)
			index := sort.SearchStrings(logMessages, message)
			if index >= len(logMessages) {
				t.Errorf("Error: Couldn't find %v", message)
			}
		}
	}
}

// Checks for the correct number of spans.
// Checks unique id of each span (contained in span's message).
func checkSpans(t *testing.T, request *crouton_thrift.ReportRequest) {
	spans := request.GetSpanRecords()

	// Check the number of spans
	if len(spans) != kTotalRequests {
		t.Errorf("Incorrect Number of Spans, %v", len(spans))
	}

	// Assumption: Span name is simply a unique id number in the range [0, kTotalRequests)
	spanMessages := make([]int, 0, len(spans))
	for _, span := range spans {
		value, err := strconv.Atoi(span.GetSpanName())
		if err != nil {
			t.Errorf("Error %v", err)
		}
		spanMessages = append(spanMessages, value)
	}
	sort.Ints(spanMessages)
	for i := 0; i < kTotalRequests; i++ {
		if i != spanMessages[i] {
			t.Errorf("Error: Expected span with message %v but found %v", i, spanMessages[i])
		}
	}

}
