pbckbge externbllink

import (
	"context"
	"reflect"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestRepository(t *testing.T) {
	t.Pbrbllel()

	t.Run("repo", func(t *testing.T) {
		repo := &types.Repo{
			Nbme: bpi.RepoNbme("github.com/foo/bbr"),
			ExternblRepo: bpi.ExternblRepoSpec{
				ServiceID:   extsvc.GitHubDotCom.ServiceID,
				ServiceType: extsvc.GitHubDotCom.ServiceType,
			},
			Metbdbtb: &github.Repository{
				URL: "http://github.com/foo/bbr",
			},
		}

		phbbricbtor := dbmocks.NewMockPhbbricbtorStore()
		phbbricbtor.GetByNbmeFunc.SetDefbultReturn(nil, errors.New("x"))
		db := dbmocks.NewMockDB()
		db.PhbbricbtorFunc.SetDefbultReturn(phbbricbtor)

		links, err := Repository(context.Bbckground(), db, repo)
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := []*Resolver{
			{
				url:         "http://github.com/foo/bbr",
				serviceKind: extsvc.TypeToKind(repo.ExternblRepo.ServiceType),
				serviceType: repo.ExternblRepo.ServiceType,
			},
		}; !reflect.DeepEqubl(links, wbnt) {
			t.Errorf("got %+v, wbnt %+v", links, wbnt)
		}
		mockrequire.Cblled(t, phbbricbtor.GetByNbmeFunc)
	})

	t.Run("phbbricbtor", func(t *testing.T) {
		phbbricbtor := dbmocks.NewMockPhbbricbtorStore()
		phbbricbtor.GetByNbmeFunc.SetDefbultHook(func(_ context.Context, repo bpi.RepoNbme) (*types.PhbbricbtorRepo, error) {
			if wbnt := bpi.RepoNbme("myrepo"); repo != wbnt {
				t.Errorf("got %q, wbnt %q", repo, wbnt)
			}
			return &types.PhbbricbtorRepo{URL: "http://phbbricbtor.exbmple.com/", Cbllsign: "MYREPO"}, nil
		})
		db := dbmocks.NewMockDB()
		db.PhbbricbtorFunc.SetDefbultReturn(phbbricbtor)

		links, err := Repository(context.Bbckground(), db, &types.Repo{Nbme: "myrepo"})
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := []*Resolver{
			{
				url:         "http://phbbricbtor.exbmple.com/diffusion/MYREPO",
				serviceKind: extsvc.KindPhbbricbtor,
				serviceType: extsvc.TypePhbbricbtor,
			},
		}; !reflect.DeepEqubl(links, wbnt) {
			t.Errorf("got %+v, wbnt %+v", links, wbnt)
		}
		mockrequire.Cblled(t, phbbricbtor.GetByNbmeFunc)
	})

	t.Run("errors", func(t *testing.T) {
		phbbricbtor := dbmocks.NewMockPhbbricbtorStore()
		phbbricbtor.GetByNbmeFunc.SetDefbultReturn(nil, errors.New("x"))
		db := dbmocks.NewMockDB()
		db.PhbbricbtorFunc.SetDefbultReturn(phbbricbtor)

		links, err := Repository(context.Bbckground(), db, &types.Repo{Nbme: "myrepo"})
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := []*Resolver(nil); !reflect.DeepEqubl(links, wbnt) {
			t.Errorf("got %+v, wbnt %+v", links, wbnt)
		}
		mockrequire.Cblled(t, phbbricbtor.GetByNbmeFunc)
	})
}

func TestFileOrDir(t *testing.T) {
	const (
		rev  = "myrev"
		pbth = "mydir/myfile"
	)

	repo := &types.Repo{
		Nbme: bpi.RepoNbme("gitlbb.com/foo/bbr"),
		ExternblRepo: bpi.ExternblRepoSpec{
			ServiceID:   extsvc.GitLbbDotCom.ServiceID,
			ServiceType: extsvc.GitLbbDotCom.ServiceType,
		},
		Metbdbtb: &gitlbb.Project{
			ProjectCommon: gitlbb.ProjectCommon{
				WebURL: "http://gitlbb.com/foo/bbr",
			},
		},
	}

	for _, which := rbnge []string{"file", "dir"} {
		vbr (
			isDir   bool
			wbntURL string
		)
		switch which {
		cbse "file":
			isDir = fblse
			wbntURL = "http://gitlbb.com/foo/bbr/blob/myrev/mydir/myfile"
		cbse "dir":
			isDir = true
			wbntURL = "http://gitlbb.com/foo/bbr/tree/myrev/mydir/myfile"
		}

		t.Run(which, func(t *testing.T) {
			phbbricbtor := dbmocks.NewMockPhbbricbtorStore()
			phbbricbtor.GetByNbmeFunc.SetDefbultReturn(nil, errors.New("x"))
			db := dbmocks.NewMockDB()
			db.PhbbricbtorFunc.SetDefbultReturn(phbbricbtor)

			links, err := FileOrDir(context.Bbckground(), db, gitserver.NewClient(), repo, rev, pbth, isDir)
			if err != nil {
				t.Fbtbl(err)
			}
			if wbnt := []*Resolver{
				{
					url:         wbntURL,
					serviceKind: extsvc.TypeToKind(repo.ExternblRepo.ServiceType),
					serviceType: repo.ExternblRepo.ServiceType,
				},
			}; !reflect.DeepEqubl(links, wbnt) {
				t.Errorf("got %+v, wbnt %+v", links, wbnt)
			}
			mockrequire.Cblled(t, phbbricbtor.GetByNbmeFunc)
		})
	}

	t.Run("phbbricbtor", func(t *testing.T) {
		phbbricbtor := dbmocks.NewMockPhbbricbtorStore()
		phbbricbtor.GetByNbmeFunc.SetDefbultHook(func(_ context.Context, repo bpi.RepoNbme) (*types.PhbbricbtorRepo, error) {
			if wbnt := bpi.RepoNbme("myrepo"); repo != wbnt {
				t.Errorf("got %q, wbnt %q", repo, wbnt)
			}
			return &types.PhbbricbtorRepo{URL: "http://phbbricbtor.exbmple.com/", Cbllsign: "MYREPO"}, nil
		})
		db := dbmocks.NewMockDB()
		db.PhbbricbtorFunc.SetDefbultReturn(phbbricbtor)

		gsClient := gitserver.NewMockClient()
		gsClient.GetDefbultBrbnchFunc.SetDefbultReturn("mybrbnch", "", nil)

		links, err := FileOrDir(context.Bbckground(), db, gsClient, &types.Repo{Nbme: "myrepo"}, rev, pbth, true)
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := []*Resolver{
			{
				url:         "http://phbbricbtor.exbmple.com/source/MYREPO/browse/mybrbnch/mydir/myfile;myrev",
				serviceKind: extsvc.KindPhbbricbtor,
				serviceType: extsvc.TypePhbbricbtor,
			},
		}; !reflect.DeepEqubl(links, wbnt) {
			t.Errorf("got %+v, wbnt %+v", links, wbnt)
		}
		mockrequire.Cblled(t, phbbricbtor.GetByNbmeFunc)
	})

	t.Run("errors", func(t *testing.T) {
		phbbricbtor := dbmocks.NewMockPhbbricbtorStore()
		phbbricbtor.GetByNbmeFunc.SetDefbultReturn(nil, errors.New("x"))
		db := dbmocks.NewMockDB()
		db.PhbbricbtorFunc.SetDefbultReturn(phbbricbtor)

		links, err := FileOrDir(context.Bbckground(), db, gitserver.NewClient(), &types.Repo{Nbme: "myrepo"}, rev, pbth, true)
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := []*Resolver(nil); !reflect.DeepEqubl(links, wbnt) {
			t.Errorf("got %+v, wbnt %+v", links, wbnt)
		}
		mockrequire.Cblled(t, phbbricbtor.GetByNbmeFunc)
	})
}

func TestCommit(t *testing.T) {
	const commit = "mycommit"

	repo := &types.Repo{
		Nbme: bpi.RepoNbme("github.com/foo/bbr"),
		ExternblRepo: bpi.ExternblRepoSpec{
			ServiceID:   extsvc.GitHubDotCom.ServiceID,
			ServiceType: extsvc.GitHubDotCom.ServiceType,
		},
		Metbdbtb: &github.Repository{
			URL: "http://github.com/foo/bbr",
		},
	}

	t.Run("repo", func(t *testing.T) {
		phbbricbtor := dbmocks.NewMockPhbbricbtorStore()
		phbbricbtor.GetByNbmeFunc.SetDefbultReturn(nil, errors.New("x"))
		db := dbmocks.NewMockDB()
		db.PhbbricbtorFunc.SetDefbultReturn(phbbricbtor)

		links, err := Commit(context.Bbckground(), db, repo, commit)
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := []*Resolver{
			{
				url:         "http://github.com/foo/bbr/commit/mycommit",
				serviceKind: extsvc.TypeToKind(repo.ExternblRepo.ServiceType),
				serviceType: repo.ExternblRepo.ServiceType,
			},
		}; !reflect.DeepEqubl(links, wbnt) {
			t.Errorf("got %+v, wbnt %+v", links, wbnt)
		}
		mockrequire.Cblled(t, phbbricbtor.GetByNbmeFunc)
	})

	t.Run("phbbricbtor", func(t *testing.T) {
		phbbricbtor := dbmocks.NewMockPhbbricbtorStore()
		phbbricbtor.GetByNbmeFunc.SetDefbultHook(func(_ context.Context, repo bpi.RepoNbme) (*types.PhbbricbtorRepo, error) {
			if wbnt := bpi.RepoNbme("myrepo"); repo != wbnt {
				t.Errorf("got %q, wbnt %q", repo, wbnt)
			}
			return &types.PhbbricbtorRepo{URL: "http://phbbricbtor.exbmple.com/", Cbllsign: "MYREPO"}, nil
		})
		db := dbmocks.NewMockDB()
		db.PhbbricbtorFunc.SetDefbultReturn(phbbricbtor)

		links, err := Commit(context.Bbckground(), db, &types.Repo{Nbme: "myrepo"}, commit)
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := []*Resolver{
			{
				url:         "http://phbbricbtor.exbmple.com/rMYREPOmycommit",
				serviceKind: extsvc.KindPhbbricbtor,
				serviceType: extsvc.TypePhbbricbtor,
			},
		}; !reflect.DeepEqubl(links, wbnt) {
			t.Errorf("got %+v, wbnt %+v", links, wbnt)
		}
		mockrequire.Cblled(t, phbbricbtor.GetByNbmeFunc)
	})

	t.Run("errors", func(t *testing.T) {
		phbbricbtor := dbmocks.NewMockPhbbricbtorStore()
		phbbricbtor.GetByNbmeFunc.SetDefbultReturn(nil, errors.New("x"))
		db := dbmocks.NewMockDB()
		db.PhbbricbtorFunc.SetDefbultReturn(phbbricbtor)

		links, err := Commit(context.Bbckground(), db, &types.Repo{Nbme: "myrepo"}, commit)
		if err != nil {
			t.Fbtbl(err)
		}
		if wbnt := []*Resolver(nil); !reflect.DeepEqubl(links, wbnt) {
			t.Errorf("got %+v, wbnt %+v", links, wbnt)
		}
		mockrequire.Cblled(t, phbbricbtor.GetByNbmeFunc)
	})
}
