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
	// failureRatio represents the maximum ratio of failure records out of last maxRecords
	// allowed within the evaluationWindow before a model is considered unavailable.
	failureRatio float64

	// maxRecords defines the maximum number of error records to keep for each model.
	maxRecords int

	// evaluationWindow defines the time window for assessing model availability.
	// If a model accumulates the number of errors equal to failureThreshold
	// within this duration, it's considered unavailable.
	evaluationWindow time.Duration

	// circuitBreakerByModel stores circuit breakers for each model, allowing concurrent access.
	circuitBreakerByModel sync.Map
}

// modelCircuitBreaker keeps track of error records for a specific model,
// implementing a circular buffer to efficiently manage error history.
type modelCircuitBreaker struct {
	mu sync.RWMutex

	headIndex int
	records   []record
}

// record represents an individual error occurrence with details about the reason for the error
// and the timestamp when it happened. This information is used to assess the model's availability.
type record struct {
	// statusCode equals the response status code or 0 in case of unknown error.
	statusCode int
	timestamp  time.Time
}

// newModelsLoadTracker initializes and returns a modelsLoadTracker with the specified
// failure threshold and timeout period for assessing model availability.
func newModelsLoadTracker() *modelsLoadTracker {
	return &modelsLoadTracker{
		failureRatio:          0.95,
		maxRecords:            100,
		evaluationWindow:      1 * time.Minute,
		circuitBreakerByModel: sync.Map{},
	}
}

func newModelCircuitBreaker(maxRecords int) *modelCircuitBreaker {
	return &modelCircuitBreaker{
		headIndex: 0,
		records:   make([]record, maxRecords),
	}
}

// record adds a new record to the model's circuit breaker.
func (mlt *modelsLoadTracker) record(model string, resp *http.Response, reqErr error) {
	var statusCode int
	if errors.Is(reqErr, context.DeadlineExceeded) {
		// special case for timeout
		statusCode = http.StatusGatewayTimeout
	} else if resp != nil {
		statusCode = resp.StatusCode
	} else {
		// We don't have a response object, so we use 0 to represent an unknown error.
		statusCode = 0
	}
	r := record{
		statusCode: statusCode,
		timestamp:  time.Now(),
	}

	var mcb *modelCircuitBreaker
	if v, exists := mlt.getCircuitBreaker(model); exists {
		mcb = v
	} else {
		mcb = mlt.createCircuitBreaker(model)
	}

	mcb.addRecord(r)
}

func (mlt *modelsLoadTracker) getCircuitBreaker(model string) (mcb *modelCircuitBreaker, exists bool) {
	if v, ok := mlt.circuitBreakerByModel.Load(model); ok {
		return v.(*modelCircuitBreaker), true
	}
	return nil, false
}

func (mlt *modelsLoadTracker) createCircuitBreaker(model string) *modelCircuitBreaker {
	v, _ := mlt.circuitBreakerByModel.LoadOrStore(model, newModelCircuitBreaker(mlt.maxRecords))
	return v.(*modelCircuitBreaker)
}

// isModelAvailable returns false if the percentage of failures for model in the timeframe
// is greater than the failureThreshold.Otherwise, returns true.
func (mlt *modelsLoadTracker) isModelAvailable(model string) bool {
	mcb, ok := mlt.getCircuitBreaker(model)
	if !ok {
		return true
	}

	return mcb.getFailureRatio(time.Now(), mlt.evaluationWindow) < mlt.failureRatio
}

func (mcb *modelCircuitBreaker) addRecord(r record) {
	mcb.mu.Lock()
	defer mcb.mu.Unlock()

	mcb.records[mcb.headIndex] = r
	mcb.headIndex++
	if mcb.headIndex >= len(mcb.records) {
		mcb.headIndex = 0
	}
}

// getFailureRatio calculates the percentage of failures within the specified evaluation window.
func (mcb *modelCircuitBreaker) getFailureRatio(now time.Time, evaluationWindow time.Duration) float64 {
	mcb.mu.RLock()
	defer mcb.mu.RUnlock()

	var failureCount int
	var reqCount int
	for _, r := range mcb.records {
		// Check if record is within the evaluation window
		if now.Sub(r.timestamp) <= evaluationWindow {
			reqCount++

			switch r.statusCode {
			case http.StatusTooManyRequests, http.StatusGatewayTimeout, http.StatusInternalServerError:
				failureCount++
			}
		}
	}

	if reqCount == 0 {
		return 0
	}

	return float64(failureCount) / float64(reqCount)
}
