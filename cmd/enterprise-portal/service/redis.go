package service

import (
	"context"
	"net"
	"os"
	"time"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// newRedisClient creates a new Redis client with tracing enabled.
func newRedisClient(msp bool, endpoint *string) (*redis.Client, error) {
	var redisOpts *redis.Options
	if msp && endpoint != nil {
		var err error
		redisOpts, err = redis.ParseURL(*endpoint)
		if err != nil {
			return nil, errors.Wrap(err, "invalid endpoint")
		}
	} else {
		// Local dev fallback
		redisOpts = &redis.Options{Addr: os.ExpandEnv("$REDIS_HOST:$REDIS_PORT")}
	}
	redisClient := redis.NewClient(redisOpts)
	redisClient.AddHook(&redisTracingWrappingHook{})
	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		return nil, errors.Wrap(err, "instrument tracing")
	}
	return redisClient, nil
}

var redisTracer = otel.Tracer("enterprise-portal/service/redis")

// redisTracingWrappingHook creates a parent trace span called "redis" to make
// Redis spans look nicer in the UI, it should be added before the real Redis
// OTEL tracing hook.
type redisTracingWrappingHook struct{}

func (h *redisTracingWrappingHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// The DialHook already has "redis." prefix, thus we do nothing.
		return next(ctx, network, addr)
	}
}

func (h *redisTracingWrappingHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		ctx, span := redisTracer.Start(ctx, "redis")
		defer span.End()
		return next(ctx, cmd)
	}
}

func (h *redisTracingWrappingHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		// The ProcessPipelineHook already has "redis." prefix, thus we do nothing.
		return next(ctx, cmds)
	}
}

// newRedisKVClient provides a redis client compatible with monorepo-isms.
func newRedisKVClient(endpoint *string) redispool.KeyValue {
	if endpoint == nil {
		// Local dev fallback: use monorepo default client
		return redispool.Cache
	}
	return redispool.NewKeyValue(*endpoint, &redigo.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		MaxActive:   1000,
	})
}
