package campaigns

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	edb "github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func testStoreCodeHost(t *testing.T, ctx context.Context, s *Store, db dbutil.DB, clock clock) {
	reposStore := edb.NewRepoStoreWithDB(db)
	esStore := edb.NewExternalServicesStoreWithDB(db)

	repo := ct.TestRepo(t, esStore, extsvc.KindGitHub)
	otherRepo := ct.TestRepo(t, esStore, extsvc.KindGitHub)
	gitlabRepo := ct.TestRepo(t, esStore, extsvc.KindGitLab)
	bitbucketRepo := ct.TestRepo(t, esStore, extsvc.KindBitbucketServer)
	awsRepo := ct.TestRepo(t, esStore, extsvc.KindAWSCodeCommit)

	if err := reposStore.Create(ctx, repo, otherRepo, gitlabRepo, bitbucketRepo, awsRepo); err != nil {
		t.Fatal(err)
	}
	deletedRepo := otherRepo.With(repos.Opt.RepoDeletedAt(clock.now()))
	if err := reposStore.Delete(ctx, deletedRepo.ID); err != nil {
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

	t.Run("GetExternalServiceID", func(t *testing.T) {
		for _, repo := range []*types.Repo{repo, otherRepo, gitlabRepo, bitbucketRepo} {
			id, err := s.GetExternalServiceID(ctx, GetExternalServiceIDOpts{
				ExternalServiceType: repo.ExternalRepo.ServiceType,
				ExternalServiceID:   repo.ExternalRepo.ServiceID,
			})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// We fetch the ExternalService and make sure that Type and URL match
			es, err := esStore.List(ctx, edb.ExternalServicesListOptions{
				IDs: []int64{id},
			})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if len(es) != 1 {
				t.Fatalf("wrong number of external services: %d", len(es))
			}
			extSvc := es[0]
			if have, want := extSvc.Kind, extsvc.TypeToKind(repo.ExternalRepo.ServiceType); have != want {
				t.Fatalf("wrong external service kind. want=%q, have=%q", want, have)
			}
		}
	})
}
