package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetMergeRequestNotes retrieves the notes for the given merge request. As the
// notes are paginated, a function is returned that may be invoked to return the
// next page of results. An empty slice and a nil error indicates that all pages
// have been returned.
func (c *Client) GetMergeRequestNotes(ctx context.Context, project *Project, iid ID) func() ([]*Note, error) {
	if MockGetMergeRequestNotes != nil {
		return MockGetMergeRequestNotes(c, ctx, project, iid)
	}

	baseURL := fmt.Sprintf("projects/%d/merge_requests/%d/notes", project.ID, iid)
	currentPage := "1"
	return func() ([]*Note, error) {
		page := []*Note{}

		// If there aren't any further pages, we'll return the empty slice we
		// just created.
		if currentPage == "" {
			return page, nil
		}

		parsedUrl, err := url.Parse(baseURL)
		if err != nil {
			return nil, err
		}
		q := parsedUrl.Query()
		q.Add("page", currentPage)
		parsedUrl.RawQuery = q.Encode()

		req, err := http.NewRequest("GET", parsedUrl.String(), nil)
		if err != nil {
			return nil, errors.Wrap(err, "creating notes request")
		}

		header, _, err := c.do(ctx, req, &page)
		if err != nil {
			return nil, errors.Wrap(err, "requesting notes page")
		}

		// If there's another page, this will be a page number. If there's not, then
		// this will be an empty string, and we can detect that next iteration
		// to short circuit.
		currentPage = header.Get("X-Next-Page")

		return page, nil
	}
}

// SystemNoteBody is a type of all known system message bodies.
type SystemNoteBody string

const (
	SystemNoteBodyReviewApproved         SystemNoteBody = "approved this merge request"
	SystemNoteBodyReviewUnapproved       SystemNoteBody = "unapproved this merge request"
	SystemNoteBodyUnmarkedWorkInProgress SystemNoteBody = "unmarked as a **Work In Progress**"
	SystemNoteBodyMarkedWorkInProgress   SystemNoteBody = "marked as a **Work In Progress**"
	SystemNoteBodyMarkedDraft            SystemNoteBody = "marked this merge request as **draft**"
	SystemNoteBodyMarkedReady            SystemNoteBody = "marked this merge request as **ready**"
)

type Note struct {
	ID        ID             `json:"id"`
	Body      SystemNoteBody `json:"body"`
	Author    User           `json:"author"`
	CreatedAt Time           `json:"created_at"`
	System    bool           `json:"system"`
}

// Notes are not strongly typed, but also provide the only real method we have
// of getting historical approval events. We'll define a couple of fake types to
// better match what other external services provide, and a function to convert
// a Note into one of those types if the Note is a system approval comment.

type ReviewApprovedEvent struct{ *Note }

func (e *ReviewApprovedEvent) Key() string {
	return fmt.Sprintf("approved:%s:%s", e.Author.Username, e.CreatedAt.Time.Truncate(time.Second))
}

type ReviewUnapprovedEvent struct{ *Note }

func (e *ReviewUnapprovedEvent) Key() string {
	return fmt.Sprintf("unapproved:%s:%s", e.Author.Username, e.CreatedAt.Time.Truncate(time.Second))
}

type MarkWorkInProgressEvent struct{ *Note }

func (e *MarkWorkInProgressEvent) Key() string {
	return fmt.Sprintf("wip:%s", e.CreatedAt.Time.Truncate(time.Second))
}

type UnmarkWorkInProgressEvent struct{ *Note }

func (e *UnmarkWorkInProgressEvent) Key() string {
	return fmt.Sprintf("unwip:%s", e.CreatedAt.Time.Truncate(time.Second))
}

type keyer interface {
	Key() string
}

// ToEvent returns a pointer to a more specific struct, or
// nil if the Note is not of a known kind.
func (n *Note) ToEvent() keyer {
	if n.System {
		switch n.Body {
		case SystemNoteBodyReviewApproved:
			return &ReviewApprovedEvent{n}
		case SystemNoteBodyReviewUnapproved:
			return &ReviewUnapprovedEvent{n}
		case SystemNoteBodyMarkedReady,
			SystemNoteBodyUnmarkedWorkInProgress:
			return &UnmarkWorkInProgressEvent{n}
		case SystemNoteBodyMarkedDraft,
			SystemNoteBodyMarkedWorkInProgress:
			return &MarkWorkInProgressEvent{n}
		}
	}

	return nil
}
