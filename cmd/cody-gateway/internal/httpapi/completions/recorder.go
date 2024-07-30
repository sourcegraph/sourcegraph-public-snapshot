package completions

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// modelsLoadTracker tracks the error records of models and determines their availability
// based on a specified failure threshold and timeout period. It uses a circuit breaker
// pattern to temporarily mark models as unavailable if they exceed the allowed number
// of errors within the defined time window.
type modelsLoadTracker struct {
	mu sync.RWMutex

	// failureThreshold represents the maximum number of error records
	// allowed within the timeout before a model is considered unavailable.
	failureThreshold int

	// timeout defines the time window for assessing model availability.
	// If a model accumulates the number of errors equal to failureThreshold
	// within this duration, it's considered unavailable.
	timeout time.Duration

	circuitBreakerByModel map[string]*modelCircuitBreaker
}

// modelCircuitBreaker keeps track of error records for a specific model,
// implementing a circular buffer to efficiently manage error history.
type modelCircuitBreaker struct {
	headIndex int
	records   []*record
}

// record represents an individual error occurrence with details about the reason for the error
// and the timestamp when it happened. This information is used to assess the model's availability.
type record struct {
	reason    string
	timestamp time.Time
}

// newModelsLoadTracker initializes and returns a modelsLoadTracker with the specified
// failure threshold and timeout period for assessing model availability.
func newModelsLoadTracker(failureThreshold int, timeout time.Duration) *modelsLoadTracker {
	return &modelsLoadTracker{
		failureThreshold:      failureThreshold,
		timeout:               timeout,
		circuitBreakerByModel: map[string]*modelCircuitBreaker{},
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

	mlt.mu.Lock()
	defer mlt.mu.Unlock()

	// Zero out the list of errors for the model, resetting the circuit breaker.
	if r == nil {
		mlt.circuitBreakerByModel[gatewayModel] = nil
		return
	}

	mcb := mlt.circuitBreakerByModel[gatewayModel]
	if mcb == nil {
		mcb = &modelCircuitBreaker{
			headIndex: 0,
			records:   make([]*record, mlt.failureThreshold),
		}
	}

	mcb.records[mcb.headIndex] = r
	mcb.headIndex++
	if mcb.headIndex >= mlt.failureThreshold {
		mcb.headIndex = 0
	}

	mlt.circuitBreakerByModel[gatewayModel] = mcb
}

// isModelAvailable checks if a model is available based on the number of failures within the specified timeout period.
// Returns false if there is at least one failure within the timeout period. Otherwise, returns true.
func (mlt *modelsLoadTracker) isModelAvailable(gatewayModel string) bool {
	mcb := mlt.circuitBreakerByModel[gatewayModel]
	if mcb == nil {
		return true
	}

	now := time.Now()
	for _, r := range mcb.records {
		if r == nil || now.Sub(r.timestamp) > mlt.timeout {
			return true
		}
	}

	return false
}
