package libhoney

// txClient handles the transmission of events to Honeycomb.
//
// Overview
//
// Create a new instance of Client.
// Set any of the public fields for which you want to override the defaults.
// Call Start() to spin up the background goroutines necessary for transmission
// Call Add(Event) to queue an event for transmission
// Ensure Stop() is called to flush all in-flight messages.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/facebookgo/muster"
)

type txClient interface {
	Start() error
	Stop() error
	Add(*Event)
}

type txDefaultClient struct {
	maxBatchSize         uint          // how many events to collect into a batch before sending
	batchTimeout         time.Duration // how often to send off batches
	maxConcurrentBatches uint          // how many batches can be inflight simultaneously
	pendingWorkCapacity  uint          // how many events to allow to pile up
	blockOnSend          bool          // whether to block or drop events when the queue fills
	blockOnResponses     bool          // whether to block or drop responses when the queue fills

	transport http.RoundTripper

	muster muster.Client
}

func (t *txDefaultClient) Start() error {
	t.muster.MaxBatchSize = t.maxBatchSize
	t.muster.BatchTimeout = t.batchTimeout
	t.muster.MaxConcurrentBatches = t.maxConcurrentBatches
	t.muster.PendingWorkCapacity = t.pendingWorkCapacity
	t.muster.BatchMaker = func() muster.Batch {
		return &batch{
			events:           make([]*Event, 0, t.maxBatchSize),
			httpClient:       &http.Client{Transport: t.transport},
			blockOnResponses: t.blockOnResponses,
		}
	}
	return t.muster.Start()
}

func (t *txDefaultClient) Stop() error {
	return t.muster.Stop()
}

func (t *txDefaultClient) Add(ev *Event) {
	// don't block if we can't send events fast enough
	sd.Gauge("queue_length", len(t.muster.Work))
	if t.blockOnSend {
		t.muster.Work <- ev
		sd.Increment("messages_queued")
	} else {
		select {
		case t.muster.Work <- ev:
			sd.Increment("messages_queued")
		default:
			sd.Increment("queue_overflow")
			r := Response{
				Err:      errors.New("queue overflow"),
				Metadata: ev.Metadata,
			}
			if t.blockOnResponses {
				responses <- r
			} else {
				select {
				case responses <- r:
				default:
				}
			}
		}
	}
}

type txTestClient struct {
	Timestamps  []time.Time
	datas       [][]byte
	sampleRates []uint
}

func (t *txTestClient) Start() error {
	t.Timestamps = make([]time.Time, 0)
	t.datas = make([][]byte, 0)
	return nil
}

func (t *txTestClient) Stop() error {
	return nil
}

func (t *txTestClient) Add(ev *Event) {
	t.Timestamps = append(t.Timestamps, ev.Timestamp)
	blob, err := json.Marshal(ev.data)
	if err != nil {
		panic(err)
	}
	t.datas = append(t.datas, blob)
	t.sampleRates = append(t.sampleRates, ev.SampleRate)
}

type batch struct {
	events           []*Event
	httpClient       *http.Client
	blockOnResponses bool

	// allows manipulation of the value of "now" for testing
	testNower   nower
	testBlocker *sync.WaitGroup
}

func (b *batch) Add(ev interface{}) {
	b.events = append(b.events, ev.(*Event))
}

func (b *batch) Fire(notifier muster.Notifier) {
	defer notifier.Done()

	for _, e := range b.events {
		b.sendRequest(e)
	}
}

// sendRequest sends an individual request to Honeycomb and returns
func (b *batch) sendRequest(e *Event) {
	start := time.Now().UTC()
	if b.testNower != nil {
		start = b.testNower.Now()
	}
	timestamp := e.Timestamp
	blob, err := json.Marshal(e.data)
	if err != nil {
		// TODO add logging or something to raise this error
		sd.Increment("json_marshal_errors")
		return
	}

	userAgent := fmt.Sprintf("libhoney-go/%s", version)
	if UserAgentAddition != "" {
		userAgent = fmt.Sprintf("%s %s", userAgent, strings.TrimSpace(UserAgentAddition))
	}

	url := fmt.Sprintf("%s/1/events/%s", e.APIHost, e.Dataset)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(blob))
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("X-Honeycomb-Team", e.WriteKey)
	req.Header.Add("X-Event-Time", timestamp.Format(time.RFC3339))
	req.Header.Add("X-Honeycomb-SampleRate", strconv.Itoa(int(e.SampleRate)))

	resp, err := b.httpClient.Do(req)

	end := time.Now().UTC()
	if b.testNower != nil {
		end = b.testNower.Now()
	}
	dur := end.Sub(start)
	evResp := Response{}
	defer func() {
		if b.blockOnResponses {
			responses <- evResp
		} else {
			select {
			case responses <- evResp:
			default:
				if b.testBlocker != nil {
					b.testBlocker.Done()
				}
			}
		}
	}()
	evResp.Duration = dur
	evResp.Metadata = e.Metadata
	if err != nil {
		// TODO add logging or something to raise this error
		sd.Increment("send_errors")
		evResp.Err = err
		return
	}
	sd.Increment("messages_sent")
	defer resp.Body.Close()
	evResp.StatusCode = resp.StatusCode
	body, _ := ioutil.ReadAll(resp.Body)
	evResp.Body = body
}

// nower to make testing easier
type nower interface {
	Now() time.Time
}
