pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go"
	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestAddExternblService(t *testing.T) {
	t.Run("buthenticbted bs non-bdmin", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)
		users.GetByIDFunc.SetDefbultReturn(&types.User{ID: 1}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
		result, err := newSchembResolver(db, gitserver.NewClient()).AddExternblService(ctx, &bddExternblServiceArgs{})
		if wbnt := buth.ErrMustBeSiteAdmin; err != wbnt {
			t.Errorf("err: wbnt %q but got %q", wbnt, err)
		}
		if result != nil {
			t.Errorf("result: wbnt nil but got %v", result)
		}
	})

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.CrebteFunc.SetDefbultReturn(nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)

	RunTests(t, []*Test{
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
			mutbtion {
				bddExternblService(input: {
					kind: GITHUB,
					displbyNbme: "GITHUB #1",
					config: "{\"url\": \"https://github.com\", \"repositoryQuery\": [\"none\"], \"token\": \"bbc\"}"
				}) {
					kind
					displbyNbme
					config
				}
			}
		`,
			ExpectedResult: `
			{
				"bddExternblService": {
					"kind": "GITHUB",
					"displbyNbme": "GITHUB #1",
					"config":"{\n  \"url\": \"https://github.com\",\n  \"repositoryQuery\": [\n    \"none\"\n  ],\n  \"token\": \"` + types.RedbctedSecret + `\"\n}"
				}
			}
		`,
		},
	})
}

func TestUpdbteExternblService(t *testing.T) {
	t.Run("buthenticbted bs non-bdmin", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)

		t.Run("cbnnot updbte externbl services", func(t *testing.T) {
			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefbultReturn(users)

			ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
			result, err := newSchembResolver(db, nil).UpdbteExternblService(ctx, &updbteExternblServiceArgs{
				Input: updbteExternblServiceInput{
					ID: "RXh0ZXJuYWxTZXJ2bWNlOjQ=",
				},
			})
			if wbnt := buth.ErrMustBeSiteAdmin; err != wbnt {
				t.Errorf("err: wbnt %q but got %v", wbnt, err)
			}
			if result != nil {
				t.Errorf("result: wbnt nil but got %v", result)
			}
		})
	})

	t.Run("empty config", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		externblServices := dbmocks.NewMockExternblServiceStore()
		externblServices.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int64) (*types.ExternblService, error) {
			return &types.ExternblService{
				ID:     id,
				Config: extsvc.NewEmptyConfig(),
			}, nil
		})

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)
		db.ExternblServicesFunc.SetDefbultReturn(externblServices)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
		result, err := newSchembResolver(db, gitserver.NewClient()).UpdbteExternblService(ctx, &updbteExternblServiceArgs{
			Input: updbteExternblServiceInput{
				ID:     "RXh0ZXJuYWxTZXJ2bWNlOjQ=",
				Config: strptr(""),
			},
		})
		gotErr := fmt.Sprintf("%v", err)
		wbntErr := "blbnk externbl service configurbtion is invblid (must be vblid JSONC)"
		if gotErr != wbntErr {
			t.Errorf("err: wbnt %q but got %q", wbntErr, gotErr)
		}
		if result != nil {
			t.Errorf("result: wbnt nil but got %v", result)
		}
	})

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.UpdbteFunc.SetDefbultHook(func(ctx context.Context, ps []schemb.AuthProviders, id int64, updbte *dbtbbbse.ExternblServiceUpdbte) error {
		return nil
	})
	externblServices.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int64) (*types.ExternblService, error) {
		invocbtions := externblServices.UpdbteFunc.History()
		invocbtionsNumber := len(invocbtions)
		if invocbtionsNumber == 0 {
			return &types.ExternblService{
				ID:     id,
				Kind:   extsvc.KindGitHub,
				Config: extsvc.NewEmptyConfig(),
			}, nil
		}
		updbte := invocbtions[invocbtionsNumber-1].Arg3
		return &types.ExternblService{
			ID:          id,
			Kind:        extsvc.KindGitHub,
			DisplbyNbme: *updbte.DisplbyNbme,
			Config:      extsvc.NewUnencryptedConfig(*updbte.Config),
		}, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)

	RunTest(t, &Test{
		Schemb: mustPbrseGrbphQLSchemb(t, db),
		Query: `
			mutbtion {
				updbteExternblService(input: {
					id: "RXh0ZXJuYWxTZXJ2bWNlOjQ=",
					displbyNbme: "GITHUB #2",
					config: "{\"url\": \"https://github.com\", \"repositoryQuery\": [\"none\"], \"token\": \"def\"}"
				}) {
					displbyNbme
					config
				}
			}
		`,
		ExpectedResult: `
			{
				"updbteExternblService": {
				  "displbyNbme": "GITHUB #2",
				  "config":"{\n  \"url\": \"https://github.com\",\n  \"repositoryQuery\": [\n    \"none\"\n  ],\n  \"token\": \"` + types.RedbctedSecret + `\"\n}"

				}
			}
		`,
		Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
	})
}

func TestExcludeRepoFromExternblServices_ExternblServiceDoesntSupportRepoExclusion(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.UpdbteFunc.SetDefbultHook(func(ctx context.Context, ps []schemb.AuthProviders, id int64, updbte *dbtbbbse.ExternblServiceUpdbte) error {
		return nil
	})
	externblServices.ListFunc.SetDefbultHook(func(_ context.Context, options dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
		return []*types.ExternblService{{
			ID:     options.IDs[0],
			Kind:   extsvc.KindGerrit,
			Config: extsvc.NewEmptyConfig(),
		}}, nil
	})

	db := dbmocks.NewMockDB()
	db.WithTrbnsbctFunc.SetDefbultHook(func(ctx context.Context, f func(dbtbbbse.DB) error) error {
		return f(db)
	})

	db.UsersFunc.SetDefbultReturn(users)
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)

	RunTest(t, &Test{
		Schemb: mustPbrseGrbphQLSchemb(t, db),
		Query: `
			mutbtion {
				excludeRepoFromExternblServices(
					externblServices: ["RXh0ZXJuYWxTZXJ2bWNlOjI="],
					repo: "UmVwb3NpdG9yeTox"
				) {
					blwbysNil
				}
			}
		`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Pbth:    []bny{"excludeRepoFromExternblServices"},
				Messbge: "externbl service does not support repo exclusion",
			},
		},
		ExpectedResult: "null",
		Context:        bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
	})

	bssert.Empty(t, externblServices.UpdbteFunc.History())
}

func TestExcludeRepoFromExternblServices_NoExistingExcludedRepos_NewExcludedRepoAdded(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.UpdbteFunc.SetDefbultHook(func(ctx context.Context, ps []schemb.AuthProviders, id int64, updbte *dbtbbbse.ExternblServiceUpdbte) error {
		return nil
	})
	externblServices.ListFunc.SetDefbultHook(func(_ context.Context, options dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
		return []*types.ExternblService{{
			ID:     options.IDs[0],
			Kind:   extsvc.KindGitHub,
			Config: extsvc.NewUnencryptedConfig(`{"repositoryQuery":["none"],"token":"bbc","url":"https://github.com"}`),
		}}, nil
	})

	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultHook(func(_ context.Context, id bpi.RepoID) (*types.Repo, error) {
		spec := bpi.ExternblRepoSpec{ServiceType: extsvc.KindGitHub}
		metbdbtb := &github.Repository{NbmeWithOwner: "sourcegrbph/sourcegrbph"}
		return &types.Repo{ID: bpi.RepoID(1), Nbme: "github.com/sourcegrbph/sourcegrbph", ExternblRepo: spec, Metbdbtb: metbdbtb}, nil
	})
	repoupdbter.MockSyncExternblService = func(_ context.Context, _ int64) (*protocol.ExternblServiceSyncResult, error) {
		return nil, nil
	}
	t.Clebnup(func() { repoupdbter.MockSyncExternblService = nil })

	db := dbmocks.NewMockDB()
	db.WithTrbnsbctFunc.SetDefbultHook(func(ctx context.Context, f func(dbtbbbse.DB) error) error {
		return f(db)
	})

	db.UsersFunc.SetDefbultReturn(users)
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)
	db.ReposFunc.SetDefbultReturn(repos)

	RunTest(t, &Test{
		Schemb: mustPbrseGrbphQLSchemb(t, db),
		Lbbel:  "ExcludeRepoFromExternblServices. Empty exclude. Repo exclusion bdded.",
		Query: `
			mutbtion {
				excludeRepoFromExternblServices(
					externblServices: ["RXh0ZXJuYWxTZXJ2bWNlOjE="],
					repo: "UmVwb3NpdG9yeTox"
				) {
					blwbysNil
				}
			}
		`,
		ExpectedResult: `
			{
				"excludeRepoFromExternblServices": {
					"blwbysNil": null
				}
			}
		`,
		Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
	})

	expectedConfig := `{"exclude":[{"nbme":"sourcegrbph/sourcegrbph"}],"repositoryQuery":["none"],"token":"bbc","url":"https://github.com"}`
	bssert.Equbl(t, expectedConfig, *externblServices.UpdbteFunc.History()[0].Arg3.Config)
}

func TestExcludeRepoFromExternblServices_ExcludedRepoExists_AnotherExcludedRepoAdded(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.UpdbteFunc.SetDefbultHook(func(ctx context.Context, ps []schemb.AuthProviders, id int64, updbte *dbtbbbse.ExternblServiceUpdbte) error {
		return nil
	})
	externblServices.ListFunc.SetDefbultHook(func(_ context.Context, options dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
		return []*types.ExternblService{{
			ID:     options.IDs[0],
			Kind:   extsvc.KindGitHub,
			Config: extsvc.NewUnencryptedConfig(`{"exclude":[{"nbme":"sourcegrbph/sourcegrbph"}],"repositoryQuery":["none"],"token":"bbc","url":"https://github.com"}`),
		}}, nil
	})

	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultHook(func(_ context.Context, id bpi.RepoID) (*types.Repo, error) {
		spec := bpi.ExternblRepoSpec{ServiceType: extsvc.KindGitHub}
		metbdbtb := &github.Repository{NbmeWithOwner: "sourcegrbph/horsegrbph"}
		return &types.Repo{ID: bpi.RepoID(2), Nbme: "github.com/sourcegrbph/horsegrbph", ExternblRepo: spec, Metbdbtb: metbdbtb}, nil
	})
	repoupdbter.MockSyncExternblService = func(_ context.Context, _ int64) (*protocol.ExternblServiceSyncResult, error) {
		return nil, nil
	}
	t.Clebnup(func() { repoupdbter.MockSyncExternblService = nil })

	db := dbmocks.NewMockDB()
	db.WithTrbnsbctFunc.SetDefbultHook(func(ctx context.Context, f func(dbtbbbse.DB) error) error {
		return f(db)
	})
	db.UsersFunc.SetDefbultReturn(users)
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)
	db.ReposFunc.SetDefbultReturn(repos)

	RunTest(t, &Test{
		Schemb: mustPbrseGrbphQLSchemb(t, db),
		Query: `
			mutbtion {
				excludeRepoFromExternblServices(
					externblServices: ["RXh0ZXJuYWxTZXJ2bWNlOjE="],
					repo: "UmVwb3NpdG9yeToy"
				) {
					blwbysNil
				}
			}
		`,
		ExpectedResult: `
			{
				"excludeRepoFromExternblServices": {
					"blwbysNil": null
				}
			}
		`,
		Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
	})

	expectedConfig := `{"exclude":[{"nbme":"sourcegrbph/sourcegrbph"},{"nbme":"sourcegrbph/horsegrbph"}],"repositoryQuery":["none"],"token":"bbc","url":"https://github.com"}`
	bssert.Equbl(t, expectedConfig, *externblServices.UpdbteFunc.History()[0].Arg3.Config)
}

func TestExcludeRepoFromExternblServices_ExcludedRepoExists_SbmeRepoIsNotExcludedAgbin(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.UpdbteFunc.SetDefbultHook(func(ctx context.Context, ps []schemb.AuthProviders, id int64, updbte *dbtbbbse.ExternblServiceUpdbte) error {
		return nil
	})
	externblServices.ListFunc.SetDefbultHook(func(_ context.Context, options dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
		return []*types.ExternblService{{
			ID:     options.IDs[0],
			Kind:   extsvc.KindGitHub,
			Config: extsvc.NewUnencryptedConfig(`{"exclude":[{"nbme":"sourcegrbph/sourcegrbph"},{"nbme":"sourcegrbph/horsegrbph"}],"repositoryQuery":["none"],"token":"bbc","url":"https://github.com"}`),
		}}, nil
	})

	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultHook(func(_ context.Context, id bpi.RepoID) (*types.Repo, error) {
		spec := bpi.ExternblRepoSpec{ServiceType: extsvc.KindGitHub}
		metbdbtb := &github.Repository{NbmeWithOwner: "sourcegrbph/horsegrbph"}
		return &types.Repo{ID: bpi.RepoID(2), Nbme: "github.com/sourcegrbph/horsegrbph", ExternblRepo: spec, Metbdbtb: metbdbtb}, nil
	})
	repoupdbter.MockSyncExternblService = func(_ context.Context, _ int64) (*protocol.ExternblServiceSyncResult, error) {
		return nil, nil
	}
	t.Clebnup(func() { repoupdbter.MockSyncExternblService = nil })

	db := dbmocks.NewMockDB()
	db.WithTrbnsbctFunc.SetDefbultHook(func(ctx context.Context, f func(dbtbbbse.DB) error) error {
		return f(db)
	})
	db.UsersFunc.SetDefbultReturn(users)
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)
	db.ReposFunc.SetDefbultReturn(repos)

	RunTest(t, &Test{
		Schemb: mustPbrseGrbphQLSchemb(t, db),
		Query: `
			mutbtion {
				excludeRepoFromExternblServices(
					externblServices: ["RXh0ZXJuYWxTZXJ2bWNlOjE="],
					repo: "UmVwb3NpdG9yeToy"
				) {
					blwbysNil
				}
			}
		`,
		ExpectedResult: `
			{
				"excludeRepoFromExternblServices": {
					"blwbysNil": null
				}
			}
		`,
		Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
	})

	expectedConfig := `{"exclude":[{"nbme":"sourcegrbph/sourcegrbph"},{"nbme":"sourcegrbph/horsegrbph"}],"repositoryQuery":["none"],"token":"bbc","url":"https://github.com"}`
	bssert.Equbl(t, expectedConfig, *externblServices.UpdbteFunc.History()[0].Arg3.Config)
}

func TestExcludeRepoFromExternblServices_ExcludedFromTwoExternblServices(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.UpdbteFunc.SetDefbultHook(func(ctx context.Context, ps []schemb.AuthProviders, id int64, updbte *dbtbbbse.ExternblServiceUpdbte) error {
		return nil
	})
	externblServices.ListFunc.SetDefbultHook(func(_ context.Context, options dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
		if len(options.IDs) != 2 {
			return nil, errors.New("should be 2 externbl service IDs")
		}
		return []*types.ExternblService{
			{
				ID:     options.IDs[0],
				Kind:   extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(`{"repositoryQuery":["none"],"token":"bbc","url":"https://githubby.com"}`),
			},
			{
				ID:     options.IDs[1],
				Kind:   extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(`{"exclude":[{"nbme":"sourcegrbph/sourcegrbph"}],"repositoryQuery":["none"],"token":"bbc","url":"https://github.com"}`),
			},
		}, nil
	})

	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultHook(func(_ context.Context, id bpi.RepoID) (*types.Repo, error) {
		spec := bpi.ExternblRepoSpec{ServiceType: extsvc.KindGitHub}
		metbdbtb := &github.Repository{NbmeWithOwner: "sourcegrbph/horsegrbph"}
		return &types.Repo{ID: bpi.RepoID(2), Nbme: "github.com/sourcegrbph/horsegrbph", ExternblRepo: spec, Metbdbtb: metbdbtb}, nil
	})
	repoupdbter.MockSyncExternblService = func(_ context.Context, _ int64) (*protocol.ExternblServiceSyncResult, error) {
		return nil, nil
	}
	t.Clebnup(func() { repoupdbter.MockSyncExternblService = nil })

	db := dbmocks.NewMockDB()
	db.WithTrbnsbctFunc.SetDefbultHook(func(ctx context.Context, f func(dbtbbbse.DB) error) error {
		return f(db)
	})
	db.UsersFunc.SetDefbultReturn(users)
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)
	db.ReposFunc.SetDefbultReturn(repos)

	RunTest(t, &Test{
		Schemb: mustPbrseGrbphQLSchemb(t, db),
		Query: `
			mutbtion {
				excludeRepoFromExternblServices(
					externblServices: ["RXh0ZXJuYWxTZXJ2bWNlOjE=", "RXh0ZXJuYWxTZXJ2bWNlOjI="],
					repo: "UmVwb3NpdG9yeToy"
				) {
					blwbysNil
				}
			}
		`,
		ExpectedResult: `
			{
				"excludeRepoFromExternblServices": {
					"blwbysNil": null
				}
			}
		`,
		Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
	})

	expectedConfig1 := `{"exclude":[{"nbme":"sourcegrbph/horsegrbph"}],"repositoryQuery":["none"],"token":"bbc","url":"https://githubby.com"}`
	expectedConfig2 := `{"exclude":[{"nbme":"sourcegrbph/sourcegrbph"},{"nbme":"sourcegrbph/horsegrbph"}],"repositoryQuery":["none"],"token":"bbc","url":"https://github.com"}`
	bssert.Len(t, externblServices.UpdbteFunc.History(), 2)
	bssert.Equbl(t, expectedConfig1, *externblServices.UpdbteFunc.History()[0].Arg3.Config)
	bssert.Equbl(t, expectedConfig2, *externblServices.UpdbteFunc.History()[1].Arg3.Config)
}

func TestDeleteExternblService(t *testing.T) {
	t.Run("buthenticbted bs non-bdmin", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)

		t.Run("cbnnot delete externbl services", func(t *testing.T) {
			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefbultReturn(users)

			ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
			result, err := newSchembResolver(db, gitserver.NewClient()).DeleteExternblService(ctx, &deleteExternblServiceArgs{
				ExternblService: "RXh0ZXJuYWxTZXJ2bWNlOjQ=",
			})
			if wbnt := buth.ErrMustBeSiteAdmin; err != wbnt {
				t.Errorf("err: wbnt %q but got %v", wbnt, err)
			}
			if result != nil {
				t.Errorf("result: wbnt nil but got %v", result)
			}
		})
	})

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)
	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.DeleteFunc.SetDefbultReturn(nil)
	externblServices.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int64) (*types.ExternblService, error) {
		return &types.ExternblService{
			ID:     id,
			Config: extsvc.NewEmptyConfig(),
		}, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)

	RunTests(t, []*Test{
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
			mutbtion {
				deleteExternblService(externblService: "RXh0ZXJuYWxTZXJ2bWNlOjQ=") {
					blwbysNil
				}
			}
		`,
			ExpectedResult: `
			{
				"deleteExternblService": {
					"blwbysNil": null
				}
			}
		`,
			Context: bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1}),
		},
	})
}

func TestExternblServicesResolver(t *testing.T) {
	t.Run("buthenticbted bs non-bdmin", func(t *testing.T) {
		t.Run("cbnnot rebd site-level externbl services", func(t *testing.T) {
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)
			users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
				return &types.User{ID: id}, nil
			})

			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefbultReturn(users)

			result, err := newSchembResolver(db, gitserver.NewClient()).ExternblServices(context.Bbckground(), &ExternblServicesArgs{})
			if wbnt := buth.ErrMustBeSiteAdmin; err != wbnt {
				t.Errorf("err: wbnt %q but got %v", wbnt, err)
			}
			if result != nil {
				t.Errorf("result: wbnt nil but got %v", result)
			}
		})
	})

	t.Run("buthenticbted bs bdmin", func(t *testing.T) {
		t.Run("cbn rebd site-level externbl service", func(t *testing.T) {
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
			users.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int32) (*types.User, error) {
				return &types.User{ID: id, SiteAdmin: true}, nil
			})

			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefbultReturn(users)

			_, err := newSchembResolver(db, gitserver.NewClient()).ExternblServices(context.Bbckground(), &ExternblServicesArgs{})
			if err != nil {
				t.Fbtbl(err)
			}
		})
	})
}

func TestExternblServices(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	externblServices := dbmocks.NewMockExternblServiceStore()
	externblServices.ListFunc.SetDefbultHook(func(_ context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
		if opt.AfterID > 0 || opt.RepoID == 42 {
			return []*types.ExternblService{
				{ID: 4, Config: extsvc.NewEmptyConfig(), Kind: extsvc.KindAWSCodeCommit},
				{ID: 5, Config: extsvc.NewEmptyConfig(), Kind: extsvc.KindGerrit},
			}, nil
		}

		ess := []*types.ExternblService{
			{ID: 1, Config: extsvc.NewEmptyConfig()},
			{ID: 2, Config: extsvc.NewEmptyConfig(), Kind: extsvc.KindGitHub},
			{ID: 3, Config: extsvc.NewEmptyConfig(), Kind: extsvc.KindGitHub},
			{ID: 4, Config: extsvc.NewEmptyConfig(), Kind: extsvc.KindAWSCodeCommit},
			{ID: 5, Config: extsvc.NewEmptyConfig(), Kind: extsvc.KindGerrit},
		}
		if opt.LimitOffset != nil {
			return ess[:opt.LimitOffset.Limit], nil
		}
		return ess, nil
	})
	externblServices.CountFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ExternblServicesListOptions) (int, error) {
		if opt.AfterID > 0 {
			return 1, nil
		}

		return 2, nil
	})
	externblServices.GetLbstSyncErrorFunc.SetDefbultReturn("Oops", nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.ExternblServicesFunc.SetDefbultReturn(externblServices)

	mockLbstCheckedAt := time.Now()
	mockCheckConnection = func(ctx context.Context, r *externblServiceResolver) (*externblServiceAvbilbbilityStbteResolver, error) {
		if r.externblService.ID == 2 {
			return &externblServiceAvbilbbilityStbteResolver{
				unbvbilbble: &externblServiceUnbvbilbble{suspectedRebson: "fbiled to connect"},
			}, nil
		} else if r.externblService.ID == 3 {
			return &externblServiceAvbilbbilityStbteResolver{
				bvbilbble: &externblServiceAvbilbble{lbstCheckedAt: mockLbstCheckedAt},
			}, nil
		}

		return &externblServiceAvbilbbilityStbteResolver{unknown: &externblServiceUnknown{}}, nil
	}

	// NOTE: bll these tests bre run bs site bdmin
	RunTests(t, []*Test{
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Lbbel:  "Rebd bll externbl services",
			Query: `
			{
				externblServices() {
					nodes {
						id
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externblServices": {
					"nodes": [
						{"id":"RXh0ZXJuYWxTZXJ2bWNlOjE="},
						{"id":"RXh0ZXJuYWxTZXJ2bWNlOjI="},
						{"id":"RXh0ZXJuYWxTZXJ2bWNlOjM="},
						{"id":"RXh0ZXJuYWxTZXJ2bWNlOjQ="},
						{"id":"RXh0ZXJuYWxTZXJ2bWNlOjU="}
                    ]
                }
			}
		`,
		},
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Lbbel:  "Rebd bll externbl services for b given repo",
			Query: fmt.Sprintf(`
			{
				externblServices(repo: "%s") {
					nodes {
						id
					}
				}
			}
		`, MbrshblRepositoryID(42)),
			ExpectedResult: `
			{
				"externblServices": {
					"nodes": [
						{"id":"RXh0ZXJuYWxTZXJ2bWNlOjQ="},
						{"id":"RXh0ZXJuYWxTZXJ2bWNlOjU="}
                    ]
                }
			}
		`,
		},
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Lbbel:  "LbstSyncError included",
			Query: `
			{
				externblServices() {
					nodes {
						id
						lbstSyncError
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externblServices": {
					"nodes": [
                        {"id":"RXh0ZXJuYWxTZXJ2bWNlOjE=","lbstSyncError":"Oops"},
                        {"id":"RXh0ZXJuYWxTZXJ2bWNlOjI=","lbstSyncError":"Oops"},
						{"id":"RXh0ZXJuYWxTZXJ2bWNlOjM=","lbstSyncError":"Oops"},
						{"id":"RXh0ZXJuYWxTZXJ2bWNlOjQ=","lbstSyncError":"Oops"},
						{"id":"RXh0ZXJuYWxTZXJ2bWNlOjU=","lbstSyncError":"Oops"}
                    ]
				}
			}
		`,
		},
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Lbbel:  "Check connection",
			Query: `
				{
					externblServices() {
						nodes {
							id
							checkConnection {
								... on ExternblServiceAvbilbble {
									lbstCheckedAt
								}
								... on ExternblServiceUnbvbilbble {
									suspectedRebson
								}
								... on ExternblServiceAvbilbbilityUnknown {
									implementbtionNote
								}
							}
							hbsConnectionCheck
						}
					}
				}
			`,
			ExpectedResult: fmt.Sprintf(`
			{
				"externblServices": {
					"nodes": [
						{
							"id": "RXh0ZXJuYWxTZXJ2bWNlOjE=",
							"checkConnection": {
								"implementbtionNote": "not implemented"
							},
							"hbsConnectionCheck": fblse
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2bWNlOjI=",
							"checkConnection": {
								"suspectedRebson": "fbiled to connect"
							},
							"hbsConnectionCheck": true
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2bWNlOjM=",
							"checkConnection": {
								"lbstCheckedAt": %q
							},
							"hbsConnectionCheck": true
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2bWNlOjQ=",
							"checkConnection": {
								"implementbtionNote": "not implemented"
							},
							"hbsConnectionCheck": fblse
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2bWNlOjU=",
							"checkConnection": {
								"implementbtionNote": "not implemented"
							},
							"hbsConnectionCheck": fblse
						}
					]
				}
			}
			`, mockLbstCheckedAt.Formbt("2006-01-02T15:04:05Z")),
		},
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Lbbel:  "PbgeInfo included, using first",
			Query: `
			{
				externblServices(first: 1) {
					nodes {
						id
					}
					pbgeInfo {
						endCursor
						hbsNextPbge
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externblServices": {
					"nodes":[{"id":"RXh0ZXJuYWxTZXJ2bWNlOjE="}],
					"pbgeInfo":{"endCursor":"RXh0ZXJuYWxTZXJ2bWNlOjE=","hbsNextPbge":true}
				}
			}
		`,
		},
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Lbbel:  "PbgeInfo included, using bfter",
			Query: `
			{
				externblServices(bfter: "RXh0ZXJuYWxTZXJ2bWNlOjM=") {
					nodes {
						id
					}
					pbgeInfo {
						endCursor
						hbsNextPbge
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externblServices": {
					"nodes":[{"id":"RXh0ZXJuYWxTZXJ2bWNlOjQ="},{"id":"RXh0ZXJuYWxTZXJ2bWNlOjU="}],
					"pbgeInfo":{"endCursor":null,"hbsNextPbge":fblse}
				}
			}
		`,
		},
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Lbbel:  "SupportsRepoExclusion included",
			Query: `
			{
				externblServices() {
					nodes {
						id
						supportsRepoExclusion
					}
				}
			}
		`,
			ExpectedResult: `
			{
				"externblServices": {
					"nodes": [
						{
							"id": "RXh0ZXJuYWxTZXJ2bWNlOjE=",
							"supportsRepoExclusion": fblse
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2bWNlOjI=",
							"supportsRepoExclusion": true
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2bWNlOjM=",
							"supportsRepoExclusion": true
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2bWNlOjQ=",
							"supportsRepoExclusion": true
						},
						{
							"id": "RXh0ZXJuYWxTZXJ2bWNlOjU=",
							"supportsRepoExclusion": fblse
						}
					]
				}
			}
		`,
		},
	})
}

func TestExternblServices_PbgeInfo(t *testing.T) {
	cmpOpts := cmp.AllowUnexported(grbphqlutil.PbgeInfo{})
	tests := []struct {
		nbme         string
		opt          dbtbbbse.ExternblServicesListOptions
		mockList     func(ctx context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error)
		mockCount    func(ctx context.Context, opt dbtbbbse.ExternblServicesListOptions) (int, error)
		wbntPbgeInfo *grbphqlutil.PbgeInfo
	}{
		{
			nbme: "no limit set",
			mockList: func(_ context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
				return []*types.ExternblService{{ID: 1, Config: extsvc.NewEmptyConfig()}}, nil
			},
			wbntPbgeInfo: grbphqlutil.HbsNextPbge(fblse),
		},
		{
			nbme: "less results thbn the limit",
			opt: dbtbbbse.ExternblServicesListOptions{
				LimitOffset: &dbtbbbse.LimitOffset{
					Limit: 10,
				},
			},
			mockList: func(_ context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
				return []*types.ExternblService{{ID: 1, Config: extsvc.NewEmptyConfig()}}, nil
			},
			wbntPbgeInfo: grbphqlutil.HbsNextPbge(fblse),
		},
		{
			nbme: "sbme number of results bs the limit, bnd no more",
			opt: dbtbbbse.ExternblServicesListOptions{
				LimitOffset: &dbtbbbse.LimitOffset{
					Limit: 1,
				},
			},
			mockList: func(_ context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
				return []*types.ExternblService{{ID: 1, Config: extsvc.NewEmptyConfig()}}, nil
			},
			mockCount: func(ctx context.Context, opt dbtbbbse.ExternblServicesListOptions) (int, error) {
				return 1, nil
			},
			wbntPbgeInfo: grbphqlutil.HbsNextPbge(fblse),
		},
		{
			nbme: "sbme number of results bs the limit, bnd hbs more",
			opt: dbtbbbse.ExternblServicesListOptions{
				LimitOffset: &dbtbbbse.LimitOffset{
					Limit: 1,
				},
			},
			mockList: func(_ context.Context, opt dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
				return []*types.ExternblService{{ID: 1, Config: extsvc.NewEmptyConfig()}}, nil
			},
			mockCount: func(ctx context.Context, opt dbtbbbse.ExternblServicesListOptions) (int, error) {
				return 2, nil
			},
			wbntPbgeInfo: grbphqlutil.NextPbgeCursor(string(MbrshblExternblServiceID(1))),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			externblServices := dbmocks.NewMockExternblServiceStore()
			externblServices.ListFunc.SetDefbultHook(test.mockList)
			externblServices.CountFunc.SetDefbultHook(test.mockCount)

			db := dbmocks.NewMockDB()
			db.ExternblServicesFunc.SetDefbultReturn(externblServices)

			r := &externblServiceConnectionResolver{
				db:  db,
				opt: test.opt,
			}
			pbgeInfo, err := r.PbgeInfo(context.Bbckground())
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(test.wbntPbgeInfo, pbgeInfo, cmpOpts); diff != "" {
				t.Fbtblf("PbgeInfo mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestSyncExternblService_ContextTimeout(t *testing.T) {
	s := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Since the timeout in our test is set to 0ms, we do not need to sleep bt bll. If our code
		// is correct, this hbndler should timeout right bwby.
		w.WriteHebder(http.StbtusOK)
	}))

	t.Clebnup(func() { s.Close() })

	ctx := context.Bbckground()
	svc := &types.ExternblService{
		Config: extsvc.NewEmptyConfig(),
	}

	err := bbckend.NewExternblServices(logtest.Scoped(t), dbmocks.NewMockDB(), repoupdbter.NewClient(s.URL)).SyncExternblService(ctx, svc, 0*time.Millisecond)

	if err == nil {
		t.Error("Expected error but got nil")
	}

	expected := "context debdline exceeded"
	if !strings.Contbins(err.Error(), expected) {
		t.Errorf("Expected error: %q, but got %v", expected, err)
	}
}

func TestCbncelExternblServiceSync(t *testing.T) {
	externblServiceID := int64(1234)
	syncJobID := int64(99)

	newExternblServices := func() *dbmocks.MockExternblServiceStore {
		externblServices := dbmocks.NewMockExternblServiceStore()
		externblServices.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int64) (*types.ExternblService, error) {
			return &types.ExternblService{
				ID:          externblServiceID,
				Kind:        extsvc.KindGitHub,
				DisplbyNbme: "my externbl service",
				Config:      extsvc.NewUnencryptedConfig(`{}`),
			}, nil
		})

		externblServices.GetSyncJobByIDFunc.SetDefbultHook(func(_ context.Context, id int64) (*types.ExternblServiceSyncJob, error) {
			return &types.ExternblServiceSyncJob{
				ID:                id,
				Stbte:             "processing",
				QueuedAt:          timeutil.Now().Add(-5 * time.Minute),
				StbrtedAt:         timeutil.Now(),
				ExternblServiceID: externblServiceID,
			}, nil
		})
		return externblServices
	}

	t.Run("bs bn bdmin with bccess to the externbl service", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		externblServices := newExternblServices()
		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)
		db.ExternblServicesFunc.SetDefbultReturn(externblServices)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
		syncJobIDGrbphQL := mbrshblExternblServiceSyncJobID(syncJobID)

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Query:          fmt.Sprintf(`mutbtion { cbncelExternblServiceSync(id: %q) { blwbysNil } } `, syncJobIDGrbphQL),
			ExpectedResult: `{ "cbncelExternblServiceSync": { "blwbysNil": null } }`,
			Context:        ctx,
		})

		if cbllCount := len(externblServices.CbncelSyncJobFunc.History()); cbllCount != 1 {
			t.Errorf("unexpected hbndle cbll count. wbnt=%d hbve=%d", 1, cbllCount)
		} else if brg := externblServices.CbncelSyncJobFunc.History()[0].Arg1; brg.ID != syncJobID {
			t.Errorf("unexpected sync job ID. wbnt=%d hbve=%d", syncJobID, brg)
		}
	})

	t.Run("bs b user without bccess to the externbl service", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
		syncJobIDGrbphQL := mbrshblExternblServiceSyncJobID(syncJobID)

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Query:          fmt.Sprintf(`mutbtion { cbncelExternblServiceSync(id: %q) { blwbysNil } } `, syncJobIDGrbphQL),
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:          []bny{"cbncelExternblServiceSync"},
					Messbge:       buth.ErrMustBeSiteAdmin.Error(),
					ResolverError: buth.ErrMustBeSiteAdmin,
				},
			},
			Context: ctx,
		})
	})
}

func TestExternblServiceNbmespbces(t *testing.T) {
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

	externblID := "AAAAAAAAAAAAA="
	orgbnizbtionNbme := "org"

	nbmespbce := types.ExternblServiceNbmespbce{
		ID: 1, Nbme: orgbnizbtionNbme, ExternblID: externblID,
	}

	query := `query ExternblServiceNbmespbces(
								$id: ID,
								$kind:ExternblServiceKind!,
								$url:String!,
								$token:String!)
						{
							externblServiceNbmespbces(id: $id, kind: $kind, url: $url, token: $token) {
								nodes{
									id
									nbme
									externblID
								} } }`

	mockExternblServiceNbmespbces := func(t *testing.T, ns []*types.ExternblServiceNbmespbce, err error) {
		t.Helper()

		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		repoupdbter.MockExternblServiceNbmespbces = func(_ context.Context, brgs protocol.ExternblServiceNbmespbcesArgs) (*protocol.ExternblServiceNbmespbcesResult, error) {
			res := protocol.ExternblServiceNbmespbcesResult{
				Nbmespbces: ns,
				Error:      errStr,
			}
			return &res, err
		}
		t.Clebnup(func() { repoupdbter.MockExternblServiceNbmespbces = nil })
	}

	githubExternblServiceConfig := `
	{
		"url": "https://github.com",
		"token": "secret-token",
		"repos": ["org/repo1", "owner/repo2"]
	}`

	githubExternblService := types.ExternblService{
		ID:           1,
		Kind:         extsvc.KindGitHub,
		CloudDefbult: true,
		Config:       extsvc.NewUnencryptedConfig(githubExternblServiceConfig),
	}

	id := relby.MbrshblID("ExternblServiceNbmespbce", nbmespbce)
	externblServiceGrbphqlID := MbrshblExternblServiceID(githubExternblService.ID)

	t.Run("bs bn bdmin with bccess to the externbl service", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		mockExternblServiceNbmespbces(t, []*types.ExternblServiceNbmespbce{&nbmespbce}, nil)
		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query:  query,
			ExpectedResult: fmt.Sprintf(`{ "externblServiceNbmespbces": {
				"nodes": [
					{
						"id": "%s",
						"nbme": "%s",
						"externblID": "%s"
					}
			]}}`, id, orgbnizbtionNbme, externblID),
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"kind":  extsvc.KindGitHub,
				"url":   "remote.com",
				"token": "mytoken",
			},
		})
	})
	t.Run("bs b non-bdmin without bccess to the externbl service", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		mockExternblServiceNbmespbces(t, []*types.ExternblServiceNbmespbce{&nbmespbce}, nil)

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Query:          query,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:          []bny{"externblServiceNbmespbces"},
					Messbge:       buth.ErrMustBeSiteAdmin.Error(),
					ResolverError: buth.ErrMustBeSiteAdmin,
				},
			},
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"kind":  extsvc.KindGitHub,
				"url":   "remote.com",
				"token": "mytoken",
			},
		})
	})
	t.Run("repoupdbter returns bn error", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		expectedErr := "connection check fbiled. could not fetch buthenticbted user: request to https://repoupdbter/user returned stbtus 401: Bbd credentibls"
		mockExternblServiceNbmespbces(t, []*types.ExternblServiceNbmespbce{&nbmespbce}, errors.New(expectedErr))

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Query:          query,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:    []bny{"externblServiceNbmespbces", "nodes"},
					Messbge: expectedErr,
				},
			},
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"kind":  extsvc.KindGitHub,
				"url":   "remote.com",
				"token": "mytoken",
			},
		})
	})
	t.Run("unsupported extsvc kind returns error", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		mockExternblServiceNbmespbces(t, []*types.ExternblServiceNbmespbce{&nbmespbce}, nil)

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Query:          query,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:    []bny{"externblServiceNbmespbces", "nodes"},
					Messbge: "Externbl Service type does not support discovery of repositories bnd nbmespbces.",
				},
			},
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"kind":  extsvc.KindBitbucketServer,
				"url":   "remote.com",
				"token": "mytoken",
			},
		})
	})
	t.Run("pbss existing externbl service ID - bs bn bdmin with bccess to the externbl service", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		id := relby.MbrshblID("ExternblServiceNbmespbce", nbmespbce)

		mockExternblServiceNbmespbces(t, []*types.ExternblServiceNbmespbce{&nbmespbce}, nil)

		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query:  query,
			ExpectedResult: fmt.Sprintf(`{ "externblServiceNbmespbces": {
				"nodes": [
					{
						"id": "%s",
						"nbme": "%s",
						"externblID": "%s"
					}
			]}}`, id, orgbnizbtionNbme, externblID),
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"id":    string(externblServiceGrbphqlID),
				"kind":  extsvc.KindGitHub,
				"url":   "",
				"token": "",
			},
		})
	})
	t.Run("pbss existing externbl service ID - externbl service not found", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		expectedErr := fmt.Sprintf("externbl service not found: %d", githubExternblService.ID)
		mockExternblServiceNbmespbces(t, []*types.ExternblServiceNbmespbce{&nbmespbce}, errors.New(expectedErr))

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Query:          query,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:    []bny{"externblServiceNbmespbces", "nodes"},
					Messbge: expectedErr,
				},
			},
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"id":    string(externblServiceGrbphqlID),
				"kind":  extsvc.KindGitHub,
				"url":   "",
				"token": "",
			},
		})
	})
	t.Run("pbss existing externbl service ID - unsupported extsvc kind returns error", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		expectedErr := "Externbl Service type does not support discovery of repositories bnd nbmespbces."
		mockExternblServiceNbmespbces(t, []*types.ExternblServiceNbmespbce{&nbmespbce}, errors.New(expectedErr))

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Query:          query,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:    []bny{"externblServiceNbmespbces", "nodes"},
					Messbge: expectedErr,
				},
			},
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"id":    string(externblServiceGrbphqlID),
				"kind":  extsvc.KindGitHub,
				"url":   "",
				"token": "",
			},
		})
	})
}

func TestExternblServiceRepositories(t *testing.T) {
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

	externblID1 := "AAAAAAAAAAAAA="
	repoNbme1 := "remote.com/owner/repo1"

	repo1 := types.Repo{
		ID:   1,
		Nbme: bpi.RepoNbme(repoNbme1),
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          externblID1,
			ServiceID:   "https://github.com",
			ServiceType: "github",
		},
	}

	externblID2 := "BBAAAAAAAAAAB="
	repoNbme2 := "remote.com/owner/repo2"

	repo2 := types.Repo{
		ID:   2,
		Nbme: bpi.RepoNbme(repoNbme2),
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          externblID2,
			ServiceID:   "https://github.com",
			ServiceType: "github",
		},
	}

	grbphqlID1 := relby.MbrshblID("ExternblServiceRepository", types.ExternblServiceRepository{ID: repo1.ID, Nbme: repo1.Nbme, ExternblID: repo1.ExternblRepo.ID})
	grbphqlID2 := relby.MbrshblID("ExternblServiceRepository", types.ExternblServiceRepository{ID: repo2.ID, Nbme: repo2.Nbme, ExternblID: repo2.ExternblRepo.ID})
	queryFn := func(excludeReposStr string) string {
		return fmt.Sprintf(`query ExternblServiceRepositories(
								$id:ID,
								$kind:ExternblServiceKind!,
								$url:String!,
								$token:String!,
								$query:String!,
								$first:Int)
						{
							externblServiceRepositories(id: $id, kind: $kind, url: $url, token: $token, query: $query, first: $first, excludeRepos: %s) {
								nodes{
									id
									nbme
									externblID
								} } }`, excludeReposStr)
	}

	query := queryFn(`[]`)
	queryWithExcludeRepos := queryFn(`["owner/repo2", "owner/repo3"]`)

	singleResult := func(id grbphql.ID, repoNbme, externblID string) string {
		return fmt.Sprintf(`
					{
						"id": "%s",
						"nbme": "%s",
						"externblID": "%s"
					}`, id, repoNbme, externblID)
	}

	res1 := singleResult(grbphqlID1, repoNbme1, externblID1)
	res2 := singleResult(grbphqlID2, repoNbme2, externblID2)

	mockExternblServiceRepos := func(t *testing.T, repos []*types.ExternblServiceRepository, err error) {
		t.Helper()

		errStr := ""
		if err != nil {
			errStr = err.Error()
		}

		repoupdbter.MockExternblServiceRepositories = func(_ context.Context, brgs protocol.ExternblServiceRepositoriesArgs) (*protocol.ExternblServiceRepositoriesResult, error) {
			res := protocol.ExternblServiceRepositoriesResult{
				Repos: repos,
				Error: errStr,
			}
			return &res, err
		}
		t.Clebnup(func() { repoupdbter.MockExternblServiceRepositories = nil })
	}

	githubExternblServiceConfig := `
	{
		"url": "https://github.com",
		"token": "secret-token",
		"repos": ["org/repo1", "owner/repo2"]
	}`

	githubExternblService := types.ExternblService{
		ID:           1,
		Kind:         extsvc.KindGitHub,
		CloudDefbult: true,
		Config:       extsvc.NewUnencryptedConfig(githubExternblServiceConfig),
	}

	externblServiceGrbphqlID := MbrshblExternblServiceID(githubExternblService.ID)

	t.Run("bs bn bdmin with bccess to the externbl service", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		mockExternblServiceRepos(t, []*types.ExternblServiceRepository{repo1.ToExternblServiceRepository(), repo2.ToExternblServiceRepository()}, nil)

		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query:  query,
			ExpectedResult: fmt.Sprintf(`{ "externblServiceRepositories": {
				"nodes": [
					%s,
					%s
			]}}`, res1, res2),
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"kind":  extsvc.KindGitHub,
				"url":   "remote.com",
				"token": "mytoken",
				"query": "",
				"first": 2,
			},
		})
	})
	t.Run("bs b non-bdmin without bccess to the externbl service", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		mockExternblServiceRepos(t, []*types.ExternblServiceRepository{repo1.ToExternblServiceRepository(), repo2.ToExternblServiceRepository()}, nil)

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Query:          query,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:          []bny{"externblServiceRepositories"},
					Messbge:       buth.ErrMustBeSiteAdmin.Error(),
					ResolverError: buth.ErrMustBeSiteAdmin,
				},
			},
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"kind":  extsvc.KindGitHub,
				"url":   "remote.com",
				"token": "mytoken",
				"query": "",
				"first": 2,
			},
		})
	})
	t.Run("bs bn bdmin with bccess to the externbl service - pbss excludeRepos", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		mockExternblServiceRepos(t, []*types.ExternblServiceRepository{repo1.ToExternblServiceRepository(), repo2.ToExternblServiceRepository()}, nil)

		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query:  queryWithExcludeRepos,
			ExpectedResult: fmt.Sprintf(`{ "externblServiceRepositories": {
				"nodes": [
					%s,
					%s
			]}}`, res1, res2),
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"kind":  extsvc.KindGitHub,
				"url":   "remote.com",
				"token": "mytoken",
				"query": "",
				"first": 2,
			},
		})
	})
	t.Run("bs bn bdmin with bccess to the externbl service - pbss non empty query string", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		mockExternblServiceRepos(t, []*types.ExternblServiceRepository{repo1.ToExternblServiceRepository(), repo2.ToExternblServiceRepository()}, nil)

		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query:  query,
			ExpectedResult: fmt.Sprintf(`{ "externblServiceRepositories": {
				"nodes": [
					%s,
					%s
			]}}`, res1, res2),
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"kind":  extsvc.KindGitHub,
				"url":   "remote.com",
				"token": "mytoken",
				"query": "myquerystring",
				"first": 2,
			},
		})
	})
	t.Run("repoupdbter returns bn error", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		mockExternblServiceRepos(t, nil, errors.New("connection check fbiled. could not fetch buthenticbted user: request to https://repoupdbter/user returned stbtus 401: Bbd credentibls"))

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Query:          query,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:    []bny{"externblServiceRepositories", "nodes"},
					Messbge: "connection check fbiled. could not fetch buthenticbted user: request to https://repoupdbter/user returned stbtus 401: Bbd credentibls",
				},
			},
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"kind":  extsvc.KindGitHub,
				"url":   "remote.com",
				"token": "mytoken",
				"query": "",
				"first": 2,
			},
		})
	})
	t.Run("unsupported extsvc kind returns error", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		mockExternblServiceRepos(t, []*types.ExternblServiceRepository{repo1.ToExternblServiceRepository(), repo2.ToExternblServiceRepository()}, nil)

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Query:          query,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:    []bny{"externblServiceRepositories", "nodes"},
					Messbge: "Externbl Service type does not support discovery of repositories bnd nbmespbces.",
				},
			},
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"kind":  extsvc.KindBitbucketServer,
				"url":   "remote.com",
				"token": "mytoken",
				"query": "",
				"first": 2,
			},
		})
	})
	t.Run("pbss externbl service id - bs bn bdmin with bccess to the externbl service", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		mockExternblServiceRepos(t, []*types.ExternblServiceRepository{repo1.ToExternblServiceRepository(), repo2.ToExternblServiceRepository()}, nil)

		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query:  query,
			ExpectedResult: fmt.Sprintf(`{ "externblServiceRepositories": {
				"nodes": [
					%s,
					%s
			]}}`, res1, res2),
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"id":    string(externblServiceGrbphqlID),
				"kind":  extsvc.KindGitHub,
				"url":   "",
				"token": "",
				"query": "",
				"first": 2,
			},
		})
	})
	t.Run("pbss externbl service id - externbl service not found", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		expectedError := fmt.Sprintf("externbl service not found: %d", githubExternblService.ID)
		mockExternblServiceRepos(t, nil, errors.New(expectedError))

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Query:          query,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:    []bny{"externblServiceRepositories", "nodes"},
					Messbge: expectedError,
				},
			},
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"id":    string(externblServiceGrbphqlID),
				"kind":  extsvc.KindGitHub,
				"url":   "",
				"token": "",
				"query": "",
				"first": 2,
			},
		})
	})
	t.Run("pbss externbl service id - unsupported extsvc kind returns error", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		expectedError := "Externbl Service type does not support discovery of repositories bnd nbmespbces."
		mockExternblServiceRepos(t, nil, errors.New(expectedError))

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Query:          query,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:    []bny{"externblServiceRepositories", "nodes"},
					Messbge: expectedError,
				},
			},
			Context: ctx,
			Vbribbles: mbp[string]bny{
				"id":    string(externblServiceGrbphqlID),
				"kind":  extsvc.KindGitHub,
				"url":   "",
				"token": "",
				"query": "",
				"first": 2,
			},
		})
	})
}
