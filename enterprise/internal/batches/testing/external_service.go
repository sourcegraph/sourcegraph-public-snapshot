package testing

import (
	"context"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type ExternalServiceWebhookFixture struct {
	AdminUser   *types.User
	RegularUser *types.User

	BitbucketServerRepos []*types.Repo
	BitbucketServerSvc   *types.ExternalService
	GitHubRepos          []*types.Repo
	GitHubSvc            *types.ExternalService
	GitLabRepos          []*types.Repo
	GitLabSvc            *types.ExternalService
	OtherRepos           []*types.Repo
	OtherSvc             *types.ExternalService

	EmptyBatchChange   *btypes.BatchChange
	OtherBatchChange   *btypes.BatchChange
	PrimaryBatchChange *btypes.BatchChange
}

// WebhookStore is an interface representing the subset of store.Store methods
// that are required for NewExternalServiceWebhookFixture, which can't just
// receive a *store.Store as that would create an import cycle.
type WebhookStore interface {
	CreateBatchChanger
	CreateBatchSpecer
	CreateChangeseter
	CreateChangesetSpecer

	// Methods from store.Store.
	Clock() func() time.Time
	DB() dbutil.DB

	// Methods from basestore.Store.
	Exec(context.Context, *sqlf.Query) error
}

// NewExternalServiceWebhookFixture sets up a scenario that's used both the
// store and resolver tests.
//
// The fixture specifically includes:
//
// 1. Two users: one admin; one regular.
// 2. A GitHub service with two repos. The first repo is visible to the regular
//    user.
// 3. A GitLab service with one repo, which is invisible to the regular user.
// 4. A BitBucket Server service with one repo and no permission restrictions.
// 5. An "other" service with one repo and no permission restrictions.
// 6. An empty batch change.
// 7. An "other" batch change with one changeset on the other repo.
// 8. A "primary" batch change with one changeset on each of:
//    (a) GitHub repo 0,
//    (b) The GitLab repo,
//    (c) The Bitbucket Server repo.
//
// This allows for permissions and pagination to be tested, as we have the
// various scenarios required.
func NewExternalServiceWebhookFixture(t *testing.T, ctx context.Context, store WebhookStore) *ExternalServiceWebhookFixture {
	// Create the fixture and set up the users.
	fixture := &ExternalServiceWebhookFixture{
		AdminUser:   CreateTestUser(t, store.DB(), true),
		RegularUser: CreateTestUser(t, store.DB(), false),
	}

	// Create the repos and external services, for now without any permissions.
	fixture.GitHubRepos, fixture.GitHubSvc = CreateGitHubSSHTestRepos(t, ctx, store.DB(), 2)
	fixture.GitLabRepos, fixture.GitLabSvc = CreateGitlabTestRepos(t, ctx, store.DB(), 1)
	fixture.BitbucketServerRepos, fixture.BitbucketServerSvc = CreateBbsSSHTestRepos(t, ctx, store.DB(), 1)
	fixture.OtherRepos, fixture.OtherSvc = CreateGitHubSSHTestRepos(t, ctx, store.DB(), 1)

	// Create the primary batch change and its changesets.
	primaryBatchSpec := CreateBatchSpec(t, ctx, store, "external-services", fixture.RegularUser.ID)
	fixture.PrimaryBatchChange = CreateBatchChange(t, ctx, store, "external-services", fixture.RegularUser.ID, primaryBatchSpec.ID)

	repos := []*types.Repo{
		fixture.GitHubRepos[0],
		fixture.GitHubRepos[1],
		fixture.GitLabRepos[0],
		fixture.BitbucketServerRepos[0],
	}
	changesets := make([]*btypes.Changeset, len(repos))
	for i, repo := range repos {
		spec := CreateChangesetSpec(t, ctx, store, TestSpecOpts{
			User:      fixture.RegularUser.ID,
			Repo:      repo.ID,
			BatchSpec: primaryBatchSpec.ID,
			HeadRef:   "main",
			Published: true,
		})
		changesets[i] = CreateChangeset(t, ctx, store, TestChangesetOpts{
			Repo:             repo.ID,
			BatchChange:      fixture.PrimaryBatchChange.ID,
			CurrentSpec:      spec.ID,
			PublicationState: btypes.ChangesetPublicationStatePublished,
		})
	}

	// Create the other batch change.
	otherBatchSpec := CreateBatchSpec(t, ctx, store, "other", fixture.RegularUser.ID)
	fixture.OtherBatchChange = CreateBatchChange(t, ctx, store, "other", fixture.RegularUser.ID, otherBatchSpec.ID)
	otherChangesetSpec := CreateChangesetSpec(t, ctx, store, TestSpecOpts{
		User:      fixture.RegularUser.ID,
		Repo:      fixture.OtherRepos[0].ID,
		BatchSpec: otherBatchSpec.ID,
		HeadRef:   "main",
		Published: true,
	})
	CreateChangeset(t, ctx, store, TestChangesetOpts{
		Repo:             fixture.OtherRepos[0].ID,
		BatchChange:      fixture.OtherBatchChange.ID,
		CurrentSpec:      otherChangesetSpec.ID,
		PublicationState: btypes.ChangesetPublicationStatePublished,
	})

	// Create the empty batch change.
	emptyBatchSpec := CreateBatchSpec(t, ctx, store, "empty", fixture.RegularUser.ID)
	fixture.EmptyBatchChange = CreateBatchChange(t, ctx, store, "empty", fixture.RegularUser.ID, emptyBatchSpec.ID)

	// Set the GitLab and GitHub repos to be private, and their associated
	// external services to be restricted. We need to do these with raw queries
	// because the external service store will recalculate Unrestricted based on
	// the configuration, and because the repo store doesn't provide an update
	// or upsert method.
	store.Exec(
		ctx,
		sqlf.Sprintf(
			"UPDATE external_services SET unrestricted = FALSE WHERE id IN (%d, %d)",
			fixture.GitHubSvc.ID, fixture.GitLabSvc.ID,
		),
	)

	store.Exec(
		ctx,
		sqlf.Sprintf(
			"UPDATE repo SET private = TRUE where id IN (%d, %d)",
			fixture.GitHubRepos[0].ID, fixture.GitLabRepos[0].ID,
		),
	)

	// For comparison purposes later, though, we need the restricted external
	// services to match what we just did in the database.
	fixture.GitHubSvc.Unrestricted = false
	fixture.GitLabSvc.Unrestricted = false

	// Set up the repo permissions to allow the regular user access to GitHub,
	// but not GitLab.
	MockRepoPermissions(t, store.DB(), fixture.RegularUser.ID, fixture.GitHubRepos[0].ID)

	return fixture
}
