package executorqueue

import (
	"fmt"
	"testing"
)

func TestNormalizeQueueAllocation(t *testing.T) {
	t.Run("Not configured", func(t *testing.T) {
		for _, testValue := range []map[string]float64{
			{},
			{"aws": 0.5},
			{"aws": 1.0},
			{"gcp": 0.5},
			{"gcp": 1.0},
			{"aws": 0.5, "gcp": 0.5},
			{"aws": 1.0, "gcp": 1.0},
		} {
			t.Run(fmt.Sprintf("%v", testValue), func(t *testing.T) {
				queueAllocation, err := normalizeQueueAllocation("", testValue, false, false)
				if err != nil {
					t.Fatalf("unexpected error: %q", err)
				}

				// any values are set back to zero
				assertAllocation(t, queueAllocation, 0, 0)
			})
		}
	})

	t.Run("AWS enabled", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			queueAllocation, err := normalizeQueueAllocation("", nil, true, false)
			if err != nil {
				t.Fatalf("unexpected error: %q", err)
			}

			// unconfigured allocations
			assertAllocation(t, queueAllocation, 1, 0)
		})

		for _, testValue := range []map[string]float64{
			{"aws": 0.5},
			{"aws": 1.0},
			{"gcp": 0.5},
			{"gcp": 1.0},
			{"aws": 0.5, "gcp": 0.5},
			{"aws": 1.0, "gcp": 1.0},
		} {
			t.Run(fmt.Sprintf("%v", testValue), func(t *testing.T) {
				queueAllocation, err := normalizeQueueAllocation("", testValue, true, false)
				if err != nil {
					t.Fatalf("unexpected error: %q", err)
				}

				// any GCP values are set back to zero
				assertAllocation(t, queueAllocation, testValue["aws"], 0)
			})
		}
	})

	t.Run("GCP enabled", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			queueAllocation, err := normalizeQueueAllocation("", nil, false, true)
			if err != nil {
				t.Fatalf("unexpected error: %q", err)
			}

			// unconfigured allocations
			assertAllocation(t, queueAllocation, 0, 1)
		})

		for _, testValue := range []map[string]float64{
			{"aws": 0.5},
			{"aws": 1.0},
			{"gcp": 0.5},
			{"gcp": 1.0},
			{"aws": 0.5, "gcp": 0.5},
			{"aws": 1.0, "gcp": 1.0},
		} {
			t.Run(fmt.Sprintf("%v", testValue), func(t *testing.T) {
				queueAllocation, err := normalizeQueueAllocation("", testValue, false, true)
				if err != nil {
					t.Fatalf("unexpected error: %q", err)
				}

				// any AWS values are set back to zero
				assertAllocation(t, queueAllocation, 0, testValue["gcp"])
			})
		}
	})

	t.Run("Multi-cloud", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			queueAllocation, err := normalizeQueueAllocation("", nil, true, true)
			if err != nil {
				t.Fatalf("unexpected error: %q", err)
			}

			// unconfigured allocations
			assertAllocation(t, queueAllocation, 1, 1)
		})

		for _, testValue := range []map[string]float64{
			{"aws": 0.5},
			{"aws": 1.0},
			{"gcp": 0.5},
			{"gcp": 1.0},
			{"aws": 0.5, "gcp": 0.5},
			{"aws": 1.0, "gcp": 1.0},
		} {
			t.Run(fmt.Sprintf("%v", testValue), func(t *testing.T) {
				queueAllocation, err := normalizeQueueAllocation("", testValue, true, true)
				if err != nil {
					t.Fatalf("unexpected error: %q", err)
				}

				assertAllocation(t, queueAllocation, testValue["aws"], testValue["gcp"])
			})
		}
	})
}

func assertAllocation(t *testing.T, queueAllocation QueueAllocation, percentageAWS, percentageGCP float64) {
	if queueAllocation.PercentageAWS != percentageAWS {
		t.Fatalf("unexpected AWS percentage. want=%.2f have=%.2f", percentageAWS, queueAllocation.PercentageAWS)
	}

	if queueAllocation.PercentageGCP != percentageGCP {
		t.Fatalf("unexpected GCP percentage. want=%.2f have=%.2f", percentageGCP, queueAllocation.PercentageGCP)
	}
}
