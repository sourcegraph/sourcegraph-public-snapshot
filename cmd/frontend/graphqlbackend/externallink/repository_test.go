package externallink

import (
	"context"
	"reflect"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestRepository(t *testing.T) {
	t.Run("repo", func(t *testing.T) {
		resetMocks()
		repo := &types.Repo{
			Name: api.RepoName("github.com/foo/bar"),
			ExternalRepo: api.ExternalRepoSpec{
				ServiceID:   extsvc.GitHubDotCom.ServiceID,
				ServiceType: extsvc.GitHubDotCom.ServiceType,
			},
			Metadata: &github.Repository{
				URL: "http://github.com/foo/bar",
			},
		}
		database.Mocks.Phabricator.GetByName = func(repo api.RepoName) (*types.PhabricatorRepo, error) {
			return nil, errors.New("x")
		}
		links, err := Repository(context.Background(), new(dbtesting.MockDB), repo)
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver{
			{
				url:         "http://github.com/foo/bar",
				serviceKind: extsvc.TypeToKind(repo.ExternalRepo.ServiceType),
				serviceType: repo.ExternalRepo.ServiceType,
			},
		}; !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})

	t.Run("phabricator", func(t *testing.T) {
		resetMocks()
		database.Mocks.Phabricator.GetByName = func(repo api.RepoName) (*types.PhabricatorRepo, error) {
			if want := api.RepoName("myrepo"); repo != want {
				t.Errorf("got %q, want %q", repo, want)
			}
			return &types.PhabricatorRepo{URL: "http://phabricator.example.com/", Callsign: "MYREPO"}, nil
		}
		links, err := Repository(context.Background(), new(dbtesting.MockDB), &types.Repo{Name: "myrepo"})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver{
			{
				url:         "http://phabricator.example.com/diffusion/MYREPO",
				serviceKind: extsvc.KindPhabricator,
				serviceType: extsvc.TypePhabricator,
			},
		}; !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})

	t.Run("errors", func(t *testing.T) {
		resetMocks()
		database.Mocks.Phabricator.GetByName = func(repo api.RepoName) (*types.PhabricatorRepo, error) {
			return nil, errors.New("x")
		}
		links, err := Repository(context.Background(), new(dbtesting.MockDB), &types.Repo{Name: "myrepo"})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver(nil); !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})
}

func TestFileOrDir(t *testing.T) {
	const (
		rev  = "myrev"
		path = "mydir/myfile"
	)

	repo := &types.Repo{
		Name: api.RepoName("gitlab.com/foo/bar"),
		ExternalRepo: api.ExternalRepoSpec{
			ServiceID:   extsvc.GitLabDotCom.ServiceID,
			ServiceType: extsvc.GitLabDotCom.ServiceType,
		},
		Metadata: &gitlab.Project{
			ProjectCommon: gitlab.ProjectCommon{
				WebURL: "http://gitlab.com/foo/bar",
			},
		},
	}

	for _, which := range []string{"file", "dir"} {
		var (
			isDir   bool
			wantURL string
		)
		switch which {
		case "file":
			isDir = false
			wantURL = "http://gitlab.com/foo/bar/blob/myrev/mydir/myfile"
		case "dir":
			isDir = true
			wantURL = "http://gitlab.com/foo/bar/tree/myrev/mydir/myfile"
		}

		t.Run(which, func(t *testing.T) {
			resetMocks()
			database.Mocks.Phabricator.GetByName = func(repo api.RepoName) (*types.PhabricatorRepo, error) {
				return nil, errors.New("x")
			}
			links, err := FileOrDir(context.Background(), new(dbtesting.MockDB), repo, rev, path, isDir)
			if err != nil {
				t.Fatal(err)
			}
			if want := []*Resolver{
				{
					url:         wantURL,
					serviceKind: extsvc.TypeToKind(repo.ExternalRepo.ServiceType),
					serviceType: repo.ExternalRepo.ServiceType,
				},
			}; !reflect.DeepEqual(links, want) {
				t.Errorf("got %+v, want %+v", links, want)
			}
		})
	}

	t.Run("phabricator", func(t *testing.T) {
		resetMocks()
		database.Mocks.Phabricator.GetByName = func(repo api.RepoName) (*types.PhabricatorRepo, error) {
			if want := api.RepoName("myrepo"); repo != want {
				t.Errorf("got %q, want %q", repo, want)
			}
			return &types.PhabricatorRepo{URL: "http://phabricator.example.com/", Callsign: "MYREPO"}, nil
		}
		git.Mocks.ExecSafe = func(params []string) ([]byte, []byte, int, error) {
			return []byte("mybranch"), nil, 0, nil
		}
		defer git.ResetMocks()
		links, err := FileOrDir(context.Background(), new(dbtesting.MockDB), &types.Repo{Name: "myrepo"}, rev, path, true)
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver{
			{
				url:         "http://phabricator.example.com/source/MYREPO/browse/mybranch/mydir/myfile;myrev",
				serviceKind: extsvc.KindPhabricator,
				serviceType: extsvc.TypePhabricator,
			},
		}; !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})

	t.Run("errors", func(t *testing.T) {
		resetMocks()
		database.Mocks.Phabricator.GetByName = func(repo api.RepoName) (*types.PhabricatorRepo, error) {
			return nil, errors.New("x")
		}
		links, err := FileOrDir(context.Background(), new(dbtesting.MockDB), &types.Repo{Name: "myrepo"}, rev, path, true)
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver(nil); !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})
}

func TestCommit(t *testing.T) {
	const commit = "mycommit"

	repo := &types.Repo{
		Name: api.RepoName("github.com/foo/bar"),
		ExternalRepo: api.ExternalRepoSpec{
			ServiceID:   extsvc.GitHubDotCom.ServiceID,
			ServiceType: extsvc.GitHubDotCom.ServiceType,
		},
		Metadata: &github.Repository{
			URL: "http://github.com/foo/bar",
		},
	}

	t.Run("repo", func(t *testing.T) {
		resetMocks()
		database.Mocks.Phabricator.GetByName = func(repo api.RepoName) (*types.PhabricatorRepo, error) {
			return nil, errors.New("x")
		}
		links, err := Commit(context.Background(), new(dbtesting.MockDB), repo, commit)
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver{
			{
				url:         "http://github.com/foo/bar/commit/mycommit",
				serviceKind: extsvc.TypeToKind(repo.ExternalRepo.ServiceType),
				serviceType: repo.ExternalRepo.ServiceType,
			},
		}; !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})

	t.Run("phabricator", func(t *testing.T) {
		resetMocks()
		database.Mocks.Phabricator.GetByName = func(repo api.RepoName) (*types.PhabricatorRepo, error) {
			if want := api.RepoName("myrepo"); repo != want {
				t.Errorf("got %q, want %q", repo, want)
			}
			return &types.PhabricatorRepo{URL: "http://phabricator.example.com/", Callsign: "MYREPO"}, nil
		}
		links, err := Commit(context.Background(), new(dbtesting.MockDB), &types.Repo{Name: "myrepo"}, commit)
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver{
			{
				url:         "http://phabricator.example.com/rMYREPOmycommit",
				serviceKind: extsvc.KindPhabricator,
				serviceType: extsvc.TypePhabricator,
			},
		}; !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})

	t.Run("errors", func(t *testing.T) {
		resetMocks()
		database.Mocks.Phabricator.GetByName = func(repo api.RepoName) (*types.PhabricatorRepo, error) {
			return nil, errors.New("x")
		}
		links, err := Commit(context.Background(), new(dbtesting.MockDB), &types.Repo{Name: "myrepo"}, commit)
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver(nil); !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})
}
