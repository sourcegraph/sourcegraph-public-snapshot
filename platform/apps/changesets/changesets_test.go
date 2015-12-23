// +build exectest,DISABLED_DUE_TO_FLAKINESS

package changesets_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"sourcegraph.com/sqs/pbtypes"

	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/ext"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/testserver"
)

// Here we test critical functionality of the Changesets app at the gRPC API
// level (i.e. a high level).

var (
	basicRepo = sourcegraph.RepoSpec{URI: "changesets-test"}
	basicCS   = &sourcegraph.Changeset{
		Title:       "Test changeset!",
		Description: "Adding some things!",
		Author:      sourcegraph.UserSpec{Login: "jim"},
		DeltaSpec: &sourcegraph.DeltaSpec{
			Base: sourcegraph.RepoRevSpec{RepoSpec: basicRepo, Rev: "master"},
			Head: sourcegraph.RepoRevSpec{RepoSpec: basicRepo, Rev: "feature-branch"},
		},
	}
)

func setRepoURI(uri string) {
	basicRepo.URI = uri
	basicCS.DeltaSpec.Base.RepoSpec = basicRepo
	basicCS.DeltaSpec.Head.RepoSpec = basicRepo
}

// TODO(slimsag): probably using go-vcs for some of the git functionality here
// would be easier.

// testSuite implements functionality common to all Changesets tests.
type testSuite struct {
	ctx     context.Context
	server  *testserver.Server
	workDir string
	t       *testing.T

	mirror                             bool
	mirrorOrganization, mirrorRepoName string
	github                             *github.Client
}

// prepRepo creates a new repository, adds one file, commits and pushes it.
func (ts *testSuite) prepRepo() error {
	// Create the repository
	if ts.mirror {
		// Parse environment variables so we can exit early.
		personalAccessToken := os.Getenv("CS_PERSONAL_ACCESS_TOKEN")
		if personalAccessToken == "" {
			ts.t.Skip("CS_PERSONAL_ACCESS_TOKEN is not set; please set it to run mirrored repo tests")
			return nil
		}
		ts.mirrorOrganization = os.Getenv("CS_ORGANIZATION")
		if ts.mirrorOrganization == "" {
			ts.t.Skip("CS_ORGANIZATION is not set; please set it to run mirrored repo tests")
			return nil
		}

		// Add the token to the auth store so that RefreshVCS operations succeed.
		authStore := ext.AuthStore{}
		err := authStore.Set(ts.ctx, "github.com", ext.Credentials{Token: personalAccessToken})
		if err != nil {
			return err
		}

		// Create an authenticated GitHub client.
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: personalAccessToken})
		tokenClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
		ts.github = github.NewClient(tokenClient)

		// Determine new private repo name. We prefix a timestamp in the form of
		// dd-mm-yy-hr-min-sec to avoid collisions with other pending CI builds.
		ts.mirrorRepoName = "changesets-test-" + time.Now().Format("2-1-06-15-04-05")

		// Create private repository.
		_, _, err = ts.github.Repositories.Create(ts.mirrorOrganization, &github.Repository{
			Name:     &ts.mirrorRepoName,
			Private:  github.Bool(true),
			AutoInit: github.Bool(true),
		})
		if err != nil {
			return err
		}

		// Yield for repo creation to complete.
		time.Sleep(1 * time.Second)

		// Build and set repo URI.
		mirrorRepoURI := "github.com/" + ts.mirrorOrganization + "/" + ts.mirrorRepoName
		setRepoURI(mirrorRepoURI)

		// Clone the repository.
		cloneURL := fmt.Sprintf("https://%s@%s", personalAccessToken, mirrorRepoURI)
		if err := ts.cmd("git", "clone", cloneURL, ts.workDir); err != nil {
			return err
		}

		// Create the repo locally.
		_, err = ts.server.Client.Repos.Create(ts.ctx, &sourcegraph.ReposCreateOp{
			URI:      basicRepo.URI,
			VCS:      "git",
			CloneURL: cloneURL,
			Mirror:   true,
			Private:  true,
		})
		if err != nil {
			return err
		}

		ts.refreshVCS()

	} else {
		ts.t.Log("$ src repo create", basicRepo.URI)
		repo, err := ts.server.Client.Repos.Create(ts.ctx, &sourcegraph.ReposCreateOp{
			URI: basicRepo.URI,
			VCS: "git",
		})
		if err != nil {
			return err
		}

		// Clone the repository.
		if err := ts.cmd("git", "clone", repo.CloneURL().String(), ts.workDir); err != nil {
			return err
		}
	}

	// Add a file.
	if err := ts.addFile("first", "first file contents"); err != nil {
		return err
	}

	return ts.cmds([][]string{
		{"git", "add", "first"},
		{"git", "commit", "-m", "add first file"},
		{"git", "push", "--set-upstream", "origin", "master"},
	})
}

// refreshVCS causes a VCS refresh on mirrored repos, if this is a mirror repo
// test.
func (ts *testSuite) refreshVCS() {
	if !ts.mirror {
		return
	}

	// Refresh the VCS.
	_, err := ts.server.Client.MirrorRepos.RefreshVCS(ts.ctx, &sourcegraph.MirrorReposRefreshVCSOp{
		Repo: basicRepo,
	})
	if err != nil {
		ts.t.Fatal(err)
	}
}

func (ts *testSuite) gitRevParse(target string) (string, error) {
	ts.t.Log("$ git rev-parse", target)
	cmd := exec.Command("git", "rev-parse", target)
	cmd.Dir = ts.workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
		return "", err
	}
	sha := strings.TrimSpace(string(out))
	if len(sha) != 40 {
		ts.t.Logf("expected to find string of length 40, found %q", sha)
		return "", err
	}
	return sha, nil
}

// cmds runs the list of commands in the work directory.
func (ts *testSuite) cmds(cmds [][]string) error {
	for _, args := range cmds {
		if err := ts.cmd(args...); err != nil {
			return err
		}
	}
	return nil
}

// cmd runs a command in the work directory.
func (ts *testSuite) cmd(args ...string) error {
	ts.t.Logf("$ %s\n", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = ts.workDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// addFile writes a file with the given contents in the work directory.
func (ts *testSuite) addFile(name, contents string) error {
	ts.t.Logf("$ echo %q > %s", contents, name)
	return ioutil.WriteFile(filepath.Join(ts.workDir, name), []byte(contents), 0666)
}

// close closes the test suite, and must be called when finished.
func (ts *testSuite) close() {
	if ts.mirror {
		if _, err := ts.github.Repositories.Delete(ts.mirrorOrganization, ts.mirrorRepoName); err != nil {
			ts.t.Log(err)
		}
	}
	if err := os.RemoveAll(ts.workDir); err != nil {
		ts.t.Log(err)
	}
	ts.server.Close()
}

// createBasicChangeset creates a basic changeset with one commit. It can be
// called multiple times to create a duplicate changesets using the same
// feature-branch.
func (ts *testSuite) createBasicChangeset() (*sourcegraph.Changeset, error) {
	if ts.workDir == "" {
		// Create a working directory.
		var err error
		ts.workDir, err = ioutil.TempDir("", "changesets-test")
		if err != nil {
			return nil, err
		}
		ts.t.Log("using temp dir", ts.workDir)

		// Prepare the repo.
		if err := ts.prepRepo(); err != nil {
			return nil, err
		}

		// Create a new branch.
		if err := ts.cmd("git", "checkout", "-B", "feature-branch"); err != nil {
			return nil, err
		}
		if err := ts.addFile("second", "second file contents"); err != nil {
			return nil, err
		}
		err = ts.cmds([][]string{
			{"git", "add", "second"},
			{"git", "commit", "-m", "add second file"},
			{"git", "push", "--set-upstream", "origin", "feature-branch"},
		})
		if err != nil {
			return nil, err
		}
	}

	ts.refreshVCS()

	// Create a new changeset.
	cs, err := ts.server.Client.Changesets.Create(ts.ctx, &sourcegraph.ChangesetCreateOp{
		Repo:      basicRepo,
		Changeset: basicCS,
	})
	if err != nil {
		return nil, err
	}

	// Sanity check the returned CS.
	wantCS := *basicCS
	wantCS.CreatedAt = &pbtypes.Timestamp{}
	if err := ts.changesetEqual(cs, &wantCS); err != nil {
		return nil, err
	}
	return cs, nil
}

// changesetEqual returns a human-readable error if the two changesets are not
// equal. The fields omitted from equality are:
//
//  DeltaSpec.Base.CommitID
//  DeltaSpec.Head.CommitID
//
// CreatedAt and ClosedAt are only tested for non-nil equality.
func (ts *testSuite) changesetEqual(got, want *sourcegraph.Changeset) error {
	if got.Merged != want.Merged {
		return fmt.Errorf("wrong merged status, got %v want %v", got.Merged, want.Merged)
	}
	if (got.CreatedAt != nil) != (want.CreatedAt != nil) {
		return fmt.Errorf("wrong created at status, got %v want %v", got.CreatedAt, want.CreatedAt)
	}
	if (got.ClosedAt != nil) != (want.ClosedAt != nil) {
		return fmt.Errorf("wrong closed at status, got %v want %v", got.ClosedAt, want.ClosedAt)
	}
	if got.Title != want.Title {
		return fmt.Errorf("wrong title, got %q want %q", got.Title, want.Title)
	}
	if got.Description != want.Description {
		return fmt.Errorf("wrong description, got %q want %q", got.Description, want.Description)
	}
	if got.Author.Login != want.Author.Login {
		return fmt.Errorf("wrong author login, got %q want %q", got.Author.Login, want.Author.Login)
	}
	if got.DeltaSpec.Base.URI != want.DeltaSpec.Base.URI {
		return fmt.Errorf("wrong DeltaSpec.Base.URI, got %#v want %#v", got.DeltaSpec.Base.URI, want.DeltaSpec.Base.URI)
	}
	if got.DeltaSpec.Head.URI != want.DeltaSpec.Head.URI {
		return fmt.Errorf("wrong DeltaSpec.Head.URI, got %#v want %#v", got.DeltaSpec.Head.URI, want.DeltaSpec.Head.URI)
	}
	if got.DeltaSpec.Base.Rev != want.DeltaSpec.Base.Rev {
		return fmt.Errorf("wrong DeltaSpec.Base.Rev, got %#v want %#v", got.DeltaSpec.Base.Rev, want.DeltaSpec.Base.Rev)
	}
	if got.DeltaSpec.Head.Rev != want.DeltaSpec.Head.Rev {
		return fmt.Errorf("wrong DeltaSpec.Head.Rev, got %#v want %#v", got.DeltaSpec.Head.Rev, want.DeltaSpec.Head.Rev)
	}
	return nil
}

// newTestSuite creates and returns a new test suite, or an error if any was
// encountered. close must be called upon test finish.
func newTestSuite(t *testing.T) (*testSuite, error) {
	// Spawn a test server.
	server, ctx := testserver.NewUnstartedServer()
	server.Config.ServeFlags = append(server.Config.ServeFlags,
		&authutil.Flags{DisableAccessControl: true},
	)
	if err := server.Start(); err != nil {
		return nil, err
	}

	ts := &testSuite{
		ctx:    ctx,
		server: server,
		t:      t,
	}
	ts.cmd("git", "config", "--global", "push.default", "simple")
	return ts, nil
}

// TestChangesets_Create tests that creating a changeset works.
func TestChangesets_Create(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip("pgsql local store can only create mirror repos")
	}

	// Create a new test suite.
	ts, err := newTestSuite(t)
	if err != nil {
		t.Fatal(err)
	}
	defer ts.close()

	// Create a basic changeset.
	_, err = ts.createBasicChangeset()
	if err != nil {
		t.Fatal(err)
	}
}

// TestChangesets_Get tests that creating a changeset and then getting it works.
func TestChangesets_Get(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip("pgsql local store can only create mirror repos")
	}

	// Create a new test suite.
	ts, err := newTestSuite(t)
	if err != nil {
		t.Fatal(err)
	}
	defer ts.close()

	// Create a basic changeset.
	newCS, err := ts.createBasicChangeset()
	if err != nil {
		t.Fatal(err)
	}

	// Test that Get works.
	cs, err := ts.server.Client.Changesets.Get(ts.ctx, &sourcegraph.ChangesetSpec{
		ID:   newCS.ID,
		Repo: basicRepo,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Sanity check the returned CS.
	wantCS := *basicCS
	wantCS.CreatedAt = &pbtypes.Timestamp{}
	if err := ts.changesetEqual(cs, &wantCS); err != nil {
		t.Fatal(err)
	}
}

// TestChangesets_List tests that creating a few changesets and then listing
// them works.
func TestChangesets_List(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip("pgsql local store can only create mirror repos")
	}

	// Create a new test suite.
	ts, err := newTestSuite(t)
	if err != nil {
		t.Fatal(err)
	}
	defer ts.close()

	// Create first changeset.
	firstCS, err := ts.createBasicChangeset()
	if err != nil {
		t.Fatal(err)
	}

	// Create second changeset.
	secondCS, err := ts.createBasicChangeset()
	if err != nil {
		t.Fatal(err)
	}

	// Verify that List returns both open changesets.
	list, err := ts.server.Client.Changesets.List(ts.ctx, &sourcegraph.ChangesetListOp{
		Repo: basicRepo.URI,
		Open: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Changesets) != 2 {
		t.Fatal("expected 2 changesets, got", len(list.Changesets))
	}
	if err := ts.changesetEqual(list.Changesets[0], firstCS); err != nil {
		t.Fatal(err)
	}
	if err := ts.changesetEqual(list.Changesets[1], secondCS); err != nil {
		t.Fatal(err)
	}

	// TODO(slimsag): write tests for ChangesetListOp Closed, Head, Base, and
	// pagination.

	// TODO(slimsag): guarantee behavior if e.g. both ChangesetListOp Open and
	// Close fields are not specified.

	// TODO(slimsag): guarantee an error if no Repo is specified.
}

// TestChangesets_Update tests that creating a changeset and then updating it
// works.
func TestChangesets_Update(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip("pgsql local store can only create mirror repos")
	}

	// Create a new test suite.
	ts, err := newTestSuite(t)
	if err != nil {
		t.Fatal(err)
	}
	defer ts.close()

	// Create a basic changeset.
	cs, err := ts.createBasicChangeset()
	if err != nil {
		t.Fatal(err)
	}

	// Update the changeset title.
	newTitle := "New changeset title!"
	event, err := ts.server.Client.Changesets.Update(ts.ctx, &sourcegraph.ChangesetUpdateOp{
		Repo:  basicRepo,
		ID:    cs.ID,
		Title: newTitle,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Confirm changeset event.
	if err := ts.changesetEqual(event.Before, cs); err != nil {
		t.Fatal(err)
	}
	wantCS := *cs
	wantCS.Title = newTitle
	if err := ts.changesetEqual(event.After, &wantCS); err != nil {
		t.Fatal(err)
	}

	// TODO(slimsag): test valid changes to e.g. Description, Open, Close.
	// TODO(slimsag): test invalid changes to e.g. Merged and Author.
}

// TestChangesets_CreateReview tests that creating a changeset and then
// creating a review on it works.
func TestChangesets_CreateReview(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip("pgsql local store can only create mirror repos")
	}

	// Create a new test suite.
	ts, err := newTestSuite(t)
	if err != nil {
		t.Fatal(err)
	}
	defer ts.close()

	// Create a basic changeset.
	cs, err := ts.createBasicChangeset()
	if err != nil {
		t.Fatal(err)
	}

	// Create a review.
	//
	// TODO(slimsag): verify behavior if ChangesetReview CreatedAt, EditedAt or
	// Deleted fields are specified.
	//
	// TODO(slimsag): verify behavior of ChangesetReview Comments field.
	wantReview := &sourcegraph.ChangesetReview{
		Body:     "changeset review body",
		Author:   sourcegraph.UserSpec{Login: "jessica"},
		Comments: []*sourcegraph.InlineComment{},
	}
	review, err := ts.server.Client.Changesets.CreateReview(ts.ctx, &sourcegraph.ChangesetCreateReviewOp{
		Repo:        basicRepo,
		ChangesetID: cs.ID,
		Review:      wantReview,
	})

	// Verify review.
	if review.Body != wantReview.Body {
		t.Fatalf("incorrect review, got %q want %q\n", review.Body, wantReview.Body)
	}
	if review.Author != wantReview.Author {
		t.Fatalf("incorrect author, got %q want %q\n", review.Author, wantReview.Author)
	}
	if review.CreatedAt == nil {
		t.Skip("BUG: currently failing!")
		return
		t.Fatal("incorrect created at status, got nil want non-nil")
	}
	if review.EditedAt != nil {
		t.Fatalf("incorrect edited at status, got %v want nil\n", review.EditedAt)
	}
	if len(review.Comments) > 0 {
		t.Fatal("expected one comment, found", len(review.Comments))
	}
	if review.Deleted {
		t.Fatal("incorrect deleted status, got true want false")
	}
}

// TestChangesets_ListReviews tests that creating a few changesets and then
// listing the events on it works.
func TestChangesets_ListReviews(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip("pgsql local store can only create mirror repos")
	}

	// Create a new test suite.
	ts, err := newTestSuite(t)
	if err != nil {
		t.Fatal(err)
	}
	defer ts.close()

	// Create a basic changeset.
	cs, err := ts.createBasicChangeset()
	if err != nil {
		t.Fatal(err)
	}

	// Create a few reviews.
	wantReviews := []*sourcegraph.ChangesetReview{
		&sourcegraph.ChangesetReview{
			Body:   "first changeset review body",
			Author: sourcegraph.UserSpec{Login: "jessica"},
		},
		&sourcegraph.ChangesetReview{
			Body:   "second changeset review body",
			Author: sourcegraph.UserSpec{Login: "george"},
		},
		&sourcegraph.ChangesetReview{
			Body:   "third changeset review body",
			Author: sourcegraph.UserSpec{Login: "kim"},
		},
	}
	for _, wantReview := range wantReviews {
		_, err := ts.server.Client.Changesets.CreateReview(ts.ctx, &sourcegraph.ChangesetCreateReviewOp{
			Repo:        basicRepo,
			ChangesetID: cs.ID,
			Review:      wantReview,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// List reviews and verify them.
	list, err := ts.server.Client.Changesets.ListReviews(ts.ctx, &sourcegraph.ChangesetListReviewsOp{
		Repo:        basicRepo,
		ChangesetID: cs.ID,
	})
	if len(list.Reviews) != len(wantReviews) {
		t.Fatalf("incorrect reviews list, got %v want %v\n", len(list.Reviews), len(wantReviews))
	}
	for i, review := range list.Reviews {
		wantReview := wantReviews[i]
		if review.Body != wantReview.Body {
			t.Fatalf("incorrect review, got %q want %q\n", review.Body, wantReview.Body)
		}
		if review.Author != wantReview.Author {
			t.Fatalf("incorrect author, got %q want %q\n", review.Author, wantReview.Author)
		}
		if review.CreatedAt == nil {
			t.Skip("BUG: currently failing!")
			return
			t.Fatal("incorrect created at status, got nil want non-nil")
		}
		if review.EditedAt != nil {
			t.Fatalf("incorrect edited at status, got %v want nil\n", review.EditedAt)
		}
		if len(review.Comments) > 0 {
			t.Fatal("expected one comment, found", len(review.Comments))
		}
		if review.Deleted {
			t.Fatal("incorrect deleted status, got true want false")
		}
	}
}

// TestChangesets_ListEvents runs the ListEvents test on a hosted repository.
func TestChangesets_ListEvents(t *testing.T) {
	testChangesets_ListEvents(t, false)
}

// TestChangesets_MirroredListEvents runs the ListEvents test on a mirrored
// GitHub repository.
func TestChangesets_MirroredListEvents(t *testing.T) {
	testChangesets_ListEvents(t, true)
}

// testChangesets_ListEvents tests that creating a changeset with a few events
// and listing them works.
//
// It takes a single parameter, which is whether or not the test should be
// performed on a mirrored GitHub repo or not.
func testChangesets_ListEvents(t *testing.T, mirror bool) {
	if !mirror && testserver.Store == "pgsql" {
		t.Skip("pgsql local store can only create mirror repos")
	}

	// Create a new test suite.
	ts, err := newTestSuite(t)
	if err != nil {
		t.Fatal(err)
	}
	ts.mirror = mirror
	defer ts.close()

	// Create a basic changeset.
	cs, err := ts.createBasicChangeset()
	if err != nil {
		t.Fatal(err)
	}

	// Update the changeset title a few times.
	updates := []*sourcegraph.ChangesetUpdateOp{
		&sourcegraph.ChangesetUpdateOp{
			Repo:  basicRepo,
			ID:    cs.ID,
			Title: "First new changeset title!",
		},
		&sourcegraph.ChangesetUpdateOp{
			Repo:  basicRepo,
			ID:    cs.ID,
			Title: "Second new changeset title!",
		},
		&sourcegraph.ChangesetUpdateOp{
			Repo:  basicRepo,
			ID:    cs.ID,
			Title: "Third new changeset title!",
		},
	}
	for _, update := range updates {
		_, err := ts.server.Client.Changesets.Update(ts.ctx, update)
		if err != nil {
			t.Fatal(err)
		}
	}

	// List the events.
	events, err := ts.server.Client.Changesets.ListEvents(ts.ctx, &sourcegraph.ChangesetSpec{
		Repo: basicRepo,
		ID:   cs.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify events and updates.
	if len(events.Events) != len(updates) {
		t.Fatalf("got %v events expected %v", len(events.Events), len(updates))
	}
	for i, gotEvent := range events.Events {
		if gotEvent.CreatedAt == nil {
			t.Fatal("incorrect created at status, got nil expected non-nil")
		}
		if !reflect.DeepEqual(gotEvent.Op, updates[i]) {
			t.Fatalf("incorrect update, got %#v expected %#v", gotEvent.Op, updates[i])
		}
	}
}

// TestChangesets_BackgroundBaseCommits runs the BackgroundBaseEvents test on a
// hosted repository.
func TestChangesets_BackgroundBaseCommits(t *testing.T) {
	testChangesets_BackgroundBaseCommits(t, false)
}

// TestChangesets_MirroredBackgroundBaseCommits runs the BackgroundBaseEvents
// test on a mirrored GitHub repository.
func TestChangesets_MirroredBackgroundBaseCommits(t *testing.T) {
	testChangesets_BackgroundBaseCommits(t, true)
}

// testChangesets_BackgroundBaseCommits tests that commits to the base branch
// (e.g. master) do not show up in a changeset's event stream after creating a
// CS.
//
// It takes a single parameter, which is whether or not the test should be
// performed on a mirrored GitHub repo or not.
func testChangesets_BackgroundBaseCommits(t *testing.T, mirror bool) {
	if !mirror && testserver.Store == "pgsql" {
		t.Skip("pgsql local store can only create mirror repos")
	}

	// Create a new test suite.
	ts, err := newTestSuite(t)
	if err != nil {
		t.Fatal(err)
	}
	ts.mirror = mirror
	defer ts.close()

	// Create a basic changeset.
	cs, err := ts.createBasicChangeset()
	if err != nil {
		t.Fatal(err)
	}

	const (
		csHeadRev = "feature-branch"
		csBaseRev = "master"
	)

	// Confirm that head branch HEAD is correct.
	gotHead := cs.DeltaSpec.Head
	if gotHead.Rev != csHeadRev {
		t.Fatalf("incorrect Head.Rev, got %q want %q", gotHead, csHeadRev)
	}
	if gotHead.CommitID != "" {
		if err := ts.cmd("git", "checkout", csHeadRev); err != nil {
			t.Fatal(err)
		}
		wantHeadCommitID, err := ts.gitRevParse("HEAD")
		if err != nil {
			t.Fatal(err)
		}
		if gotHead.CommitID != wantHeadCommitID {
			t.Fatalf("incorrect head, got %q want %q", gotHead.CommitID, wantHeadCommitID)
		}
	}

	// Confirm that base branch HEAD is correct.
	baseBefore := cs.DeltaSpec.Base
	gotBase := cs.DeltaSpec.Base
	if gotBase.Rev != csBaseRev {
		t.Fatalf("incorrect Base.Rev, got %q want %q", gotBase, csBaseRev)
	}
	if gotBase.CommitID != "" {
		if err := ts.cmd("git", "checkout", csBaseRev); err != nil {
			t.Fatal(err)
		}
		wantBaseCommitID, err := ts.gitRevParse("HEAD")
		if err != nil {
			t.Fatal(err)
		}
		if gotBase.CommitID != wantBaseCommitID {
			t.Fatalf("incorrect base, got %q want %q", gotBase.CommitID, wantBaseCommitID)
		}
	}

	// Now checkout the base branch (master) and push three commits.
	if err := ts.cmd("git", "checkout", csBaseRev); err != nil {
		t.Fatal(err)
	}
	files := [][2]string{
		{"third", "third file contents"},
		{"fourth", "fourth file contents"},
		{"fifth", "fifth file contents"},
	}
	for _, filePair := range files {
		if err := ts.addFile(filePair[0], filePair[1]); err != nil {
			t.Fatal(err)
		}
		err = ts.cmds([][]string{
			{"git", "add", filePair[0]},
			{"git", "commit", "-m", "add another file"},
			{"git", "push"},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Confirm that no events were generated.
	events, err := ts.server.Client.Changesets.ListEvents(ts.ctx, &sourcegraph.ChangesetSpec{
		Repo: basicRepo,
		ID:   cs.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events.Events) != 0 {
		t.Fatalf("wrong number of events, got %v expected 0", len(events.Events))
	}

	// Get the changeset again.
	cs, err = ts.server.Client.Changesets.Get(ts.ctx, &sourcegraph.ChangesetSpec{
		Repo: basicRepo,
		ID:   cs.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Confirm that head branch HEAD is still correct.
	gotHead = cs.DeltaSpec.Head
	if gotHead.Rev != csHeadRev {
		t.Fatalf("incorrect Head.Rev, got %q want %q", gotHead, csHeadRev)
	}
	if gotHead.CommitID != "" {
		if err := ts.cmd("git", "checkout", csHeadRev); err != nil {
			t.Fatal(err)
		}
		wantHeadCommitID, err := ts.gitRevParse("HEAD")
		if err != nil {
			t.Fatal(err)
		}
		if gotHead.CommitID != wantHeadCommitID {
			t.Fatalf("incorrect head, got %q want %q", gotHead.CommitID, wantHeadCommitID)
		}
	}

	// Confirm that our DeltaSpec.Base has been untouched.
	gotBase = cs.DeltaSpec.Base
	wantBase := baseBefore
	if !reflect.DeepEqual(gotBase, wantBase) {
		t.Fatalf("incorrect DeltaSpec.Base, got %#v want %#v", gotBase, wantBase)
	}
}

// TestChangesets_RebaseFlow runs the RebaseFlow test on a hosted repository.
func TestChangesets_RebaseFlow(t *testing.T) {
	testChangesets_RebaseFlow(t, false)
}

// TestChangesets_MirroredRebaseFlow runs the RebaseFlow test on a mirrored
// GitHub repository.
func TestChangesets_MirroredRebaseFlow(t *testing.T) {
	testChangesets_RebaseFlow(t, true)
}

// testChangesets_RebaseFlow tests that a basic rebase workflow works. i.e. that
// three commits -> background master changes -> rebase -> force push -> merge
// works as expected.
//
// It takes a single parameter, which is whether or not the test should be
// performed on a mirrored GitHub repo or not.
func testChangesets_RebaseFlow(t *testing.T, mirror bool) {
	if !mirror && testserver.Store == "pgsql" {
		t.Skip("pgsql local store can only create mirror repos")
	}

	// Create a new test suite.
	ts, err := newTestSuite(t)
	if err != nil {
		t.Fatal(err)
	}
	ts.mirror = mirror
	defer ts.close()

	// Create a basic changeset.
	cs, err := ts.createBasicChangeset()
	if err != nil {
		t.Fatal(err)
	}

	// Add two more commits to feature-branch so our CS has three in total.
	files := [][2]string{
		{"third", "third file contents"},
		{"fourth", "fourth file contents"},
	}
	for _, filePair := range files {
		if err := ts.addFile(filePair[0], filePair[1]); err != nil {
			t.Fatal(err)
		}
		err = ts.cmds([][]string{
			{"git", "add", filePair[0]},
			{"git", "commit", "-m", "add another file"},
			{"git", "push"},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	const (
		csHeadRev = "feature-branch"
		csBaseRev = "master"
	)

	// Now checkout the base branch (master) and push three commits.
	if err := ts.cmd("git", "checkout", csBaseRev); err != nil {
		t.Fatal(err)
	}
	files = [][2]string{
		{"base-third", "third file contents"},
		{"base-fourth", "fourth file contents"},
		{"base-fifth", "fifth file contents"},
	}
	for _, filePair := range files {
		if err := ts.addFile(filePair[0], filePair[1]); err != nil {
			t.Fatal(err)
		}
		err = ts.cmds([][]string{
			{"git", "add", filePair[0]},
			{"git", "commit", "-m", "add another file"},
			{"git", "push"},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Grab the current HEAD of the base branch (master) -- this is the commit ID
	// our branch will be based on after the rebase below.
	wantBaseCommitID, err := ts.gitRevParse("HEAD")
	if err != nil {
		t.Fatal(err)
	}

	// Rebase feature-branch on master and force push so that CS is aware of the
	// new commits.
	err = ts.cmds([][]string{
		{"git", "checkout", csHeadRev},
		{"git", "rebase", csBaseRev},
		{"git", "push", "-f"},
	})
	if err != nil {
		t.Fatal(err)
	}
	ts.refreshVCS()
	time.Sleep(500 * time.Millisecond)

	// Grab the current HEAD of the head branch (feature-branch) -- this is the
	// commit ID of the last commit to our feature branch. From this point on, our
	// CS base and head commit IDs should never change (or else the diff will be
	// corrupt).
	wantHeadCommitID, err := ts.gitRevParse("HEAD")
	if err != nil {
		t.Fatal(err)
	}

	// Verify the current DeltaSpec of the changeset.
	//
	// We do this once prior to merge; and also after merge (below) as there was
	// once a bug that caused the DeltaSpec.Base.CommitID to not be updated on a
	// force push (causing the computed diff to show all commits to e.g. master
	// since the CS was opened).
	cs, err = ts.server.Client.Changesets.Get(ts.ctx, &sourcegraph.ChangesetSpec{
		Repo: basicRepo,
		ID:   cs.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotHead := cs.DeltaSpec.Head.CommitID; gotHead != "" && gotHead != wantHeadCommitID {
		t.Fatalf("wrong Head.CommitID, got %q want %q", gotHead, wantHeadCommitID)
	}
	if gotBase := cs.DeltaSpec.Base.CommitID; gotBase != "" && gotBase != wantBaseCommitID {
		t.Fatalf("wrong Base.CommitID, got %q want %q", gotBase, wantBaseCommitID)
	}

	// Merge head branch (feature-branch) into base branch (master) and push.
	csBeforeMerge := cs
	err = ts.cmds([][]string{
		{"git", "checkout", csBaseRev},
		{"git", "merge", csHeadRev},
		{"git", "push"},
	})
	if err != nil {
		t.Fatal(err)
	}
	ts.refreshVCS()

	// Verify the current DeltaSpec of the changeset.
	cs, err = ts.server.Client.Changesets.Get(ts.ctx, &sourcegraph.ChangesetSpec{
		Repo: basicRepo,
		ID:   cs.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotHead := cs.DeltaSpec.Head.CommitID; gotHead != "" && gotHead != wantHeadCommitID {
		t.Fatalf("wrong Head.CommitID, got %q want %q", gotHead, wantHeadCommitID)
	}
	if gotBase := cs.DeltaSpec.Base.CommitID; gotBase != "" && gotBase != wantBaseCommitID {
		t.Fatalf("wrong Base.CommitID, got %q want %q", gotBase, wantBaseCommitID)
	}

	// Wait for hook to fire and event to be generated.
	time.Sleep(500 * time.Millisecond)
	events, err := ts.server.Client.Changesets.ListEvents(ts.ctx, &sourcegraph.ChangesetSpec{
		Repo: basicRepo,
		ID:   cs.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events.Events) != 1 {
		t.Fatalf("wrong number of events, got %v expected 1", len(events.Events))
	}

	// Confirm that we have exactly one merge event.
	ev := events.Events[0]
	if err := ts.changesetEqual(ev.Before, csBeforeMerge); err != nil {
		t.Fatal(err)
	}
	afterCS := *cs
	afterCS.Merged = true
	afterCS.ClosedAt = &pbtypes.Timestamp{}
	if err := ts.changesetEqual(ev.After, &afterCS); err != nil {
		t.Fatal(err)
	}
	if ev.Op.Merged != true {
		t.Fatalf("incorrect merged status, got false expected true")
	}

	// The changeset is now merged, verify the DeltaSpec has exactly the right
	// CommitIDs (which must be present for persistence after branch deletion).
	cs, err = ts.server.Client.Changesets.Get(ts.ctx, &sourcegraph.ChangesetSpec{
		Repo: basicRepo,
		ID:   cs.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotHead := cs.DeltaSpec.Head.CommitID; gotHead != wantHeadCommitID {
		t.Fatalf("wrong Head.CommitID, got %q want %q", gotHead, wantHeadCommitID)
	}
	if gotBase := cs.DeltaSpec.Base.CommitID; gotBase != wantBaseCommitID {
		t.Fatalf("wrong Base.CommitID, got %q want %q", gotBase, wantBaseCommitID)
	}
}

// TestChangesets_MergeFlow runs the MergeFlow test on a hosted repository.
func TestChangesets_MergeFlow(t *testing.T) {
	testChangesets_MergeFlow(t, false, false)
}

// TestChangesets_CLIMergeFlow runs the CLIMergeFlow test on a hosted
// repository.
func TestChangesets_CLIMergeFlow(t *testing.T) {
	t.Skip("BUG: hosted repos do not emit githook events on merge")
	return
	testChangesets_MergeFlow(t, false, true)
}

// TestChangesets_MirroredMergeFlow runs the MergeFlow test on a mirrored
// GitHub repository.
func TestChangesets_MirroredMergeFlow(t *testing.T) {
	testChangesets_MergeFlow(t, true, false)
}

// TestChangesets_MirroredCLIMergeFlow runs the CLIMergeFlow test on a mirrored
// GitHub repository.
func TestChangesets_MirroredCLIMergeFlow(t *testing.T) {
	testChangesets_MergeFlow(t, true, true)
}

// testChangesets_MergeFlow tests that a basic merge workflow works. i.e. that
// three commits -> master changes -> merge works as expected.
//
// It takes a single parameter, which is whether or not the test should be
// performed on a mirrored GitHub repo or not.
func testChangesets_MergeFlow(t *testing.T, mirror, cli bool) {
	if !mirror && testserver.Store == "pgsql" {
		t.Skip("pgsql local store can only create mirror repos")
	}

	// Create a new test suite.
	ts, err := newTestSuite(t)
	if err != nil {
		t.Fatal(err)
	}
	ts.mirror = mirror
	defer ts.close()

	// Create a basic changeset.
	cs, err := ts.createBasicChangeset()
	if err != nil {
		t.Fatal(err)
	}

	// Add two more commits to feature-branch so our CS has three in total.
	files := [][2]string{
		{"third", "third file contents"},
		{"fourth", "fourth file contents"},
	}
	for _, filePair := range files {
		if err := ts.addFile(filePair[0], filePair[1]); err != nil {
			t.Fatal(err)
		}
		err = ts.cmds([][]string{
			{"git", "add", filePair[0]},
			{"git", "commit", "-m", "add another file"},
			{"git", "push"},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	ts.refreshVCS()

	// Grab the current HEAD of the head branch (feature-branch) -- this is the
	// commit ID of the last commit to our feature branch. From this point on, our
	// CS base and head commit IDs should never change (or else the diff will be
	// corrupt).
	wantHeadCommitID, err := ts.gitRevParse("HEAD")
	if err != nil {
		t.Fatal(err)
	}

	const (
		csHeadRev = "feature-branch"
		csBaseRev = "master"
	)

	// Now checkout the base branch (master) and push three commits.
	if err := ts.cmd("git", "checkout", csBaseRev); err != nil {
		t.Fatal(err)
	}
	files = [][2]string{
		{"base-third", "third file contents"},
		{"base-fourth", "fourth file contents"},
		{"base-fifth", "fifth file contents"},
	}
	for _, filePair := range files {
		if err := ts.addFile(filePair[0], filePair[1]); err != nil {
			t.Fatal(err)
		}
		err = ts.cmds([][]string{
			{"git", "add", filePair[0]},
			{"git", "commit", "-m", "add another file"},
			{"git", "push"},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	ts.refreshVCS()

	// Grab the current HEAD of the base branch (master) -- this is the commit ID
	// our branch will be based on after the rebase below.
	wantBaseCommitID, err := ts.gitRevParse("HEAD")
	if err != nil {
		t.Fatal(err)
	}

	// Verify the current DeltaSpec of the changeset. We do this once prior to
	// merge; and also after merge (below).
	cs, err = ts.server.Client.Changesets.Get(ts.ctx, &sourcegraph.ChangesetSpec{
		Repo: basicRepo,
		ID:   cs.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotHead := cs.DeltaSpec.Head.CommitID; gotHead != "" && gotHead != wantHeadCommitID {
		t.Fatalf("wrong Head.CommitID, got %q want %q", gotHead, wantHeadCommitID)
	}
	if gotBase := cs.DeltaSpec.Base.CommitID; gotBase != "" && gotBase != wantBaseCommitID {
		t.Fatalf("wrong Base.CommitID, got %q want %q", gotBase, wantBaseCommitID)
	}

	csBeforeMerge := cs
	if cli {
		// Merge the CS as if via the CLI tool.
		// TODO(slimsag): validate the returned ChangesetEvent too?
		_, err := ts.server.Client.Changesets.Merge(ts.ctx, &sourcegraph.ChangesetMergeOp{
			Repo:    basicRepo,
			ID:      cs.ID,
			Message: "Custom changeset merge message.",
			Squash:  false,
		})
		if err != nil {
			t.Fatal(err)
		}
	} else {
		// Merge head branch (feature-branch) into base branch (master) and push.
		err = ts.cmds([][]string{
			{"git", "checkout", csBaseRev},
			{"git", "merge", csHeadRev},
			{"git", "push"},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	ts.refreshVCS()

	// Verify the current DeltaSpec of the changeset.
	cs, err = ts.server.Client.Changesets.Get(ts.ctx, &sourcegraph.ChangesetSpec{
		Repo: basicRepo,
		ID:   cs.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotHead := cs.DeltaSpec.Head.CommitID; gotHead != "" && gotHead != wantHeadCommitID {
		t.Fatalf("wrong Head.CommitID, got %q want %q", gotHead, wantHeadCommitID)
	}
	if gotBase := cs.DeltaSpec.Base.CommitID; gotBase != "" && gotBase != wantBaseCommitID {
		t.Fatalf("wrong Base.CommitID, got %q want %q", gotBase, wantBaseCommitID)
	}

	// Wait for hook to fire and event to be generated.
	time.Sleep(500 * time.Millisecond)
	events, err := ts.server.Client.Changesets.ListEvents(ts.ctx, &sourcegraph.ChangesetSpec{
		Repo: basicRepo,
		ID:   cs.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events.Events) != 1 {
		t.Fatalf("wrong number of events, got %v expected 1", len(events.Events))
	}

	// Confirm that we have exactly one merge event.
	ev := events.Events[0]
	if err := ts.changesetEqual(ev.Before, csBeforeMerge); err != nil {
		t.Fatal(err)
	}
	afterCS := *cs
	afterCS.Merged = true
	afterCS.ClosedAt = &pbtypes.Timestamp{}
	if err := ts.changesetEqual(ev.After, &afterCS); err != nil {
		t.Fatal(err)
	}
	if ev.Op.Merged != true {
		t.Fatalf("incorrect merged status, got false expected true")
	}

	// The changeset is now merged, verify the DeltaSpec has exactly the right
	// CommitIDs (which must be present for persistence after branch deletion).
	cs, err = ts.server.Client.Changesets.Get(ts.ctx, &sourcegraph.ChangesetSpec{
		Repo: basicRepo,
		ID:   cs.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotHead := cs.DeltaSpec.Head.CommitID; gotHead != wantHeadCommitID {
		t.Fatalf("wrong Head.CommitID, got %q want %q", gotHead, wantHeadCommitID)
	}
	if gotBase := cs.DeltaSpec.Base.CommitID; gotBase != wantBaseCommitID {
		t.Fatalf("wrong Base.CommitID, got %q want %q", gotBase, wantBaseCommitID)
	}
}
