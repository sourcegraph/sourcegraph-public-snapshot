package campaigns

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExecChangesetJob(t *testing.T) {
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time { return now.UTC().Truncate(time.Microsecond) }

	dbtesting.SetupGlobalTestDB(t)

	tests := []struct {
		name string

		createRepoExtSvc  func(t *testing.T, ctx context.Context, now time.Time, s *Store) (*repos.Repo, *repos.ExternalService)
		changesetMetadata func(now time.Time, c *cmpgn.Campaign, headRef string) interface{}

		existsOnCodehost bool
		existsInDB       bool
	}{
		{
			name:              "GitHub_NewChangeset",
			createRepoExtSvc:  createGitHubRepo,
			changesetMetadata: buildGithubPR,
		},
		{
			name:              "GitHub_ChangesetExistsOnCodehost",
			createRepoExtSvc:  createGitHubRepo,
			changesetMetadata: buildGithubPR,
			existsOnCodehost:  true,
		},

		{
			name:              "GitHub_ChangesetExistsInDB",
			createRepoExtSvc:  createGitHubRepo,
			changesetMetadata: buildGithubPR,
			existsInDB:        true,
		},
		{
			name:              "BitbucketServer_NewChangeset",
			createRepoExtSvc:  createBitbucketServerRepo,
			changesetMetadata: buildBitbucketServerPR,
		},
		{
			name:              "BitbucketServer_ChangesetExistsOnCodehost",
			createRepoExtSvc:  createBitbucketServerRepo,
			changesetMetadata: buildBitbucketServerPR,
			existsOnCodehost:  true,
		},

		{
			name:              "BitbucketServer_ChangesetExistsInDB",
			createRepoExtSvc:  createBitbucketServerRepo,
			changesetMetadata: buildBitbucketServerPR,
			existsInDB:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tx := dbtest.NewTx(t, dbconn.Global)
			s := NewStoreWithClock(tx, clock)

			repo, extSvc := tc.createRepoExtSvc(t, ctx, now, s)
			campaign, patch := createCampaignPatch(t, ctx, now, s, repo)

			headRef := "refs/heads/" + campaign.Branch
			baseRef := patch.BaseRef

			meta := tc.changesetMetadata(now, campaign, headRef)

			oldCreatedAt := now.Add(-5 * time.Second)
			if tc.existsInDB {
				// We simulate that the Changeset with the same external ID
				// for the same repository already exists in the DB, but with
				// empty metadata, so we can later check that it was properly
				// updated.
				// We also leave CreatedByCampaign false to ensure that it is set on
				// update too.
				ch := &cmpgn.Changeset{
					RepoID:    repo.ID,
					CreatedAt: oldCreatedAt,
					UpdatedAt: now.Add(-5 * time.Second),
				}
				// This sets ExternalID, which we need to trigger the
				// AlreadyExistsError.
				if err := ch.SetMetadata(meta); err != nil {
					t.Fatal(err)
				}
				// Now we can remove metadata.
				ch.Metadata = nil

				if err := s.CreateChangesets(ctx, ch); err != nil {
					t.Fatal(err)
				}
			}

			gitClient := &ct.FakeGitserverClient{Response: headRef, ResponseErr: nil}

			sourcer := repos.NewFakeSourcer(nil, &ct.FakeChangesetSource{
				Svc:             extSvc,
				Err:             nil,
				ChangesetExists: tc.existsOnCodehost,
				WantHeadRef:     headRef,
				WantBaseRef:     baseRef,
				FakeMetadata:    meta,
			})

			changesetJob := &cmpgn.ChangesetJob{CampaignID: campaign.ID, PatchID: patch.ID}
			if err := s.CreateChangesetJob(ctx, changesetJob); err != nil {
				t.Fatal(err)
			}

			err := ExecChangesetJob(ctx, campaign, changesetJob, ExecChangesetJobOpts{
				Clock: clock, Store: s, GitClient: gitClient, Sourcer: sourcer, ExternalURL: "http://localhost",
			})
			if err != nil {
				t.Fatal(err)
			}

			changesetJob, err = s.GetChangesetJob(ctx, GetChangesetJobOpts{ID: changesetJob.ID})
			if err != nil {
				t.Fatal(err)
			}

			if changesetJob.ChangesetID == 0 {
				t.Fatalf("ChangesetJob has not ChangesetID set")
			}

			wantChangeset := &cmpgn.Changeset{
				RepoID:              repo.ID,
				CampaignIDs:         []int64{campaign.ID},
				ExternalBranch:      headRef,
				ExternalState:       cmpgn.ChangesetExternalStateOpen,
				ExternalReviewState: cmpgn.ChangesetReviewStatePending,
				ExternalCheckState:  cmpgn.ChangesetCheckStateUnknown,
				CreatedAt:           now,
				UpdatedAt:           now,
				CreatedByCampaign:   true,
			}
			err = wantChangeset.SetMetadata(meta)
			if err != nil {
				t.Fatal(err)
			}

			if tc.existsInDB {
				// If it was already in the DB we want to make sure that all
				// other fields are updated, but not CreatedAt.
				wantChangeset.CreatedAt = oldCreatedAt
			}

			assertChangesetInDB(t, ctx, s, changesetJob.ChangesetID, wantChangeset)

			wantEvents := wantChangeset.Events()
			for _, e := range wantEvents {
				e.ChangesetID = changesetJob.ChangesetID
				e.UpdatedAt = now
				e.CreatedAt = now
			}
			assertChangesetEventsInDB(t, ctx, s, changesetJob.ChangesetID, wantEvents)
		})
	}
}

const testDiff = `diff --git foobar.c foobar.c
index d75b080..cf04b5b 100644
--- foobar.c
+++ foobar.c
@@ -1 +1 @@
-onto monto(int argc, char *argv[]) { printf("Nice."); }
+int main(int argc, char *argv[]) { printf("Nice."); }
`

func createGitHubRepo(t *testing.T, ctx context.Context, now time.Time, s *Store) (*repos.Repo, *repos.ExternalService) {
	t.Helper()

	reposStore := repos.NewDBStore(s.DB(), sql.TxOptions{})

	ext := &repos.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub",
		Config: marshalJSON(t, &schema.GitHubConnection{
			Url:   "https://github.com",
			Token: "SECRETTOKEN",
		}),
	}

	if err := reposStore.UpsertExternalServices(ctx, ext); err != nil {
		t.Fatal(err)
	}

	repo := testRepo(0, extsvc.TypeGitHub)
	repo.Sources = map[string]*repos.SourceInfo{ext.URN(): {
		ID: ext.URN(),
	}}
	if err := reposStore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	return repo, ext
}

func createBitbucketServerRepo(t *testing.T, ctx context.Context, now time.Time, s *Store) (*repos.Repo, *repos.ExternalService) {
	t.Helper()

	reposStore := repos.NewDBStore(s.DB(), sql.TxOptions{})

	ext := &repos.ExternalService{
		Kind:        extsvc.KindBitbucketServer,
		DisplayName: "Bitbucket Server",
		Config: marshalJSON(t, &schema.BitbucketServerConnection{
			Url:   "https://bbs.example.com",
			Token: "SECRETTOKEN",
		}),
	}

	if err := reposStore.UpsertExternalServices(ctx, ext); err != nil {
		t.Fatal(err)
	}

	repo := testRepo(0, extsvc.TypeBitbucketServer)
	repo.Sources = map[string]*repos.SourceInfo{ext.URN(): {
		ID: ext.URN(),
	}}
	if err := reposStore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	return repo, ext
}

func createCampaignPatch(t *testing.T, ctx context.Context, now time.Time, s *Store, repo *repos.Repo) (*cmpgn.Campaign, *cmpgn.Patch) {
	t.Helper()

	patchSet := &cmpgn.PatchSet{}
	if err := s.CreatePatchSet(ctx, patchSet); err != nil {
		t.Fatal(err)
	}

	patch := &cmpgn.Patch{
		RepoID:     repo.ID,
		PatchSetID: patchSet.ID,
		Diff:       testDiff,
		Rev:        "f00b4r",
		BaseRef:    "refs/heads/master",
	}
	if err := s.CreatePatch(ctx, patch); err != nil {
		t.Fatal(err)
	}

	campaign := &cmpgn.Campaign{
		Name:            "Remove dead code",
		Description:     "This campaign removes dead code.",
		Branch:          "dead-code-b-gone",
		AuthorID:        888,
		NamespaceUserID: 888,
		PatchSetID:      patchSet.ID,
		ClosedAt:        now,
	}
	if err := s.CreateCampaign(ctx, campaign); err != nil {
		t.Fatal(err)
	}

	return campaign, patch
}

func buildGithubPR(now time.Time, c *cmpgn.Campaign, headRef string) interface{} {
	return &github.PullRequest{
		ID:          "FOOBARID",
		Title:       c.Name,
		Body:        c.Description,
		HeadRefName: git.AbbreviateRef(headRef),
		Number:      12345,
		State:       "OPEN",
		TimelineItems: []github.TimelineItem{
			{Type: "PullRequestCommit", Item: &github.PullRequestCommit{
				Commit: github.Commit{
					OID:           "new-f00bar",
					PushedDate:    now,
					CommittedDate: now,
				},
			}},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func buildBitbucketServerPR(now time.Time, c *cmpgn.Campaign, headRef string) interface{} {
	return &bitbucketserver.PullRequest{
		ID:          999,
		Title:       c.Name,
		Description: c.Description,
		State:       "OPEN",
		FromRef: bitbucketserver.Ref{
			ID: git.AbbreviateRef(headRef),
		},
	}
}

func assertChangesetInDB(t *testing.T, ctx context.Context, s *Store, id int64, want *cmpgn.Changeset) {
	t.Helper()

	changeset, err := s.GetChangeset(ctx, GetChangesetOpts{ID: id})
	if err != nil {
		t.Fatal(err)
	}

	diff := cmp.Diff(want, changeset, cmpopts.IgnoreFields(cmpgn.Changeset{}, "ID"))
	if diff != "" {
		t.Fatal(diff)
	}
}

func assertChangesetEventsInDB(t *testing.T, ctx context.Context, s *Store, changesetID int64, want []*cmpgn.ChangesetEvent) {
	t.Helper()

	events, _, err := s.ListChangesetEvents(ctx, ListChangesetEventsOpts{
		Limit:        -1,
		ChangesetIDs: []int64{changesetID},
	})
	if err != nil {
		t.Fatal(err)
	}

	diff := cmp.Diff(want, events, cmpopts.IgnoreFields(cmpgn.ChangesetEvent{}, "ID"))
	if diff != "" {
		t.Fatal(diff)
	}
}
