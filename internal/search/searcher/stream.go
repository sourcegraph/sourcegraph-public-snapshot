package searcher

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type StreamDecoder struct {
	OnMatches func([]*protocol.FileMatch)
	OnDone    func(EventDone)
	OnUnknown func(event, data []byte)
}

func (rr StreamDecoder) ReadAll(r io.Reader) error {
	dec := streamhttp.NewDecoder(r)
	for dec.Scan() {
		event := dec.Event()
		data := dec.Data()
		if bytes.Equal(event, []byte("matches")) {
			if rr.OnMatches == nil {
				continue
			}
			var d []*protocol.FileMatch
			if err := json.Unmarshal(data, &d); err != nil {
				return errors.Wrap(err, "decode matches payload")
			}
			rr.OnMatches(d)
		} else if bytes.Equal(event, []byte("done")) {
			if rr.OnDone == nil {
				continue
			}
			var e EventDone
			if err := json.Unmarshal(data, &e); err != nil {
				return errors.Wrap(err, "decode done payload")
			}
			rr.OnDone(e)
			break // done will always be the last event
		} else {
			if rr.OnUnknown == nil {
				continue
			}
			rr.OnUnknown(event, data)
		}
	}
	return dec.Err()
}

type EventDone struct {
	LimitHit bool   `json:"limit_hit"`
	Error    string `json:"error"`
}
