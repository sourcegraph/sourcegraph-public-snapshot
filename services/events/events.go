package events

import (
	"github.com/AaronO/go-git-http"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

type EventID string

const GitPushEvent EventID = "git.push"
const GitCreateBranchEvent EventID = "git.create"
const GitDeleteBranchEvent EventID = "git.delete"

type GitPayload struct {
	Actor       sourcegraph.UserSpec
	Repo        int32
	IgnoreBuild bool
	Event       githttp.Event
}
