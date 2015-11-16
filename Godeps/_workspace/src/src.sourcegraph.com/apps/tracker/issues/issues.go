package issues

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"golang.org/x/net/context"
)

type RepoSpec struct {
	URI string // URI is clean '/'-separated URI. E.g, "user/repo".
}

func (rs RepoSpec) String() string {
	return rs.URI
}

type Service interface {
	List(ctx context.Context, repo RepoSpec, opt IssueListOptions) ([]Issue, error)
	Count(ctx context.Context, repo RepoSpec, opt IssueListOptions) (uint64, error)

	Get(ctx context.Context, repo RepoSpec, id uint64) (Issue, error)

	ListComments(ctx context.Context, repo RepoSpec, id uint64, opt interface{}) ([]Comment, error)
	ListEvents(ctx context.Context, repo RepoSpec, id uint64, opt interface{}) ([]Event, error)

	Create(ctx context.Context, repo RepoSpec, issue Issue) (Issue, error)
	CreateComment(ctx context.Context, repo RepoSpec, id uint64, comment Comment) (Comment, error)

	Edit(ctx context.Context, repo RepoSpec, id uint64, ir IssueRequest) (Issue, error)
	EditComment(ctx context.Context, repo RepoSpec, id uint64, comment Comment) (Comment, error)

	// TODO: This doesn't belong here, does it?
	CurrentUser(ctx context.Context) (*User, error)
}

// Issue represents an issue on a repository.
type Issue struct {
	ID    uint64
	State State
	Title string
	Comment
	Reference *Reference
	Replies   int // Number of replies to this issue (not counting the mandatory issue description comment).
}

// Comment represents a comment left on an issue.
type Comment struct {
	ID        uint64
	User      User
	CreatedAt time.Time
	Body      string
	Editable  bool // Editable represents whether the current user (if any) can perform edit operations on this comment (or the encompassing issue).
}

// User represents a user.
type User struct {
	Login     string
	AvatarURL template.URL
	HTMLURL   template.URL
}

// IssueRequest is a request to edit an issue.
// To edit the body, use EditComment with comment ID 0.
type IssueRequest struct {
	State *State
	Title *string
}

// State represents the issue state.
type State string

const (
	OpenState   State = "open"
	ClosedState State = "closed"
)

func (i Issue) Validate() error {
	if strings.TrimSpace(i.Title) == "" {
		return fmt.Errorf("title can't be blank or all whitespace")
	}
	if ref := i.Reference; ref != nil {
		if ref.CommitID == "" {
			return fmt.Errorf("commit ID is required for reference")
		}
	}
	return nil
}

func (ir IssueRequest) Validate() error {
	if ir.State != nil {
		switch *ir.State {
		case OpenState, ClosedState:
		default:
			return fmt.Errorf("bad state")
		}
	}
	if ir.Title != nil {
		if strings.TrimSpace(*ir.Title) == "" {
			return fmt.Errorf("title can't be blank or all whitespace")
		}
	}
	return nil
}

func (c Comment) Validate() error {
	// TODO: Issue descriptions can have blank bodies, support that (primarily for editing comments).
	if strings.TrimSpace(c.Body) == "" {
		return fmt.Errorf("comment body can't be blank or all whitespace")
	}
	return nil
}
