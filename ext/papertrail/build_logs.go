package papertrail

import (
	"fmt"
	"log"
	"time"

	"github.com/sourcegraph/go-papertrail/papertrail"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

// buildLogs is a Papertrail-backed implementation of the build logs
// store.
type buildLogs struct{}

var _ store.BuildLogs = (*buildLogs)(nil)

func (s *buildLogs) Get(ctx context.Context, build sourcegraph.BuildSpec, tag, minID string, minTime, maxTime time.Time) (*sourcegraph.LogEntries, error) {
	pOpt := papertrail.SearchOptions{
		Query:   "program:" + tag,
		MinID:   minID,
		MinTime: minTime,
		MaxTime: maxTime,
	}

	var allEvents []*papertrail.Event
	var maxID string

	// The Papertrail API returns 30-100 log lines per request. To get the full
	// log, we need to iterate back in time and make multiple requests.
	const maxDuration = 15 * time.Second
	start := time.Now()
	for {
		if time.Since(start) > maxDuration {
			log.Printf("Truncated log fetch for tag %q after %s (%d lines)", tag, maxDuration, len(allEvents))
			allEvents = append(allEvents, &papertrail.Event{Message: fmt.Sprintf("*******************************************************\n\n*** NOTE: log truncated to newest %d lines (fetching stopped after %s); older log lines will not be visible ***\n*******************************************************\n", len(allEvents), maxDuration)})
			break
		}

		e0s, _, err := client(ctx).Search(pOpt)
		if err != nil {
			return nil, err
		}

		// only set maxID on first iteration because subsequent iterations go
		// *backwards* in time and will have responses with lower max_id values
		if maxID == "" {
			maxID = e0s.MaxID
		}

		// append in reverse
		for i := len(e0s.Events) - 1; i >= 0; i-- {
			allEvents = append(allEvents, e0s.Events[i])
		}

		if e0s.ReachedBeginning || len(e0s.Events) == 0 {
			break
		}

		// continue going back in time to get all events
		pOpt.MaxID = e0s.MinID

		time.Sleep(250 * time.Millisecond)
	}

	es := &sourcegraph.LogEntries{
		Entries: make([]string, len(allEvents)),
		MaxID:   maxID,
	}
	for i := len(allEvents) - 1; i >= 0; i-- {
		e0 := allEvents[i]
		var prog string
		if e0.Program != nil {
			prog = *e0.Program
		}
		es.Entries[len(es.Entries)-i-1] = fmt.Sprintf("%s %s %s: %s", e0.ReceivedAt, e0.SourceName, prog, e0.Message)
	}
	return es, nil
}
