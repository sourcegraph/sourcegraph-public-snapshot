package syncjobs

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

// Status describes the outcome of an authz sync job.
type Status struct {
	JobType   string    `json:"job_type"`
	JobID     int32     `json:"job_id"`
	Completed time.Time `json:"completed"`

	// Status is one of "ERROR" or "SUCCESS"
	Status  string `json:"status"`
	Message string `json:"message"`

	// Per-provider states during the sync job
	Providers []database.PermissionSyncCodeHostState `json:"providers"`
}
