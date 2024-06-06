package redislock

import (
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

func TestOnlyOne(t *testing.T) {
	rcache.SetupForTest(t)

	logger := logtest.NoOp(t)
	client := redis.NewClient(&redis.Options{Addr: rcache.TestAddr})

	// If the algorithm is correct, there should be no data race detected in tests,
	// and the final count should be 10.
	count := 0
	for i := 0; i < 10; i++ {
		err := OnlyOne(logger, client, "test", 3*time.Second, func() error {
			count++
			return nil
		})
		require.NoError(t, err)
	}
	assert.Equal(t, 10, count)
}
