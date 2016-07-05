package backend

import (
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/platform"
)

func TestReposService_Get(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepo := &sourcegraph.Repo{
		ID:      1,
		URI:     "github.com/u/r",
		HTMLURL: "http://github.com/u/r",
		Mirror:  true,
	}

	mock.githubRepos.MockGet_Return(ctx, &sourcegraph.Repo{})

	calledGet := mock.stores.Repos.MockGet_Return(t, wantRepo)
	calledUpdate := mock.stores.Repos.MockUpdate(t, 1)

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	// Should not be called because mock GitHub has same data as mock DB.
	if *calledUpdate {
		t.Error("calledUpdate")
	}
	if !reflect.DeepEqual(repo, wantRepo) {
		t.Errorf("got %+v, want %+v", repo, wantRepo)
	}
}

func TestReposService_Get_UpdateMeta(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepo := &sourcegraph.Repo{
		ID:      1,
		URI:     "github.com/u/r",
		HTMLURL: "http://github.com/u/r",
		Mirror:  true,
	}

	mock.githubRepos.MockGet_Return(ctx, &sourcegraph.Repo{
		Description: "This is a repository",
	})

	calledGet := mock.stores.Repos.MockGet_Return(t, wantRepo)
	calledUpdate := mock.stores.Repos.MockUpdate(t, 1)

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledUpdate {
		t.Error("!calledUpdate")
	}
	if !reflect.DeepEqual(repo, wantRepo) {
		t.Errorf("got %+v, want %+v", repo, wantRepo)
	}
}

func TestReposService_Get_UnauthedUpdateMeta(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	// Remove auth from testContext
	ctx = authpkg.WithActor(ctx, authpkg.Actor{})
	ctx = accesscontrol.WithInsecureSkip(ctx, false)

	wantRepo := &sourcegraph.Repo{
		ID:      1,
		URI:     "github.com/u/r",
		HTMLURL: "http://github.com/u/r",
		Mirror:  true,
	}

	mock.githubRepos.MockGet_Return(ctx, &sourcegraph.Repo{
		Description: "This is a repository",
	})

	calledGet := mock.stores.Repos.MockGet_Return(t, wantRepo)
	var calledUpdate bool
	mock.stores.Repos.Update_ = func(ctx context.Context, op store.RepoUpdate) error {
		if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Repos.Update", op.Repo); err != nil {
			return err
		}
		calledUpdate = true
		if op.ReposUpdateOp.Repo != wantRepo.ID {
			t.Errorf("got repo %q, want %q", op.ReposUpdateOp.Repo, wantRepo.ID)
			return grpc.Errorf(codes.NotFound, "repo %v not found", wantRepo.ID)
		}
		return nil
	}

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !calledUpdate {
		t.Error("!calledUpdate")
	}
	if !reflect.DeepEqual(repo, wantRepo) {
		t.Errorf("got %+v, want %+v", repo, wantRepo)
	}
}

func TestReposService_Get_NonGitHub(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepo := &sourcegraph.Repo{
		ID:      1,
		URI:     "r",
		HTMLURL: "http://example.com/r",
		Mirror:  true,
	}

	mock.githubRepos.MockGet_Return(ctx, &sourcegraph.Repo{})

	calledGet := mock.stores.Repos.MockGet_Return(t, wantRepo)
	calledUpdate := mock.stores.Repos.MockUpdate(t, 1)

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if *calledUpdate {
		t.Error("calledUpdate")
	}
	if !reflect.DeepEqual(repo, wantRepo) {
		t.Errorf("got %+v, want %+v", repo, wantRepo)
	}
}

func TestRepos_Create_New(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepo := &sourcegraph.Repo{
		ID:      1,
		URI:     "r",
		Name:    "r",
		HTMLURL: "http://example.com/r",
	}

	calledCreate := false
	mock.stores.Repos.Create_ = func(ctx context.Context, repo *sourcegraph.Repo) (int32, error) {
		calledCreate = true
		if repo.URI != wantRepo.URI {
			t.Errorf("got uri %#v, want %#v", repo.URI, wantRepo.URI)
		}
		return wantRepo.ID, nil
	}
	mock.stores.Repos.MockGet(t, 1)

	_, err := s.Create(ctx, &sourcegraph.ReposCreateOp{
		Op: &sourcegraph.ReposCreateOp_New{New: &sourcegraph.ReposCreateOp_NewRepo{
			URI: "r",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !calledCreate {
		t.Error("!calledCreate")
	}
}

func TestRepos_Create_Origin(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepo := &sourcegraph.Repo{
		ID:      1,
		URI:     "github.com/a/b",
		HTMLURL: "http://example.com/github.com/a/b",
		Origin: &sourcegraph.Origin{
			ID:         "123",
			Service:    sourcegraph.Origin_GitHub,
			APIBaseURL: "https://api.github.com",
		},
	}

	calledGet := false
	mock.githubRepos.GetByID_ = func(ctx context.Context, id int) (*sourcegraph.Repo, error) {
		if want := 123; id != want {
			t.Errorf("got id %d, want %d", id, want)
		}
		calledGet = true
		return &sourcegraph.Repo{Origin: &sourcegraph.Origin{ID: "123", Service: sourcegraph.Origin_GitHub}}, nil
	}

	calledCreate := false
	mock.stores.Repos.Create_ = func(ctx context.Context, repo *sourcegraph.Repo) (int32, error) {
		calledCreate = true
		if !reflect.DeepEqual(repo.Origin, wantRepo.Origin) {
			t.Errorf("got repo origin %#v, want %#v", repo.Origin, wantRepo.Origin)
		}
		return wantRepo.ID, nil
	}
	mock.stores.Repos.MockGet(t, 1)

	_, err := s.Create(ctx, &sourcegraph.ReposCreateOp{
		Op: &sourcegraph.ReposCreateOp_Origin{Origin: wantRepo.Origin},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !calledGet {
		t.Error("!calledGet")
	}
	if !calledCreate {
		t.Error("!calledCreate")
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

	mock.stores.Repos.MockGetByURI(t, "r", 1)
	calledConfigsGet := mock.stores.RepoConfigs.MockGet_Return(t, 1, wantRepoConfig)

	conf, err := s.GetConfig(ctx, &sourcegraph.RepoSpec{ID: 1})
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

	mock.stores.Repos.MockGetByURI(t, "r", 1)
	calledConfigsGet := mock.stores.RepoConfigs.MockGet_Return(t, 1, &sourcegraph.RepoConfig{Apps: []string{"a"}})
	var calledConfigsUpdate bool
	mock.stores.RepoConfigs.Update_ = func(ctx context.Context, repo int32, conf sourcegraph.RepoConfig) error {
		if want := []string{"a", "b"}; !reflect.DeepEqual(conf.Apps, want) {
			t.Errorf("got %#v, want Apps %v", conf, want)
		}
		calledConfigsUpdate = true
		return nil
	}

	_, err := s.ConfigureApp(ctx, &sourcegraph.RepoConfigureAppOp{
		Repo:   1,
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

	mock.stores.Repos.MockGetByURI(t, "r", 1)
	calledConfigsGet := mock.stores.RepoConfigs.MockGet_Return(t, 1, &sourcegraph.RepoConfig{Apps: []string{"a", "b"}})
	var calledConfigsUpdate bool
	mock.stores.RepoConfigs.Update_ = func(ctx context.Context, repo int32, conf sourcegraph.RepoConfig) error {
		if want := []string{"a"}; !reflect.DeepEqual(conf.Apps, want) {
			t.Errorf("got %#v, want Apps %v", conf, want)
		}
		calledConfigsUpdate = true
		return nil
	}

	_, err := s.ConfigureApp(ctx, &sourcegraph.RepoConfigureAppOp{
		Repo:   1,
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
