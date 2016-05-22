package backend

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/platform"
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
