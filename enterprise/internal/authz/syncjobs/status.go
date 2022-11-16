package syncjobs

import "time"

// Status describes the outcome of an authz sync job.
type Status struct {
	JobType   string    `json:"job_type"`
	JobID     int32     `json:"job_id"`
	Completed time.Time `json:"completed"`

	// Status is one of "ERROR" or "SUCCESS"
	Status  string `json:"status"`
	Message string `json:"message"`

	// Per-provider states during the sync job
	Providers []ProviderStatus `json:"providers"`
}

// ProviderStatus describes the state of a provider during an authz sync job.
type ProviderStatus struct {
	ProviderID   string `json:"provider_id"`
	ProviderType string `json:"provider_type"`

	// Status is one of "ERROR" or "SUCCESS"
	Status  string `json:"status"`
	Message string `json:"message"`
}
