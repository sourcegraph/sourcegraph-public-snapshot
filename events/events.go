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
	Actor           sourcegraph.UserSpec
	Repo            sourcegraph.RepoSpec
	ContentEncoding string
	IgnoreBuild     bool
	Event           githttp.Event
}

const ChangesetCreateEvent EventID = "changeset.create"
const ChangesetUpdateEvent EventID = "changeset.update"
const ChangesetReviewEvent EventID = "changeset.review"
const ChangesetMergeEvent EventID = "changeset.merge"
const ChangesetCloseEvent EventID = "changeset.close"

type ChangesetPayload struct {
	Actor     sourcegraph.UserSpec
	ID        int64
	Repo      string
	Title     string
	URL       string
	Changeset *sourcegraph.Changeset
	Review    *sourcegraph.ChangesetReview
	Update    *sourcegraph.ChangesetUpdateOp
	Merge     *sourcegraph.ChangesetMergeOp
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
