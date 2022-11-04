package testing

import (
	"context"
	"testing"

	"github.com/sourcegraph/go-diff/diff"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/batches"
)

type TestSpecOpts struct {
	ID        int64
	User      int32
	Repo      api.RepoID
	BatchSpec int64

	// If this is non-blank, the changesetSpec will be an import/track spec for
	// the changeset with the given ExternalID in the given repo.
	ExternalID string

	// If this is set, the changesetSpec will be a "create commit on this
	// branch" changeset spec.
	HeadRef string

	// If this is set along with headRef, the changesetSpec will have Published
	// set.
	Published any

	Title             string
	Body              string
	CommitMessage     string
	CommitDiff        string
	CommitAuthorEmail string
	CommitAuthorName  string

	BaseRev string
	BaseRef string

	Typ btypes.ChangesetSpecType
}

var TestChangsetSpecDiffStat = &diff.Stat{Added: 10, Changed: 5, Deleted: 2}

func BuildChangesetSpec(t *testing.T, opts TestSpecOpts) *btypes.ChangesetSpec {
	t.Helper()

	published := batches.PublishedValue{Val: opts.Published}
	if !published.Valid() {
		t.Fatalf("invalid value for published passed, got %v (%T)", opts.Published, opts.Published)
	}

	if opts.Typ == "" {
		t.Fatal("empty typ on changeset spec in test helper")
	}

	spec := &btypes.ChangesetSpec{
		ID:                opts.ID,
		UserID:            opts.User,
		BaseRepoID:        opts.Repo,
		BatchSpecID:       opts.BatchSpec,
		BaseRev:           opts.BaseRev,
		BaseRef:           opts.BaseRef,
		ExternalID:        opts.ExternalID,
		HeadRef:           opts.HeadRef,
		Published:         published,
		Title:             opts.Title,
		Body:              opts.Body,
		CommitMessage:     opts.CommitMessage,
		Diff:              []byte(opts.CommitDiff),
		CommitAuthorEmail: opts.CommitAuthorEmail,
		CommitAuthorName:  opts.CommitAuthorName,
		DiffStatAdded:     TestChangsetSpecDiffStat.Added,
		DiffStatChanged:   TestChangsetSpecDiffStat.Changed,
		DiffStatDeleted:   TestChangsetSpecDiffStat.Deleted,
		Type:              opts.Typ,
	}

	return spec
}

type CreateChangesetSpecer interface {
	CreateChangesetSpec(ctx context.Context, changesetSpecs ...*btypes.ChangesetSpec) error
}

func CreateChangesetSpec(
	t *testing.T,
	ctx context.Context,
	store CreateChangesetSpecer,
	opts TestSpecOpts,
) *btypes.ChangesetSpec {
	t.Helper()

	spec := BuildChangesetSpec(t, opts)

	if err := store.CreateChangesetSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}

	return spec
}
