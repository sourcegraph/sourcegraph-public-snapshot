package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// DiscussionThread mirrors the underlying discussion_threads field types exactly.
// It intentionally does not try to e.g. alleviate null fields.
type DiscussionThread struct {
	ID           int64
	AuthorUserID int32
	Title        string
	CreatedAt    time.Time
	ArchivedAt   *time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
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
