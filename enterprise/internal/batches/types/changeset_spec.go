package types

import (
	"io"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

func NewChangesetSpecFromRaw(rawSpec string) (*ChangesetSpec, error) {
	spec, err := batcheslib.ParseChangesetSpec([]byte(rawSpec))
	if err != nil {
		return nil, err
	}

	return NewChangesetSpecFromSpec(spec)
}

func NewChangesetSpecFromSpec(spec *batcheslib.ChangesetSpec) (*ChangesetSpec, error) {
	diff, err := spec.Diff()
	if err != nil {
		return nil, err
	}
	var typ ChangesetSpecType
	if spec.IsImportingExisting() {
		typ = ChangesetSpecTypeExisting
	} else {
		typ = ChangesetSpecTypeBranch
	}
	baseRepoID, err := graphqlbackend.UnmarshalRepositoryID(graphql.ID(spec.BaseRepository))
	if err != nil {
		return nil, err
	}
	headRepoID, err := graphqlbackend.UnmarshalRepositoryID(graphql.ID(spec.HeadRepository))
	if err != nil {
		return nil, err
	}
	commitMsg, err := spec.CommitMessage()
	if err != nil {
		return nil, err
	}
	authorName, err := spec.AuthorName()
	if err != nil {
		return nil, err
	}
	authorEmail, err := spec.AuthorEmail()
	if err != nil {
		return nil, err
	}
	c := &ChangesetSpec{
		Diff:              diff,
		Typ:               typ,
		BaseRepoID:        baseRepoID,
		ExternalID:        spec.ExternalID,
		HeadRef:           spec.HeadRef,
		Title:             spec.Title,
		HeadRepoID:        headRepoID,
		BaseRev:           spec.BaseRev,
		BaseRef:           spec.BaseRef,
		Body:              spec.Body,
		Published:         spec.Published,
		CommitMessage:     commitMsg,
		CommitAuthorName:  authorName,
		CommitAuthorEmail: authorEmail,
	}
	c.computeForkNamespace()
	return c, c.computeDiffStat()
}

type ChangesetSpecType string

const (
	ChangesetSpecTypeBranch   ChangesetSpecType = "branch"
	ChangesetSpecTypeExisting ChangesetSpecType = "existing"
)

type ChangesetSpec struct {
	ID     int64
	RandID string

	// Spec *batcheslib.ChangesetSpec

	Typ  ChangesetSpecType
	Diff string

	DiffStatAdded   int32
	DiffStatChanged int32
	DiffStatDeleted int32

	BatchSpecID int64
	BaseRepoID  api.RepoID
	HeadRepoID  api.RepoID
	UserID      int32

	CreatedAt time.Time
	UpdatedAt time.Time

	ExternalID        string
	BaseRev           string
	BaseRef           string
	HeadRef           string
	Title             string
	Body              string
	Published         batcheslib.PublishedValue
	CommitMessage     string
	CommitAuthorName  string
	CommitAuthorEmail string

	ForkNamespace *string
}

// Clone returns a clone of a ChangesetSpec.
func (cs *ChangesetSpec) Clone() *ChangesetSpec {
	cc := *cs
	return &cc
}

// computeDiffStat parses the Diff of the ChangesetSpecDescription and sets the
// diff stat fields that can be retrieved with DiffStat().
// If the Diff is invalid or parsing failed, an error is returned.
func (cs *ChangesetSpec) computeDiffStat() error {
	if cs.Typ == ChangesetSpecTypeImporting {
		return nil
	}

	stats := diff.Stat{}
	reader := diff.NewMultiFileDiffReader(strings.NewReader(cs.Diff))
	for {
		fileDiff, err := reader.ReadFile()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		stat := fileDiff.Stat()
		stats.Added += stat.Added
		stats.Deleted += stat.Deleted
		stats.Changed += stat.Changed
	}

	cs.DiffStatAdded = stats.Added
	cs.DiffStatDeleted = stats.Deleted
	cs.DiffStatChanged = stats.Changed

	return nil
}

// computeForkNamespace calculates the namespace that the changeset spec will be
// forked into, if any.
func (cs *ChangesetSpec) computeForkNamespace() {
	// Right now, we only look at the global enforceForks setting, but we will
	// likely base this off the description eventually as well.
	if conf.Get().BatchChangesEnforceForks {
		cs.setForkToUser()
	}
}

// DiffStat returns a *diff.Stat.
func (cs *ChangesetSpec) DiffStat() diff.Stat {
	return diff.Stat{
		Added:   cs.DiffStatAdded,
		Deleted: cs.DiffStatDeleted,
		Changed: cs.DiffStatChanged,
	}
}

// ChangesetSpecTTL specifies the TTL of ChangesetSpecs that haven't been
// attached to a BatchSpec.
// It's lower than BatchSpecTTL because ChangesetSpecs should be attached to
// a BatchSpec immediately after having been created, whereas a BatchSpec
// might take a while to be complete and might also go through a lengthy review
// phase.
const ChangesetSpecTTL = 2 * 24 * time.Hour

// ExpiresAt returns the time when the ChangesetSpec will be deleted if not
// attached to a BatchSpec.
func (cs *ChangesetSpec) ExpiresAt() time.Time {
	return cs.CreatedAt.Add(ChangesetSpecTTL)
}

// ChangesetSpecs is a slice of *ChangesetSpecs.
type ChangesetSpecs []*ChangesetSpec

// IDs returns the unique RepoIDs of all changeset specs in the slice.
func (cs ChangesetSpecs) RepoIDs() []api.RepoID {
	repoIDMap := make(map[api.RepoID]struct{})
	for _, c := range cs {
		repoIDMap[c.BaseRepoID] = struct{}{}
	}
	repoIDs := make([]api.RepoID, 0)
	for id := range repoIDMap {
		repoIDs = append(repoIDs, id)
	}
	return repoIDs
}

// changesetSpecForkNamespaceUser is the sentinel value used in the database to
// indicate that the changeset spec should be forked into the user's namespace,
// which we don't know at spec upload time.
const changesetSpecForkNamespaceUser = "<user>"

// IsFork returns true if the changeset spec should be pushed to a fork.
func (cs *ChangesetSpec) IsFork() bool {
	return cs.ForkNamespace != nil
}

// GetForkNamespace returns the namespace if the changeset spec should be pushed
// to a named fork, or nil if the changeset spec shouldn't be pushed to a fork
// _or_ should be pushed to a fork in the user's default namespace.
func (cs *ChangesetSpec) GetForkNamespace() *string {
	if cs.ForkNamespace != nil && *cs.ForkNamespace != changesetSpecForkNamespaceUser {
		return cs.ForkNamespace
	}
	return nil
}

func (cs *ChangesetSpec) setForkToUser() {
	s := changesetSpecForkNamespaceUser
	cs.ForkNamespace = &s
}
