package ratelimit

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Handle(t *testing.T) {
	obsCtx := observation.TestContextTB(t)
	logger := obsCtx.Logger
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	prefix := "__test__" + t.Name()
	ctx := context.Background()
	ten := int32(10)
	url := "https://github.com/"
	redisHost := "127.0.0.1:6379"
	codeHost := &types.CodeHost{
		Kind:                        extsvc.KindGitHub,
		URL:                         url,
		APIRateLimitQuota:           &ten,
		APIRateLimitIntervalSeconds: &ten,
		GitRateLimitQuota:           &ten,
		GitRateLimitIntervalSeconds: &ten,
	}
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
	assertValFromRedis := func(key string, expectedVal int32, kv redispool.KeyValue) {
		val, err := kv.Get(key).Int()
		assert.NoError(t, err)
		assert.Equal(t, expectedVal, int32(val))
	}

	kv := redispool.NewKeyValue(redisHost, &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
	})
	rateLimiter, err := redispool.NewRateLimiterWithPoolAndPrefix(pool, prefix)
	require.NoError(t, err)

	t.Cleanup(func() {
		c := pool.Get()
		err := deleteAllKeysWithPrefix(c, prefix)
		if err != nil {
			t.Logf("Failed to clear redis: %+v\n", err)
		}
		c.Close()
	})
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
		codeHostStore:  db.CodeHosts(),
		redisKeyPrefix: prefix,
		ratelimiter:    rateLimiter,
	}
	err = h.Handle(ctx, obsCtx)
	assert.NoError(t, err)
	apiCapKey, apiReplenishmentKey, gitCapKey, gitReplenishmentKey := redispool.GetCodeHostRateLimiterConfigKeys(prefix, url)
	assertValFromRedis(apiCapKey, ten, kv)
	assertValFromRedis(apiReplenishmentKey, ten, kv)
	assertValFromRedis(gitCapKey, ten, kv)
	assertValFromRedis(gitReplenishmentKey, ten, kv)

	// Update the rate limit config in Postgres to use defaults/maxes
	thirtySixHundred := int32(3600)
	maxInt := int32(math.MaxInt32)
	one := int32(1)
	// This should default to max int32
	codeHost.APIRateLimitQuota = nil
	// This should default to  3600
	codeHost.APIRateLimitIntervalSeconds = nil
	codeHost.GitRateLimitQuota = &maxInt
	// This should default to 1
	codeHost.GitRateLimitIntervalSeconds = nil
	err = db.CodeHosts().Update(ctx, codeHost)
	assert.NoError(t, err)
	err = h.Handle(ctx, obsCtx)
	assert.NoError(t, err)

	// Check updated values are in Redis
	assertValFromRedis(apiCapKey, maxInt, kv)
	assertValFromRedis(apiReplenishmentKey, thirtySixHundred, kv)
	assertValFromRedis(gitCapKey, maxInt, kv)
	assertValFromRedis(gitReplenishmentKey, one, kv)
}

func deleteAllKeysWithPrefix(c redis.Conn, prefix string) error {
	const script = `
redis.replicate_commands()
local cursor = '0'
local prefix = ARGV[1]
local batchSize = ARGV[2]
local result = ''
repeat
	local keys = redis.call('SCAN', cursor, 'MATCH', prefix, 'COUNT', batchSize)
	if #keys[2] > 0
	then
		result = redis.call('DEL', unpack(keys[2]))
	end

	cursor = keys[1]
until cursor == '0'
return result
`

	_, err := c.Do("EVAL", script, 0, prefix+":*", 100)
	return err
}
