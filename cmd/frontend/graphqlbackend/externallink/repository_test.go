package externallink

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestRepository(t *testing.T) {
	t.Run("repo-updater info with no repo.ExternalRepo", func(t *testing.T) {
		resetMocks()
		repoupdater.MockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return &protocol.RepoLookupResult{
				Repo: &protocol.RepoInfo{
					Links: &protocol.RepoLinks{
						Root: "http://example.com/myrepo",
					},
				},
			}, nil
		}
		db.Mocks.Phabricator.GetByURI = func(repo api.RepoURI) (*types.PhabricatorRepo, error) {
			return nil, errors.New("x")
		}
		links, err := Repository(context.Background(), &types.Repo{URI: "myrepo"})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver{
			{
				url:         "http://example.com/myrepo",
				serviceType: "",
			},
		}; !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})

	t.Run("repo-updater info with repo.ExternalRepo", func(t *testing.T) {
		resetMocks()
		externalRepoSpec := api.ExternalRepoSpec{
			ID:          "myid",
			ServiceType: "github",
		}
		repoupdater.MockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			if !reflect.DeepEqual(args.ExternalRepo, &externalRepoSpec) {
				t.Errorf("got %+v, want %+v", args.ExternalRepo, externalRepoSpec)
			}
			return &protocol.RepoLookupResult{
				Repo: &protocol.RepoInfo{
					Links: &protocol.RepoLinks{
						Root: "https://github.example.com/myorg/myrepo",
					},
					ExternalRepo: args.ExternalRepo,
				},
			}, nil
		}
		db.Mocks.Phabricator.GetByURI = func(repo api.RepoURI) (*types.PhabricatorRepo, error) {
			return nil, errors.New("x")
		}
		links, err := Repository(context.Background(), &types.Repo{
			URI:          "myrepo",
			ExternalRepo: &externalRepoSpec,
		})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver{
			{
				url:         "https://github.example.com/myorg/myrepo",
				serviceType: "github",
			},
		}; !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})

	t.Run("phabricator", func(t *testing.T) {
		resetMocks()
		repoupdater.MockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return &protocol.RepoLookupResult{}, nil
		}
		db.Mocks.Phabricator.GetByURI = func(repo api.RepoURI) (*types.PhabricatorRepo, error) {
			if want := api.RepoURI("myrepo"); repo != want {
				t.Errorf("got %q, want %q", repo, want)
			}
			return &types.PhabricatorRepo{URL: "http://phabricator.example.com/", Callsign: "MYREPO"}, nil
		}
		links, err := Repository(context.Background(), &types.Repo{URI: "myrepo"})
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver{
			{
				url:         "http://phabricator.example.com/diffusion/MYREPO",
				serviceType: "phabricator",
			},
		}; !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})

	t.Run("errors", func(t *testing.T) {
		resetMocks()
		repoupdater.MockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return nil, errors.New("x")
		}
		db.Mocks.Phabricator.GetByURI = func(repo api.RepoURI) (*types.PhabricatorRepo, error) {
			return nil, errors.New("x")
		}
		links, err := Repository(context.Background(), &types.Repo{URI: "myrepo"})
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

	for _, which := range []string{"file", "dir"} {
		var (
			repoLinks protocol.RepoLinks
			isDir     bool
			wantURL   string
		)
		switch which {
		case "file":
			repoLinks = protocol.RepoLinks{Blob: "http://example.com/myrepo@{rev}/file/{path}"}
			isDir = false
			wantURL = "http://example.com/myrepo@myrev/file/mydir/myfile"
		case "dir":
			repoLinks = protocol.RepoLinks{Tree: "http://example.com/myrepo@{rev}/dir/{path}"}
			isDir = true
			wantURL = "http://example.com/myrepo@myrev/dir/mydir/myfile"
		}

		t.Run(which, func(t *testing.T) {
			t.Run("repo-updater info with no repo.ExternalRepo", func(t *testing.T) {
				resetMocks()
				repoupdater.MockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
					return &protocol.RepoLookupResult{
						Repo: &protocol.RepoInfo{
							Links: &repoLinks,
						},
					}, nil
				}
				db.Mocks.Phabricator.GetByURI = func(repo api.RepoURI) (*types.PhabricatorRepo, error) {
					return nil, errors.New("x")
				}
				links, err := FileOrDir(context.Background(), &types.Repo{URI: "myrepo"}, rev, path, isDir)
				if err != nil {
					t.Fatal(err)
				}
				if want := []*Resolver{
					{
						url:         wantURL,
						serviceType: "",
					},
				}; !reflect.DeepEqual(links, want) {
					t.Errorf("got %+v, want %+v", links, want)
				}
			})

			t.Run("repo-updater info with repo.ExternalRepo", func(t *testing.T) {
				resetMocks()
				externalRepoSpec := api.ExternalRepoSpec{
					ID:          "myid",
					ServiceType: "github",
				}
				repoupdater.MockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
					if !reflect.DeepEqual(args.ExternalRepo, &externalRepoSpec) {
						t.Errorf("got %+v, want %+v", args.ExternalRepo, externalRepoSpec)
					}
					return &protocol.RepoLookupResult{
						Repo: &protocol.RepoInfo{
							Links:        &repoLinks,
							ExternalRepo: args.ExternalRepo,
						},
					}, nil
				}
				db.Mocks.Phabricator.GetByURI = func(repo api.RepoURI) (*types.PhabricatorRepo, error) {
					return nil, errors.New("x")
				}
				links, err := FileOrDir(context.Background(), &types.Repo{
					URI:          "myrepo",
					ExternalRepo: &externalRepoSpec,
				}, rev, path, isDir)
				if err != nil {
					t.Fatal(err)
				}
				if want := []*Resolver{
					{
						url:         wantURL,
						serviceType: "github",
					},
				}; !reflect.DeepEqual(links, want) {
					t.Errorf("got %+v, want %+v", links, want)
				}
			})
		})
	}

	t.Run("phabricator", func(t *testing.T) {
		resetMocks()
		repoupdater.MockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return &protocol.RepoLookupResult{}, nil
		}
		db.Mocks.Phabricator.GetByURI = func(repo api.RepoURI) (*types.PhabricatorRepo, error) {
			if want := api.RepoURI("myrepo"); repo != want {
				t.Errorf("got %q, want %q", repo, want)
			}
			return &types.PhabricatorRepo{URL: "http://phabricator.example.com/", Callsign: "MYREPO"}, nil
		}
		git.Mocks.ExecSafe = func(params []string) ([]byte, []byte, int, error) {
			return []byte("mybranch"), nil, 0, nil
		}
		defer git.ResetMocks()
		links, err := FileOrDir(context.Background(), &types.Repo{URI: "myrepo"}, rev, path, true)
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver{
			{
				url:         "http://phabricator.example.com/source/MYREPO/browse/mybranch/mydir/myfile;myrev",
				serviceType: "phabricator",
			},
		}; !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})

	t.Run("errors", func(t *testing.T) {
		resetMocks()
		repoupdater.MockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return nil, errors.New("x")
		}
		db.Mocks.Phabricator.GetByURI = func(repo api.RepoURI) (*types.PhabricatorRepo, error) {
			return nil, errors.New("x")
		}
		links, err := FileOrDir(context.Background(), &types.Repo{URI: "myrepo"}, rev, path, true)
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

	t.Run("repo-updater info with no repo.ExternalRepo", func(t *testing.T) {
		resetMocks()
		repoupdater.MockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return &protocol.RepoLookupResult{
				Repo: &protocol.RepoInfo{
					Links: &protocol.RepoLinks{
						Commit: "http://example.com/myrepo/commit/{commit}",
					},
				},
			}, nil
		}
		db.Mocks.Phabricator.GetByURI = func(repo api.RepoURI) (*types.PhabricatorRepo, error) {
			return nil, errors.New("x")
		}
		links, err := Commit(context.Background(), &types.Repo{URI: "myrepo"}, commit)
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver{
			{
				url:         "http://example.com/myrepo/commit/mycommit",
				serviceType: "",
			},
		}; !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})

	t.Run("repo-updater info with repo.ExternalRepo", func(t *testing.T) {
		resetMocks()
		externalRepoSpec := api.ExternalRepoSpec{
			ID:          "myid",
			ServiceType: "github",
		}
		repoupdater.MockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			if !reflect.DeepEqual(args.ExternalRepo, &externalRepoSpec) {
				t.Errorf("got %+v, want %+v", args.ExternalRepo, externalRepoSpec)
			}
			return &protocol.RepoLookupResult{
				Repo: &protocol.RepoInfo{
					Links: &protocol.RepoLinks{
						Commit: "https://github.example.com/myorg/myrepo/commit/{commit}",
					},
					ExternalRepo: args.ExternalRepo,
				},
			}, nil
		}
		db.Mocks.Phabricator.GetByURI = func(repo api.RepoURI) (*types.PhabricatorRepo, error) {
			return nil, errors.New("x")
		}
		links, err := Commit(context.Background(), &types.Repo{
			URI:          "myrepo",
			ExternalRepo: &externalRepoSpec,
		}, commit)
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver{
			{
				url:         "https://github.example.com/myorg/myrepo/commit/mycommit",
				serviceType: "github",
			},
		}; !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})

	t.Run("phabricator", func(t *testing.T) {
		resetMocks()
		repoupdater.MockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return &protocol.RepoLookupResult{}, nil
		}
		db.Mocks.Phabricator.GetByURI = func(repo api.RepoURI) (*types.PhabricatorRepo, error) {
			if want := api.RepoURI("myrepo"); repo != want {
				t.Errorf("got %q, want %q", repo, want)
			}
			return &types.PhabricatorRepo{URL: "http://phabricator.example.com/", Callsign: "MYREPO"}, nil
		}
		links, err := Commit(context.Background(), &types.Repo{URI: "myrepo"}, commit)
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver{
			{
				url:         "http://phabricator.example.com/rMYREPOmycommit",
				serviceType: "phabricator",
			},
		}; !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})

	t.Run("errors", func(t *testing.T) {
		resetMocks()
		repoupdater.MockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return nil, errors.New("x")
		}
		db.Mocks.Phabricator.GetByURI = func(repo api.RepoURI) (*types.PhabricatorRepo, error) {
			return nil, errors.New("x")
		}
		links, err := Commit(context.Background(), &types.Repo{URI: "myrepo"}, commit)
		if err != nil {
			t.Fatal(err)
		}
		if want := []*Resolver(nil); !reflect.DeepEqual(links, want) {
			t.Errorf("got %+v, want %+v", links, want)
		}
	})
}
