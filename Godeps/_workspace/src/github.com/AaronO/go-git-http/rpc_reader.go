package githttp

import (
	"io"
	"regexp"
)

// RpcReader scans for events in the incoming rpc request data
type RpcReader struct {
	// Underlaying reader (to relay calls to)
	io.Reader

	// Rpc type (upload-pack or receive-pack)
	Rpc string

	// List of events RpcReader has picked up through scanning
	// these events do not contain the "Dir" attribute
	Events []Event

	pktLineParser pktLineParser
}

// Regexes to detect types of actions (fetch, push, etc ...)
var (
	receivePackRegex = regexp.MustCompile("([0-9a-fA-F]{40}) ([0-9a-fA-F]{40}) refs\\/(heads|tags)\\/(.*?)( |00|\x00)")
	uploadPackRegex  = regexp.MustCompile(`^want ([0-9a-fA-F]{40})`)
)

// Implement the io.Reader interface
func (r *RpcReader) Read(p []byte) (n int, err error) {
	// Relay call
	n, err = r.Reader.Read(p)

	// Scan for events
	if n > 0 {
		r.scan(p[:n])
	}

	return n, err
}

func (r *RpcReader) scan(data []byte) {
	if r.pktLineParser.state == done {
		return
	}

	r.pktLineParser.Feed(data)

	// If parsing has just finished, process its output once.
	if r.pktLineParser.state == done {
		if r.pktLineParser.Error != nil {
			return
		}

		// When we get here, we're done collecting all pkt-lines successfully
		// and can now extract relevant events.
		var events []Event
		switch r.Rpc {
		case "receive-pack":
			events = scanPush(r.pktLineParser.Total)
		case "upload-pack":
			events = scanFetch(r.pktLineParser.Total)
		}
		r.Events = append(r.Events, events...)
	}
}

func scanFetch(data []byte) []Event {
	matches := uploadPackRegex.FindAllStringSubmatch(string(data), -1)

	if matches == nil {
		return nil
	}

	var events []Event
	for _, m := range matches {
		events = append(events, Event{
			Type:   FETCH,
			Commit: m[1],
		})
	}

	return events
}

func scanPush(data []byte) []Event {
	matches := receivePackRegex.FindAllStringSubmatch(string(data), -1)

	if matches == nil {
		return nil
	}

	var events []Event
	for _, m := range matches {
		e := Event{
			Last:   m[1],
			Commit: m[2],
		}

		// Handle pushes to branches and tags differently
		if m[3] == "heads" {
			e.Type = PUSH
			e.Branch = m[4]
		} else {
			e.Type = TAG
			e.Tag = m[4]
		}

		events = append(events, e)
	}

	return events
}
