package notif

import (
	"github.com/AaronO/go-git-http"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/events"
)

const GitPushEvent events.EventID = "git.push"
const GitCreateEvent events.EventID = "git.create"
const GitDeleteEvent events.EventID = "git.delete"

type GitPayload struct {
	Actor           sourcegraph.UserSpec
	Repo            sourcegraph.RepoSpec
	ContentEncoding string
	Event           githttp.Event
}

const DiscussionCreateEvent events.EventID = "discussion.create"
const DiscussionCommentEvent events.EventID = "discussion.comment"

type DiscussionPayload struct {
	Actor      sourcegraph.UserSpec
	ID         int64
	Repo       string
	Title      string
	URL        string
	Discussion *sourcegraph.Discussion
	Comment    *sourcegraph.DiscussionComment
}
