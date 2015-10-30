package notif

import (
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/events"
)

const ChangesetCreateEvent events.EventID = "changeset.create"
const ChangesetUpdateEvent events.EventID = "changeset.update"
const ChangesetReviewEvent events.EventID = "changeset.review"
const ChangesetCloseEvent events.EventID = "changeset.close"

const DiscussionCreateEvent events.EventID = "discussion.create"
const DiscussionCommentEvent events.EventID = "discussion.comment"

type Payload struct {
	Type          events.EventID
	UserSpec      sourcegraph.UserSpec
	ActionContent string
	ActionType    string
	ObjectID      int64
	ObjectRepo    string
	ObjectTitle   string
	ObjectType    string
	ObjectURL     string
	Object        interface{}
}
