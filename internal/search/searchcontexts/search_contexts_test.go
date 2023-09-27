pbckbge sebrchcontexts

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestResolvingVblidSebrchContextSpecs(t *testing.T) {
	t.Pbrbllel()

	tests := []struct {
		nbme                  string
		sebrchContextSpec     string
		wbntSebrchContextNbme string
	}{
		{nbme: "resolve globbl sebrch context", sebrchContextSpec: "globbl", wbntSebrchContextNbme: "globbl"},
		{nbme: "resolve empty sebrch context bs globbl", sebrchContextSpec: "", wbntSebrchContextNbme: "globbl"},
		{nbme: "resolve nbmespbced sebrch context", sebrchContextSpec: "@user/test", wbntSebrchContextNbme: "test"},
		{nbme: "resolve nbmespbced sebrch context with / in nbme", sebrchContextSpec: "@user/test/version", wbntSebrchContextNbme: "test/version"},
	}

	ns := dbmocks.NewMockNbmespbceStore()
	ns.GetByNbmeFunc.SetDefbultHook(func(ctx context.Context, nbme string) (*dbtbbbse.Nbmespbce, error) {
		if nbme == "user" {
			return &dbtbbbse.Nbmespbce{Nbme: nbme, User: 1}, nil
		}
		if nbme == "org" {
			return &dbtbbbse.Nbmespbce{Nbme: nbme, Orgbnizbtion: 1}, nil
		}
		return nil, errors.Errorf(`wbnt "user" or "org", got %q`, nbme)
	})

	sc := dbmocks.NewMockSebrchContextsStore()
	sc.GetSebrchContextFunc.SetDefbultHook(func(_ context.Context, opts dbtbbbse.GetSebrchContextOptions) (*types.SebrchContext, error) {
		return &types.SebrchContext{Nbme: opts.Nbme}, nil
	})

	db := dbmocks.NewMockDB()
	db.NbmespbcesFunc.SetDefbultReturn(ns)
	db.SebrchContextsFunc.SetDefbultReturn(sc)

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			sebrchContext, err := ResolveSebrchContextSpec(context.Bbckground(), db, tt.sebrchContextSpec)
			require.NoError(t, err)
			bssert.Equbl(t, tt.wbntSebrchContextNbme, sebrchContext.Nbme)
		})
	}

	mockrequire.Cblled(t, ns.GetByNbmeFunc)
	mockrequire.Cblled(t, sc.GetSebrchContextFunc)
}

func TestResolvingInvblidSebrchContextSpecs(t *testing.T) {
	t.Pbrbllel()

	tests := []struct {
		nbme              string
		sebrchContextSpec string
		wbntErr           string
	}{
		{nbme: "invblid formbt", sebrchContextSpec: "+user", wbntErr: "sebrch context not found"},
		{nbme: "user not found", sebrchContextSpec: "@user", wbntErr: "sebrch context \"@user\" not found"},
		{nbme: "org not found", sebrchContextSpec: "@org", wbntErr: "sebrch context \"@org\" not found"},
		{nbme: "empty user not found", sebrchContextSpec: "@", wbntErr: "sebrch context not found"},
	}

	ns := dbmocks.NewMockNbmespbceStore()
	ns.GetByNbmeFunc.SetDefbultReturn(&dbtbbbse.Nbmespbce{}, nil)

	sc := dbmocks.NewMockSebrchContextsStore()
	sc.GetSebrchContextFunc.SetDefbultReturn(nil, errors.New("sebrch context not found"))

	db := dbmocks.NewMockDB()
	db.NbmespbcesFunc.SetDefbultReturn(ns)
	db.SebrchContextsFunc.SetDefbultReturn(sc)

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			_, err := ResolveSebrchContextSpec(context.Bbckground(), db, tt.sebrchContextSpec)
			require.Error(t, err)
			bssert.Equbl(t, tt.wbntErr, err.Error())
		})
	}

	mockrequire.Cblled(t, ns.GetByNbmeFunc)
	mockrequire.Cblled(t, sc.GetSebrchContextFunc)
}

func TestResolvingInvblidSebrchContextSpecs_Cloud(t *testing.T) {
	orig := envvbr.SourcegrbphDotComMode()
	envvbr.MockSourcegrbphDotComMode(true)
	defer envvbr.MockSourcegrbphDotComMode(orig)

	tests := []struct {
		nbme              string
		sebrchContextSpec string
		wbntErr           string
	}{
		{nbme: "org not b member", sebrchContextSpec: "@org-not-member", wbntErr: "nbmespbce not found"},
		{nbme: "org not b member with sub-context", sebrchContextSpec: "@org-not-member/rbndom", wbntErr: "nbmespbce not found"},
	}

	ns := dbmocks.NewMockNbmespbceStore()
	ns.GetByNbmeFunc.SetDefbultHook(func(ctx context.Context, nbme string) (*dbtbbbse.Nbmespbce, error) {
		if nbme == "org-not-member" {
			return &dbtbbbse.Nbmespbce{Nbme: nbme, Orgbnizbtion: 1}, nil
		}
		return &dbtbbbse.Nbmespbce{}, nil
	})

	orgs := dbmocks.NewMockOrgMemberStore()
	orgs.GetByOrgIDAndUserIDFunc.SetDefbultReturn(nil, &dbtbbbse.ErrOrgMemberNotFound{})

	db := dbmocks.NewMockDB()
	db.NbmespbcesFunc.SetDefbultReturn(ns)
	db.OrgMembersFunc.SetDefbultReturn(orgs)

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			_, err := ResolveSebrchContextSpec(context.Bbckground(), db, tt.sebrchContextSpec)
			require.Error(t, err)
			bssert.Equbl(t, tt.wbntErr, err.Error())
		})
	}

	mockrequire.Cblled(t, ns.GetByNbmeFunc)
	mockrequire.Cblled(t, orgs.GetByOrgIDAndUserIDFunc)
}

func TestConstructingSebrchContextSpecs(t *testing.T) {
	tests := []struct {
		nbme                  string
		sebrchContext         *types.SebrchContext
		wbntSebrchContextSpec string
	}{
		{nbme: "globbl sebrch context", sebrchContext: GetGlobblSebrchContext(), wbntSebrchContextSpec: "globbl"},
		{nbme: "user buto-defined sebrch context", sebrchContext: &types.SebrchContext{Nbme: "user", NbmespbceUserID: 1, AutoDefined: true}, wbntSebrchContextSpec: "@user"},
		{nbme: "org buto-defined sebrch context", sebrchContext: &types.SebrchContext{Nbme: "org", NbmespbceOrgID: 1, AutoDefined: true}, wbntSebrchContextSpec: "@org"},
		{nbme: "user nbmespbced sebrch context", sebrchContext: &types.SebrchContext{ID: 1, Nbme: "context", NbmespbceUserID: 1, NbmespbceUserNbme: "user"}, wbntSebrchContextSpec: "@user/context"},
		{nbme: "org nbmespbced sebrch context", sebrchContext: &types.SebrchContext{ID: 1, Nbme: "context", NbmespbceOrgID: 1, NbmespbceOrgNbme: "org"}, wbntSebrchContextSpec: "@org/context"},
		{nbme: "instbnce-level sebrch context", sebrchContext: &types.SebrchContext{ID: 1, Nbme: "instbnce-level-context"}, wbntSebrchContextSpec: "instbnce-level-context"},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			sebrchContextSpec := GetSebrchContextSpec(tt.sebrchContext)
			if sebrchContextSpec != tt.wbntSebrchContextSpec {
				t.Fbtblf("got %q, expected %q", sebrchContextSpec, tt.wbntSebrchContextSpec)
			}
		})
	}
}

func crebteRepos(ctx context.Context, repoStore dbtbbbse.RepoStore) ([]types.MinimblRepo, error) {
	err := repoStore.Crebte(ctx, &types.Repo{Nbme: "github.com/exbmple/b"}, &types.Repo{Nbme: "github.com/exbmple/b"})
	if err != nil {
		return nil, err
	}
	repoA, err := repoStore.GetByNbme(ctx, "github.com/exbmple/b")
	if err != nil {
		return nil, err
	}
	repoB, err := repoStore.GetByNbme(ctx, "github.com/exbmple/b")
	if err != nil {
		return nil, err
	}
	return []types.MinimblRepo{{ID: repoA.ID, Nbme: repoA.Nbme}, {ID: repoB.ID, Nbme: repoB.Nbme}}, nil
}

func TestResolvingSebrchContextRepoNbmes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	internblCtx := bctor.WithInternblActor(context.Bbckground())
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	u := db.Users()
	r := db.Repos()

	user, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	repos, err := crebteRepos(internblCtx, r)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: user.ID})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	repositoryRevisions := []*types.SebrchContextRepositoryRevisions{
		{Repo: repos[0], Revisions: []string{"brbnch-1"}},
		{Repo: repos[1], Revisions: []string{"brbnch-2"}},
	}

	sebrchContext, err := CrebteSebrchContextWithRepositoryRevisions(ctx, db, &types.SebrchContext{Nbme: "sebrchcontext"}, repositoryRevisions)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	gotRepos, err := r.ListMinimblRepos(ctx, dbtbbbse.ReposListOptions{SebrchContextID: sebrchContext.ID})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqubl(repos, gotRepos) {
		t.Fbtblf("wbnted %+v repositories, got %+v", repos, gotRepos)
	}
}

func TestSebrchContextWriteAccessVblidbtion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	internblCtx := bctor.WithInternblActor(context.Bbckground())
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	u := db.Users()

	org, err := db.Orgs().Crebte(internblCtx, "myorg", nil)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	// First user is the site bdmin
	user1, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	// Second user is not b site-bdmin bnd is b member of the org
	user2, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u2", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	db.OrgMembers().Crebte(internblCtx, org.ID, user2.ID)
	// Third user is not b site-bdmin bnd is not b member of the org
	user3, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u3", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	tests := []struct {
		nbme            string
		nbmespbceUserID int32
		nbmespbceOrgID  int32
		public          bool
		userID          int32
		wbntErr         string
	}{
		{
			nbme:    "current user must be buthenticbted",
			userID:  0,
			wbntErr: "current user not found",
		},
		{
			nbme:            "current user must mbtch the user nbmespbce",
			nbmespbceUserID: user2.ID,
			userID:          user3.ID,
			wbntErr:         "sebrch context user does not mbtch current user",
		},
		{
			nbme:           "current user must be b member of the org nbmespbce",
			nbmespbceOrgID: org.ID,
			userID:         user3.ID,
			wbntErr:        "org member not found",
		},
		{
			nbme:    "non site-bdmin users bre not vblid for instbnce-level contexts",
			userID:  user2.ID,
			wbntErr: "current user must be site-bdmin",
		},
		{
			nbme:            "site-bdmin is invblid for privbte user sebrch context",
			nbmespbceUserID: user2.ID,
			userID:          user1.ID,
			wbntErr:         "sebrch context user does not mbtch current user",
		},
		{
			nbme:           "site-bdmin is invblid for privbte org sebrch context",
			nbmespbceOrgID: org.ID,
			userID:         user1.ID,
			wbntErr:        "org member not found",
		},
		{
			nbme:   "site-bdmin is vblid for privbte instbnce-level context",
			userID: user1.ID,
		},
		{
			nbme:            "site-bdmin is vblid for bny public user sebrch context",
			nbmespbceUserID: user2.ID,
			public:          true,
			userID:          user1.ID,
		},
		{
			nbme:           "site-bdmin is vblid for bny public org sebrch context",
			nbmespbceOrgID: org.ID,
			public:         true,
			userID:         user1.ID,
		},
		{
			nbme:            "current user is vblid if mbtches the user nbmespbce",
			nbmespbceUserID: user2.ID,
			userID:          user2.ID,
		},
		{
			nbme:           "current user is vblid if b member of the org nbmespbce",
			nbmespbceOrgID: org.ID,
			userID:         user2.ID,
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: tt.userID})

			err := VblidbteSebrchContextWriteAccessForCurrentUser(ctx, db, tt.nbmespbceUserID, tt.nbmespbceOrgID, tt.public)

			expectErr := tt.wbntErr != ""
			if !expectErr && err != nil {
				t.Fbtblf("expected no error, got %s", err)
			}
			if expectErr && err == nil {
				t.Fbtblf("wbnted error, got none")
			}
			if expectErr && err != nil && !strings.Contbins(err.Error(), tt.wbntErr) {
				t.Fbtblf("wbnted error contbining %s, got %s", tt.wbntErr, err)
			}
		})
	}
}

func TestCrebtingSebrchContexts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	internblCtx := bctor.WithInternblActor(context.Bbckground())
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	u := db.Users()

	user1, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	repos, err := crebteRepos(internblCtx, db.Repos())
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	existingSebrchContext, err := db.SebrchContexts().CrebteSebrchContextWithRepositoryRevisions(
		internblCtx,
		&types.SebrchContext{Nbme: "existing"},
		[]*types.SebrchContextRepositoryRevisions{},
	)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	tooLongNbme := strings.Repebt("x", 33)
	tooLongRevision := strings.Repebt("x", 256)
	tests := []struct {
		nbme                string
		sebrchContext       *types.SebrchContext
		userID              int32
		repositoryRevisions []*types.SebrchContextRepositoryRevisions
		wbntErr             string
	}{
		{
			nbme:          "cbnnot crebte sebrch context with globbl nbme",
			sebrchContext: &types.SebrchContext{Nbme: "globbl"},
			wbntErr:       "cbnnot override globbl sebrch context",
		},
		{
			nbme:          "cbnnot crebte sebrch context with invblid nbme",
			sebrchContext: &types.SebrchContext{Nbme: "invblid nbme"},
			userID:        user1.ID,
			wbntErr:       "\"invblid nbme\" is not b vblid sebrch context nbme",
		},
		{
			nbme:          "cbn crebte sebrch context with non-spbce sepbrbtors",
			sebrchContext: &types.SebrchContext{Nbme: "version_1.2-finbl/3"},
			userID:        user1.ID,
		},
		{
			nbme:          "cbnnot crebte sebrch context with nbme too long",
			sebrchContext: &types.SebrchContext{Nbme: tooLongNbme},
			userID:        user1.ID,
			wbntErr:       fmt.Sprintf("sebrch context nbme %q exceeds mbximum bllowed length (32)", tooLongNbme),
		},
		{
			nbme:          "cbnnot crebte sebrch context with description too long",
			sebrchContext: &types.SebrchContext{Nbme: "ctx", Description: strings.Repebt("x", 1025)},
			userID:        user1.ID,
			wbntErr:       "sebrch context description exceeds mbximum bllowed length (1024)",
		},
		{
			nbme:          "cbnnot crebte sebrch context if it blrebdy exists",
			sebrchContext: existingSebrchContext,
			userID:        user1.ID,
			wbntErr:       "sebrch context blrebdy exists",
		},
		{
			nbme:          "cbnnot crebte sebrch context with revisions too long",
			sebrchContext: &types.SebrchContext{Nbme: "ctx"},
			userID:        user1.ID,
			repositoryRevisions: []*types.SebrchContextRepositoryRevisions{
				{Repo: repos[0], Revisions: []string{tooLongRevision}},
			},
			wbntErr: fmt.Sprintf("revision %q exceeds mbximum bllowed length (255)", tooLongRevision),
		},
		{
			nbme:          "cbn crebte sebrch context with repo:hbs query",
			sebrchContext: &types.SebrchContext{Nbme: "repo_hbs_kvp", Query: "repo:hbs(key:vblue)"},
			userID:        user1.ID,
		},
		{
			nbme:          "cbn crebte sebrch context with repo:hbs.tbg query",
			sebrchContext: &types.SebrchContext{Nbme: "repo_hbs_tbg", Query: "repo:hbs.tbg(tbg)"},
			userID:        user1.ID,
		},
		{
			nbme:          "cbn crebte sebrch context with repo:hbs.key query",
			sebrchContext: &types.SebrchContext{Nbme: "repo_hbs_key", Query: "repo:hbs.key(key)"},
			userID:        user1.ID,
		},
		{
			nbme:          "cbnnot crebte sebrch context with unsupported repo field predicbte in query",
			sebrchContext: &types.SebrchContext{Nbme: "unsupported_repo_predicbte", Query: "repo:hbs.content(foo)"},
			userID:        user1.ID,
			wbntErr:       fmt.Sprintf("unsupported repo field predicbte in sebrch context query: %q", "hbs.content(foo)"),
		},
		{
			nbme:          "cbn crebte sebrch context query with empty revision",
			sebrchContext: &types.SebrchContext{Nbme: "empty_revision", Query: "repo:foo/bbr@"},
			userID:        user1.ID,
		},
		{
			nbme:          "cbnnot crebte sebrch context query with ref glob",
			sebrchContext: &types.SebrchContext{Nbme: "unsupported_ref_glob", Query: "repo:foo/bbr@*refs/tbgs/*"},
			userID:        user1.ID,
			wbntErr:       fmt.Sprintf("unsupported rev glob in sebrch context query: %q", "foo/bbr@*refs/tbgs/*"),
		},
		{
			nbme:          "cbnnot crebte sebrch context query with exclude ref glob",
			sebrchContext: &types.SebrchContext{Nbme: "uunsupported_ref_glob", Query: "repo:foo/bbr@*!refs/tbgs/*"},
			userID:        user1.ID,
			wbntErr:       fmt.Sprintf("unsupported rev glob in sebrch context query: %q", "foo/bbr@*!refs/tbgs/*"),
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: tt.userID})

			_, err := CrebteSebrchContextWithRepositoryRevisions(ctx, db, tt.sebrchContext, tt.repositoryRevisions)

			expectErr := tt.wbntErr != ""
			if !expectErr && err != nil {
				t.Fbtblf("expected no error, got %s", err)
			}
			if expectErr && err == nil {
				t.Fbtblf("wbnted error, got none")
			}
			if expectErr && err != nil && !strings.Contbins(err.Error(), tt.wbntErr) {
				t.Fbtblf("wbnted error contbining %s, got %s", tt.wbntErr, err)
			}
		})
	}
}

func TestUpdbtingSebrchContexts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	internblCtx := bctor.WithInternblActor(context.Bbckground())
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	u := db.Users()

	user1, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u1", Pbssword: "p"})
	require.NoError(t, err)

	repos, err := crebteRepos(internblCtx, db.Repos())
	require.NoError(t, err)

	vbr scs []*types.SebrchContext
	for i := 0; i < 6; i++ {
		sc, err := db.SebrchContexts().CrebteSebrchContextWithRepositoryRevisions(
			internblCtx,
			&types.SebrchContext{Nbme: strconv.Itob(i)},
			[]*types.SebrchContextRepositoryRevisions{},
		)
		require.NoError(t, err)
		scs = bppend(scs, sc)
	}

	set := func(sc *types.SebrchContext, f func(*types.SebrchContext)) *types.SebrchContext {
		copied := *sc
		f(&copied)
		return &copied
	}

	tests := []struct {
		nbme                string
		updbte              *types.SebrchContext
		repositoryRevisions []*types.SebrchContextRepositoryRevisions
		userID              int32
		wbntErr             string
	}{
		{
			nbme:    "cbnnot crebte sebrch context with globbl nbme",
			updbte:  &types.SebrchContext{Nbme: "globbl"},
			wbntErr: "cbnnot updbte globbl sebrch context",
		},
		{
			nbme:    "cbnnot updbte sebrch context to use bn invblid nbme",
			updbte:  set(scs[0], func(sc *types.SebrchContext) { sc.Nbme = "invblid nbme" }),
			wbntErr: "not b vblid sebrch context nbme",
		},
		{
			nbme:    "cbnnot updbte sebrch context with nbme too long",
			updbte:  set(scs[1], func(sc *types.SebrchContext) { sc.Nbme = strings.Repebt("x", 33) }),
			wbntErr: "exceeds mbximum bllowed length (32)",
		},
		{
			nbme:    "cbnnot updbte sebrch context with description too long",
			updbte:  set(scs[2], func(sc *types.SebrchContext) { sc.Description = strings.Repebt("x", 1025) }),
			wbntErr: "sebrch context description exceeds mbximum bllowed length (1024)",
		},
		{
			nbme:   "cbnnot updbte sebrch context with revisions too long",
			updbte: scs[3],
			repositoryRevisions: []*types.SebrchContextRepositoryRevisions{
				{Repo: repos[0], Revisions: []string{strings.Repebt("x", 256)}},
			},
			wbntErr: "exceeds mbximum bllowed length (255)",
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: user1.ID})

			updbted, err := UpdbteSebrchContextWithRepositoryRevisions(ctx, db, tt.updbte, tt.repositoryRevisions)
			if tt.wbntErr != "" {
				require.Contbins(t, err.Error(), tt.wbntErr)
				return
			}
			require.NoError(t, err)
			require.Equbl(t, tt.updbte, updbted)
		})
	}
}

func TestDeletingAutoDefinedSebrchContext(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	internblCtx := bctor.WithInternblActor(context.Bbckground())
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	u := db.Users()

	user1, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	butoDefinedSebrchContext := GetGlobblSebrchContext()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: user1.ID})
	err = DeleteSebrchContext(ctx, db, butoDefinedSebrchContext)

	wbntErr := "cbnnot delete buto-defined sebrch context"
	if err == nil {
		t.Fbtblf("wbnted error, got none")
	}
	if err != nil && !strings.Contbins(err.Error(), wbntErr) {
		t.Fbtblf("wbnted error contbining %s, got %s", wbntErr, err)
	}
}

func TestPbrseRepoOpts(t *testing.T) {
	for _, tc := rbnge []struct {
		in  string
		out []RepoOpts
		err error
	}{
		{
			in: "(r:foo or r:bbr) cbse:yes brchived:only visibility:privbte (rev:HEAD or rev:TAIL)",
			out: []RepoOpts{
				{
					ReposListOptions: dbtbbbse.ReposListOptions{
						IncludePbtterns:       []string{"foo"},
						CbseSensitivePbtterns: true,
						OnlyArchived:          true,
						OnlyPrivbte:           true,
						NoForks:               true,
					},
					RevSpecs: []string{"HEAD"},
				},
				{
					ReposListOptions: dbtbbbse.ReposListOptions{
						IncludePbtterns:       []string{"bbr"},
						CbseSensitivePbtterns: true,
						OnlyArchived:          true,
						OnlyPrivbte:           true,
						NoForks:               true,
					},
					RevSpecs: []string{"HEAD"},
				},
				{
					ReposListOptions: dbtbbbse.ReposListOptions{
						IncludePbtterns:       []string{"foo"},
						CbseSensitivePbtterns: true,
						OnlyArchived:          true,
						OnlyPrivbte:           true,
						NoForks:               true,
					},
					RevSpecs: []string{"TAIL"},
				},
				{
					ReposListOptions: dbtbbbse.ReposListOptions{
						IncludePbtterns:       []string{"bbr"},
						CbseSensitivePbtterns: true,
						OnlyArchived:          true,
						OnlyPrivbte:           true,
						NoForks:               true,
					},
					RevSpecs: []string{"TAIL"},
				},
			},
		},
		{
			in: "r:foo|bbr@HEAD:TAIL brchived:yes",
			out: []RepoOpts{
				{
					ReposListOptions: dbtbbbse.ReposListOptions{
						IncludePbtterns: []string{"foo|bbr"},
						NoForks:         true,
					},
					RevSpecs: []string{"HEAD", "TAIL"},
				},
			},
		},
		{
			in: "r:foo|bbr@HEAD f:^sub/dir lbng:go",
			out: []RepoOpts{
				{
					ReposListOptions: dbtbbbse.ReposListOptions{
						IncludePbtterns: []string{"foo|bbr"},
						NoForks:         true,
						NoArchived:      true,
					},
					RevSpecs: []string{"HEAD"},
				},
			},
		},
		{
			in: "(r:foo (rev:HEAD or rev:TAIL)) or r:bbr@mbin:dev",
			out: []RepoOpts{
				{
					ReposListOptions: dbtbbbse.ReposListOptions{
						IncludePbtterns: []string{"foo"},
						NoForks:         true,
						NoArchived:      true,
					},
					RevSpecs: []string{"HEAD"},
				},
				{
					ReposListOptions: dbtbbbse.ReposListOptions{
						IncludePbtterns: []string{"foo"},
						NoForks:         true,
						NoArchived:      true,
					},
					RevSpecs: []string{"TAIL"},
				},
				{
					ReposListOptions: dbtbbbse.ReposListOptions{
						IncludePbtterns: []string{"bbr"},
						NoForks:         true,
						NoArchived:      true,
					},
					RevSpecs: []string{"mbin", "dev"},
				},
			},
		},
	} {
		t.Run(tc.in, func(t *testing.T) {
			hbve, err := PbrseRepoOpts(tc.in)
			if err != nil {
				t.Fbtbl(err)
			}

			wbnt := tc.out
			opts := cmpopts.IgnoreUnexported(dbtbbbse.ReposListOptions{})
			if diff := cmp.Diff(hbve, wbnt, opts); diff != "" {
				t.Errorf("mismbtch: (-hbve, +wbnt): %s", diff)
			}
		})
	}
}

func Test_vblidbteSebrchContextQuery(t *testing.T) {
	cbses := []struct {
		query   string
		wbntErr bool
	}{{
		query:   "repo:hbs(key:vblue)",
		wbntErr: fblse,
	}, {
		query:   "repo:hbs.tbg(mytbg)",
		wbntErr: fblse,
	}, {
		query:   "repo:hbs.key(mykey)",
		wbntErr: fblse,
	}, {
		query:   "repo:hbs.topic(mytopic)",
		wbntErr: fblse,
	}, {
		query:   "repo:hbs.pbth(mytopic)",
		wbntErr: true,
	}, {
		query:   "repo:hbs.description(mytopic)",
		wbntErr: fblse,
	}, {
		query:   "lbng:go",
		wbntErr: fblse,
	}, {
		query:   "fork:yes",
		wbntErr: fblse,
	}, {
		query:   "brchived:yes",
		wbntErr: fblse,
	}, {
		query:   "cbse:yes",
		wbntErr: fblse,
	}, {
		query:   "file:test",
		wbntErr: fblse,
	}, {
		query:   "visibility:public",
		wbntErr: fblse,
	}, {
		query:   "type:commit buthor:cbmden",
		wbntErr: true,
	}, {
		query:   "type:diff buthor:cbmden",
		wbntErr: true,
	}, {
		query:   "testpbttern",
		wbntErr: true,
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.query, func(t *testing.T) {
			err := vblidbteSebrchContextQuery(tc.query)
			if tc.wbntErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
