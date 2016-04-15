package local

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	vcstesting "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
	"sourcegraph.com/sourcegraph/sourcegraph/platform"
)

func TestReposService_Get(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepo := &sourcegraph.Repo{
		URI:     "r",
		HTMLURL: "http://example.com/r",
	}

	calledGet := mock.stores.Repos.MockGet(t, "r")

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{URI: "r"})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !reflect.DeepEqual(repo, wantRepo) {
		t.Errorf("got %+v, want %+v", repo, wantRepo)
	}
}

func TestReposService_List(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepos := &sourcegraph.RepoList{
		Repos: []*sourcegraph.Repo{
			{URI: "r1", HTMLURL: "http://example.com/r1"},
			{URI: "r2", HTMLURL: "http://example.com/r2"},
		},
	}

	calledList := mock.stores.Repos.MockList(t, "r1", "r2")

	repos, err := s.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !*calledList {
		t.Error("!calledList")
	}
	if !reflect.DeepEqual(repos, wantRepos) {
		t.Errorf("got %+v, want %+v", repos, wantRepos)
	}
}

func TestReposService_resolveRepoRev_noRevSpecified_getsDefaultBranch(t *testing.T) {
	ctx, mock := testContext()

	wantRepoRev := &sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "r"},
		Rev:      "b",
		CommitID: strings.Repeat("a", 40),
	}

	calledGet := mock.servers.Repos.MockGet_Return(t, &sourcegraph.Repo{URI: "r", DefaultBranch: "b"})
	var calledVCSRepoResolveRevision bool
	mock.stores.RepoVCS.MockOpen(t, "r", vcstesting.MockRepository{
		ResolveRevision_: func(rev string) (vcs.CommitID, error) {
			calledVCSRepoResolveRevision = true
			return vcs.CommitID(wantRepoRev.CommitID), nil
		},
	})

	repoRev := &sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "r"},
		// (no rev/branch specified)
	}
	if err := resolveRepoRev(ctx, repoRev); err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !calledVCSRepoResolveRevision {
		t.Error("!calledVCSRepoResolveRevision")
	}
	if !reflect.DeepEqual(repoRev, wantRepoRev) {
		t.Errorf("got %+v, want %+v", repoRev, wantRepoRev)
	}
}

func TestReposService_resolveRepoRev_noCommitIDSpecified_resolvesRev(t *testing.T) {
	ctx, mock := testContext()

	wantRepoRev := &sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "r"},
		Rev:      "b",
		CommitID: strings.Repeat("a", 40),
	}

	calledGet := mock.stores.Repos.MockGet(t, "r")
	var calledVCSRepoResolveRevision bool
	mock.stores.RepoVCS.MockOpen(t, "r", vcstesting.MockRepository{
		ResolveRevision_: func(rev string) (vcs.CommitID, error) {
			calledVCSRepoResolveRevision = true
			return vcs.CommitID(wantRepoRev.CommitID), nil
		},
	})

	repoRev := &sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "r"},
		Rev:      "b",
		// (no commit ID specified)
	}
	if err := resolveRepoRev(ctx, repoRev); err != nil {
		t.Fatal(err)
	}
	if *calledGet {
		t.Error("calledGet needlessly")
	}
	if !calledVCSRepoResolveRevision {
		t.Error("!calledVCSRepoResolveRevision")
	}
	if !reflect.DeepEqual(repoRev, wantRepoRev) {
		t.Errorf("got %+v, want %+v", repoRev, wantRepoRev)
	}
}

func TestReposService_resolveRepoRev_revSpecIsAlreadyResolved_noop(t *testing.T) {
	ctx, mock := testContext()

	calledGet := mock.stores.Repos.MockGet(t, "r")
	// TODO(nodb-ctx): check that the VCS opener is never used

	wantRepoRev := &sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "r"},
		Rev:      "b",
		CommitID: strings.Repeat("a", 40),
	}

	repoRev := &sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "r"},
		Rev:      "b",
		CommitID: strings.Repeat("a", 40),
	}
	if err := resolveRepoRev(ctx, repoRev); err != nil {
		t.Fatal(err)
	}
	if *calledGet {
		t.Error("calledGet needlessly")
	}
	if !reflect.DeepEqual(repoRev, wantRepoRev) {
		t.Errorf("got %+v, want %+v", repoRev, wantRepoRev)
	}
}

func TestReposService_GetConfig(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepoConfig := &sourcegraph.RepoConfig{
		Apps: []string{"a", "b"},
	}

	calledConfigsGet := mock.stores.RepoConfigs.MockGet_Return(t, "r", wantRepoConfig)

	conf, err := s.GetConfig(ctx, &sourcegraph.RepoSpec{URI: "r"})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledConfigsGet {
		t.Error("!calledConfigsGet")
	}
	if !reflect.DeepEqual(conf, wantRepoConfig) {
		t.Errorf("got %+v, want %+v", conf, wantRepoConfig)
	}
}

func TestReposService_ConfigureApp_Enable(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	// Add dummy app.
	platform.Apps["b"] = struct{}{}
	defer func() {
		delete(platform.Apps, "b")
	}()

	calledConfigsGet := mock.stores.RepoConfigs.MockGet_Return(t, "r", &sourcegraph.RepoConfig{Apps: []string{"a"}})
	var calledConfigsUpdate bool
	mock.stores.RepoConfigs.Update_ = func(ctx context.Context, repo string, conf sourcegraph.RepoConfig) error {
		if want := []string{"a", "b"}; !reflect.DeepEqual(conf.Apps, want) {
			t.Errorf("got %#v, want Apps %v", conf, want)
		}
		calledConfigsUpdate = true
		return nil
	}

	_, err := s.ConfigureApp(ctx, &sourcegraph.RepoConfigureAppOp{
		Repo:   sourcegraph.RepoSpec{URI: "r"},
		App:    "b",
		Enable: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledConfigsGet {
		t.Error("!calledConfigsGet")
	}
	if !calledConfigsUpdate {
		t.Error("!calledConfigsUpdate")
	}
}

func TestReposService_ConfigureApp_Disable(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	// Add dummy app.
	platform.Apps["b"] = struct{}{}
	defer func() {
		delete(platform.Apps, "b")
	}()

	calledConfigsGet := mock.stores.RepoConfigs.MockGet_Return(t, "r", &sourcegraph.RepoConfig{Apps: []string{"a", "b"}})
	var calledConfigsUpdate bool
	mock.stores.RepoConfigs.Update_ = func(ctx context.Context, repo string, conf sourcegraph.RepoConfig) error {
		if want := []string{"a"}; !reflect.DeepEqual(conf.Apps, want) {
			t.Errorf("got %#v, want Apps %v", conf, want)
		}
		calledConfigsUpdate = true
		return nil
	}

	_, err := s.ConfigureApp(ctx, &sourcegraph.RepoConfigureAppOp{
		Repo:   sourcegraph.RepoSpec{URI: "r"},
		App:    "b",
		Enable: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledConfigsGet {
		t.Error("!calledConfigsGet")
	}
	if !calledConfigsUpdate {
		t.Error("!calledConfigsUpdate")
	}
}
