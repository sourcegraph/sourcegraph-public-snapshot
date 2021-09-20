package batches

import (
	"github.com/cockroachdb/errors"

	jsonutil "github.com/sourcegraph/sourcegraph/lib/batches/json"
	"github.com/sourcegraph/sourcegraph/lib/batches/schema"
)

// ErrHeadBaseMismatch is returned by (*ChangesetSpec).UnmarshalValidate() if
// the head and base repositories do not match (a case which we do not support
// yet).
var ErrHeadBaseMismatch = errors.New("headRepository does not match baseRepository")

// ParseChangesetSpec unmarshals the RawSpec into Spec and validates it against
// the ChangesetSpec schema and does additional semantic validation.
func ParseChangesetSpec(rawSpec []byte) (*ChangesetSpec, error) {
	spec := &ChangesetSpec{}
	err := jsonutil.UnmarshalValidate(schema.ChangesetSpecJSON, rawSpec, &spec)
	if err != nil {
		return nil, err
	}

	headRepo := spec.HeadRepository
	baseRepo := spec.BaseRepository
	if headRepo != "" && baseRepo != "" && headRepo != baseRepo {
		return nil, ErrHeadBaseMismatch
	}

	return spec, nil
}

type ChangesetSpec struct {
	// BaseRepository is the GraphQL ID of the base repository.
	BaseRepository string `json:"baseRepository,omitempty"`

	// If this is not empty, the description is a reference to an existing
	// changeset and the rest of these fields are empty.
	ExternalID string `json:"externalID,omitempty"`

	BaseRev string `json:"baseRev,omitempty"`
	BaseRef string `json:"baseRef,omitempty"`

	// HeadRepository is the GraphQL ID of the head repository.
	HeadRepository string `json:"headRepository,omitempty"`
	HeadRef        string `json:"headRef,omitempty"`

	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`

	Commits []GitCommitDescription `json:"commits,omitempty"`

	Published PublishedValue `json:"published,omitempty"`
}

type GitCommitDescription struct {
	Message     string `json:"message,omitempty"`
	Diff        string `json:"diff,omitempty"`
	AuthorName  string `json:"authorName,omitempty"`
	AuthorEmail string `json:"authorEmail,omitempty"`
}

// Type returns the ChangesetSpecDescriptionType of the ChangesetSpecDescription.
func (d *ChangesetSpec) Type() ChangesetSpecDescriptionType {
	if d.ExternalID != "" {
		return ChangesetSpecDescriptionTypeExisting
	}
	return ChangesetSpecDescriptionTypeBranch
}

// IsExisting returns whether the description is of type
// ChangesetSpecDescriptionTypeExisting.
func (d *ChangesetSpec) IsImportingExisting() bool {
	return d.Type() == ChangesetSpecDescriptionTypeExisting
}

// IsBranch returns whether the description is of type
// ChangesetSpecDescriptionTypeBranch.
func (d *ChangesetSpec) IsBranch() bool {
	return d.Type() == ChangesetSpecDescriptionTypeBranch
}

// ChangesetSpecDescriptionType tells the consumer what the type of a
// ChangesetSpecDescription is without having to look into the description.
// Useful in the GraphQL when a HiddenChangesetSpec is returned.
type ChangesetSpecDescriptionType string

// Valid ChangesetSpecDescriptionTypes kinds
const (
	ChangesetSpecDescriptionTypeExisting ChangesetSpecDescriptionType = "EXISTING"
	ChangesetSpecDescriptionTypeBranch   ChangesetSpecDescriptionType = "BRANCH"
)

// ErrNoCommits is returned by (*ChangesetSpecDescription).Diff if the
// description doesn't have any commits descriptions.
var ErrNoCommits = errors.New("changeset description doesn't contain commit descriptions")

// Diff returns the Diff of the first GitCommitDescription in Commits. If the
// ChangesetSpecDescription doesn't have Commits it returns ErrNoCommits.
//
// We currently only support a single commit in Commits. Once we support more,
// this method will need to be revisited.
func (d *ChangesetSpec) Diff() (string, error) {
	if len(d.Commits) == 0 {
		return "", ErrNoCommits
	}
	return d.Commits[0].Diff, nil
}

// CommitMessage returns the Message of the first GitCommitDescription in Commits. If the
// ChangesetSpecDescription doesn't have Commits it returns ErrNoCommits.
//
// We currently only support a single commit in Commits. Once we support more,
// this method will need to be revisited.
func (d *ChangesetSpec) CommitMessage() (string, error) {
	if len(d.Commits) == 0 {
		return "", ErrNoCommits
	}
	return d.Commits[0].Message, nil
}

// AuthorName returns the author name of the first GitCommitDescription in Commits. If the
// ChangesetSpecDescription doesn't have Commits it returns ErrNoCommits.
//
// We currently only support a single commit in Commits. Once we support more,
// this method will need to be revisited.
func (d *ChangesetSpec) AuthorName() (string, error) {
	if len(d.Commits) == 0 {
		return "", ErrNoCommits
	}
	return d.Commits[0].AuthorName, nil
}

// AuthorEmail returns the author email of the first GitCommitDescription in Commits. If the
// ChangesetSpecDescription doesn't have Commits it returns ErrNoCommits.
//
// We currently only support a single commit in Commits. Once we support more,
// this method will need to be revisited.
func (d *ChangesetSpec) AuthorEmail() (string, error) {
	if len(d.Commits) == 0 {
		return "", ErrNoCommits
	}
	return d.Commits[0].AuthorEmail, nil
}
