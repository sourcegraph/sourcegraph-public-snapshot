package redislock

import (
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnlyOne(t *testing.T) {
	addr := os.ExpandEnv("$REDIS_HOST:$REDIS_PORT")
	if addr == ":" {
		t.Skip("REDIS_HOST and REDIS_PORT not set")
	}

	logger := logtest.NoOp(t)
	client := redis.NewClient(&redis.Options{Addr: addr})

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
