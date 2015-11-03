package notif

import (
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/events"
)

const ChangesetCreateEvent events.EventID = "changeset.create"
const ChangesetUpdateEvent events.EventID = "changeset.update"
const ChangesetReviewEvent events.EventID = "changeset.review"
const ChangesetCloseEvent events.EventID = "changeset.close"

type ChangesetPayload struct {
	UserSpec  sourcegraph.UserSpec
	ID        int64
	Repo      string
	Title     string
	URL       string
	Changeset *sourcegraph.Changeset
	Review    *sourcegraph.ChangesetReview
	Update    *sourcegraph.ChangesetUpdateOp
}

const DiscussionCreateEvent events.EventID = "discussion.create"
const DiscussionCommentEvent events.EventID = "discussion.comment"

type DiscussionPayload struct {
	UserSpec   sourcegraph.UserSpec
	ID         int64
	Repo       string
	Title      string
	URL        string
	Discussion *sourcegraph.Discussion
	Comment    *sourcegraph.DiscussionComment
}
