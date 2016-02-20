// Package issues provides an issue tracker service definition.
package issues

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"golang.org/x/net/context"
)

// RepoSpec is a specification for a repository.
type RepoSpec struct {
	URI string // URI is clean '/'-separated URI. E.g, "user/repo".
}

// String implements fmt.Stringer.
func (rs RepoSpec) String() string {
	return rs.URI
}

// Service defines methods of an issue tracking service.
type Service interface {
	// List issues.
	List(ctx context.Context, repo RepoSpec, opt IssueListOptions) ([]Issue, error)
	// Count issues.
	Count(ctx context.Context, repo RepoSpec, opt IssueListOptions) (uint64, error)

	// Get an issue.
	Get(ctx context.Context, repo RepoSpec, id uint64) (Issue, error)

	// ListComments lists comments for specified issue id.
	ListComments(ctx context.Context, repo RepoSpec, id uint64, opt interface{}) ([]Comment, error)
	// ListEvents lists events for specified issue id.
	ListEvents(ctx context.Context, repo RepoSpec, id uint64, opt interface{}) ([]Event, error)

	// Create a new issue.
	Create(ctx context.Context, repo RepoSpec, issue Issue) (Issue, error)
	// CreateComment creates a new comment for specified issue id.
	CreateComment(ctx context.Context, repo RepoSpec, id uint64, comment Comment) (Comment, error)

	// Edit the specified issue id.
	Edit(ctx context.Context, repo RepoSpec, id uint64, ir IssueRequest) (Issue, []Event, error)
	// EditComment edits comment of specified issue id.
	EditComment(ctx context.Context, repo RepoSpec, id uint64, cr CommentRequest) (Comment, error)

	// Search searches issues.
	Search(ctx context.Context, opt SearchOptions) (SearchResponse, error)

	// TODO: This doesn't belong here; it should be factored out into a platform Users service that is provided to this service.
	CurrentUser(ctx context.Context) (*User, error)
}

// CopierFrom is an optional interface that allows copying issues between services.
type CopierFrom interface {
	// CopyFrom copies all issues from src for specified repo.
	CopyFrom(src Service, repo RepoSpec) error
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
	Reactions []Reaction
	Editable  bool // Editable represents whether the current user (if any) can perform edit operations on this comment (or the encompassing issue).
}

// Reaction represents a single reaction to a comment, backed by 1 or more users.
type Reaction struct {
	Reaction EmojiID
	Users    []User // Length is 1 or more.
}

// UserSpec is a specification for a user.
type UserSpec struct {
	ID     uint64
	Domain string
}

// User represents a user, including their details.
type User struct {
	UserSpec
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

// EmojiID is the id of a reaction. For example, "+1".
type EmojiID string

// CommentRequest is a request to edit a comment.
type CommentRequest struct {
	ID       uint64
	Body     *string  // If not nil, set the body.
	Reaction *EmojiID // If not nil, toggle this reaction.
}

// State represents the issue state.
type State string

const (
	// OpenState is when an issue is open.
	OpenState State = "open"
	// ClosedState is when an issue is closed.
	ClosedState State = "closed"
)

// SearchOptions are options for search.
type SearchOptions struct {
	Query   string
	Repo    RepoSpec
	Page    int
	PerPage int
}

// SearchResult represents a single search result.
type SearchResult struct {
	ID        string
	Title     template.HTML
	Comment   template.HTML
	User      User
	CreatedAt time.Time
	State     State
}

// SearchResponse is the return value of doing a search.
type SearchResponse struct {
	Results []SearchResult
	Total   uint64
}

// Validate returns non-nil error if the issue is invalid.
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

// Validate returns non-nil error if the issue request is invalid.
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

// Validate returns non-nil error if the comment is invalid.
func (c Comment) Validate() error {
	// TODO: Issue descriptions can have blank bodies, support that (primarily for editing comments).
	if strings.TrimSpace(c.Body) == "" {
		return fmt.Errorf("comment body can't be blank or all whitespace")
	}
	return nil
}

// Validate validates the comment edit request, returning an non-nil error if it's invalid.
// requiresEdit reports if the edit request needs edit rights or if it can be done by anyone that can react.
func (cr CommentRequest) Validate() (requiresEdit bool, err error) {
	if cr.Body != nil {
		requiresEdit = true

		// TODO: Issue descriptions can have blank bodies, support that (primarily for editing comments).
		if strings.TrimSpace(*cr.Body) == "" {
			return requiresEdit, fmt.Errorf("comment body can't be blank or all whitespace")
		}
	}
	if cr.Reaction != nil {
		// TODO: Maybe validate that the emojiID is one of supported ones.
		//       Or maybe not (unsupported ones can be handled by frontend component).
		//       That way custom emoji can be added/removed, etc. Figure out what the best thing to do is and do it.
	}
	return requiresEdit, nil
}
