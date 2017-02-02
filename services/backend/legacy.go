package backend

import (
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

func init() {
	accesscontrol.Repos = localstore.Repos
}
