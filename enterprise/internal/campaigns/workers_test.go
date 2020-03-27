package campaigns

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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

	// Setup global test db in dbconn.Global
	dbtesting.SetupGlobalTestDB(t)
	// Wrap test db in transaction that's rolled back at end of test
	tx := dbtest.NewTx(t, dbconn.Global)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time { return now.UTC().Truncate(time.Microsecond) }

	s := NewStoreWithClock(tx, clock)
	// Create repositories and external service
	repo, githubExtSvc := createGitHubRepo(t, ctx, now, s)
	// Create PatchSet, Patch, Campaign and ChangesetJob
	campaign, patch := createCampaignPatch(t, ctx, now, s, repo)
	// Create dummy GitHub PR
	createdHeadRef := "refs/heads/" + campaign.Branch
	pr := githubPR(now, campaign.Name, campaign.Description, createdHeadRef)

	// Setup the dependencies
	gitClient := &dummyGitserverClient{
		response:    createdHeadRef,
		responseErr: nil,
	}

	var (
		wantHeadRef  = createdHeadRef
		wantBaseRef  = patch.BaseRef
		wantMetadata = pr
	)

	fakeSource := fakeChangesetSource{
		svc:          githubExtSvc,
		err:          nil,
		exists:       false,
		wantHeadRef:  wantHeadRef,
		wantBaseRef:  wantBaseRef,
		fakeMetadata: pr,
	}

	sourcer := repos.NewFakeSourcer(nil, fakeSource)

	// Create and execute the ChangesetJob
	changesetJob := &cmpgn.ChangesetJob{CampaignID: campaign.ID, PatchID: patch.ID}
	if err := s.CreateChangesetJob(ctx, changesetJob); err != nil {
		t.Fatal(err)
	}

	err := ExecChangesetJob(ctx, clock, s, gitClient, sourcer, campaign, changesetJob)
	if err != nil {
		t.Fatal(err)
	}

	// Reload ChangesetJob
	changesetJob, err = s.GetChangesetJob(ctx, GetChangesetJobOpts{ID: changesetJob.ID})
	if err != nil {
		t.Fatal(err)
	}

	if changesetJob.ChangesetID == 0 {
		t.Fatalf("ChangesetJob has not ChangesetID set")
	}

	// Load newly created Changeset
	changeset, err := s.GetChangeset(ctx, GetChangesetOpts{ID: changesetJob.ChangesetID})
	if err != nil {
		t.Fatal(err)
	}

	if want, have := pr.HeadRefName, changeset.ExternalBranch; have != want {
		t.Fatalf("wrong changeset.ExternalBranch. want=%s, have=%s", want, have)
	}

	haveMetadata := changeset.Metadata.(*github.PullRequest)
	if diff := cmp.Diff(wantMetadata, haveMetadata); diff != "" {
		t.Fatal(diff)
	}

	// Load newly created ChangesetEvents
	events, _, err := s.ListChangesetEvents(ctx, ListChangesetEventsOpts{
		Limit:        -1,
		ChangesetIDs: []int64{changeset.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	if want, have := 1, len(events); want != have {
		t.Fatalf("wrong number of ChangesetEvents. want=%d, have=%d", want, have)
	}

	if want, have := cmpgn.ChangesetEventKindGitHubCommit, events[0].Kind; want != have {
		t.Fatalf("wrong event. want=%s, have=%s", want, have)
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
func (s fakeChangesetSource) UpdateChangeset(ctx context.Context, c *repos.Changeset) error {
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

func githubPR(now time.Time, title, body, headRef string) *github.PullRequest {
	githubActor := github.Actor{
		AvatarURL: "https://avatars2.githubusercontent.com/u/1185253",
		Login:     "mrnugget",
		URL:       "https://github.com/mrnugget",
	}

	return &github.PullRequest{
		ID:           "FOOBARID",
		Title:        title,
		Body:         body,
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
