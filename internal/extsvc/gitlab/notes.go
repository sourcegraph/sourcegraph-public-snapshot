package gitlab

import (
	"context"
	"fmt"
	"time"
)

// GetMergeRequestNotes retrieves the notes for the given merge request. As the
// notes are paginated, a function is returned that may be invoked to return the
// next page of results. An empty slice and a nil error indicates that all pages
// have been returned.
func (c *Client) GetMergeRequestNotes(ctx context.Context, project *Project, iid ID) func() ([]*Note, error) {
	if MockGetMergeRequestNotes != nil {
		return MockGetMergeRequestNotes(c, ctx, project, iid)
	}

	pr := c.newPaginatedResult("GET", fmt.Sprintf("projects/%d/merge_requests/%d/notes", project.ID, iid), func() interface{} { return []*Note{} })
	return func() ([]*Note, error) {
		page, err := pr.next(ctx)
		return page.([]*Note), err
	}
}

type Note struct {
	// As with the other types here, we're only decoding enough fields here for
	// the required campaign functionality.
	ID        ID        `json:"id"`
	Body      string    `json:"body"`
	Author    User      `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	System    bool      `json:"system"`
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
