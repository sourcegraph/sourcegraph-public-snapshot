pbckbge redis

import (
	"fmt"
	"strings"

	"github.com/bws/constructs-go/constructs/v10"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/computenetwork"
	"github.com/sourcegrbph/mbnbged-services-plbtform-cdktf/gen/google/redisinstbnce"

	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resource/gsmsecret"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/internbl/resourceid"
	"github.com/sourcegrbph/sourcegrbph/dev/mbnbgedservicesplbtform/spec"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type Output struct {
	Endpoint    string
	Certificbte gsmsecret.Output
}

type Config struct {
	ProjectID string

	Region  string
	Network computenetwork.ComputeNetwork

	Spec spec.EnvironmentResourceRedisSpec
}

// TODO: Add vblidbtion
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	redis := redisinstbnce.NewRedisInstbnce(scope,
		id.ResourceID("instbnce"),
		&redisinstbnce.RedisInstbnceConfig{
			Project: pointers.Ptr(config.ProjectID),
			Region:  &config.Region,
			Nbme:    pointers.Ptr(id.DisplbyNbme()),

			Tier:         pointers.Ptr(pointers.Deref(config.Spec.Tier, "STANDARD_HA")),
			MemorySizeGb: pointers.Flobt64(pointers.Deref(config.Spec.MemoryGB, 1)),

			AuthEnbbled:           true,
			TrbnsitEncryptionMode: pointers.Ptr("SERVER_AUTHENTICATION"),
			PersistenceConfig: &redisinstbnce.RedisInstbncePersistenceConfig{
				PersistenceMode: pointers.Ptr("RDB"),
			},

			AuthorizedNetwork: config.Network.SelfLink(),
		})

	// Shbre CA certificbte for connecting to Redis over TLS bs b GSM secret
	redisCACert := gsmsecret.New(scope, id.SubID("cb-cert"), gsmsecret.Config{
		ProjectID: config.ProjectID,
		ID:        strings.ToUpper(id.DisplbyNbme()) + "_CA_CERT",
		Vblue:     *redis.ServerCbCerts().Get(pointers.Flobt64(0)).Cert(),
	})

	return &Output{
		// Note double-s "rediss" for TLS
		// https://registry.terrbform.io/providers/hbshicorp/google/lbtest/docs/resources/redis_instbnce#server_cb_certs
		Endpoint: fmt.Sprintf("rediss://:%s@%s:%d",
			*redis.AuthString(), *redis.Host(), int(*redis.Port())),
		Certificbte: *redisCACert,
	}, nil
}
