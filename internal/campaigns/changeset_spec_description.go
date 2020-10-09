package campaigns

import (
	"encoding/json"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
)

type Publish struct{ Val interface{} }

// True is true if the enclosed value is a bool being true.
func (p *Publish) True() bool {
	if b, ok := p.Val.(bool); ok {
		return b
	}
	return false
}

// False is true if the enclosed value is a bool being false.
func (p Publish) False() bool {
	if b, ok := p.Val.(bool); ok {
		return !b
	}
	return false
}

// Draft is true if the enclosed value is a string being "draft".
func (p Publish) Draft() bool {
	if s, ok := p.Val.(string); ok {
		return strings.EqualFold(s, "draft")
	}
	return false
}

// Valid returns whether the enclosed value is of any of the permitted types.
func (p *Publish) Valid() bool {
	return p.True() || p.False() || p.Draft()
}

func (p Publish) MarshalJSON() ([]byte, error) {
	if !p.Valid() {
		if p.Val == nil {
			v := "null"
			return []byte(v), nil
		}
		return nil, errors.New("invalid value")
	}
	if p.True() {
		v := "true"
		return []byte(v), nil
	}
	if p.False() {
		v := "false"
		return []byte(v), nil
	}
	v := `"draft"`
	return []byte(v), nil
}

func (p *Publish) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &p.Val); err != nil {
		return err
	}
	return nil
}

func (p Publish) UnmarshalYAML(b []byte) error {
	if err := json.Unmarshal(b, &p.Val); err != nil {
		return err
	}
	return nil
}

func (p *Publish) UnmarshalGraphQL(input interface{}) error {
	p.Val = input
	return nil
}

func (p *Publish) ImplementsGraphQLType(name string) bool {
	return name == "PublishedTriple"
}

type ChangesetSpecDescription struct {
	BaseRepository graphql.ID `json:"baseRepository,omitempty"`

	// If this is not empty, the description is a reference to an existing
	// changeset and the rest of these fields are empty.
	ExternalID string `json:"externalID,omitempty"`

	BaseRev string `json:"baseRev,omitempty"`
	BaseRef string `json:"baseRef,omitempty"`

	HeadRepository graphql.ID `json:"headRepository,omitempty"`
	HeadRef        string     `json:"headRef,omitempty"`

	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`

	Commits []GitCommitDescription `json:"commits,omitempty"`

	Published Publish `json:"published,omitempty"`
}

// Type returns the ChangesetSpecDescriptionType of the ChangesetSpecDescription.
func (d *ChangesetSpecDescription) Type() ChangesetSpecDescriptionType {
	if d.ExternalID != "" {
		return ChangesetSpecDescriptionTypeExisting
	}
	return ChangesetSpecDescriptionTypeBranch
}

// IsExisting returns whether the description is of type
// ChangesetSpecDescriptionTypeExisting.
func (d *ChangesetSpecDescription) IsImportingExisting() bool {
	return d.Type() == ChangesetSpecDescriptionTypeExisting
}

// IsBranch returns whether the description is of type
// ChangesetSpecDescriptionTypeBranch.
func (d *ChangesetSpecDescription) IsBranch() bool {
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
func (d *ChangesetSpecDescription) Diff() (string, error) {
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
func (d *ChangesetSpecDescription) CommitMessage() (string, error) {
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
func (d *ChangesetSpecDescription) AuthorName() (string, error) {
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
func (d *ChangesetSpecDescription) AuthorEmail() (string, error) {
	if len(d.Commits) == 0 {
		return "", ErrNoCommits
	}
	return d.Commits[0].AuthorEmail, nil
}
