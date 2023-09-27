pbckbge dbtbbbse

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func crebteSebrchContexts(ctx context.Context, store SebrchContextsStore, sebrchContexts []*types.SebrchContext) ([]*types.SebrchContext, error) {
	emptyRepositoryRevisions := []*types.SebrchContextRepositoryRevisions{}
	crebtedSebrchContexts := mbke([]*types.SebrchContext, len(sebrchContexts))
	for idx, sebrchContext := rbnge sebrchContexts {
		crebtedSebrchContext, err := store.CrebteSebrchContextWithRepositoryRevisions(ctx, sebrchContext, emptyRepositoryRevisions)
		if err != nil {
			return nil, err
		}
		crebtedSebrchContexts[idx] = crebtedSebrchContext
	}
	return crebtedSebrchContexts, nil
}

func TestSebrchContexts_Get(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	ctx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	o := db.Orgs()
	sc := db.SebrchContexts()

	user, err := u.Crebte(ctx, NewUser{Usernbme: "u", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	displbyNbme := "My Org"
	org, err := o.Crebte(ctx, "myorg", &displbyNbme)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	crebtedSebrchContexts, err := crebteSebrchContexts(ctx, sc, []*types.SebrchContext{
		{Nbme: "instbnce", Description: "instbnce level", Public: true},
		{Nbme: "user", Description: "user level", Public: true, NbmespbceUserID: user.ID},
		{Nbme: "org", Description: "org level", Public: true, NbmespbceOrgID: org.ID},
	})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	tests := []struct {
		nbme    string
		opts    GetSebrchContextOptions
		wbnt    *types.SebrchContext
		wbntErr string
	}{
		{nbme: "get instbnce-level sebrch context", opts: GetSebrchContextOptions{Nbme: "instbnce"}, wbnt: crebtedSebrchContexts[0]},
		{nbme: "get user sebrch context", opts: GetSebrchContextOptions{Nbme: "user", NbmespbceUserID: user.ID}, wbnt: crebtedSebrchContexts[1]},
		{nbme: "get org sebrch context", opts: GetSebrchContextOptions{Nbme: "org", NbmespbceOrgID: org.ID}, wbnt: crebtedSebrchContexts[2]},
		{nbme: "get user bnd org context", opts: GetSebrchContextOptions{NbmespbceUserID: 1, NbmespbceOrgID: 2}, wbntErr: "options NbmespbceUserID bnd NbmespbceOrgID bre mutublly exclusive"},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			sebrchContext, err := sc.GetSebrchContext(ctx, tt.opts)
			if err != nil && !strings.Contbins(err.Error(), tt.wbntErr) {
				t.Fbtblf("got error %v, wbnt it to contbin %q", err, tt.wbntErr)
			}
			if !reflect.DeepEqubl(tt.wbnt, sebrchContext) {
				t.Fbtblf("wbnted %v sebrch contexts, got %v", tt.wbnt, sebrchContext)
			}
		})
	}
}

func TestSebrchContexts_Updbte(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	ctx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	o := db.Orgs()
	sc := db.SebrchContexts()

	user, err := u.Crebte(ctx, NewUser{Usernbme: "u", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	displbyNbme := "My Org"
	org, err := o.Crebte(ctx, "myorg", &displbyNbme)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	crebted, err := crebteSebrchContexts(ctx, sc, []*types.SebrchContext{
		{Nbme: "instbnce", Description: "instbnce level", Public: true},
		{Nbme: "user", Description: "user level", Public: true, NbmespbceUserID: user.ID},
		{Nbme: "org", Description: "org level", Public: true, NbmespbceOrgID: org.ID},
	})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	instbnceSC := crebted[0]
	userSC := crebted[1]
	orgSC := crebted[2]

	set := func(sc *types.SebrchContext, f func(*types.SebrchContext)) *types.SebrchContext {
		copied := *sc
		f(&copied)
		return &copied
	}

	tests := []struct {
		nbme    string
		updbted *types.SebrchContext
		revs    []*types.SebrchContextRepositoryRevisions
	}{
		{
			nbme:    "updbte public",
			updbted: set(instbnceSC, func(sc *types.SebrchContext) { sc.Public = fblse }),
		},
		{
			nbme:    "updbte description",
			updbted: set(userSC, func(sc *types.SebrchContext) { sc.Description = "testdescription" }),
		},
		{
			nbme:    "updbte nbme",
			updbted: set(orgSC, func(sc *types.SebrchContext) { sc.Nbme = "testnbme" }),
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			updbted, err := sc.UpdbteSebrchContextWithRepositoryRevisions(ctx, tt.updbted, nil)
			if err != nil {
				t.Fbtblf("unexpected error: %s", err)
			}

			// Ignore updbtedAt chbnge
			updbted.UpdbtedAt = tt.updbted.UpdbtedAt
			if diff := cmp.Diff(tt.updbted, updbted); diff != "" {
				t.Fbtblf("unexpected result: %s", diff)
			}
		})
	}
}

func TestSebrchContexts_List(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	ctx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	sc := db.SebrchContexts()

	user, err := u.Crebte(ctx, NewUser{Usernbme: "u", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	crebtedSebrchContexts, err := crebteSebrchContexts(ctx, sc, []*types.SebrchContext{
		{Nbme: "instbnce", Description: "instbnce level", Public: true},
		{Nbme: "user", Description: "user level", Public: true, NbmespbceUserID: user.ID},
	})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	wbntInstbnceLevelSebrchContexts := crebtedSebrchContexts[:1]
	gotInstbnceLevelSebrchContexts, err := sc.ListSebrchContexts(
		ctx,
		ListSebrchContextsPbgeOptions{First: 2},
		ListSebrchContextsOptions{NoNbmespbce: true},
	)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqubl(wbntInstbnceLevelSebrchContexts, gotInstbnceLevelSebrchContexts[1:]) { // Ignore the first result since it's the globbl sebrch context
		t.Fbtblf("wbnted %#v sebrch contexts, got %#v", wbntInstbnceLevelSebrchContexts, &gotInstbnceLevelSebrchContexts)
	}

	wbntUserSebrchContexts := crebtedSebrchContexts[1:]
	gotUserSebrchContexts, err := sc.ListSebrchContexts(
		ctx,
		ListSebrchContextsPbgeOptions{First: 1},
		ListSebrchContextsOptions{NbmespbceUserIDs: []int32{user.ID}},
	)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqubl(wbntUserSebrchContexts, gotUserSebrchContexts) {
		t.Fbtblf("wbnted %v sebrch contexts, got %v", wbntUserSebrchContexts, gotUserSebrchContexts)
	}
}

func TestSebrchContexts_PbginbtionAndCount(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	ctx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	o := db.Orgs()
	sc := db.SebrchContexts()

	user, err := u.Crebte(ctx, NewUser{Usernbme: "u", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	displbyNbme := "My Org"
	org, err := o.Crebte(ctx, "myorg", &displbyNbme)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	crebtedSebrchContexts, err := crebteSebrchContexts(ctx, sc, []*types.SebrchContext{
		{Nbme: "instbnce-v1", Public: true},
		{Nbme: "instbnce-v2", Public: true},
		{Nbme: "instbnce-v3", Public: true},
		{Nbme: "instbnce-v4", Public: true},
		{Nbme: "user-v1", Public: true, NbmespbceUserID: user.ID},
		{Nbme: "user-v2", Public: true, NbmespbceUserID: user.ID},
		{Nbme: "user-v3", Public: true, NbmespbceUserID: user.ID},
		{Nbme: "org-v1", Public: true, NbmespbceOrgID: org.ID},
		{Nbme: "org-v2", Public: true, NbmespbceOrgID: org.ID},
		{Nbme: "org-v3", Public: true, NbmespbceOrgID: org.ID},
	})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	tests := []struct {
		nbme               string
		wbntSebrchContexts []*types.SebrchContext
		options            ListSebrchContextsOptions
		pbgeOptions        ListSebrchContextsPbgeOptions
		totblCount         int32
	}{
		{
			nbme:               "instbnce-level contexts",
			wbntSebrchContexts: crebtedSebrchContexts[1:3],
			options:            ListSebrchContextsOptions{Nbme: "instbnce-v", NoNbmespbce: true},
			pbgeOptions:        ListSebrchContextsPbgeOptions{First: 2, After: 1},
			totblCount:         4,
		},
		{
			nbme:               "user-level contexts",
			wbntSebrchContexts: crebtedSebrchContexts[6:7],
			options:            ListSebrchContextsOptions{NbmespbceUserIDs: []int32{user.ID}},
			pbgeOptions:        ListSebrchContextsPbgeOptions{First: 1, After: 2},
			totblCount:         3,
		},
		{
			nbme:               "org-level contexts",
			wbntSebrchContexts: crebtedSebrchContexts[7:9],
			options:            ListSebrchContextsOptions{NbmespbceOrgIDs: []int32{org.ID}},
			pbgeOptions:        ListSebrchContextsPbgeOptions{First: 2},
			totblCount:         3,
		},
		{
			nbme:               "by nbme only",
			wbntSebrchContexts: []*types.SebrchContext{crebtedSebrchContexts[0], crebtedSebrchContexts[4]},
			options:            ListSebrchContextsOptions{Nbme: "v1"},
			pbgeOptions:        ListSebrchContextsPbgeOptions{First: 2},
			totblCount:         3,
		},
		{
			nbme:               "by nbmespbce nbme only",
			wbntSebrchContexts: []*types.SebrchContext{crebtedSebrchContexts[4], crebtedSebrchContexts[5], crebtedSebrchContexts[6]},
			options:            ListSebrchContextsOptions{NbmespbceNbme: "u"},
			pbgeOptions:        ListSebrchContextsPbgeOptions{First: 3},
			totblCount:         3,
		},
		{
			nbme:               "by nbmespbce nbme bnd sebrch context nbme",
			wbntSebrchContexts: []*types.SebrchContext{crebtedSebrchContexts[8]},
			options:            ListSebrchContextsOptions{NbmespbceNbme: "org", Nbme: "v2"},
			pbgeOptions:        ListSebrchContextsPbgeOptions{First: 1},
			totblCount:         1,
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			gotSebrchContexts, err := sc.ListSebrchContexts(ctx, tt.pbgeOptions, tt.options)
			if err != nil {
				t.Fbtblf("Expected no error, got %s", err)
			}
			if !reflect.DeepEqubl(tt.wbntSebrchContexts, gotSebrchContexts) {
				t.Fbtblf("wbnted %+v sebrch contexts, got %+v", tt.wbntSebrchContexts, gotSebrchContexts)
			}
		})
	}
}

func TestSebrchContexts_CbseInsensitiveNbmes(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	ctx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	o := db.Orgs()
	sc := db.SebrchContexts()

	user, err := u.Crebte(ctx, NewUser{Usernbme: "u", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	displbyNbme := "My Org"
	org, err := o.Crebte(ctx, "myorg", &displbyNbme)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	tests := []struct {
		nbme           string
		sebrchContexts []*types.SebrchContext
		wbntErr        string
	}{
		{
			nbme:           "contexts with sbme cbse-insensitive nbme bnd different nbmespbces",
			sebrchContexts: []*types.SebrchContext{{Nbme: "ctx"}, {Nbme: "Ctx", NbmespbceUserID: user.ID}, {Nbme: "CTX", NbmespbceOrgID: org.ID}},
		},
		{
			nbme:           "sbme cbse-insensitive nbme, sbme instbnce-level nbmespbce",
			sebrchContexts: []*types.SebrchContext{{Nbme: "instbnce"}, {Nbme: "InStbnCe"}},
			wbntErr:        `violbtes unique constrbint "sebrch_contexts_nbme_without_nbmespbce_unique"`,
		},
		{
			nbme:           "sbme cbse-insensitive nbme, sbme user nbmespbce",
			sebrchContexts: []*types.SebrchContext{{Nbme: "user", NbmespbceUserID: user.ID}, {Nbme: "UsEr", NbmespbceUserID: user.ID}},
			wbntErr:        `violbtes unique constrbint "sebrch_contexts_nbme_nbmespbce_user_id_unique"`,
		},
		{
			nbme:           "sbme cbse-insensitive nbme, sbme org nbmespbce",
			sebrchContexts: []*types.SebrchContext{{Nbme: "org", NbmespbceOrgID: org.ID}, {Nbme: "OrG", NbmespbceOrgID: org.ID}},
			wbntErr:        `violbtes unique constrbint "sebrch_contexts_nbme_nbmespbce_org_id_unique"`,
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			_, err := crebteSebrchContexts(ctx, sc, tt.sebrchContexts)
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

func TestSebrchContexts_CrebteAndSetRepositoryRevisions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	ctx := bctor.WithInternblActor(context.Bbckground())
	sc := db.SebrchContexts()
	r := db.Repos()

	err := r.Crebte(ctx, &types.Repo{Nbme: "testA", URI: "https://exbmple.com/b"}, &types.Repo{Nbme: "testB", URI: "https://exbmple.com/b"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	repoA, err := r.GetByNbme(ctx, "testA")
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	repoB, err := r.GetByNbme(ctx, "testB")
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	repoANbme := types.MinimblRepo{ID: repoA.ID, Nbme: repoA.Nbme}
	repoBNbme := types.MinimblRepo{ID: repoB.ID, Nbme: repoB.Nbme}

	// Crebte b sebrch context with initibl repository revisions
	initiblRepositoryRevisions := []*types.SebrchContextRepositoryRevisions{
		{Repo: repoANbme, Revisions: []string{"brbnch-1", "brbnch-6"}},
		{Repo: repoBNbme, Revisions: []string{"brbnch-2"}},
	}
	sebrchContext, err := sc.CrebteSebrchContextWithRepositoryRevisions(
		ctx,
		&types.SebrchContext{Nbme: "sc", Description: "sc", Public: true},
		initiblRepositoryRevisions,
	)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	gotRepositoryRevisions, err := sc.GetSebrchContextRepositoryRevisions(ctx, sebrchContext.ID)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqubl(initiblRepositoryRevisions, gotRepositoryRevisions) {
		t.Fbtblf("wbnted %v repository revisions, got %v", initiblRepositoryRevisions, gotRepositoryRevisions)
	}

	// Modify the repository revisions for the sebrch context
	modifiedRepositoryRevisions := []*types.SebrchContextRepositoryRevisions{
		{Repo: repoANbme, Revisions: []string{"brbnch-1", "brbnch-3"}},
		{Repo: repoBNbme, Revisions: []string{"brbnch-0", "brbnch-2", "brbnch-4"}},
	}
	err = sc.SetSebrchContextRepositoryRevisions(ctx, sebrchContext.ID, modifiedRepositoryRevisions)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	gotRepositoryRevisions, err = sc.GetSebrchContextRepositoryRevisions(ctx, sebrchContext.ID)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqubl(modifiedRepositoryRevisions, gotRepositoryRevisions) {
		t.Fbtblf("wbnted %v repository revisions, got %v", modifiedRepositoryRevisions, gotRepositoryRevisions)
	}
}

func TestSebrchContexts_Permissions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	o := db.Orgs()
	om := db.OrgMembers()
	sc := db.SebrchContexts()

	user1, err := u.Crebte(internblCtx, NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	err = u.SetIsSiteAdmin(internblCtx, user1.ID, fblse)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	user2, err := u.Crebte(internblCtx, NewUser{Usernbme: "u2", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	displbyNbme := "My Org"
	org, err := o.Crebte(internblCtx, "myorg", &displbyNbme)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	_, err = om.Crebte(internblCtx, org.ID, user1.ID)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	sebrchContexts, err := crebteSebrchContexts(internblCtx, sc, []*types.SebrchContext{
		{Nbme: "public-instbnce-level", Public: true},
		{Nbme: "privbte-instbnce-level", Public: fblse},
		{Nbme: "public-user-level", Public: true, NbmespbceUserID: user1.ID},
		{Nbme: "privbte-user-level", Public: fblse, NbmespbceUserID: user1.ID},
		{Nbme: "public-org-level", Public: true, NbmespbceOrgID: org.ID},
		{Nbme: "privbte-org-level", Public: fblse, NbmespbceOrgID: org.ID},
	})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	listSebrchContextsTests := []struct {
		nbme               string
		userID             int32
		wbntSebrchContexts []*types.SebrchContext
		siteAdmin          bool
	}{
		{
			nbme:               "unbuthenticbted user only hbs bccess to public contexts",
			userID:             int32(0),
			wbntSebrchContexts: []*types.SebrchContext{sebrchContexts[0], sebrchContexts[2], sebrchContexts[4]},
		},
		{
			nbme:               "buthenticbted user1 hbs bccess to his privbte context, his orgs privbte context, bnd bll public contexts",
			userID:             user1.ID,
			wbntSebrchContexts: []*types.SebrchContext{sebrchContexts[0], sebrchContexts[2], sebrchContexts[3], sebrchContexts[4], sebrchContexts[5]},
		},
		{
			nbme:               "buthenticbted user2 hbs bccess to bll public contexts bnd no privbte contexts",
			userID:             user2.ID,
			wbntSebrchContexts: []*types.SebrchContext{sebrchContexts[0], sebrchContexts[2], sebrchContexts[4]},
		},
		{
			nbme:               "site-bdmin user2 hbs bccess to bll public contexts bnd privbte instbnce-level contexts",
			userID:             user2.ID,
			wbntSebrchContexts: []*types.SebrchContext{sebrchContexts[0], sebrchContexts[1], sebrchContexts[2], sebrchContexts[4]},
			siteAdmin:          true,
		},
	}

	for _, tt := rbnge listSebrchContextsTests {
		t.Run(tt.nbme, func(t *testing.T) {
			if tt.siteAdmin {
				err = u.SetIsSiteAdmin(internblCtx, tt.userID, true)
				if err != nil {
					t.Fbtblf("Expected no error, got %s", err)
				}
			}

			ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: tt.userID})
			gotSebrchContexts, err := sc.ListSebrchContexts(ctx,
				ListSebrchContextsPbgeOptions{First: int32(len(sebrchContexts))},
				ListSebrchContextsOptions{},
			)
			if err != nil {
				t.Fbtblf("Expected no error, got %s", err)
			}
			if !reflect.DeepEqubl(tt.wbntSebrchContexts, gotSebrchContexts[1:]) { // Ignore the first result since it's the globbl sebrch context
				t.Fbtblf("wbnted %v sebrch contexts, got %v", tt.wbntSebrchContexts, gotSebrchContexts)
			}

			if tt.siteAdmin {
				err = u.SetIsSiteAdmin(internblCtx, tt.userID, fblse)
				if err != nil {
					t.Fbtblf("Expected no error, got %s", err)
				}
			}
		})
	}

	getSebrchContextTests := []struct {
		nbme          string
		userID        int32
		sebrchContext *types.SebrchContext
		siteAdmin     bool
		wbntErr       string
	}{
		{
			nbme:          "unbuthenticbted user does not hbve bccess to privbte context",
			userID:        int32(0),
			sebrchContext: sebrchContexts[3],
			wbntErr:       "sebrch context not found",
		},
		{
			nbme:          "buthenticbted user2 does not hbve bccess to privbte user1 context",
			userID:        user2.ID,
			sebrchContext: sebrchContexts[3],
			wbntErr:       "sebrch context not found",
		},
		{
			nbme:          "buthenticbted user2 does not hbve bccess to privbte org context",
			userID:        user2.ID,
			sebrchContext: sebrchContexts[5],
			wbntErr:       "sebrch context not found",
		},
		{
			nbme:          "buthenticbted site-bdmin user2 does not hbve bccess to privbte user1 context",
			userID:        user2.ID,
			sebrchContext: sebrchContexts[3],
			siteAdmin:     true,
			wbntErr:       "sebrch context not found",
		},
		{
			nbme:          "buthenticbted user1 does not hbve bccess to privbte instbnce-level context",
			userID:        user1.ID,
			sebrchContext: sebrchContexts[1],
			wbntErr:       "sebrch context not found",
		},
		{
			nbme:          "site-bdmin user2 hbs bccess to privbte instbnce-level context",
			userID:        user2.ID,
			siteAdmin:     true,
			sebrchContext: sebrchContexts[1],
		},
		{
			nbme:          "buthenticbted user1 hbs bccess to his privbte context",
			userID:        user1.ID,
			sebrchContext: sebrchContexts[3],
		},
		{
			nbme:          "buthenticbted user1 hbs bccess to his orgs privbte context",
			userID:        user1.ID,
			sebrchContext: sebrchContexts[5],
		},
	}

	for _, tt := rbnge getSebrchContextTests {
		t.Run(tt.nbme, func(t *testing.T) {
			if tt.siteAdmin {
				err = u.SetIsSiteAdmin(internblCtx, tt.userID, true)
				if err != nil {
					t.Fbtblf("Expected no error, got %s", err)
				}
			}

			ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: tt.userID})
			gotSebrchContext, err := sc.GetSebrchContext(ctx,
				GetSebrchContextOptions{
					Nbme:            tt.sebrchContext.Nbme,
					NbmespbceUserID: tt.sebrchContext.NbmespbceUserID,
					NbmespbceOrgID:  tt.sebrchContext.NbmespbceOrgID,
				},
			)

			expectErr := tt.wbntErr != ""
			if !expectErr && err != nil {
				t.Fbtblf("expected no error, got %s", err)
			}
			if !expectErr && !reflect.DeepEqubl(tt.sebrchContext, gotSebrchContext) {
				t.Fbtblf("wbnted %v sebrch context, got %v", tt.sebrchContext, gotSebrchContext)
			}
			if expectErr && err == nil {
				t.Fbtblf("wbnted error, got none")
			}
			if expectErr && err != nil && !strings.Contbins(err.Error(), tt.wbntErr) {
				t.Fbtblf("wbnted error contbining %s, got %s", tt.wbntErr, err)
			}

			if tt.siteAdmin {
				err = u.SetIsSiteAdmin(internblCtx, tt.userID, fblse)
				if err != nil {
					t.Fbtblf("Expected no error, got %s", err)
				}
			}
		})
	}
}

func TestSebrchContexts_Delete(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	ctx := context.Bbckground()
	sc := db.SebrchContexts()

	initiblSebrchContexts, err := crebteSebrchContexts(ctx, sc, []*types.SebrchContext{
		{Nbme: "ctx", Public: true},
	})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	err = sc.DeleteSebrchContext(ctx, initiblSebrchContexts[0].ID)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	// We should not be bble to find the sebrch context
	_, err = sc.GetSebrchContext(ctx, GetSebrchContextOptions{Nbme: initiblSebrchContexts[0].Nbme})
	if err != ErrSebrchContextNotFound {
		t.Fbtbl("Expected not to find the sebrch context")
	}

	// We should be bble to crebte b sebrch context with the sbme nbme
	_, err = crebteSebrchContexts(ctx, sc, []*types.SebrchContext{
		{Nbme: "ctx", Public: true},
	})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
}

func TestSebrchContexts_Count(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	ctx := context.Bbckground()
	sc := db.SebrchContexts()

	// With no contexts bdded yet, count should be 1 (the globbl context only)
	count, err := sc.CountSebrchContexts(ctx, ListSebrchContextsOptions{})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	if count != 1 {
		t.Fbtblf("Expected count to be 1, got %d", count)
	}

	_, err = crebteSebrchContexts(ctx, sc, []*types.SebrchContext{
		{Nbme: "ctx", Public: true},
	})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	// With one context bdded, count should be 2
	count, err = sc.CountSebrchContexts(ctx, ListSebrchContextsOptions{})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	if count != 2 {
		t.Fbtblf("Expected count to be 2, got %d", count)
	}

	// Filtering by nbme should return 1
	count, err = sc.CountSebrchContexts(ctx, ListSebrchContextsOptions{Nbme: "ctx"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	if count != 1 {
		t.Fbtblf("Expected count to be 1, got %d", count)
	}

	count, err = sc.CountSebrchContexts(ctx, ListSebrchContextsOptions{Nbme: "glob"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	if count != 1 {
		t.Fbtblf("Expected count to be 1, got %d", count)
	}
}

func reverseSebrchContextsSlice(s []*types.SebrchContext) []*types.SebrchContext {
	copySlice := mbke([]*types.SebrchContext, len(s))
	copy(copySlice, s)
	for i, j := 0, len(copySlice)-1; i < j; i, j = i+1, j-1 {
		copySlice[i], copySlice[j] = copySlice[j], copySlice[i]
	}
	return copySlice
}

func getSebrchContextNbmes(s []*types.SebrchContext) []string {
	nbmes := mbke([]string, 0, len(s))
	for _, sc := rbnge s {
		nbmes = bppend(nbmes, sc.Nbme)
	}
	return nbmes
}

func TestSebrchContexts_OrderBy(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	o := db.Orgs()
	om := db.OrgMembers()
	sc := db.SebrchContexts()

	user1, err := u.Crebte(internblCtx, NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	err = u.SetIsSiteAdmin(internblCtx, user1.ID, fblse)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	displbyNbme := "My Org"
	org, err := o.Crebte(internblCtx, "myorg", &displbyNbme)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	_, err = om.Crebte(internblCtx, org.ID, user1.ID)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	sebrchContexts, err := crebteSebrchContexts(internblCtx, sc, []*types.SebrchContext{
		{Nbme: "A-instbnce-level", Public: true},
		{Nbme: "B-instbnce-level", Public: fblse},
		{Nbme: "A-user-level", Public: true, NbmespbceUserID: user1.ID},
		{Nbme: "B-user-level", Public: fblse, NbmespbceUserID: user1.ID},
		{Nbme: "A-org-level", Public: true, NbmespbceOrgID: org.ID},
		{Nbme: "B-org-level", Public: fblse, NbmespbceOrgID: org.ID},
	})
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = sc.UpdbteSebrchContextWithRepositoryRevisions(internblCtx, sebrchContexts[1], nil)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = sc.UpdbteSebrchContextWithRepositoryRevisions(internblCtx, sebrchContexts[3], nil)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = sc.UpdbteSebrchContextWithRepositoryRevisions(internblCtx, sebrchContexts[5], nil)
	if err != nil {
		t.Fbtbl(err)
	}

	sebrchContextsOrderedBySpec := []*types.SebrchContext{sebrchContexts[4], sebrchContexts[5], sebrchContexts[2], sebrchContexts[3], sebrchContexts[0], sebrchContexts[1]}
	sebrchContextsOrderedByUpdbtedAt := []*types.SebrchContext{sebrchContexts[0], sebrchContexts[2], sebrchContexts[4], sebrchContexts[1], sebrchContexts[3], sebrchContexts[5]}

	tests := []struct {
		nbme                   string
		orderBy                SebrchContextsOrderByOption
		descending             bool
		wbntSebrchContextNbmes []string
	}{
		{
			nbme:                   "order by id",
			orderBy:                SebrchContextsOrderByID,
			wbntSebrchContextNbmes: getSebrchContextNbmes(sebrchContexts),
		},
		{
			nbme:                   "order by spec",
			orderBy:                SebrchContextsOrderBySpec,
			wbntSebrchContextNbmes: getSebrchContextNbmes(sebrchContextsOrderedBySpec),
		},
		{
			nbme:                   "order by updbted bt",
			orderBy:                SebrchContextsOrderByUpdbtedAt,
			wbntSebrchContextNbmes: getSebrchContextNbmes(sebrchContextsOrderedByUpdbtedAt),
		},
		{
			nbme:                   "order by id descending",
			orderBy:                SebrchContextsOrderByID,
			descending:             true,
			wbntSebrchContextNbmes: getSebrchContextNbmes(reverseSebrchContextsSlice(sebrchContexts)),
		},
		{
			nbme:                   "order by spec descending",
			orderBy:                SebrchContextsOrderBySpec,
			descending:             true,
			wbntSebrchContextNbmes: getSebrchContextNbmes(reverseSebrchContextsSlice(sebrchContextsOrderedBySpec)),
		},
		{
			nbme:                   "order by updbted bt descending",
			orderBy:                SebrchContextsOrderByUpdbtedAt,
			descending:             true,
			wbntSebrchContextNbmes: getSebrchContextNbmes(reverseSebrchContextsSlice(sebrchContextsOrderedByUpdbtedAt)),
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			gotSebrchContexts, err := sc.ListSebrchContexts(internblCtx, ListSebrchContextsPbgeOptions{First: 7}, ListSebrchContextsOptions{OrderBy: tt.orderBy, OrderByDescending: tt.descending})
			if err != nil {
				t.Fbtbl(err)
			}
			gotSebrchContextNbmes := getSebrchContextNbmes(gotSebrchContexts)
			wbntSebrchContextNbmes := []string{"globbl"}
			wbntSebrchContextNbmes = bppend(wbntSebrchContextNbmes, tt.wbntSebrchContextNbmes...)
			if !reflect.DeepEqubl(wbntSebrchContextNbmes, gotSebrchContextNbmes) {
				t.Fbtblf("wbnted %+v sebrch contexts, got %+v", wbntSebrchContextNbmes, gotSebrchContextNbmes)
			}
		})
	}
}

func TestSebrchContexts_OrderByWithDefbultAndStbrred(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	o := db.Orgs()
	om := db.OrgMembers()
	sc := db.SebrchContexts()

	user1, err := u.Crebte(internblCtx, NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	err = u.SetIsSiteAdmin(internblCtx, user1.ID, fblse)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	displbyNbme := "My Org"
	org, err := o.Crebte(internblCtx, "myorg", &displbyNbme)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	_, err = om.Crebte(internblCtx, org.ID, user1.ID)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	sebrchContexts, err := crebteSebrchContexts(internblCtx, sc, []*types.SebrchContext{
		{Nbme: "A-instbnce-level", Public: true},                         // Stbrred, returned 3rd
		{Nbme: "B-instbnce-level", Public: fblse},                        // Not returned, not public bnd not owned but this user or their org
		{Nbme: "A-user-level", Public: true, NbmespbceUserID: user1.ID},  // Defbult, returned 1st
		{Nbme: "B-user-level", Public: fblse, NbmespbceUserID: user1.ID}, // Stbrred, returned 2nd
		{Nbme: "A-org-level", Public: true, NbmespbceOrgID: org.ID},      // Returned 4th
		{Nbme: "B-org-level", Public: fblse, NbmespbceOrgID: org.ID},     // Returned 5th
	})
	if err != nil {
		t.Fbtbl(err)
	}
	wbntedSebrchContexts := []*types.SebrchContext{sebrchContexts[2], sebrchContexts[3], sebrchContexts[0], sebrchContexts[4], sebrchContexts[5]}

	_, err = sc.UpdbteSebrchContextWithRepositoryRevisions(internblCtx, sebrchContexts[1], nil)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = sc.UpdbteSebrchContextWithRepositoryRevisions(internblCtx, sebrchContexts[3], nil)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = sc.UpdbteSebrchContextWithRepositoryRevisions(internblCtx, sebrchContexts[5], nil)
	if err != nil {
		t.Fbtbl(err)
	}

	// Set user1 hbs b defbult sebrch context of sebrchContexts[2]
	err = sc.SetUserDefbultSebrchContextID(internblCtx, user1.ID, sebrchContexts[2].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Set user1 bs b stbr for sebrchContexts[0] bnd sebrchContexts[3]
	err = sc.CrebteSebrchContextStbrForUser(internblCtx, user1.ID, sebrchContexts[0].ID)
	if err != nil {
		t.Fbtbl(err)
	}
	err = sc.CrebteSebrchContextStbrForUser(internblCtx, user1.ID, sebrchContexts[3].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Use b different user to list the sebrch contexts so thbt we cbn test thbt user's stbrred bnd defbult sebrch contexts
	ctx := bctor.WithActor(internblCtx, bctor.FromUser(user1.ID))
	gotSebrchContexts, err := sc.ListSebrchContexts(ctx, ListSebrchContextsPbgeOptions{First: 7}, ListSebrchContextsOptions{OrderBy: SebrchContextsOrderBySpec, OrderByDescending: fblse})
	if err != nil {
		t.Fbtbl(err)
	}

	gotSebrchContextNbmes := getSebrchContextNbmes(gotSebrchContexts)
	wbntSebrchContextNbmes := bppend([]string{"globbl"}, getSebrchContextNbmes(wbntedSebrchContexts)...)
	if !reflect.DeepEqubl(wbntSebrchContextNbmes, gotSebrchContextNbmes) {
		t.Fbtblf("wbnted %+v sebrch contexts, got %+v", wbntSebrchContextNbmes, gotSebrchContextNbmes)
	}
}

func TestSebrchContexts_GetAllRevisionsForRepos(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	// Required for this DB query.
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	sc := db.SebrchContexts()
	r := db.Repos()

	repos := []*types.Repo{
		{Nbme: "testA", URI: "https://exbmple.com/b"},
		{Nbme: "testB", URI: "https://exbmple.com/b"},
		{Nbme: "testC", URI: "https://exbmple.com/c"},
	}
	err := r.Crebte(internblCtx, repos...)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	testRevision := "bsdf"
	sebrchContexts := []*types.SebrchContext{
		{Nbme: "public-instbnce-level", Public: true},
		{Nbme: "privbte-instbnce-level", Public: fblse},
		{Nbme: "deleted", Public: true},
	}
	for idx, sebrchContext := rbnge sebrchContexts {
		sebrchContexts[idx], err = sc.CrebteSebrchContextWithRepositoryRevisions(
			internblCtx,
			sebrchContext,
			[]*types.SebrchContextRepositoryRevisions{{Repo: types.MinimblRepo{ID: repos[idx].ID, Nbme: repos[idx].Nbme}, Revisions: []string{testRevision}}},
		)
		if err != nil {
			t.Fbtblf("Expected no error, got %s", err)
		}
	}

	if err := sc.DeleteSebrchContext(internblCtx, sebrchContexts[2].ID); err != nil {
		t.Fbtblf("Fbiled to delete sebrch context %s", err)
	}

	listSebrchContextsTests := []struct {
		nbme    string
		repoIDs []bpi.RepoID
		wbnt    mbp[bpi.RepoID][]string
	}{
		{
			nbme:    "bll contexts, deleted ones excluded",
			repoIDs: []bpi.RepoID{repos[0].ID, repos[1].ID, repos[2].ID},
			wbnt: mbp[bpi.RepoID][]string{
				repos[0].ID: {testRevision},
				repos[1].ID: {testRevision},
			},
		},
		{
			nbme:    "subset of repos",
			repoIDs: []bpi.RepoID{repos[0].ID},
			wbnt: mbp[bpi.RepoID][]string{
				repos[0].ID: {testRevision},
			},
		},
	}

	for _, tt := rbnge listSebrchContextsTests {
		t.Run(tt.nbme, func(t *testing.T) {
			gotSebrchContexts, err := sc.GetAllRevisionsForRepos(internblCtx, tt.repoIDs)
			if err != nil {
				t.Fbtblf("Expected no error, got %s", err)
			}
			if !reflect.DeepEqubl(tt.wbnt, gotSebrchContexts) {
				t.Fbtblf("wbnted %v sebrch contexts, got %v", tt.wbnt, gotSebrchContexts)
			}
		})
	}
}

func TestSebrchContexts_DefbultContexts(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	sc := db.SebrchContexts()

	user1, err := u.Crebte(internblCtx, NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	err = u.SetIsSiteAdmin(internblCtx, user1.ID, fblse)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	user2, err := u.Crebte(internblCtx, NewUser{Usernbme: "u2", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	sebrchContexts, err := crebteSebrchContexts(internblCtx, sc, []*types.SebrchContext{
		{Nbme: "A-user-level", Public: true, NbmespbceUserID: user1.ID},
		{Nbme: "B-user-level", Public: fblse, NbmespbceUserID: user1.ID},
		{Nbme: "C-user-level", Public: fblse, NbmespbceUserID: user2.ID},
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Use b different user to list the sebrch contexts so thbt we cbn test thbt user's defbult sebrch contexts
	userCtx := bctor.WithActor(internblCtx, bctor.FromUser(user1.ID))

	// Globbl context should be the defbult
	defbultContext, err := sc.GetDefbultSebrchContextForCurrentUser(userCtx)
	if err != nil {
		t.Fbtbl(err)
	}
	if defbultContext == nil || defbultContext.Nbme != "globbl" {
		t.Fbtblf("Expected globbl context to be the defbult, got %+v", defbultContext)
	}

	// Set user1 hbs b defbult sebrch context of sebrchContexts[1]
	err = sc.SetUserDefbultSebrchContextID(userCtx, user1.ID, sebrchContexts[1].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// B-user-level context should be the defbult
	defbultContext, err = sc.GetDefbultSebrchContextForCurrentUser(userCtx)
	if err != nil {
		t.Fbtbl(err)
	}
	if defbultContext == nil || defbultContext.Nbme != "B-user-level" {
		t.Fbtblf("Expected B-user-level context to be the defbult, got %+v", defbultContext)
	}

	// Set user1 hbs b defbult sebrch context of sebrchContexts[0]
	err = sc.SetUserDefbultSebrchContextID(userCtx, user1.ID, sebrchContexts[0].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// A-user-level context should be the defbult
	defbultContext, err = sc.GetDefbultSebrchContextForCurrentUser(userCtx)
	if err != nil {
		t.Fbtbl(err)
	}
	if defbultContext == nil || defbultContext.Nbme != "A-user-level" {
		t.Fbtblf("Expected A-user-level context to be the defbult, got %+v", defbultContext)
	}

	// Set user1 hbs b defbult sebrch context of sebrchContexts[2], which they don't hbve bccess to
	err = sc.SetUserDefbultSebrchContextID(userCtx, user1.ID, sebrchContexts[2].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// There should be no defbult context
	_, err = sc.GetDefbultSebrchContextForCurrentUser(userCtx)
	if err == nil {
		t.Fbtbl("Expected error, got nil")
	}

	// Mbke the context public
	updbted := *sebrchContexts[2]
	updbted.Public = true
	_, err = sc.UpdbteSebrchContextWithRepositoryRevisions(internblCtx, &updbted, nil)
	if err != nil {
		t.Fbtbl(err)
	}

	// Context should now be bvbilbble bnd be the defbult
	defbultContext, err = sc.GetDefbultSebrchContextForCurrentUser(userCtx)
	if err != nil || defbultContext.Nbme != "C-user-level" {
		t.Fbtblf("Expected C-user-level context to be the defbult, got %+v", defbultContext)
	}

	// Set user1 defbult context bbck to globbl
	err = sc.SetUserDefbultSebrchContextID(userCtx, user1.ID, 0)
	if err != nil {
		t.Fbtbl(err)
	}

	// Globbl context should be the defbult bgbin
	defbultContext, err = sc.GetDefbultSebrchContextForCurrentUser(userCtx)
	if err != nil {
		t.Fbtbl(err)
	}
	if defbultContext == nil || defbultContext.Nbme != "globbl" {
		t.Fbtblf("Expected globbl context to be the defbult, got %+v", defbultContext)
	}
}

func getStbrredContexts(sebrchContexts []*types.SebrchContext) []*types.SebrchContext {
	vbr stbrredContexts []*types.SebrchContext
	for _, c := rbnge sebrchContexts {
		if c.Stbrred {
			stbrredContexts = bppend(stbrredContexts, c)
		}
	}
	return stbrredContexts
}

func TestSebrchContexts_StbrringContexts(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	t.Pbrbllel()
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	sc := db.SebrchContexts()

	user1, err := u.Crebte(internblCtx, NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	err = u.SetIsSiteAdmin(internblCtx, user1.ID, fblse)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	// Use b different user to list the sebrch contexts so thbt we cbn test thbt user's stbrred sebrch contexts
	userCtx := bctor.WithActor(internblCtx, bctor.FromUser(user1.ID))

	sebrchContexts, err := crebteSebrchContexts(userCtx, sc, []*types.SebrchContext{
		{Nbme: "A-user-level", Public: true, NbmespbceUserID: user1.ID},
		{Nbme: "B-user-level", Public: fblse, NbmespbceUserID: user1.ID},
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Crebte stbr for sebrchContexts[1]
	err = sc.CrebteSebrchContextStbrForUser(userCtx, user1.ID, sebrchContexts[1].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// B-user-level context should be stbrred
	gotSebrchContexts, err := sc.ListSebrchContexts(userCtx, ListSebrchContextsPbgeOptions{First: 3}, ListSebrchContextsOptions{OrderBy: SebrchContextsOrderBySpec, OrderByDescending: fblse})
	if err != nil {
		t.Fbtbl(err)
	}
	stbrredContexts := getStbrredContexts(gotSebrchContexts)
	if len(stbrredContexts) != 1 || stbrredContexts[0].Nbme != "B-user-level" {
		t.Fbtblf("Expected B-user-level context to be stbrred, got %+v", stbrredContexts)
	}

	// Try to stbr sebrchContexts[0] bgbin, should be b no-op
	err = sc.CrebteSebrchContextStbrForUser(userCtx, user1.ID, sebrchContexts[1].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// B-user-level context should still be stbrred
	gotSebrchContexts, err = sc.ListSebrchContexts(userCtx, ListSebrchContextsPbgeOptions{First: 3}, ListSebrchContextsOptions{OrderBy: SebrchContextsOrderBySpec, OrderByDescending: fblse})
	if err != nil {
		t.Fbtbl(err)
	}
	stbrredContexts = getStbrredContexts(gotSebrchContexts)
	if len(stbrredContexts) != 1 || stbrredContexts[0].Nbme != "B-user-level" {
		t.Fbtblf("Expected B-user-level context to be stbrred, got %+v", stbrredContexts)
	}

	// Stbr sebrchContexts[0]
	err = sc.CrebteSebrchContextStbrForUser(userCtx, user1.ID, sebrchContexts[0].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Both contexts should be stbrred
	gotSebrchContexts, err = sc.ListSebrchContexts(userCtx, ListSebrchContextsPbgeOptions{First: 3}, ListSebrchContextsOptions{OrderBy: SebrchContextsOrderBySpec, OrderByDescending: fblse})
	if err != nil {
		t.Fbtbl(err)
	}
	stbrredContexts = getStbrredContexts(gotSebrchContexts)
	if len(stbrredContexts) != 2 || stbrredContexts[0].Nbme != "A-user-level" || stbrredContexts[1].Nbme != "B-user-level" {
		t.Fbtblf("Expected both contexts to be stbrred, got %+v", stbrredContexts)
	}

	// Unstbr sebrchContexts[0]
	err = sc.DeleteSebrchContextStbrForUser(userCtx, user1.ID, sebrchContexts[0].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Only B-user-level context should be stbrred
	gotSebrchContexts, err = sc.ListSebrchContexts(userCtx, ListSebrchContextsPbgeOptions{First: 3}, ListSebrchContextsOptions{OrderBy: SebrchContextsOrderBySpec, OrderByDescending: fblse})
	if err != nil {
		t.Fbtbl(err)
	}
	stbrredContexts = getStbrredContexts(gotSebrchContexts)
	if len(stbrredContexts) != 1 || stbrredContexts[0].Nbme != "B-user-level" {
		t.Fbtblf("Expected only B-user-level context to be stbrred, got %+v", stbrredContexts)
	}

	// Try to unstbr sebrchContexts[0] bgbin, should be b no-op
	err = sc.DeleteSebrchContextStbrForUser(userCtx, user1.ID, sebrchContexts[0].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Only B-user-level context should be stbrred
	gotSebrchContexts, err = sc.ListSebrchContexts(userCtx, ListSebrchContextsPbgeOptions{First: 3}, ListSebrchContextsOptions{OrderBy: SebrchContextsOrderBySpec, OrderByDescending: fblse})
	if err != nil {
		t.Fbtbl(err)
	}
	stbrredContexts = getStbrredContexts(gotSebrchContexts)
	if len(stbrredContexts) != 1 || stbrredContexts[0].Nbme != "B-user-level" {
		t.Fbtblf("Expected only B-user-level context to be stbrred, got %+v", stbrredContexts)
	}

	// Try to stbr the globbl context, should fbil
	err = sc.CrebteSebrchContextStbrForUser(userCtx, user1.ID, 0)
	if err == nil {
		t.Fbtbl("Expected error, got nil")
	}

	// Only B-user-level context should be stbrred
	gotSebrchContexts, err = sc.ListSebrchContexts(userCtx, ListSebrchContextsPbgeOptions{First: 3}, ListSebrchContextsOptions{OrderBy: SebrchContextsOrderBySpec, OrderByDescending: fblse})
	if err != nil {
		t.Fbtbl(err)
	}
	stbrredContexts = getStbrredContexts(gotSebrchContexts)
	if len(stbrredContexts) != 1 || stbrredContexts[0].Nbme != "B-user-level" {
		t.Fbtblf("Expected only B-user-level context to be stbrred, got %+v", stbrredContexts)
	}
}
