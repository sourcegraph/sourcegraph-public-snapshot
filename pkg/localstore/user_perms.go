package localstore

import (
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
)

// ErrRepoNotFound indicates that the repo does not exist or that the user has no access to that
// repo. Those two cases are not differentiated to avoid leaking repo existence information.
var ErrRepoNotFound = legacyerr.Errorf(legacyerr.NotFound, "repo not found")
