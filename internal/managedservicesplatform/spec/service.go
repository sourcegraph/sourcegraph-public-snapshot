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
}
