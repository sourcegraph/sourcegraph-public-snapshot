package ratelimit

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Handle(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	prefix := "__test__" + t.Name()
	url := "https://github.com/"
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

	assertValFromRedis := func(kv redispool.KeyValue, key string, expectedVal int32) {
		val, err := kv.Get(key).Int()
		assert.NoError(t, err)
		assert.Equal(t, expectedVal, int32(val))
	}

	kv := redispool.NewKeyValue(redisHost, &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
	})
	rateLimiter, err := redispool.NewTestRateLimiterWithPoolAndPrefix(pool, prefix)
	require.NoError(t, err)

	rateLimitConfigValues := int32(10)
	codeHost := &types.CodeHost{
		Kind:                        extsvc.KindGitHub,
		URL:                         url,
		APIRateLimitQuota:           &rateLimitConfigValues,
		APIRateLimitIntervalSeconds: &rateLimitConfigValues,
		GitRateLimitQuota:           &rateLimitConfigValues,
		GitRateLimitIntervalSeconds: &rateLimitConfigValues,
	}
	err = db.CodeHosts().Create(ctx, codeHost)
	require.NoError(t, err)

	// Create the external service so that the first code host appears when the handler calls GetByURL.
	confGet := func() *conf.Unified { return &conf.Unified{} }
	extsvcConfig := extsvc.NewUnencryptedConfig(`{"url": "https://github.com/", "repositoryQuery": ["none"], "token": "abc"}`)
	err = db.ExternalServices().Create(ctx, confGet, &types.ExternalService{
		CodeHostID: &codeHost.ID,
		Kind:       codeHost.Kind,
		Config:     extsvcConfig,
	})
	require.NoError(t, err)

	// Create the handler to start the test
	h := handler{
		codeHostStore: db.CodeHosts(),
		rateLimiter:   ratelimit.NewCodeHostRateLimiter(rateLimiter),
		logger:        logger,
	}
	err = h.Handle(ctx)
	assert.NoError(t, err)

	bucketTypeAPI := "api_tokens"
	bucketTypeGit := "git_tokens"
	configKeyBucketQuota := "config:bucket_quota"
	configKeyReplenishmentIntervalSeconds := "config:bucket_replenishment_interval_seconds"
	apiCapKey := fmt.Sprintf("%s:%s:%s:%s", prefix, url, bucketTypeAPI, configKeyBucketQuota)
	apiReplenishmentKey := fmt.Sprintf("%s:%s:%s:%s", prefix, url, bucketTypeAPI, configKeyReplenishmentIntervalSeconds)
	gitCapKey := fmt.Sprintf("%s:%s:%s:%s", prefix, url, bucketTypeGit, configKeyBucketQuota)
	gitReplenishmentKey := fmt.Sprintf("%s:%s:%s:%s", prefix, url, bucketTypeGit, configKeyReplenishmentIntervalSeconds)
	assertValFromRedis(kv, apiCapKey, rateLimitConfigValues)
	assertValFromRedis(kv, apiReplenishmentKey, rateLimitConfigValues)
	assertValFromRedis(kv, gitCapKey, rateLimitConfigValues)
	assertValFromRedis(kv, gitReplenishmentKey, rateLimitConfigValues)

	// Update the rate limit config in Postgres to use defaults/maxes
	maxRateLimitQuota := int32(math.MaxInt32)

	// This should default to max int32
	codeHost.APIRateLimitQuota = nil
	// This should default to  3600
	codeHost.APIRateLimitIntervalSeconds = nil
	codeHost.GitRateLimitQuota = &maxRateLimitQuota
	// This should default to 1
	codeHost.GitRateLimitIntervalSeconds = nil
	err = db.CodeHosts().Update(ctx, codeHost)
	assert.NoError(t, err)
	err = h.Handle(ctx)
	assert.NoError(t, err)

	// Check updated values are in Redis
	defaultAPIRateLimitReplenishmentInterval := int32(3600)
	defaultGitRateLimitReplenishmentInterval := int32(1)
	assertValFromRedis(kv, apiCapKey, maxRateLimitQuota)
	assertValFromRedis(kv, apiReplenishmentKey, defaultAPIRateLimitReplenishmentInterval)
	assertValFromRedis(kv, gitCapKey, maxRateLimitQuota)
	assertValFromRedis(kv, gitReplenishmentKey, defaultGitRateLimitReplenishmentInterval)
}
