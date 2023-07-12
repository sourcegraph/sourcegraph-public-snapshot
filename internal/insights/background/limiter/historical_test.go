package limiter_test

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/insights/background/limiter"
)

func TestGetHistoricLimit(t *testing.T) {
	t.Run("Is singleton", func(t *testing.T) {
		limiter1 := limiter.HistoricalWorkRate()
		limiter2 := limiter.HistoricalWorkRate()

		if limiter1 != limiter2 {
			t.Error("both limiters should be the same instance")
		}
	})

}
