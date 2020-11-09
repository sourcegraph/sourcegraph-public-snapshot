package campaigns

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func testStoreCodeHost(t *testing.T, ctx context.Context, s *Store, reposStore repos.Store, clock clock) {
	repo := ct.TestRepo(t, reposStore, extsvc.KindGitHub)
	otherRepo := ct.TestRepo(t, reposStore, extsvc.KindGitHub)
	gitlabRepo := ct.TestRepo(t, reposStore, extsvc.KindGitLab)
	bitbucketRepo := ct.TestRepo(t, reposStore, extsvc.KindBitbucketServer)
	awsRepo := ct.TestRepo(t, reposStore, extsvc.KindAWSCodeCommit)

	if err := reposStore.InsertRepos(ctx, repo, otherRepo, gitlabRepo, bitbucketRepo, awsRepo); err != nil {
		t.Fatal(err)
	}
	deletedRepo := otherRepo.With(repos.Opt.RepoDeletedAt(clock.now()))
	if err := reposStore.DeleteRepos(ctx, deletedRepo.ID); err != nil {
		t.Fatal(err)
	}

	t.Run("ListCodeHosts", func(t *testing.T) {
		have, err := s.ListCodeHosts(ctx)
		if err != nil {
			t.Fatal(err)
		}
		want := []*campaigns.CodeHost{
			{
				ExternalServiceType: extsvc.TypeBitbucketServer,
				ExternalServiceID:   "https://bitbucketserver.com/",
			},
			{
				ExternalServiceType: extsvc.TypeGitHub,
				ExternalServiceID:   "https://github.com/",
			},
			{
				ExternalServiceType: extsvc.TypeGitLab,
				ExternalServiceID:   "https://gitlab.com/",
			},
		}
		if diff := cmp.Diff(have, want); diff != "" {
			t.Fatalf("Invalid code hosts returned. %s", diff)
		}
	})
}
