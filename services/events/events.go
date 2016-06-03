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
	Repo        string
	IgnoreBuild bool
	Event       githttp.Event
}
