package types

import (
	"bytes"
	"io"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	godiff "github.com/sourcegraph/go-diff/diff"

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
	var baseRepoID api.RepoID
	err := relay.UnmarshalSpec(graphql.ID(spec.BaseRepository), &baseRepoID)
	if err != nil {
		return nil, err
	}

	c := &ChangesetSpec{
		BaseRepoID: baseRepoID,
		ExternalID: spec.ExternalID,
		Title:      spec.Title,
		Body:       spec.Body,
		Published:  spec.Published,
	}

	if spec.IsImportingExisting() {
		c.Type = ChangesetSpecTypeExisting
	} else {
		var headRepoID api.RepoID
		err := relay.UnmarshalSpec(graphql.ID(spec.HeadRepository), &headRepoID)
		if err != nil {
			return nil, err
		}
		if baseRepoID != headRepoID {
			return nil, batcheslib.ErrHeadBaseMismatch
		}

		diff, err := spec.Diff()
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
		c.Type = ChangesetSpecTypeBranch
		c.Diff = diff
		c.HeadRef = spec.HeadRef
		c.BaseRev = spec.BaseRev
		c.BaseRef = spec.BaseRef
		c.CommitMessage = commitMsg
		c.CommitAuthorName = authorName
		c.CommitAuthorEmail = authorEmail
	}

	c.computeForkNamespace(spec.Fork)
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

	Type ChangesetSpecType

	DiffStatAdded   int32
	DiffStatDeleted int32

	BatchSpecID int64
	BaseRepoID  api.RepoID
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
	Diff              []byte
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
	if cs.Type == ChangesetSpecTypeExisting {
		return nil
	}

	stats := godiff.Stat{}
	reader := godiff.NewMultiFileDiffReader(bytes.NewReader(cs.Diff))
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

	cs.DiffStatAdded = stats.Added + stats.Changed
	cs.DiffStatDeleted = stats.Deleted + stats.Changed

	return nil
}

// computeForkNamespace calculates the namespace that the changeset spec will be
// forked into, if any.
func (cs *ChangesetSpec) computeForkNamespace(forkFromSpec *bool) {
	// If the fork property is unspecified in the batch spec changesetTemplate,
	// we only look at the global enforceForks setting. But if the property *is*
	// specified, it takes precedence.
	if forkFromSpec == nil {
		if conf.Get().BatchChangesEnforceForks {
			cs.setForkToUser()
		}
	} else {
		if *forkFromSpec {
			cs.setForkToUser()
		}
	}
}

// DiffStat returns a *diff.Stat.
func (cs *ChangesetSpec) DiffStat() godiff.Stat {
	return godiff.Stat{
		Added:   cs.DiffStatAdded,
		Deleted: cs.DiffStatDeleted,
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
