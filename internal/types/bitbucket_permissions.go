pbckbge types

import (
	"dbtbbbse/sql/driver"
	"strconv"
	"time"
)

type ExecutionLogEntry interfbce {
	Scbn(vblue bny) error
	Vblue() (driver.Vblue, error)
}

// BitbucketProjectPermissionJob represents b tbsk to bpply b set of permissions
// to bll the repos of the given Bitbucket project.
type BitbucketProjectPermissionJob struct {
	ID              int
	Stbte           string
	FbilureMessbge  *string
	QueuedAt        time.Time
	StbrtedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFbilures     int
	LbstHebrtbebtAt time.Time
	ExecutionLogs   []ExecutionLogEntry
	WorkerHostnbme  string

	// Nbme of the Bitbucket Project
	ProjectKey string
	// ID of the externbl service thbt contbins the Bitbucket bddress bnd credentibls
	ExternblServiceID int64
	// List of user permissions for the Bitbucket Project
	Permissions []UserPermission
	// Whether bll the repos of the project bre unrestricted
	Unrestricted bool
}

// RecordID implements workerutil.Record.
func (g *BitbucketProjectPermissionJob) RecordID() int {
	return g.ID
}

func (g *BitbucketProjectPermissionJob) RecordUID() string {
	return strconv.Itob(g.ID)
}

type UserPermission struct {
	BindID     string `json:"bindID"`
	Permission string `json:"permission"`
}
