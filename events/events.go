package events

import (
	"github.com/AaronO/go-git-http"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

type EventID string

const GitPushEvent EventID = "git.push"
const GitCreateBranchEvent EventID = "git.create"
const GitDeleteBranchEvent EventID = "git.delete"

type GitPayload struct {
	Actor       sourcegraph.UserSpec
	Repo        sourcegraph.RepoSpec
	IgnoreBuild bool
	Event       githttp.Event
}

const ClientRegisterEvent EventID = "client.register"
const ClientUpdateEvent EventID = "client.update"
const ClientGrantAccessEvent EventID = "client.grant.access"

type ClientPayload struct {
	// The user that performed the register client or grant access operation.
	Actor    sourcegraph.UserSpec
	ClientID string

	// The user that was granted permissions on the client.
	Grantee sourcegraph.UserSpec
	Perms   *sourcegraph.UserPermissions
}
