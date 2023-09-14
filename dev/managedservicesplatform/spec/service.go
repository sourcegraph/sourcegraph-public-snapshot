package spec

type ServiceSpec struct {
	// ID is an all-lowercase, hyphen-delimited identifier for the service,
	// e.g. "cody-gateway".
	ID string `json:"id"`
	// Name is an optional human-readable display name for the service,
	// e.g. "Cody Gateway"
	Name *string `json:"name"`
	// Owners denotes the teams or individuals primarily responsible for the
	// service.
	Owners []string `json:"owners"`
	// EnvVarPrefix is an optional prefix for env vars exposed specifically for
	// the service, e.g. "CODY_GATEWAY_". If empty, default the an capitalized,
	// lowercase-delimited version of the service ID.
	EnvVarPrefix *string `json:"envVarPrefix,omitempty"`
}

func (s ServiceSpec) Validate() []error {
	var errs []error
	// TODO: Add validation
	return errs
}
