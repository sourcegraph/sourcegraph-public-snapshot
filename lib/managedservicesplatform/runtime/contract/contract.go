package contract

import (
	"net/http"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// Contract loads standardized MSP-provisioned (Managed Services Platform)
// configuration.
type Contract struct {
	// MSP indicates if we are running in a live Managed Services Platform
	// environment. In local development, this should generally be false.
	MSP bool
	// EnvironmentID is the ID of the MSP environment this service is deployed
	// in. In local development, this should be 'unknown' if ENVIRONMENT_ID is
	// not set.
	EnvironmentID string
	// Port is the port the service must listen on.
	Port int
	// ExternalDNSName is the DNS name the service uses, if one is configured.
	ExternalDNSName *string

	// RedisEndpoint is the full connection string of a MSP Redis instance if
	// provisioned, including any prerequisite authentication.
	RedisEndpoint *string

	// PostgreSQL has helpers and configuration for MSP PostgreSQL instances.
	PostgreSQL postgreSQLContract

	// BigQuery has embedded helpers and configuration for MSP-provisioned
	// BigQuery datasets and tables.
	BigQuery bigQueryContract

	// Diagnostics embedded helpers and configuration for MSP-provisioned
	// diagnostic services.
	Diagnostics diagnosticsContract

	// internal configuration for MSP internals that are not exposed to service
	// developers.
	internal internalContract
}

type ServiceMetadataProvider interface {
	Name() string
	Version() string
}

type internalContract struct {
	// logger for use in contract internals only.
	logger log.Logger
	// service is a reference to the service that is being configured.
	service ServiceMetadataProvider
}

// New returns a new Contract instance from configuration parsed from the Env
// instance. Values are expected per the 'MSP contract'.
func New(logger log.Logger, service ServiceMetadataProvider, env *Env) Contract {
	defaultGCPProjectID := pointers.Deref(env.GetOptional("GOOGLE_CLOUD_PROJECT", "GCP project ID"), "")
	internal := internalContract{
		logger:  logger,
		service: service,
	}
	isMSP := env.GetBool("MSP", "false", "indicates if we are running in a MSP environment")

	return Contract{
		MSP:             isMSP,
		EnvironmentID:   env.Get("ENVIRONMENT_ID", "unknown", "MSP Service Environment ID"),
		Port:            env.GetInt("PORT", "", "service port"),
		ExternalDNSName: env.GetOptional("EXTERNAL_DNS_NAME", "external DNS name provisioned for the service"),
		RedisEndpoint:   env.GetOptional("REDIS_ENDPOINT", "full Redis address, including any prerequisite authentication"),

		PostgreSQL: loadPostgreSQLContract(env),
		BigQuery:   loadBigQueryContract(env),

		Diagnostics: loadDiagnosticsContract(logger, env, defaultGCPProjectID, internal, isMSP),

		internal: internal,
	}
}

func extractBearer(h http.Header) (string, error) {
	var token string

	if authHeader := h.Get("Authorization"); authHeader != "" {
		typ := strings.SplitN(authHeader, " ", 2)
		if len(typ) != 2 {
			return "", errors.New("token type missing in Authorization header")
		}
		if strings.ToLower(typ[0]) != "bearer" {
			return "", errors.Newf("invalid token type %s", typ[0])
		}

		token = typ[1]
	}

	return token, nil
}
