package completions

import (
	"context"
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// modelAvailabilityCheckWindow defines the time window for assessing model availability.
// If a model accumulates too many errors within this duration, it's considered unavailable.
const modelAvailabilityCheckWindow = 5 * time.Minute

// errorThresholdForUnavailability is the maximum number of error records
// allowed within the modelAvailabilityCheckWindow before a model is
// considered unavailable.
const errorThresholdForUnavailability = 10

type modelsLoadTracker struct {
	records map[string][]*record
}

type record struct {
	reason    string
	timestamp time.Time
}

func newModelsLoadTracker() *modelsLoadTracker {
	return &modelsLoadTracker{
		records: map[string][]*record{},
	}
}

// record adds a new record to the modelsLoadTracker if a request error occurred due to
// a timeout (deadline exceeded) or if the response status code is 429 (Too Many Requests).
// If neither of these conditions are met, it resets the error records for the given model.
func (mlt *modelsLoadTracker) record(gatewayModel string, resp *http.Response, reqErr error) {
	var r *record

	if errors.Is(reqErr, context.DeadlineExceeded) {
		r = &record{
			reason:    "timeout",
			timestamp: time.Now(),
		}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		r = &record{
			reason:    "rate limit exceeded",
			timestamp: time.Now(),
		}
	}

	if r == nil {
		mlt.records[gatewayModel] = nil
		return
		// r = &record{
		// 	reason:    "success",
		// 	timestamp: time.Now(),
		// }
	}

	records := mlt.records[gatewayModel]
	records = append([]*record{r}, records...)
	if len(records) > errorThresholdForUnavailability {
		records = records[:errorThresholdForUnavailability]
	}

	mlt.records[gatewayModel] = records
}

// isModelAvailable checks if a model is available based on recent request failures.
// It returns false if the last errorThresholdForUnavailability requests to the
// specified model within modelAvailabilityCheckWindow failed. Otherwise, it returns true.
func (mlt *modelsLoadTracker) isModelAvailable(gatewayModel string) bool {
	var count int
	now := time.Now()

	for _, r := range mlt.records[gatewayModel] {
		if now.Sub(r.timestamp) > modelAvailabilityCheckWindow {
			continue
		}
		count++
		if count >= errorThresholdForUnavailability {
			return false
		}
	}

	return true
}
