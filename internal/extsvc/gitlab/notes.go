package gitlab

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// GetMergeRequestNotes retrieves the notes for the given merge request. As the
// notes are paginated, a function is returned that may be invoked to return the
// next page of results. An empty slice and a nil error indicates that all pages
// have been returned.
func (c *Client) GetMergeRequestNotes(ctx context.Context, project *Project, iid ID) func() ([]*Note, error) {
	if MockGetMergeRequestNotes != nil {
		return MockGetMergeRequestNotes(c, ctx, project, iid)
	}

	url := fmt.Sprintf("projects/%d/merge_requests/%d/notes", project.ID, iid)
	return func() ([]*Note, error) {
		page := []*Note{}

		// If there aren't any further pages, we'll return the empty slice we
		// just created.
		if url == "" {
			return page, nil
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, errors.Wrap(err, "creating notes request")
		}

		header, _, err := c.do(ctx, req, &page)
		if err != nil {
			return nil, errors.Wrap(err, "requesting notes page")
		}

		// If there's another page, this will be a URL. If there's not, then
		// this will be an empty string, and we can detect that next iteration
		// to short circuit.
		url = header.Get("X-Next-Page")

		return page, nil
	}
}

type Note struct {
	ID        ID     `json:"id"`
	Body      string `json:"body"`
	Author    User   `json:"author"`
	CreatedAt Time   `json:"created_at"`
	System    bool   `json:"system"`
}

func (n *Note) Key() string {
	return fmt.Sprintf("Note:%d", n.ID)
}

// Notes are not strongly typed, but also provide the only real method we have
// of getting historical approval events. We'll define a couple of fake types to
// better match what other external services provide, and a function to convert
// a Note into one of those types if the Note is a system approval comment.

type ReviewApproved struct{ *Note }
type ReviewUnapproved struct{ *Note }

// ToReview returns a pointer to a ReviewApproved or ReviewUnapproved struct, or
// nil if the Note is not a review note.
func (n *Note) ToReview() interface{} {
	if n.System {
		switch n.Body {
		case "approved this merge request":
			return &ReviewApproved{n}
		case "unapproved this merge request":
			return &ReviewUnapproved{n}
		}
	}

	return nil
}
