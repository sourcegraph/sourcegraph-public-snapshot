package papertrail

import (
	"fmt"
	"log"
	"time"

	"context"

	"github.com/jpillora/backoff"
	"github.com/sourcegraph/go-papertrail/papertrail"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

const maxAttempts = 5

// buildLogs is a Papertrail-backed implementation of the build logs
// store.
type buildLogs struct{}

var _ store.BuildLogs = (*buildLogs)(nil)

func (s *buildLogs) Get(ctx context.Context, task sourcegraph.TaskSpec, minID string, minTime, maxTime time.Time) (*sourcegraph.LogEntries, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "BuildLogs.Get", task.Build.Repo); err != nil {
		return nil, err
	}
	pOpt := papertrail.SearchOptions{
		Query:   "program:" + task.IDString(),
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
			log.Printf("Truncated log fetch for build logs for %s after %s (%d lines)", task.IDString(), maxDuration, len(allEvents))
			allEvents = append(allEvents, &papertrail.Event{Message: fmt.Sprintf("*******************************************************\n\n*** NOTE: log truncated to newest %d lines (fetching stopped after %s); older log lines will not be visible ***\n*******************************************************\n", len(allEvents), maxDuration)})
			break
		}

		b := &backoff.Backoff{
			Min:    500 * time.Millisecond,
			Jitter: true,
		}
		var e0s *papertrail.SearchResponse
		var err error
		for i := 0; i < maxAttempts; i++ {
			e0s, _, err = client(ctx).Search(pOpt)
			if err == nil {
				break
			}
			time.Sleep(b.Duration())
		}
		if err != nil {
			return nil, err
		}

		// only set maxID on first iteration because subsequent iterations go
		// *backwards* in time and will have responses with lower max_id values
		if maxID == "" {
			maxID = e0s.MaxID
		}

		if len(e0s.Events) == 1 && pOpt.MaxID == e0s.MinID {
			// Papertrail seems to not set reached_beginning reliably
			// and returns the same log line repeatedly; detect that
			// case and break here.
			break
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
		es.Entries[len(es.Entries)-i-1] = e0.Message
	}
	return es, nil
}
