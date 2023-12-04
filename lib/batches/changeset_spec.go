package batches

import (
	"encoding/json"
	"reflect"
	"strconv"

	jsonutil "github.com/sourcegraph/sourcegraph/lib/batches/json"
	"github.com/sourcegraph/sourcegraph/lib/batches/schema"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

// ParseChangesetSpecExternalID attempts to parse the ID of a changeset in the
// batch spec that should be imported.
func ParseChangesetSpecExternalID(id any) (string, error) {
	var sid string

	switch tid := id.(type) {
	case string:
		sid = tid
	case int, int8, int16, int32, int64:
		sid = strconv.FormatInt(reflect.ValueOf(id).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		sid = strconv.FormatUint(reflect.ValueOf(id).Uint(), 10)
	case float32:
		sid = strconv.FormatFloat(float64(tid), 'f', -1, 32)
	case float64:
		sid = strconv.FormatFloat(tid, 'f', -1, 64)
	default:
		return "", NewValidationError(errors.Newf("cannot convert value of type %T into a valid external ID: expected string or int", id))
	}

	return sid, nil
}

// Note: When modifying this struct, make sure to reflect the new fields below in
// the customized MarshalJSON method.

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
	Fork  *bool  `json:"fork,omitempty"`

	Commits []GitCommitDescription `json:"commits,omitempty"`

	Published PublishedValue `json:"published,omitempty"`
}

// MarshalJSON overwrites the default behavior of the json lib while unmarshalling
// a *ChangesetSpec. We explicitly only set Published, when it's non-nil. Due to
// it not being a pointer, omitempty does nothing. That causes it to fail schema
// validation.
// TODO: This is the easiest workaround for now, without risking breaking anything
// right before the release. Ideally, we split up this type into two separate ones
// in the future.
// See https://github.com/sourcegraph/sourcegraph/issues/25968.
func (c *ChangesetSpec) MarshalJSON() ([]byte, error) {
	v := struct {
		BaseRepository string                 `json:"baseRepository,omitempty"`
		ExternalID     string                 `json:"externalID,omitempty"`
		BaseRev        string                 `json:"baseRev,omitempty"`
		BaseRef        string                 `json:"baseRef,omitempty"`
		HeadRepository string                 `json:"headRepository,omitempty"`
		HeadRef        string                 `json:"headRef,omitempty"`
		Title          string                 `json:"title,omitempty"`
		Body           string                 `json:"body,omitempty"`
		Commits        []GitCommitDescription `json:"commits,omitempty"`
		Published      *PublishedValue        `json:"published,omitempty"`
		Fork           *bool                  `json:"fork,omitempty"`
	}{
		BaseRepository: c.BaseRepository,
		ExternalID:     c.ExternalID,
		BaseRev:        c.BaseRev,
		BaseRef:        c.BaseRef,
		HeadRepository: c.HeadRepository,
		HeadRef:        c.HeadRef,
		Title:          c.Title,
		Body:           c.Body,
		Commits:        c.Commits,
		Fork:           c.Fork,
	}
	if !c.Published.Nil() {
		v.Published = &c.Published
	}
	return json.Marshal(&v)
}

type GitCommitDescription struct {
	Version     int    `json:"version,omitempty"`
	Message     string `json:"message,omitempty"`
	Diff        []byte `json:"diff,omitempty"`
	AuthorName  string `json:"authorName,omitempty"`
	AuthorEmail string `json:"authorEmail,omitempty"`
}

func (a GitCommitDescription) MarshalJSON() ([]byte, error) {
	if a.Version == 2 {
		return json.Marshal(v2GitCommitDescription(a))
	}
	return json.Marshal(v1GitCommitDescription{
		Message:     a.Message,
		Diff:        string(a.Diff),
		AuthorName:  a.AuthorName,
		AuthorEmail: a.AuthorEmail,
	})
}

func (a *GitCommitDescription) UnmarshalJSON(data []byte) error {
	var version versionGitCommitDescription
	if err := json.Unmarshal(data, &version); err != nil {
		return err
	}
	if version.Version == 2 {
		var v2 v2GitCommitDescription
		if err := json.Unmarshal(data, &v2); err != nil {
			return err
		}
		a.Version = v2.Version
		a.Message = v2.Message
		a.Diff = v2.Diff
		a.AuthorName = v2.AuthorName
		a.AuthorEmail = v2.AuthorEmail
		return nil
	}
	var v1 v1GitCommitDescription
	if err := json.Unmarshal(data, &v1); err != nil {
		return err
	}
	a.Message = v1.Message
	a.Diff = []byte(v1.Diff)
	a.AuthorName = v1.AuthorName
	a.AuthorEmail = v1.AuthorEmail
	return nil
}

type versionGitCommitDescription struct {
	Version int `json:"version,omitempty"`
}

type v2GitCommitDescription struct {
	Version     int    `json:"version,omitempty"`
	Message     string `json:"message,omitempty"`
	Diff        []byte `json:"diff,omitempty"`
	AuthorName  string `json:"authorName,omitempty"`
	AuthorEmail string `json:"authorEmail,omitempty"`
}

type v1GitCommitDescription struct {
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

// IsImportingExisting returns whether the description is of type
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
func (d *ChangesetSpec) Diff() ([]byte, error) {
	if len(d.Commits) == 0 {
		return nil, ErrNoCommits
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
