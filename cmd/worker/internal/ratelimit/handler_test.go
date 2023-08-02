package ratelimit

import (
	"context"
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
	codeHost := &types.CodeHost{
		Kind:                        extsvc.KindGitHub,
		URL:                         url,
		APIRateLimitQuota:           &ten,
		APIRateLimitIntervalSeconds: &ten,
		GitRateLimitQuota:           &ten,
		GitRateLimitIntervalSeconds: &ten,
	}
	t.Cleanup(func() {
		pool := &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", "127.0.0.1:6379")
			},
		}

		c := pool.Get()
		defer c.Close()
		err := deleteAllKeysWithPrefix(c, prefix)
		if err != nil {
			t.Logf("Failed to clear redis: %+v\n", err)
		}
	})
	err := db.CodeHosts().Create(ctx, codeHost)
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

	assertValFromRedis := func(key string, expectedVal int32, kv redispool.KeyValue) {
		val, err := kv.Get(key).Int()
		assert.NoError(t, err)
		assert.Equal(t, expectedVal, int32(val))
	}

	kv := redispool.NewKeyValue("127.0.0.1:6379", &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
	})
	h := handler{
		codeHostStore:  db.CodeHosts(),
		redisKeyPrefix: prefix,
		kv:             kv,
	}
	err = h.Handle(ctx, obsCtx)
	assert.NoError(t, err)
	apiCapKey, apiReplenishmentKey, gitCapKey, gitReplenishmentKey := redispool.GetCodeHostRateLimiterConfigKeys(h.redisKeyPrefix, url)
	assertValFromRedis(apiCapKey, ten, kv)
	assertValFromRedis(apiReplenishmentKey, ten, kv)
	assertValFromRedis(gitCapKey, ten, kv)
	assertValFromRedis(gitReplenishmentKey, ten, kv)

	// Update the rate limit config in Postgres
	thousand := int32(1000)
	fiveHundered := int32(500)
	twoHundred := int32(200)
	oneHundred := int32(100)
	codeHost.APIRateLimitQuota = &thousand
	codeHost.APIRateLimitIntervalSeconds = &fiveHundered
	codeHost.GitRateLimitQuota = &twoHundred
	codeHost.GitRateLimitIntervalSeconds = &oneHundred
	err = db.CodeHosts().Update(ctx, codeHost)
	assert.NoError(t, err)
	err = h.Handle(ctx, obsCtx)
	assert.NoError(t, err)

	//Check Updated values are in Redis
	assertValFromRedis(apiCapKey, thousand, kv)
	assertValFromRedis(apiReplenishmentKey, fiveHundered, kv)
	assertValFromRedis(gitCapKey, twoHundred, kv)
	assertValFromRedis(gitReplenishmentKey, oneHundred, kv)
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
