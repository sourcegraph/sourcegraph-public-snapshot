package gitlab

import (
	"fmt"
	"time"
)

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
