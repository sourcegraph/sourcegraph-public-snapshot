package transmission

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// WriterSender implements the Sender interface by marshalling events to JSON
// and writing to STDOUT, or to the writer W if one is specified.
type WriterSender struct {
	W io.Writer

	BlockOnResponses  bool
	ResponseQueueSize uint
	responses         chan Response

	sync.Mutex
}

func (w *WriterSender) Start() error {
	if w.ResponseQueueSize == 0 {
		w.ResponseQueueSize = 100
	}
	w.responses = make(chan Response, w.ResponseQueueSize)
	return nil
}

func (w *WriterSender) Stop() error { return nil }

func (w *WriterSender) Flush() error { return nil }

func (w *WriterSender) Add(ev *Event) {
	var m []byte
	tPointer := &(ev.Timestamp)
	if ev.Timestamp.IsZero() {
		tPointer = nil
	}

	// don't include sample rate if it's 1; this is the default
	sampleRate := ev.SampleRate
	if sampleRate == 1 {
		sampleRate = 0
	}

	m, _ = json.Marshal(struct {
		Data       map[string]interface{} `json:"data"`
		SampleRate uint                   `json:"samplerate,omitempty"`
		Timestamp  *time.Time             `json:"time,omitempty"`
		Dataset    string                 `json:"dataset,omitempty"`
	}{ev.Data, sampleRate, tPointer, ev.Dataset})
	m = append(m, '\n')

	w.Lock()
	defer w.Unlock()
	if w.W == nil {
		w.W = os.Stdout
	}
	w.W.Write(m)
	resp := Response{
		// TODO what makes sense to set in the response here?
		Metadata: ev.Metadata,
	}
	w.SendResponse(resp)
}

func (w *WriterSender) TxResponses() chan Response {
	return w.responses
}

func (w *WriterSender) SendResponse(r Response) bool {
	if w.BlockOnResponses {
		w.responses <- r
	} else {
		select {
		case w.responses <- r:
		default:
			return true
		}
	}
	return false
}
