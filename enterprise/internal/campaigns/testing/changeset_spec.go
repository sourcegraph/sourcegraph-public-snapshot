package testing

import (
	"context"
	"testing"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

type TestSpecOpts struct {
	User         int32
	Repo         api.RepoID
	CampaignSpec int64

	// If this is non-blank, the changesetSpec will be an import/track spec for
	// the changeset with the given ExternalID in the given repo.
	ExternalID string

	// If this is set, the changesetSpec will be a "create commit on this
	// branch" changeset spec.
	HeadRef string

	// If this is set along with headRef, the changesetSpec will have Published
	// set.
	Published interface{}

	Title             string
	Body              string
	CommitMessage     string
	CommitDiff        string
	CommitAuthorEmail string
	CommitAuthorName  string

	BaseRev string
	BaseRef string
}

var TestChangsetSpecDiffStat = &diff.Stat{Added: 10, Changed: 5, Deleted: 2}

func BuildChangesetSpec(t *testing.T, opts TestSpecOpts) *campaigns.ChangesetSpec {
	t.Helper()

	published := campaigns.PublishedValue{Val: opts.Published}
	if opts.Published == nil {
		// Set false as the default.
		published.Val = false
	}
	if !published.Valid() {
		t.Fatalf("invalid value for published passed, got %v (%T)", opts.Published, opts.Published)
	}

	spec := &campaigns.ChangesetSpec{
		UserID:         opts.User,
		RepoID:         opts.Repo,
		CampaignSpecID: opts.CampaignSpec,
		Spec: &campaigns.ChangesetSpecDescription{
			BaseRepository: graphqlbackend.MarshalRepositoryID(opts.Repo),

			BaseRev: opts.BaseRev,
			BaseRef: opts.BaseRef,

			ExternalID: opts.ExternalID,
			HeadRef:    opts.HeadRef,
			Published:  published,

			Title: opts.Title,
			Body:  opts.Body,

			Commits: []campaigns.GitCommitDescription{
				{
					Message:     opts.CommitMessage,
					Diff:        opts.CommitDiff,
					AuthorEmail: opts.CommitAuthorEmail,
					AuthorName:  opts.CommitAuthorName,
				},
			},
		},
		DiffStatAdded:   TestChangsetSpecDiffStat.Added,
		DiffStatChanged: TestChangsetSpecDiffStat.Changed,
		DiffStatDeleted: TestChangsetSpecDiffStat.Deleted,
	}

	return spec
}

type CreateChangesetSpecer interface {
	CreateChangesetSpec(ctx context.Context, changesetSpec *campaigns.ChangesetSpec) error
}

func CreateChangesetSpec(
	t *testing.T,
	ctx context.Context,
	store CreateChangesetSpecer,
	opts TestSpecOpts,
) *campaigns.ChangesetSpec {
	t.Helper()

	spec := BuildChangesetSpec(t, opts)

	if err := store.CreateChangesetSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}

	return spec
}
