// Package contract implements configuration and helpers based on environment
// variables expected to be provided by MSP infrastructure (the 'MSP contract').
// It generally has provisions for local development built-in, typically toggled
// by the 'MSP=false' environment variable.
//
// Service implementors should generally default to implementing interfaces
// expected by lib/managedservicesplatform/runtime instead of using package-level
// constructors provided by this package - these are exported only for programs
// that haven't yet, or can't, migrate to the runtime package.
//
// Simple example usage if you need to integrate the contract package directly:
//
//	// Parse the environment into an Env instance.
//	e, _ := contract.ParseEnv([]string{"MSP=true"})
//
//	// Extract Contract instance from Env configuration.
//	c := contract.New(logger, service, e)
//
//	// Also load other custom configuration here from Env you want here
//
//	// Check for errors on Env retrieval (missing/invalid values, etc.)
//	if err := e.Validate(); err != nil { ... }
//
//	// Use Contract helpers and configuration values
//	writer, _ := c.BigQuery.GetTableWriter(ctx, "my-table")
//	writer.Write(...)
//
// For more help, please reach out to #discuss-core-services or refer to
// go/msp: https://handbook.sourcegraph.com/departments/engineering/teams/core-services/managed-services/platform/
package contract

import (
	"net/http"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// Contract loads standardized MSP-provisioned (Managed Services Platform)
// configuration. Most configuration is sourced from environment variables -
// refer to each "sub-contract"'s docstrings for more details. All environment
// variables referenced as part of the MSP contract, including whether they
// are required or not and their defaults, are reported when running a service
// using the MSP runtime with the '-help' flag.
//
// The "sub-contract" types (e.g. Contract.PostgreSQL's type being the private
// postgreSQLContract) are intentionally unavailable to callers, despite being
// exported on Contract. This is intentional: callers should not be passing
// around the sub-contract types directly, as they are implementation details.
// Instead, prefer to create your own parameterizations or create useful types
// (for example, open a PostgreSQL connection), and pass that around instead.
// Even though these types are not exported, their respective exported methods
// and properties are.
//
// For help with the Contract type, and the contract package in general, please
// refer to the package docs or reach out to #discuss-core-services directly.
type Contract struct {
	// MSP indicates if we are running in a live Managed Services Platform
	// environment. In local development, this should generally be false.
	MSP bool
	// EnvironmentID is the ID of the MSP environment this service is deployed
	// in. In local development, this should be 'unknown' if ENVIRONMENT_ID is
	// not set.
	EnvironmentID string

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

// ServiceContract loads standardized MSP-provisioned (Managed Services Platform)
// configuration for a service.
type ServiceContract struct {
	// Port is the port the service must listen on.
	Port int
	// ExternalDNSName is the DNS name the service uses, if one is configured.
	ExternalDNSName *string

	Contract
}

// JobContract loads standardized MSP-provisioned (Managed Services Platform)
// configuration for a job.
type JobContract struct {
	Contract
}

type ServiceMetadataProvider interface {
	// Name is the service name, typically the all-lowercase, dash-delimited,
	// machine-friendly 'id' of the service in its corresponding MSP service
	// specification (e.g. 'telemetry-gateway')
	Name() string
	// Version should indicate the stamped version of the running service
	// program. It is implementation-dependent - the value gets included in
	// logs, traces, error reports, and so on, so choose any format that is
	// operationally useful. It is also reported when running a service using
	// the MSP runtime with the '-help' flag.
	Version() string
}

type internalContract struct {
	// logger for use in contract internals only.
	logger log.Logger
	// service is a reference to the service that is being configured.
	service ServiceMetadataProvider
	// environmentID is the ID of the MSP environment this service is deployed in.
	environmentID string
}

// NewService returns a new Contract instance from configuration parsed from the Env
// instance. Values are expected per the 'MSP contract'.
func NewService(logger log.Logger, service ServiceMetadataProvider, env *Env) ServiceContract {
	return ServiceContract{
		Port:            env.GetInt("PORT", "", "service port"),
		ExternalDNSName: env.GetOptional("EXTERNAL_DNS_NAME", "external DNS name provisioned for the service"),
		Contract:        newBase(logger, service, env),
	}
}

// NewJob returns a new Contract instance from configuration parsed from the Env
// instance. Values are expected per the 'MSP contract'.
func NewJob(logger log.Logger, service ServiceMetadataProvider, env *Env) JobContract {
	return JobContract{
		Contract: newBase(logger, service, env),
	}
}

func newBase(logger log.Logger, service ServiceMetadataProvider, env *Env) Contract {
	defaultGCPProjectID := pointers.Deref(env.GetOptional("GOOGLE_CLOUD_PROJECT", "GCP project ID"), "")
	internal := internalContract{
		logger:        logger,
		service:       service,
		environmentID: env.Get("ENVIRONMENT_ID", "unknown", "MSP Service Environment ID"),
	}
	isMSP := env.GetBool("MSP", "false", "indicates if we are running in a MSP environment")

	return Contract{
		MSP:           isMSP,
		EnvironmentID: internal.environmentID,
		RedisEndpoint: env.GetOptional("REDIS_ENDPOINT", "full Redis address, including any prerequisite authentication"),

		PostgreSQL: loadPostgreSQLContract(env, isMSP),
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
			return "", errors.Newf("invalid token scheme %s", typ[0])
		}

		token = typ[1]
	}

	return token, nil
}
