package runtime

import (
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/internal/opentelemetry"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// Contract loads standardized MSP-provisioned (Managed Services Platform)
// configuration.
type Contract struct {
	// MSP indicates if we are running in a MSP environment.
	MSP bool
	// Port is the port the service must listen on.
	Port int
	// ExternalDNSName is the DNS name the service uses, if one is configured.
	ExternalDNSName *string
	// RedisEndpoint is the full Redis address, including any prerequisite
	// authentication.
	RedisEndpoint *string

	// PostgreSQL has helpers and configuration for MSP PostgreSQL instances.
	PostgreSQL postgreSQLContract

	// BigQuery has embedded helpers and configuration for MSP-provisioned
	// BigQuery datasets and tables.
	BigQuery bigQueryContract

	// internal configuration for MSP internals that are not exposed to service
	// developers.
	internal internalContract
}

type internalContract struct {
	opentelemetry opentelemetry.Config
	sentryDSN     *string
}

func newContract(env *Env) Contract {
	defaultGCPProjectID := pointers.Deref(env.GetOptional("GOOGLE_CLOUD_PROJECT", "GCP project ID"), "")

	return Contract{
		MSP:             env.GetBool("MSP", "false", "indicates if we are running in a MSP environment"),
		Port:            env.GetInt("PORT", "", "service port"),
		ExternalDNSName: env.GetOptional("EXTERNAL_DNS_NAME", "external DNS name provisioned for the service"),
		RedisEndpoint:   env.GetOptional("REDIS_ENDPOINT", "full Redis address, including any prerequisite authentication"),

		PostgreSQL: loadPostgreSQLContract(env),
		BigQuery:   loadBigQueryContract(env),

		internal: internalContract{
			opentelemetry: opentelemetry.Config{
				GCPProjectID: pointers.Deref(
					env.GetOptional("OTEL_GCP_PROJECT_ID", "GCP project ID for OpenTelemetry export"),
					defaultGCPProjectID),
			},
			sentryDSN: env.GetOptional("SENTRY_DSN", "Sentry error reporting DSN"),
		},
	}
}
