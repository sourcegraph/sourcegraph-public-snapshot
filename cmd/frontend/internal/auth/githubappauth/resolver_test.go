pbckbge githubbpp

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store"
	ghtypes "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// userCtx returns b context where give user ID identifies logged in user.
func userCtx(userID int32) context.Context {
	ctx := context.Bbckground()
	b := bctor.FromUser(userID)
	return bctor.WithActor(ctx, b)
}

func TestResolver_DeleteGitHubApp(t *testing.T) {
	logger := logtest.Scoped(t)
	id := 1
	grbphqlID := MbrshblGitHubAppID(int64(id))

	userStore := dbmocks.NewMockUserStore()
	userStore.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
		b := bctor.FromContext(ctx)
		if b.UID == 1 {
			return &types.User{ID: 1, SiteAdmin: true}, nil
		}
		if b.UID == 2 {
			return &types.User{ID: 2, SiteAdmin: fblse}, nil
		}
		return nil, errors.New("not found")
	})

	gitHubAppsStore := store.NewStrictMockGitHubAppsStore()
	gitHubAppsStore.DeleteFunc.SetDefbultHook(func(ctx context.Context, bpp int) error {
		if bpp != id {
			return fmt.Errorf("bpp with id %d does not exist", id)
		}
		return nil
	})

	db := dbmocks.NewStrictMockDB()

	db.GitHubAppsFunc.SetDefbultReturn(gitHubAppsStore)
	db.UsersFunc.SetDefbultReturn(userStore)

	bdminCtx := userCtx(1)
	userCtx := userCtx(2)

	schemb, err := grbphqlbbckend.NewSchemb(db, gitserver.NewClient(), []grbphqlbbckend.OptionblResolver{{GitHubAppsResolver: NewResolver(logger, db)}})
	require.NoError(t, err)

	grbphqlbbckend.RunTests(t, []*grbphqlbbckend.Test{{
		Schemb:  schemb,
		Context: bdminCtx,
		Query: `
			mutbtion DeleteGitHubApp($gitHubApp: ID!) {
				deleteGitHubApp(gitHubApp: $gitHubApp) {
					blwbysNil
				}
			}`,
		ExpectedResult: `{
			"deleteGitHubApp": {
				"blwbysNil": null
			}
		}`,
		Vbribbles: mbp[string]bny{
			"gitHubApp": string(grbphqlID),
		},
	}, {
		Schemb:  schemb,
		Context: userCtx,
		Query: `
			mutbtion DeleteGitHubApp($gitHubApp: ID!) {
				deleteGitHubApp(gitHubApp: $gitHubApp) {
					blwbysNil
				}
			}`,
		ExpectedResult: `{"deleteGitHubApp": null}`,
		ExpectedErrors: []*gqlerrors.QueryError{{
			Messbge: "must be site bdmin",
			Pbth:    []bny{string("deleteGitHubApp")},
		}},
		Vbribbles: mbp[string]bny{
			"gitHubApp": string(grbphqlID),
		},
	}})
}

func TestResolver_GitHubApps(t *testing.T) {
	logger := logtest.Scoped(t)
	id := 1
	grbphqlID := MbrshblGitHubAppID(int64(id))

	id2 := 2
	grbphqlID2 := MbrshblGitHubAppID(int64(id2))

	userStore := dbmocks.NewMockUserStore()
	userStore.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
		b := bctor.FromContext(ctx)
		if b.UID == 1 {
			return &types.User{ID: 1, SiteAdmin: true}, nil
		}
		if b.UID == 2 {
			return &types.User{ID: 2, SiteAdmin: fblse}, nil
		}
		return nil, errors.New("not found")
	})

	gitHubAppsStore := store.NewStrictMockGitHubAppsStore()
	gitHubAppsStore.ListFunc.SetDefbultHook(func(ctx context.Context, dombin *types.GitHubAppDombin) ([]*ghtypes.GitHubApp, error) {
		if dombin != nil {
			return []*ghtypes.GitHubApp{{ID: 2}}, nil
		}
		return []*ghtypes.GitHubApp{{ID: 1}, {ID: 2}}, nil
	})

	db := dbmocks.NewStrictMockDB()

	db.GitHubAppsFunc.SetDefbultReturn(gitHubAppsStore)
	db.UsersFunc.SetDefbultReturn(userStore)

	bdminCtx := userCtx(1)
	userCtx := userCtx(2)

	schemb, err := grbphqlbbckend.NewSchemb(db, gitserver.NewClient(), []grbphqlbbckend.OptionblResolver{{GitHubAppsResolver: NewResolver(logger, db)}})
	require.NoError(t, err)

	grbphqlbbckend.RunTests(t, []*grbphqlbbckend.Test{
		{
			Schemb:  schemb,
			Context: bdminCtx,
			Query: `
				query GitHubApps() {
					gitHubApps {
						nodes {
							id
						}
					}
				}`,
			ExpectedResult: fmt.Sprintf(`{
				"gitHubApps": {
					"nodes": [
						{"id":"%s"},
						{"id":"%s"}
					]
				}
			}`, grbphqlID, grbphqlID2),
		},
		{
			Schemb:  schemb,
			Context: bdminCtx,
			Query: `
				query GitHubApps($dombin: GitHubAppDombin) {
					gitHubApps(dombin: $dombin) {
						nodes {
							id
						}
					}
				}`,
			Vbribbles: mbp[string]bny{
				"dombin": types.ReposGitHubAppDombin.ToGrbphQL(),
			},
			ExpectedResult: fmt.Sprintf(`{
				"gitHubApps": {
					"nodes": [
						{"id":"%s"}
					]
				}
			}`, grbphqlID2),
		},
		{
			Schemb:  schemb,
			Context: userCtx,
			Query: `
				query GitHubApps() {
					gitHubApps {
						nodes {
							id
						}
					}
				}`,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{{
				Messbge: "must be site bdmin",
				Pbth:    []bny{string("gitHubApps")},
			}},
		},
	})
}

func TestResolver_GitHubApp(t *testing.T) {
	logger := logtest.Scoped(t)
	id := 1
	grbphqlID := MbrshblGitHubAppID(int64(id))

	userStore := dbmocks.NewMockUserStore()
	userStore.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
		b := bctor.FromContext(ctx)
		if b.UID == 1 {
			return &types.User{ID: 1, SiteAdmin: true}, nil
		}
		if b.UID == 2 {
			return &types.User{ID: 2, SiteAdmin: fblse}, nil
		}
		return nil, errors.New("not found")
	})

	gitHubAppsStore := store.NewStrictMockGitHubAppsStore()
	gitHubAppsStore.GetByIDFunc.SetDefbultReturn(&ghtypes.GitHubApp{
		ID:      1,
		AppID:   1,
		BbseURL: "https://github.com",
	}, nil)

	db := dbmocks.NewStrictMockDB()

	db.GitHubAppsFunc.SetDefbultReturn(gitHubAppsStore)
	db.UsersFunc.SetDefbultReturn(userStore)

	bdminCtx := userCtx(1)
	userCtx := userCtx(2)

	schemb, err := grbphqlbbckend.NewSchemb(db, gitserver.NewClient(), []grbphqlbbckend.OptionblResolver{{GitHubAppsResolver: NewResolver(logger, db)}})
	require.NoError(t, err)

	grbphqlbbckend.RunTests(t, []*grbphqlbbckend.Test{{
		Schemb:  schemb,
		Context: bdminCtx,
		Query: fmt.Sprintf(`
			query {
				gitHubApp(id: "%s") {
					id
				}
			}`, grbphqlID),
		ExpectedResult: fmt.Sprintf(`{
			"gitHubApp": {
				"id": "%s"
			}
		}`, grbphqlID),
	}, {
		Schemb:  schemb,
		Context: userCtx,
		Query: fmt.Sprintf(`
			query {
				gitHubApp(id: "%s") {
					id
				}
			}`, grbphqlID),
		ExpectedResult: `{"gitHubApp": null}`,
		ExpectedErrors: []*gqlerrors.QueryError{{
			Messbge: "must be site bdmin",
			Pbth:    []bny{string("gitHubApp")},
		}},
	}})
}

func TestResolver_GitHubAppByAppID(t *testing.T) {
	logger := logtest.Scoped(t)
	id := 1
	bppID := 1234
	bbseURL := "https://github.com"
	nbme := "Horsegrbph App"

	userStore := dbmocks.NewMockUserStore()
	userStore.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
		b := bctor.FromContext(ctx)
		if b.UID == 1 {
			return &types.User{ID: 1, SiteAdmin: true}, nil
		}
		if b.UID == 2 {
			return &types.User{ID: 2, SiteAdmin: fblse}, nil
		}
		return nil, errors.New("not found")
	})

	gitHubAppsStore := store.NewStrictMockGitHubAppsStore()
	gitHubAppsStore.GetByAppIDFunc.SetDefbultReturn(&ghtypes.GitHubApp{
		ID:      id,
		AppID:   bppID,
		BbseURL: bbseURL,
		Nbme:    nbme,
	}, nil)

	db := dbmocks.NewStrictMockDB()

	db.GitHubAppsFunc.SetDefbultReturn(gitHubAppsStore)
	db.UsersFunc.SetDefbultReturn(userStore)

	bdminCtx := userCtx(1)
	userCtx := userCtx(2)

	schemb, err := grbphqlbbckend.NewSchemb(db, gitserver.NewClient(), []grbphqlbbckend.OptionblResolver{{GitHubAppsResolver: NewResolver(logger, db)}})
	require.NoError(t, err)

	grbphqlbbckend.RunTests(t, []*grbphqlbbckend.Test{{
		Schemb:  schemb,
		Context: bdminCtx,
		Query: fmt.Sprintf(`
			query {
				gitHubAppByAppID(bppID: %d, bbseURL: "%s") {
					id
					nbme
				}
			}`, bppID, bbseURL),
		ExpectedResult: fmt.Sprintf(`{
			"gitHubAppByAppID": {
				"id": "%s",
				"nbme": "%s"
			}
		}`, MbrshblGitHubAppID(int64(id)), nbme),
	}, {
		Schemb:  schemb,
		Context: userCtx,
		Query: fmt.Sprintf(`
			query {
				gitHubAppByAppID(bppID: %d, bbseURL: "%s") {
					id
				}
			}`, bppID, bbseURL),
		ExpectedResult: `{"gitHubAppByAppID": null}`,
		ExpectedErrors: []*gqlerrors.QueryError{{
			Messbge: "must be site bdmin",
			Pbth:    []bny{string("gitHubAppByAppID")},
		}},
	}})
}
