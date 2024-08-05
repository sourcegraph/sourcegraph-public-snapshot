package completions

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestModelLoadTracker(t *testing.T) {
	t.Run("creates proper circuit breakers", func(t *testing.T) {
		model1 := "openai/gpt-4o"
		model2 := "openai/gpt-3.5"
		maxRecords := 10

		mlt := modelsLoadTracker{
			failureRatio:          0.2,
			maxRecords:            maxRecords,
			evaluationWindow:      1 * time.Minute,
			circuitBreakerByModel: sync.Map{},
		}

		for _, m := range []string{model1, model2} {
			if _, exists := mlt.getCircuitBreaker(m); exists {
				t.Errorf("Circuit breaker for model %q should not exist", m)
			}
		}

		mlt.record(model1, &http.Response{StatusCode: http.StatusOK}, nil)

		mcb, exists := mlt.getCircuitBreaker(model1)
		require.True(t, exists)
		require.Equal(t, maxRecords, len(mcb.records))

		if _, exists := mlt.getCircuitBreaker(model2); exists {
			t.Errorf("Circuit breaker for model %q should not exist", model2)
		}
	})

	t.Run("adds records with proper status codes", func(t *testing.T) {
		model := "openai/gpt-4o"
		toRecord := []struct {
			resp   *http.Response
			reqErr error
		}{
			{
				resp:   &http.Response{StatusCode: http.StatusOK},
				reqErr: nil,
			},
			{
				resp:   &http.Response{StatusCode: http.StatusTooManyRequests},
				reqErr: nil,
			},
			{
				resp:   &http.Response{StatusCode: http.StatusOK},
				reqErr: nil,
			},
			{
				resp:   &http.Response{StatusCode: http.StatusInternalServerError},
				reqErr: nil,
			},
			{
				resp:   nil,
				reqErr: context.DeadlineExceeded,
			},
			{
				resp:   nil,
				reqErr: context.Canceled,
			},
			{
				resp:   nil,
				reqErr: errors.New("unknown error"),
			},
		}

		mlt := modelsLoadTracker{
			failureRatio:          0.2,
			maxRecords:            10,
			evaluationWindow:      1 * time.Minute,
			circuitBreakerByModel: sync.Map{},
		}

		for _, r := range toRecord {
			mlt.record(model, r.resp, r.reqErr)
		}

		mcb, exists := mlt.getCircuitBreaker(model)
		require.True(t, exists)

		for i, r := range toRecord {
			var wantStatusCode int
			if errors.Is(r.reqErr, context.DeadlineExceeded) {
				wantStatusCode = http.StatusGatewayTimeout
			} else if r.reqErr != nil {
				wantStatusCode = 0
			} else {
				wantStatusCode = r.resp.StatusCode
			}

			require.Equal(t, wantStatusCode, mcb.records[i].statusCode)
		}
	})

	t.Run("performs is model available check", func(t *testing.T) {
		now := time.Now()
		testCases := []struct {
			model   string
			records []record
			want    bool
		}{
			{
				model: "openai/gpt-4o",
				records: []record{
					{statusCode: http.StatusOK, timestamp: now},
					{statusCode: http.StatusTooManyRequests, timestamp: now},
					{statusCode: http.StatusOK, timestamp: now},
					{statusCode: http.StatusOK, timestamp: now},
				},
				want: true,
			},
			{
				model: "openai/gpt-4o",
				records: []record{
					{statusCode: http.StatusOK, timestamp: now},
					{statusCode: http.StatusTooManyRequests, timestamp: now},
					{statusCode: http.StatusTooManyRequests, timestamp: now},
					{statusCode: http.StatusOK, timestamp: now},
				},
				want: false,
			},
			{
				model: "openai/gpt-4o",
				records: []record{
					{statusCode: http.StatusOK, timestamp: now},
					{statusCode: http.StatusTooManyRequests, timestamp: now},
					{statusCode: http.StatusTooManyRequests, timestamp: now},
					{statusCode: http.StatusInternalServerError, timestamp: now},
				},
				want: false,
			},
			{
				model:   "openai/gpt-4o",
				records: []record{},
				want:    true,
			},
		}

		mlt := modelsLoadTracker{
			failureRatio:          0.5,
			maxRecords:            10,
			evaluationWindow:      1 * time.Minute,
			circuitBreakerByModel: sync.Map{},
		}
		for _, tc := range testCases {
			// Mock the circuit breaker with records - we test record addition and failure ration calculation separately.
			mcb := mlt.createCircuitBreaker(tc.model)
			mcb.records = tc.records

			require.Equal(t, tc.want, mlt.isModelAvailable(tc.model))
		}
	})
}

func TestModelCircuitBreaker(t *testing.T) {
	t.Run("overwites old records if max capacity is reached", func(t *testing.T) {
		mcb := newModelCircuitBreaker(2)

		mcb.addRecord(record{statusCode: http.StatusOK, timestamp: time.Now()})
		mcb.addRecord(record{statusCode: http.StatusTooManyRequests, timestamp: time.Now()})
		require.Equal(t, http.StatusOK, mcb.records[0].statusCode)
		require.Equal(t, http.StatusTooManyRequests, mcb.records[1].statusCode)

		mcb.addRecord(record{statusCode: http.StatusInternalServerError, timestamp: time.Now()})
		require.Equal(t, http.StatusInternalServerError, mcb.records[0].statusCode)
		require.Equal(t, http.StatusTooManyRequests, mcb.records[1].statusCode)

		mcb.addRecord(record{statusCode: http.StatusGatewayTimeout, timestamp: time.Now()})
		require.Equal(t, http.StatusInternalServerError, mcb.records[0].statusCode)
		require.Equal(t, http.StatusGatewayTimeout, mcb.records[1].statusCode)
	})

	t.Run("calculates the average failure rate", func(t *testing.T) {
		now := time.Now()
		evaluationWindow := 1 * time.Minute
		testCases := []struct {
			records []record
			want    float64
		}{
			{
				records: []record{
					// outside evaluationWindow
					{statusCode: http.StatusOK, timestamp: now.Add(-120 * time.Second)},
					{statusCode: http.StatusTooManyRequests, timestamp: now.Add(-90 * time.Second)},

					// within evaluationWindow
					{statusCode: http.StatusOK, timestamp: now.Add(-60 * time.Second)},
					{statusCode: http.StatusOK, timestamp: now.Add(-50 * time.Second)},
					{statusCode: http.StatusGatewayTimeout, timestamp: now.Add(-40 * time.Second)},
					{statusCode: http.StatusInternalServerError, timestamp: now.Add(-30 * time.Second)},
					{statusCode: http.StatusGatewayTimeout, timestamp: now.Add(-20 * time.Second)},
					{statusCode: http.StatusInternalServerError, timestamp: now.Add(-10 * time.Second)},
					{statusCode: http.StatusGatewayTimeout, timestamp: now.Add(-50 * time.Second)},
					{statusCode: http.StatusGatewayTimeout, timestamp: now},
				},
				want: 0.75,
			},
			{
				records: []record{
					// outside evaluationWindow
					{statusCode: http.StatusTooManyRequests, timestamp: now.Add(-120 * time.Second)},
					{statusCode: http.StatusTooManyRequests, timestamp: now.Add(-90 * time.Second)},

					// within evaluationWindow
					{statusCode: http.StatusOK, timestamp: now.Add(-60 * time.Second)},
					{statusCode: http.StatusOK, timestamp: now.Add(-40 * time.Second)},
					{statusCode: http.StatusGatewayTimeout, timestamp: now.Add(-10 * time.Second)},
					{statusCode: http.StatusOK, timestamp: now},
				},
				want: 0.25,
			},
			{
				records: []record{
					// outside evaluationWindow
					{statusCode: http.StatusTooManyRequests, timestamp: now.Add(-120 * time.Second)},
					{statusCode: http.StatusTooManyRequests, timestamp: now.Add(-90 * time.Second)},
					{statusCode: http.StatusOK, timestamp: now.Add(-70 * time.Second)},

					// no records within evaluationWindow
				},
				want: 0,
			},
			{
				records: []record{
					// outside evaluationWindow
					{statusCode: http.StatusTooManyRequests, timestamp: now.Add(-120 * time.Second)},
					{statusCode: http.StatusTooManyRequests, timestamp: now.Add(-90 * time.Second)},

					// within evaluationWindow
					{statusCode: http.StatusOK, timestamp: now.Add(-60 * time.Second)},
					{statusCode: http.StatusOK, timestamp: now.Add(-40 * time.Second)},
					{statusCode: http.StatusOK, timestamp: now.Add(-10 * time.Second)},
					{statusCode: http.StatusOK, timestamp: now},
				},
				want: 0,
			},
		}

		for _, tc := range testCases {
			mcb := newModelCircuitBreaker(10)
			for _, r := range tc.records {
				mcb.addRecord(r)
			}
			require.Equal(t, tc.want, mcb.getFailureRatio(now, evaluationWindow))
		}

	})
}
