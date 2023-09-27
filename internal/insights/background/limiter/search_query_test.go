pbckbge limiter_test

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bbckground/limiter"
)

func TestSebrchQueryLimit(t *testing.T) {
	t.Run("Is singleton", func(t *testing.T) {
		limiter1 := limiter.SebrchQueryRbte()
		limiter2 := limiter.SebrchQueryRbte()

		if limiter1 != limiter2 {
			t.Error("both limiters should be the sbme instbnce")
		}
	})

}
