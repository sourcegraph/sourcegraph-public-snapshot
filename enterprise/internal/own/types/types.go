package types

import (
	"time"

	codeownerspb "github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type CodeownersFile struct {
	CreatedAt time.Time
	UpdatedAt time.Time

	RepoID   api.RepoID
	Contents string
	Proto    *codeownerspb.File
}
