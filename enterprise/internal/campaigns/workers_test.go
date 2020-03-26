package campaigns

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExecChangesetJob(t *testing.T) {
	ctx := context.Background()

	// Setup global test db in dbconn.Global
	dbtesting.SetupGlobalTestDB(t)
	// Wrap test db in transaction that's rolled back at end of test
	tx := dbtest.NewTx(t, dbconn.Global)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}

	// Create repositories and external service
	reposStore := repos.NewDBStore(tx, sql.TxOptions{})

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

	var rs []*repos.Repo
	for i := 0; i < 4; i++ {
		r := testRepo(i, github.ServiceType)
		r.Sources = map[string]*repos.SourceInfo{githubExtSvc.URN(): {
			ID:       githubExtSvc.URN(),
			CloneURL: "https://TOKENTOKENTOKEN@github.com/foobar/foobar",
		}}
		rs = append(rs, r)
	}
	if err := reposStore.UpsertRepos(ctx, rs...); err != nil {
		t.Fatal(err)
	}

	// Create PatchSet, Patch, Campaign and ChangesetJob
	s := NewStoreWithClock(tx, clock)

	patchSet := &cmpgn.PatchSet{}
	if err := s.CreatePatchSet(ctx, patchSet); err != nil {
		t.Fatal(err)
	}

	patch := &cmpgn.Patch{
		RepoID:     rs[0].ID,
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

	changesetJob := &cmpgn.ChangesetJob{
		CampaignID: campaign.ID,
		PatchID:    patch.ID,
	}
	if err := s.CreateChangesetJob(ctx, changesetJob); err != nil {
		t.Fatal(err)
	}

	// Setup the dependencies
	gitClient := &dummyGitserverClient{response: "refs/heads/campaigns/TEST-REF", responseErr: nil}

	githubActor := github.Actor{
		AvatarURL: "https://avatars2.githubusercontent.com/u/1185253",
		Login:     "mrnugget",
		URL:       "https://github.com/mrnugget",
	}

	githubPR := &github.PullRequest{
		ID:           "FOOBARID",
		Title:        campaign.Name,
		Body:         campaign.Description,
		URL:          "https://github.com/sourcegraph/sourcegraph/pull/12345",
		Number:       12345,
		State:        "OPEN",
		Author:       githubActor,
		Participants: []github.Actor{githubActor},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	fakeSource := fakeChangesetSource{
		svc:          githubExtSvc,
		err:          nil,
		exists:       false,
		wantHeadRef:  gitClient.response,
		wantBaseRef:  patch.BaseRef,
		fakeMetadata: githubPR,
	}

	sourcer := repos.NewFakeSourcer(nil, fakeSource)

	// Execute the ChangesetJob
	err := ExecChangesetJob(ctx, clock, s, gitClient, sourcer, campaign, changesetJob)
	if err != nil {
		t.Fatal(err)
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
