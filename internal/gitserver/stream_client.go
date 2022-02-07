package gitserver

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type StreamSearchDecoder struct {
	OnMatches func(protocol.SearchEventMatches)
	OnDone    func(protocol.SearchEventDone)
	OnUnknown func(event, data []byte)
}

func (s StreamSearchDecoder) ReadAll(r io.Reader) error {
	dec := http.NewDecoder(r)

	for dec.Scan() {
		event := dec.Event()
		data := dec.Data()

		if bytes.Equal(event, []byte("matches")) {
			if s.OnMatches == nil {
				continue
			}
			var e protocol.SearchEventMatches
			if err := json.Unmarshal(data, &e); err != nil {
				return errors.Errorf("failed to decode matches payload: %w", err)
			}
			s.OnMatches(e)
		} else if bytes.Equal(event, []byte("done")) {
			var e protocol.SearchEventDone
			if err := json.Unmarshal(data, &e); err != nil {
				return errors.Errorf("failed to decode matches payload: %w", err)
			}
			s.OnDone(e)
		}
	}

	return dec.Err()
}
