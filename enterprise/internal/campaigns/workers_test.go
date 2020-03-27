package campaigns

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExecChangesetJob(t *testing.T) {
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time { return now.UTC().Truncate(time.Microsecond) }

	dbtesting.SetupGlobalTestDB(t)

	codehosts := []struct {
		name string

		createRepoExtSvc  func(t *testing.T, ctx context.Context, now time.Time, s *Store) (*repos.Repo, *repos.ExternalService)
		changesetMetadata func(now time.Time, c *cmpgn.Campaign, headRef string) interface{}

		wantChangeset func(now time.Time, r *repos.Repo, c *cmpgn.Campaign, headRef string, metadata interface{}) *cmpgn.Changeset
		wantEvents    func(now time.Time, changesetID int64, metadata interface{}) []*cmpgn.ChangesetEvent
	}{
		{
			name:              "GitHub",
			createRepoExtSvc:  createGitHubRepo,
			changesetMetadata: buildGithubPR,
			wantChangeset: func(now time.Time, r *repos.Repo, c *cmpgn.Campaign, headRef string, metadata interface{}) *cmpgn.Changeset {
				want := &cmpgn.Changeset{
					RepoID:              r.ID,
					CampaignIDs:         []int64{c.ID},
					ExternalBranch:      headRef,
					ExternalState:       cmpgn.ChangesetStateOpen,
					ExternalReviewState: cmpgn.ChangesetReviewStatePending,
					ExternalCheckState:  cmpgn.ChangesetCheckStateUnknown,
					CreatedAt:           now,
					UpdatedAt:           now,
				}
				want.SetMetadata(metadata)
				return want
			},
			wantEvents: func(now time.Time, changesetID int64, metadata interface{}) []*cmpgn.ChangesetEvent {
				pr := metadata.(*github.PullRequest)

				return []*cmpgn.ChangesetEvent{
					{
						ChangesetID: changesetID,
						Kind:        cmpgn.ChangesetEventKindGitHubCommit,
						Key:         pr.TimelineItems[0].Item.(*github.PullRequestCommit).Commit.OID,
						UpdatedAt:   now,
						CreatedAt:   now,
						Metadata:    pr.TimelineItems[0].Item,
					},
				}
			},
		},
	}

	for _, codehostTest := range codehosts {
		subtests := []struct {
			name          string
			alreadyExists bool
		}{
			{name: "ChangesetAlreadyExists", alreadyExists: true},
			{name: "NewChangeset", alreadyExists: false},
		}

		for _, tc := range subtests {
			t.Run(codehostTest.name+tc.name, func(t *testing.T) {
				tx := dbtest.NewTx(t, dbconn.Global)
				s := NewStoreWithClock(tx, clock)

				repo, extSvc := codehostTest.createRepoExtSvc(t, ctx, now, s)
				campaign, patch := createCampaignPatch(t, ctx, now, s, repo)

				headRef := "refs/heads/" + campaign.Branch
				baseRef := patch.BaseRef

				pr := codehostTest.changesetMetadata(now, campaign, headRef)

				gitClient := &dummyGitserverClient{response: headRef, responseErr: nil}

				sourcer := repos.NewFakeSourcer(nil, fakeChangesetSource{
					svc:          extSvc,
					err:          nil,
					exists:       tc.alreadyExists,
					wantHeadRef:  headRef,
					wantBaseRef:  baseRef,
					fakeMetadata: pr,
				})

				changesetJob := &cmpgn.ChangesetJob{CampaignID: campaign.ID, PatchID: patch.ID}
				if err := s.CreateChangesetJob(ctx, changesetJob); err != nil {
					t.Fatal(err)
				}

				err := ExecChangesetJob(ctx, clock, s, gitClient, sourcer, campaign, changesetJob)
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

				wantChangeset := codehostTest.wantChangeset(now, repo, campaign, headRef, pr)
				assertChangesetInDB(t, ctx, s, changesetJob.ChangesetID, wantChangeset)

				wantEvents := codehostTest.wantEvents(now, changesetJob.ChangesetID, pr)
				assertChangesetEventsInDB(t, ctx, s, changesetJob.ChangesetID, wantEvents)
			})
		}
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

type fakeChangesetSource struct {
	svc *repos.ExternalService

	wantHeadRef string
	wantBaseRef string

	fakeMetadata interface{}
	exists       bool
	err          error
}

func (s fakeChangesetSource) CreateChangeset(ctx context.Context, c *repos.Changeset) (bool, error) {
	if s.err != nil {
		return s.exists, s.err
	}

	if c.HeadRef != s.wantHeadRef {
		return s.exists, fmt.Errorf("wrong HeadRef. want=%s, have=%s", s.wantHeadRef, c.HeadRef)
	}

	if c.BaseRef != s.wantBaseRef {
		return s.exists, fmt.Errorf("wrong BaseRef. want=%s, have=%s", s.wantBaseRef, c.BaseRef)
	}

	c.SetMetadata(s.fakeMetadata)

	return s.exists, s.err
}

func (s fakeChangesetSource) UpdateChangeset(ctx context.Context, c *repos.Changeset) error {
	if s.err != nil {
		return s.err
	}

	if c.BaseRef != s.wantBaseRef {
		return fmt.Errorf("wrong BaseRef. want=%s, have=%s", s.wantBaseRef, c.BaseRef)
	}

	c.SetMetadata(s.fakeMetadata)
	return nil
}

var fakeNotImplemented = errors.New("not implement in fakeChangesetSource")

func (s fakeChangesetSource) ListRepos(ctx context.Context, results chan repos.SourceResult) {
	results <- repos.SourceResult{Source: s, Err: fakeNotImplemented}
}

func (s fakeChangesetSource) ExternalServices() repos.ExternalServices {
	return repos.ExternalServices{s.svc}
}
func (s fakeChangesetSource) LoadChangesets(ctx context.Context, cs ...*repos.Changeset) error {
	return fakeNotImplemented
}
func (s fakeChangesetSource) CloseChangeset(ctx context.Context, c *repos.Changeset) error {
	return fakeNotImplemented
}

func createGitHubRepo(t *testing.T, ctx context.Context, now time.Time, s *Store) (*repos.Repo, *repos.ExternalService) {
	t.Helper()

	reposStore := repos.NewDBStore(s.DB(), sql.TxOptions{})

	githubExtSvc := &repos.ExternalService{
		Kind:        "GITHUB",
		DisplayName: "GitHub",
		Config: marshalJSON(t, &schema.GitHubConnection{
			Url:   "https://github.com",
			Token: os.Getenv("GITHUB_TOKEN"),
			Repos: []string{},
		}),
	}

	if err := reposStore.UpsertExternalServices(ctx, githubExtSvc); err != nil {
		t.Fatal(t)
	}

	repo := testRepo(0, github.ServiceType)
	repo.Sources = map[string]*repos.SourceInfo{githubExtSvc.URN(): {
		ID:       githubExtSvc.URN(),
		CloneURL: "https://TOKENTOKENTOKEN@github.com/foobar/foobar",
	}}
	if err := reposStore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	return repo, githubExtSvc
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

var githubActor = github.Actor{
	AvatarURL: "https://avatars2.githubusercontent.com/u/1185253",
	Login:     "mrnugget",
	URL:       "https://github.com/mrnugget",
}

func buildGithubPR(now time.Time, c *cmpgn.Campaign, headRef string) interface{} {
	return &github.PullRequest{
		ID:           "FOOBARID",
		Title:        c.Name,
		Body:         c.Description,
		HeadRefName:  git.AbbreviateRef(headRef),
		URL:          "https://github.com/sourcegraph/sourcegraph/pull/12345",
		Number:       12345,
		State:        "OPEN",
		Author:       githubActor,
		Participants: []github.Actor{githubActor},
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
