package events

import (
	"github.com/AaronO/go-git-http"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

type EventID string

const GitPushEvent EventID = "git.push"
const GitCreateBranchEvent EventID = "git.create"
const GitDeleteBranchEvent EventID = "git.delete"

type GitPayload struct {
	Actor           sourcegraph.UserSpec
	Repo            sourcegraph.RepoSpec
	ContentEncoding string
	Event           githttp.Event
}

const DiscussionCreateEvent EventID = "discussion.create"
const DiscussionCommentEvent EventID = "discussion.comment"

type DiscussionPayload struct {
	Actor      sourcegraph.UserSpec
	ID         int64
	Repo       string
	Title      string
	URL        string
	Discussion *sourcegraph.Discussion
	Comment    *sourcegraph.DiscussionComment
}

const ChangesetCreateEvent EventID = "changeset.create"
const ChangesetUpdateEvent EventID = "changeset.update"
const ChangesetReviewEvent EventID = "changeset.review"
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
}
