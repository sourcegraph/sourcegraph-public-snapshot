package local

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"golang.org/x/tools/godoc/vfs"

	"strings"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	vcstesting "sourcegraph.com/sourcegraph/go-vcs/vcs/testing"
	"sourcegraph.com/sourcegraph/rwvfs"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/platform"
)

func TestReposService_Get(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepo := &sourcegraph.Repo{
		URI:         "r",
		HTMLURL:     "http://example.com/r",
		Permissions: &sourcegraph.RepoPermissions{Read: true},
	}

	calledGet := mock.stores.Repos.MockGet(t, "r")
	calledGetPerms := mock.stores.Repos.MockGetPerms_Read()

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{URI: "r"})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledGetPerms {
		t.Error("!calledGetPerms")
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
			{URI: "r1", HTMLURL: "http://example.com/r1", Permissions: &sourcegraph.RepoPermissions{Read: true}},
			{URI: "r2", HTMLURL: "http://example.com/r2", Permissions: &sourcegraph.RepoPermissions{Read: true}},
		},
	}

	calledList := mock.stores.Repos.MockList(t, "r1", "r2")
	calledGetPerms := mock.stores.Repos.MockGetPerms_Read()

	repos, err := s.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !*calledList {
		t.Error("!calledList")
	}
	if !*calledGetPerms {
		t.Error("!calledGetPerms")
	}
	if !reflect.DeepEqual(repos, wantRepos) {
		t.Errorf("got %+v, want %+v", repos, wantRepos)
	}
}

func TestReposService_resolveRepoRev_noRevSpecified_getsDefaultBranch(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepoRev := &sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "r"},
		Rev:      "b",
		CommitID: strings.Repeat("a", 40),
	}

	calledGet := mock.stores.Repos.MockGet_Return(t, &sourcegraph.Repo{URI: "r", DefaultBranch: "b"})
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
	if err := s.resolveRepoRev(ctx, repoRev); err != nil {
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
	var s repos
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
	if err := s.resolveRepoRev(ctx, repoRev); err != nil {
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
	var s repos
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
	if err := s.resolveRepoRev(ctx, repoRev); err != nil {
		t.Fatal(err)
	}
	if *calledGet {
		t.Error("calledGet needlessly")
	}
	if !reflect.DeepEqual(repoRev, wantRepoRev) {
		t.Errorf("got %+v, want %+v", repoRev, wantRepoRev)
	}
}

func TestReposService_GetReadme(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantReadme := &sourcegraph.Readme{
		Path: "README.txt",
		HTML: "<pre>hello</pre>",
	}

	var calledVCSRepoFileSystem bool
	mock.stores.RepoVCS.MockOpen(t, "r", vcstesting.MockRepository{
		FileSystem_: func(at vcs.CommitID) (vfs.FileSystem, error) {
			calledVCSRepoFileSystem = true
			return rwvfs.Map(map[string]string{"README.txt": "hello"}), nil
		},
	})

	readme, err := s.GetReadme(ctx, &sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "r"},
		Rev:      "v",
		CommitID: strings.Repeat("a", 40),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !calledVCSRepoFileSystem {
		t.Error("!calledVCSRepoFileSystem")
	}
	if !reflect.DeepEqual(readme, wantReadme) {
		t.Errorf("got %+v, want %+v", readme, wantReadme)
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
	platform.Frames()["b"] = platform.RepoFrame{}
	defer func() {
		delete(platform.Frames(), "b")
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
	platform.Frames()["b"] = platform.RepoFrame{}
	defer func() {
		delete(platform.Frames(), "b")
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
