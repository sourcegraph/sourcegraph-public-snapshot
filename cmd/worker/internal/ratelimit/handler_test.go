package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestHandler_Handle(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))

	prefix := "__test__" + t.Name()
	redisHost := "127.0.0.1:6379"

	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisHost)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	t.Cleanup(func() {
		c := pool.Get()
		err := redispool.DeleteAllKeysWithPrefix(c, prefix)
		if err != nil {
			t.Logf("Failed to clear redis: %+v\n", err)
		}
		c.Close()
	})

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			GitMaxCodehostRequestsPerSecond: pointers.Ptr(1),
		},
	})
	defer conf.Mock(nil)

	// Create the external service so that the first code host appears when the handler calls GetByURL.
	confGet := func() *conf.Unified { return &conf.Unified{} }
	extsvcConfig := extsvc.NewUnencryptedConfig(`{"url": "https://github.com/", "token":"abc", "repositoryQuery": ["none"], "rateLimit": {"enabled": true, "requestsPerHour": 150}}`)
	svc := &types.ExternalService{
		Kind:   extsvc.KindGitHub,
		Config: extsvcConfig,
	}
	err := db.ExternalServices().Create(ctx, confGet, svc)
	require.NoError(t, err)

	// Create the handler to start the test
	h := handler{
		externalServiceStore: db.ExternalServices(),
		newRateLimiterFunc: func(bucketName string) ratelimit.GlobalLimiter {
			return ratelimit.NewTestGlobalRateLimiter(pool, prefix, bucketName)
		},
		logger: logger,
	}
	err = h.Handle(ctx)
	assert.NoError(t, err)

	info, err := ratelimit.GetGlobalLimiterStateFromPool(ctx, pool, prefix)
	require.NoError(t, err)

	if diff := cmp.Diff(map[string]ratelimit.GlobalLimiterInfo{
		svc.URN(): {
			Burst:             10,
			Limit:             150,
			Interval:          time.Hour,
			LastReplenishment: time.Unix(0, 0),
		},
		ratelimit.GitRPSLimiterBucketName: {
			Burst:             10,
			Limit:             1,
			Interval:          time.Second,
			LastReplenishment: time.Unix(0, 0),
		},
	}, info); diff != "" {
		t.Fatal(diff)
	}
}
