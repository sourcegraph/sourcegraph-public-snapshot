package redis

import (
	"fmt"
	"strings"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computenetwork"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/redisinstance"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/internal/pointer"
)

type Output struct {
	Endpoint    string
	Certificate gsmsecret.Output
}

type Config struct {
	Project project.Project

	Region  string
	Network computenetwork.ComputeNetwork

	Spec spec.EnvironmentResourceRedisSpec
}

// TODO: Add validation
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	redis := redisinstance.NewRedisInstance(scope,
		id.ResourceID("instance"),
		&redisinstance.RedisInstanceConfig{
			Project: config.Project.ProjectId(),
			Region:  &config.Region,
			Name:    pointer.Value(id.DisplayName()),

			Tier:         pointer.Value(pointer.IfNil(config.Spec.Tier, "STANDARD_HA")),
			MemorySizeGb: pointer.Float64(pointer.IfNil(config.Spec.MemoryGB, 1)),

			AuthEnabled:           true,
			TransitEncryptionMode: pointer.Value("SERVER_AUTHENTICATION"),
			PersistenceConfig: &redisinstance.RedisInstancePersistenceConfig{
				PersistenceMode: pointer.Value("RDB"),
			},

			AuthorizedNetwork: config.Network.SelfLink(),
		})

	// Share CA certificate for connecting to Redis over TLS as a GSM secret
	redisCACert := gsmsecret.New(scope, id.SubID("ca-cert"), gsmsecret.Config{
		Project: config.Project,
		ID:      strings.ToUpper(id.DisplayName()) + "_CA_CERT",
		Value:   *redis.ServerCaCerts().Get(pointer.Float64(0)).Cert(),
	})

	return &Output{
		// Note double-s "rediss" for TLS
		// https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/redis_instance#server_ca_certs
		Endpoint: fmt.Sprintf("rediss://:%s@%s:%d",
			*redis.AuthString(), *redis.Host(), int(*redis.Port())),
		Certificate: *redisCACert,
	}, nil
}
