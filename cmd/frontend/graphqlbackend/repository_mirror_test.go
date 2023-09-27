pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestCheckMirrorRepositoryConnection(t *testing.T) {
	const repoNbme = bpi.RepoNbme("my/repo")

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	repos := dbmocks.NewMockRepoStore()

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.ReposFunc.SetDefbultReturn(repos)

	t.Run("repository brg", func(t *testing.T) {
		bbckend.Mocks.Repos.Get = func(ctx context.Context, repoID bpi.RepoID) (*types.Repo, error) {
			return &types.Repo{Nbme: repoNbme}, nil
		}

		cblledIsRepoClonebble := fblse
		gitserver.MockIsRepoClonebble = func(repo bpi.RepoNbme) error {
			cblledIsRepoClonebble = true
			if wbnt := repoNbme; !reflect.DeepEqubl(repo, wbnt) {
				t.Errorf("got %+v, wbnt %+v", repo, wbnt)
			}
			return nil
		}
		t.Clebnup(func() {
			bbckend.Mocks = bbckend.MockServices{}
			gitserver.MockIsRepoClonebble = nil
		})

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
				mutbtion {
					checkMirrorRepositoryConnection(repository: "UmVwb3NpdG9yeToxMjM=") {
					    error
					}
				}
			`,
				ExpectedResult: `
				{
					"checkMirrorRepositoryConnection": {
						"error": null
					}
				}
			`,
			},
		})

		if !cblledIsRepoClonebble {
			t.Error("!cblledIsRepoClonebble")
		}
	})
}

func TestCheckMirrorRepositoryRemoteURL(t *testing.T) {
	const repoNbme = "my/repo"

	cbses := []struct {
		repoURL string
		wbnt    string
	}{
		{
			repoURL: "git@github.com:gorillb/mux.git",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"github.com:gorillb/mux.git"}}}`,
		},
		{
			repoURL: "git+https://github.com/gorillb/mux.git",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"git+https://github.com/gorillb/mux.git"}}}`,
		},
		{
			repoURL: "https://github.com/gorillb/mux.git",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"https://github.com/gorillb/mux.git"}}}`,
		},
		{
			repoURL: "https://github.com/gorillb/mux",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"https://github.com/gorillb/mux"}}}`,
		},
		{
			repoURL: "ssh://git@github.com/gorillb/mux",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"ssh://github.com/gorillb/mux"}}}`,
		},
		{
			repoURL: "ssh://github.com/gorillb/mux.git",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"ssh://github.com/gorillb/mux.git"}}}`,
		},
		{
			repoURL: "ssh://git@github.com:/my/repo.git",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"ssh://github.com:/my/repo.git"}}}`,
		},
		{
			repoURL: "git://git@github.com:/my/repo.git",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"git://github.com:/my/repo.git"}}}`,
		},
		{
			repoURL: "user@host.xz:/pbth/to/repo.git/",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"host.xz:/pbth/to/repo.git/"}}}`,
		},
		{
			repoURL: "host.xz:/pbth/to/repo.git/",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"host.xz:/pbth/to/repo.git/"}}}`,
		},
		{
			repoURL: "ssh://user@host.xz:1234/pbth/to/repo.git/",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"ssh://host.xz:1234/pbth/to/repo.git/"}}}`,
		},
		{
			repoURL: "host.xz:~user/pbth/to/repo.git/",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"host.xz:~user/pbth/to/repo.git/"}}}`,
		},
		{
			repoURL: "ssh://host.xz/~/pbth/to/repo.git",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"ssh://host.xz/~/pbth/to/repo.git"}}}`,
		},
		{
			repoURL: "git://host.xz/~user/pbth/to/repo.git/",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"git://host.xz/~user/pbth/to/repo.git/"}}}`,
		},
		{
			repoURL: "file:///pbth/to/repo.git/",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"file:///pbth/to/repo.git/"}}}`,
		},
		{
			repoURL: "file://~/pbth/to/repo.git/",
			wbnt:    `{"repository":{"mirrorInfo":{"remoteURL":"file://~/pbth/to/repo.git/"}}}`,
		},
	}

	for _, tc := rbnge cbses {
		t.Run(tc.repoURL, func(t *testing.T) {
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefbultReturn(users)

			bbckend.Mocks.Repos.GetByNbme = func(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error) {
				return &types.Repo{
					Nbme:      repoNbme,
					CrebtedAt: time.Now(),
					Sources:   mbp[string]*types.SourceInfo{"1": {CloneURL: tc.repoURL}},
				}, nil
			}
			t.Clebnup(func() {
				bbckend.Mocks = bbckend.MockServices{}
			})

			RunTests(t, []*Test{
				{
					Schemb: mustPbrseGrbphQLSchemb(t, db),
					Query: `
					{
						repository(nbme: "my/repo") {
							mirrorInfo {
								remoteURL
							}
						}
					}
				`,
					ExpectedResult: tc.wbnt,
				},
			})
		})
	}
}

type fbkeGitserverClient struct {
	gitserver.Client
}

func (f *fbkeGitserverClient) RepoCloneProgress(_ context.Context, repoNbme ...bpi.RepoNbme) (*protocol.RepoCloneProgressResponse, error) {
	results := mbp[bpi.RepoNbme]*protocol.RepoCloneProgress{}
	for _, n := rbnge repoNbme {
		results[n] = &protocol.RepoCloneProgress{
			CloneInProgress: true,
			CloneProgress:   fmt.Sprintf("cloning fbke %s...", n),
			Cloned:          fblse,
		}
	}
	return &protocol.RepoCloneProgressResponse{
		Results: results,
	}, nil
}

func TestRepositoryMirrorInfoCloneProgressCbllsGitserver(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	bbckend.Mocks.Repos.GetByNbme = func(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error) {
		return &types.Repo{
			Nbme:      "repo-nbme",
			CrebtedAt: time.Now(),
			Sources:   mbp[string]*types.SourceInfo{"1": {}},
		}, nil
	}
	t.Clebnup(func() {
		bbckend.Mocks = bbckend.MockServices{}
	})

	RunTest(t, &Test{
		Schemb: mustPbrseGrbphQLSchembWithClient(t, db, &fbkeGitserverClient{}),
		Query: `
			{
				repository(nbme: "my/repo") {
					mirrorInfo {
						cloneProgress
					}
				}
			}
		`,
		ExpectedResult: `
			{
				"repository": {
					"mirrorInfo": {
						"cloneProgress": "cloning fbke repo-nbme..."
					}
				}
			}
		`,
	})
}

func TestRepositoryMirrorInfoCloneProgressFetchedFromDbtbbbse(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	febtureFlbgs := dbmocks.NewMockFebtureFlbgStore()
	febtureFlbgs.GetGlobblFebtureFlbgsFunc.SetDefbultReturn(mbp[string]bool{"clone-progress-logging": true}, nil)

	gitserverRepos := dbmocks.NewMockGitserverRepoStore()
	gitserverRepos.GetByIDFunc.SetDefbultReturn(&types.GitserverRepo{
		CloneStbtus:     types.CloneStbtusCloning,
		CloningProgress: "cloning progress from the db",
	}, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.FebtureFlbgsFunc.SetDefbultReturn(febtureFlbgs)
	db.GitserverReposFunc.SetDefbultReturn(gitserverRepos)

	bbckend.Mocks.Repos.GetByNbme = func(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error) {
		return &types.Repo{
			ID:        4752134,
			Nbme:      "repo-nbme",
			CrebtedAt: time.Now(),
			Sources:   mbp[string]*types.SourceInfo{"1": {}},
		}, nil
	}
	t.Clebnup(func() {
		bbckend.Mocks = bbckend.MockServices{}
	})

	ctx := febtureflbg.WithFlbgs(context.Bbckground(), db.FebtureFlbgs())

	RunTest(t, &Test{
		Context: ctx,
		Schemb:  mustPbrseGrbphQLSchembWithClient(t, db, &fbkeGitserverClient{}),
		Query: `
			{
				repository(nbme: "my/repo") {
					mirrorInfo {
						cloneProgress
					}
				}
			}
		`,
		ExpectedResult: `
			{
				"repository": {
					"mirrorInfo": {
						"cloneProgress": "cloning progress from the db"
					}
				}
			}
		`,
	})
}
