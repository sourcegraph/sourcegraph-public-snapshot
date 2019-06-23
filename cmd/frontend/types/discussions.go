package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// DiscussionThread mirrors the underlying discussion_threads field types exactly.
// It intentionally does not try to e.g. alleviate null fields.
type DiscussionThread struct {
	ID           int64
	ProjectID    int64
	AuthorUserID int32
	Title        string
	Settings     *string
	Type         ThreadType
	Status       ThreadStatus
	CreatedAt    time.Time
	ArchivedAt   *time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// ThreadType enumerates the possible thread types.
type ThreadType string

const (
	ThreadTypeThread    ThreadType = "THREAD"
	ThreadTypeCheck                = "CHECK"
	ThreadTypeChangeset            = "CHANGESET"
)

// IsValidThreadType reports whether t is a valid thread type.
func IsValidThreadType(t string) bool {
	return ThreadType(t) == ThreadTypeThread || ThreadType(t) == ThreadTypeCheck || ThreadType(t) == ThreadTypeChangeset
}

// ThreadStatus enumerates the possible thread statuses.
type ThreadStatus string

const (
	ThreadStatusPreview    ThreadStatus = "PREVIEW"
	ThreadStatusOpenActive              = "OPEN_ACTIVE"
	ThreadStatusInactive                = "INACTIVE"
	ThreadStatusClosed                  = "CLOSED"
)

// IsValidThreadStatus reports whether t is a valid thread status.
func IsValidThreadStatus(t string) bool {
	return ThreadStatus(t) == ThreadStatusPreview || ThreadStatus(t) == ThreadStatusOpenActive || ThreadStatus(t) == ThreadStatusInactive || ThreadStatus(t) == ThreadStatusClosed
}

// DiscussionThreadTargetRepo mirrors the underlying discussion_threads_target_repo field types exactly.
// It intentionally does not try to e.g. alleviate null fields.
type DiscussionThreadTargetRepo struct {
	ID       int64
	ThreadID int64
	RepoID   api.RepoID
	Path     *string
	Branch   *string
	Revision *string

	StartLine      *int32
	EndLine        *int32
	StartCharacter *int32
	EndCharacter   *int32
	LinesBefore    *[]string
	Lines          *[]string
	LinesAfter     *[]string

	IsIgnored bool
}

// HasSelection tells if the selection fields are present or not. If one field
// is present, all must be present.
func (d *DiscussionThreadTargetRepo) HasSelection() bool {
	return d.StartLine != nil || d.EndLine != nil || d.StartCharacter != nil || d.EndCharacter != nil || d.LinesBefore != nil || d.Lines != nil || d.LinesAfter != nil
}

// DiscussionComment mirrors the underlying discussion_comments field types exactly.
// It intentionally does not try to e.g. alleviate null fields.
type DiscussionComment struct {
	ID           int64
	ThreadID     int64
	AuthorUserID int32
	Contents     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
	Reports      []string
}
