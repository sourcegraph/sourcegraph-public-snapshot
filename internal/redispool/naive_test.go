package redispool_test

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

func TestInMemoryKeyValue(t *testing.T) {
	testKeyValue(t, redispool.MemoryKeyValue())
}
