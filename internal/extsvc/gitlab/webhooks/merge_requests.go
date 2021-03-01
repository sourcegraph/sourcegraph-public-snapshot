package webhooks

import (
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

// There's a bit going on in this file in terms of types, so here's a high
// level overview of what happens.
//
// When we get a webhook event of kind "merge_request" from GitLab, we want to
// eventually unmarshal it into one of the specific, exported types below, such
// as MergeRequestApprovedEvent or MergeRequestCloseEvent. To do so, we need to
// look at the "action" field embedded within the merge request in the event.
//
// We don't really want to have to unmarshal the JSON an extra time or copy the
// fairly sizable MergeRequest and User structs again, so what we do instead is
// unmarshal it initially into mergeRequestEvent. This unmarshals all of the
// fields that we need to construct the eventual typed event, but only exists
// for as long as it takes to go from the initial unmarshal into
// mergeRequestEvent until its downcast() method is called. That method looks
// at the "action" and then constructs the final struct, moving the pointer
// fields across from mergeRequestEvent into the MergeRequestEventCommon struct
// they all embed.

// MergeRequestEventCommon is the common type that underpins the types defined
// for specific merge request actions.
type MergeRequestEventCommon struct {
	EventCommon

	MergeRequest *gitlab.MergeRequest     `json:"merge_request"`
	User         *gitlab.User             `json:"user"`
	Labels       *[]gitlab.Label          `json:"labels"`
	Changes      mergeRequestEventChanges `json:"changes"`
}

type mergeRequestEventChanges struct {
	Title struct {
		Previous string `json:"previous"`
		Current  string `json:"current"`
	} `json:"title"`
	UpdatedAt struct {
		Current gitlab.Time `json:"current"`
	} `json:"updated_at"`
}

// MergeRequestEventCommonContainer is a common interface for types that embed
// MergeRequestEvent to provide a method that can return the embedded
// MergeRequestEvent.
type MergeRequestEventCommonContainer interface {
	ToEventCommon() *MergeRequestEventCommon
}

type keyer interface {
	Key() string
}

// UpsertableWebhookEvent is a common interface for types that embed
// ToEvent to provide a method that can return a changeset event
// derived from the webhook payload.
type UpsertableWebhookEvent interface {
	MergeRequestEventCommonContainer
	ToEvent() keyer
}

// Type guards:
var _ UpsertableWebhookEvent = (*MergeRequestCloseEvent)(nil)
var _ UpsertableWebhookEvent = (*MergeRequestMergeEvent)(nil)
var _ UpsertableWebhookEvent = (*MergeRequestReopenEvent)(nil)
var _ UpsertableWebhookEvent = (*MergeRequestDraftEvent)(nil)
var _ UpsertableWebhookEvent = (*MergeRequestUndraftEvent)(nil)

type MergeRequestApprovedEvent struct{ MergeRequestEventCommon }
type MergeRequestUnapprovedEvent struct{ MergeRequestEventCommon }
type MergeRequestUpdateEvent struct{ MergeRequestEventCommon }

type MergeRequestCloseEvent struct{ MergeRequestEventCommon }
type MergeRequestMergeEvent struct{ MergeRequestEventCommon }
type MergeRequestReopenEvent struct{ MergeRequestEventCommon }
type MergeRequestUndraftEvent struct{ MergeRequestEventCommon }
type MergeRequestDraftEvent struct{ MergeRequestEventCommon }

func (e *MergeRequestApprovedEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}
func (e *MergeRequestUnapprovedEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}
func (e *MergeRequestUpdateEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}

func (e *MergeRequestUndraftEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}

func (e *MergeRequestUndraftEvent) ToEvent() keyer {
	user := gitlab.User{}
	if e.User != nil {
		user = *e.User
	}
	return &gitlab.UnmarkWorkInProgressEvent{
		Note: &gitlab.Note{
			Body:      gitlab.SystemNoteBodyUnmarkedWorkInProgress,
			System:    true,
			CreatedAt: e.Changes.UpdatedAt.Current,
			Author:    user,
		},
	}
}

func (e *MergeRequestDraftEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}

func (e *MergeRequestDraftEvent) ToEvent() keyer {
	user := gitlab.User{}
	if e.User != nil {
		user = *e.User
	}
	return &gitlab.MarkWorkInProgressEvent{
		Note: &gitlab.Note{
			Body:      gitlab.SystemNoteBodyMarkedWorkInProgress,
			System:    true,
			CreatedAt: e.Changes.UpdatedAt.Current,
			Author:    user,
		},
	}
}

func (e *MergeRequestCloseEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}

func (e *MergeRequestCloseEvent) ToEvent() keyer {
	user := gitlab.User{}
	if e.User != nil {
		user = *e.User
	}
	return &gitlab.MergeRequestClosedEvent{
		ResourceStateEvent: &gitlab.ResourceStateEvent{
			User:         user,
			CreatedAt:    e.Changes.UpdatedAt.Current,
			ResourceType: "merge_request",
			ResourceID:   e.MergeRequest.ID,
			State:        gitlab.ResourceStateEventStateClosed,
		},
	}
}

func (e *MergeRequestMergeEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}

func (e *MergeRequestMergeEvent) ToEvent() keyer {
	user := gitlab.User{}
	if e.User != nil {
		user = *e.User
	}
	return &gitlab.MergeRequestMergedEvent{
		ResourceStateEvent: &gitlab.ResourceStateEvent{
			User:         user,
			CreatedAt:    e.Changes.UpdatedAt.Current,
			ResourceType: "merge_request",
			ResourceID:   e.MergeRequest.ID,
			State:        gitlab.ResourceStateEventStateMerged,
		},
	}
}

func (e *MergeRequestReopenEvent) ToEventCommon() *MergeRequestEventCommon {
	return &e.MergeRequestEventCommon
}

func (e *MergeRequestReopenEvent) ToEvent() keyer {
	user := gitlab.User{}
	if e.User != nil {
		user = *e.User
	}
	return &gitlab.MergeRequestReopenedEvent{
		ResourceStateEvent: &gitlab.ResourceStateEvent{
			User:         user,
			CreatedAt:    e.Changes.UpdatedAt.Current,
			ResourceType: "merge_request",
			ResourceID:   e.MergeRequest.ID,
			State:        gitlab.ResourceStateEventStateReopened,
		},
	}
}

// mergeRequestEvent is an internal type used for initially unmarshalling the
// typed event before it is downcast into a more specific type later based on
// the "action" field in the JSON.
type mergeRequestEvent struct {
	EventCommon

	User    *gitlab.User             `json:"user"`
	Labels  *[]gitlab.Label          `json:"labels"`
	Changes mergeRequestEventChanges `json:"changes"`

	ObjectAttributes mergeRequestEventObjectAttributes `json:"object_attributes"`
}

type mergeRequestEventObjectAttributes struct {
	*gitlab.MergeRequest
	Action string `json:"action"`
}

func (mre *mergeRequestEvent) downcast() (interface{}, error) {
	e := MergeRequestEventCommon{
		EventCommon:  mre.EventCommon,
		MergeRequest: mre.ObjectAttributes.MergeRequest,
		User:         mre.User,
		Labels:       mre.Labels,
		Changes:      mre.Changes,
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

	case "update":
		if prev, curr := e.Changes.Title.Previous, e.Changes.Title.Current; prev != "" && curr != "" {
			if gitlab.IsWIP(prev) && !gitlab.IsWIP(curr) {
				return &MergeRequestUndraftEvent{e}, nil
			} else if !gitlab.IsWIP(prev) && gitlab.IsWIP(curr) {
				return &MergeRequestDraftEvent{e}, nil
			}
		}
		return &MergeRequestUpdateEvent{e}, nil
	}

	return nil, errors.Wrapf(ErrObjectKindUnknown, "unknown merge request event action: %s", mre.ObjectAttributes.Action)
}
