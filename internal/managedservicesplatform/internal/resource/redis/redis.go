package redis

import (
	"fmt"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/redisinstance"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/internal/pointer"
)

type Output struct {
	Endpoint string
}

type Config struct {
	Spec spec.EnvironmentResourceRedisSpec
}

func New(scope constructs.Construct, name string, config Config) (*Output, error) {
	redis := redisinstance.NewRedisInstance(scope, &name, &redisinstance.RedisInstanceConfig{
		Tier:         pointer.Value(config.Spec.Tier),
		MemorySizeGb: pointer.Float64(config.Spec.MemoryGB),
	})

	return &Output{
		// Note double-s "rediss" for TLS
		Endpoint: fmt.Sprintf("rediss://:%s@%s:%d",
			*redis.AuthString(), *redis.Host(), int(*redis.Port())),
	}, nil
}
