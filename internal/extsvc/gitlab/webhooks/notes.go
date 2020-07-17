package webhooks

import (
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

// noteEvent objects represents note (or comment) events. These can apply to
// commits, merge requests, issues, or snippets. Only merge requests are fully
// unmarshalled right now.
type noteEvent struct {
	EventCommon

	Note *gitlab.Note `json:"object_attributes"`

	Commit       *struct{ ID string } `json:"commit"`
	MergeRequest *gitlab.MergeRequest `json:"merge_request"`
	Issue        *struct{ ID string } `json:"issue"`
	Snippet      *struct{ ID string } `json:"snippet"`
}

type NoteMergeRequestEvent struct {
	EventCommon

	Note         *gitlab.Note
	MergeRequest *gitlab.MergeRequest
}

var ErrNoteEventCannotDowncast = errors.New("cannot downcast note event of unknown type")

func (ne *noteEvent) downcast() (interface{}, error) {
	if ne.MergeRequest != nil {
		return &NoteMergeRequestEvent{
			EventCommon:  ne.EventCommon,
			Note:         ne.Note,
			MergeRequest: ne.MergeRequest,
		}, nil
	}

	return nil, ErrNoteEventCannotDowncast
}
