package webhooks

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

// MergeRequestEvent is the common type that underpins the types defined for
// specific merge request actions.
type MergeRequestEvent struct {
	EventCommon

	MergeRequest *gitlab.MergeRequest `json:"merge_request"`
	User         *gitlab.User         `json:"user"`
	Labels       *[]gitlab.Label      `json:"labels"`
}

type MergeRequestEventContainer interface {
	ToEvent() *MergeRequestEvent
}

type MergeRequestApprovedEvent struct{ MergeRequestEvent }
type MergeRequestCloseEvent struct{ MergeRequestEvent }
type MergeRequestMergeEvent struct{ MergeRequestEvent }
type MergeRequestReopenEvent struct{ MergeRequestEvent }
type MergeRequestUnapprovedEvent struct{ MergeRequestEvent }

func (e *MergeRequestApprovedEvent) ToEvent() *MergeRequestEvent   { return &e.MergeRequestEvent }
func (e *MergeRequestCloseEvent) ToEvent() *MergeRequestEvent      { return &e.MergeRequestEvent }
func (e *MergeRequestMergeEvent) ToEvent() *MergeRequestEvent      { return &e.MergeRequestEvent }
func (e *MergeRequestReopenEvent) ToEvent() *MergeRequestEvent     { return &e.MergeRequestEvent }
func (e *MergeRequestUnapprovedEvent) ToEvent() *MergeRequestEvent { return &e.MergeRequestEvent }

func (e *MergeRequestCloseEvent) Key() string  { return e.key("Close") }
func (e *MergeRequestMergeEvent) Key() string  { return e.key("Merge") }
func (e *MergeRequestReopenEvent) Key() string { return e.key("Reopen") }

func (e *MergeRequestEvent) key(prefix string) string {
	// We can't key solely off the merge request ID because it may be reopened
	// and closed multiple times. Instead, we'll use the CreatedAt field.
	return fmt.Sprintf("MergeRequest:%s:%d:%s", prefix, e.MergeRequest.IID, e.MergeRequest.CreatedAt.Format(time.RFC3339))
}

type mergeRequestEvent struct {
	EventCommon

	User   *gitlab.User    `json:"user"`
	Labels *[]gitlab.Label `json:"labels"`

	ObjectAttributes mergeRequestEventObjectAttributes `json:"object_attributes"`
}

type mergeRequestEventObjectAttributes struct {
	*gitlab.MergeRequest
	Action string `json:"action"`
}

func (mre *mergeRequestEvent) downcast() (interface{}, error) {
	e := MergeRequestEvent{
		EventCommon:  mre.EventCommon,
		MergeRequest: mre.ObjectAttributes.MergeRequest,
		User:         mre.User,
		Labels:       mre.Labels,
	}

	// These action values are completely undocumented in GitLab's webhook
	// documentation: indeed, the _existence_ of the action field is only
	// implied by the examples. Nevertheless, we don't really have any other
	// option but to rely on it, since there's no other way to access the
	// information on what has changed when we get a webhook, since the payload
	// is otherwise untyped and the webhook types are far too coarsely grained
	// to be able to infer anything.
	switch mre.ObjectAttributes.Action {
	case "approved":
		return &MergeRequestApprovedEvent{e}, nil

	case "close":
		return &MergeRequestCloseEvent{e}, nil

	case "merge":
		return &MergeRequestMergeEvent{e}, nil

	case "reopen":
		return &MergeRequestReopenEvent{e}, nil

	case "unapproved":
		return &MergeRequestUnapprovedEvent{e}, nil
	}

	return nil, errors.Wrapf(ErrObjectKindUnknown, "unknown merge request event action: %s", mre.ObjectAttributes.Action)
}
