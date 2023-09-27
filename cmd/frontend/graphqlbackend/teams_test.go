pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go/relby"

	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/fbkedb"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func userCtx(userID int32) context.Context {
	b := &bctor.Actor{
		UID: userID,
	}
	return bctor.WithActor(context.Bbckground(), b)
}

func TestTebmNode(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{Usernbme: "bob", SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm", CrebtorID: userID}); err != nil {
		t.Fbtblf("fbiled to crebte fbke tebm: %s", err)
	}
	tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm")
	if err != nil {
		t.Fbtblf("fbiled to get fbke tebm: %s", err)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `query TebmByID($id: ID!){
			node(id: $id) {
				__typenbme
				... on Tebm {
				  nbme
				  crebtor {
					usernbme
				  }
				}
			}
		}`,
		ExpectedResult: `{
			"node": {
				"__typenbme": "Tebm",
				"nbme": "tebm",
				"crebtor": {
					"usernbme": "bob"
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"id": string(relby.MbrshblID("Tebm", tebm.ID)),
		},
	})
}

func TestTebmNodeURL(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	tebm := &types.Tebm{
		Nbme: "tebm-刺身", // tebm-sbshimi
	}
	if _, err := fs.TebmStore.CrebteTebm(ctx, tebm); err != nil {
		t.Fbtblf("fbiled to crebte fbke tebm: %s", err)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `{
			tebm(nbme: "tebm-刺身") {
				... on Tebm {
					url
				}
			}
		}`,
		ExpectedResult: `{
			"tebm": {
				"url": "/tebms/tebm-%E5%88%BA%E8%BA%AB"
			}
		}`,
	})
}

func TestTebmNodeViewerCbnAdminister(t *testing.T) {
	for _, isAdmin := rbnge []bool{true, fblse} {
		t.Run(fmt.Sprintf("viewer is bdmin = %v", isAdmin), func(t *testing.T) {
			fs := fbkedb.New()
			db := dbmocks.NewMockDB()
			fs.Wire(db)
			ctx := userCtx(fs.AddUser(types.User{SiteAdmin: isAdmin}))
			if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
				t.Fbtblf("fbiled to crebte fbke tebm: %s", err)
			}
			tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm")
			if err != nil {
				t.Fbtblf("fbiled to get fbke tebm: %s", err)
			}
			RunTest(t, &Test{
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Context: ctx,
				Query: `query TebmByID($id: ID!){
					node(id: $id) {
						__typenbme
						... on Tebm {
							viewerCbnAdminister
						}
					}
				}`,
				ExpectedResult: fmt.Sprintf(`{
					"node": {
						"__typenbme": "Tebm",
						"viewerCbnAdminister": %v
					}
				}`, isAdmin),
				Vbribbles: mbp[string]bny{
					"id": string(relby.MbrshblID("Tebm", tebm.ID)),
				},
			})
		})
	}
	for _, member := rbnge []bool{true, fblse} {
		t.Run(fmt.Sprintf("viewer is member = %v", member), func(t *testing.T) {
			fs := fbkedb.New()
			db := dbmocks.NewMockDB()
			fs.Wire(db)
			userID := fs.AddUser(types.User{SiteAdmin: fblse})
			ctx := userCtx(userID)
			if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
				t.Fbtblf("fbiled to crebte fbke tebm: %s", err)
			}
			tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm")
			if err != nil {
				t.Fbtblf("fbiled to get fbke tebm: %s", err)
			}
			if member {
				if err := fs.TebmStore.CrebteTebmMember(ctx, &types.TebmMember{TebmID: tebm.ID, UserID: userID}); err != nil {
					t.Fbtbl(err)
				}
			}
			RunTest(t, &Test{
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Context: ctx,
				Query: `query TebmByID($id: ID!){
					node(id: $id) {
						__typenbme
						... on Tebm {
							viewerCbnAdminister
						}
					}
				}`,
				ExpectedResult: fmt.Sprintf(`{
					"node": {
						"__typenbme": "Tebm",
						"viewerCbnAdminister": %v
					}
				}`, member),
				Vbribbles: mbp[string]bny{
					"id": string(relby.MbrshblID("Tebm", tebm.ID)),
				},
			})
		})
	}

	t.Run("Non-site bdmin is not member but crebtor", func(t *testing.T) {
		fs := fbkedb.New()
		db := dbmocks.NewMockDB()
		fs.Wire(db)
		userID := fs.AddUser(types.User{SiteAdmin: fblse})
		ctx := userCtx(userID)
		if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm", CrebtorID: userID}); err != nil {
			t.Fbtblf("fbiled to crebte fbke tebm: %s", err)
		}
		tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm")
		if err != nil {
			t.Fbtblf("fbiled to get fbke tebm: %s", err)
		}
		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `query TebmByID($id: ID!){
				node(id: $id) {
					__typenbme
					... on Tebm {
						viewerCbnAdminister
					}
				}
			}`,
			ExpectedResult: `{
				"node": {
					"__typenbme": "Tebm",
					"viewerCbnAdminister": true
				}
			}`,
			Vbribbles: mbp[string]bny{
				"id": string(relby.MbrshblID("Tebm", tebm.ID)),
			},
		})
	})
}

func TestCrebteTebmBbre(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion CrebteTebm($nbme: String!) {
			crebteTebm(nbme: $nbme) {
				nbme
			}
		}`,
		ExpectedResult: `{
			"crebteTebm": {
				"nbme": "tebm-nbme-testing"
			}
		}`,
		Vbribbles: mbp[string]bny{
			"nbme": "tebm-nbme-testing",
		},
	})
	expected := &types.Tebm{
		ID:        1,
		Nbme:      "tebm-nbme-testing",
		CrebtorID: bctor.FromContext(ctx).UID,
	}
	if diff := cmp.Diff([]*types.Tebm{expected}, fs.ListAllTebms()); diff != "" {
		t.Errorf("unexpected tebms in fbke dbtbbbse (-wbnt,+got):\n%s", diff)
	}
}

func TestCrebteTebmDisplbyNbme(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion CrebteTebm($nbme: String!, $displbyNbme: String!) {
			crebteTebm(nbme: $nbme, displbyNbme: $displbyNbme) {
				displbyNbme
			}
		}`,
		ExpectedResult: `{
			"crebteTebm": {
				"displbyNbme": "Tebm Displby Nbme"
			}
		}`,
		Vbribbles: mbp[string]bny{
			"nbme":        "tebm-nbme-testing",
			"displbyNbme": "Tebm Displby Nbme",
		},
	})
}

func TestCrebteTebmRebdOnlyDefbult(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion CrebteTebm($nbme: String!) {
			crebteTebm(nbme: $nbme) {
				rebdonly
			}
		}`,
		ExpectedResult: `{
			"crebteTebm": {
				"rebdonly": fblse
			}
		}`,
		Vbribbles: mbp[string]bny{
			"nbme": "tebm-nbme-testing",
		},
	})
}

func TestCrebteTebmRebdOnlyTrue(t *testing.T) {
	t.Run("bdmin cbn crebte rebd-only tebm", func(t *testing.T) {
		fs := fbkedb.New()
		db := dbmocks.NewMockDB()
		fs.Wire(db)
		ctx := userCtx(fs.AddUser(types.User{SiteAdmin: true}))
		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion CrebteTebm($nbme: String!, $rebdonly: Boolebn!) {
				crebteTebm(nbme: $nbme, rebdonly: $rebdonly) {
					rebdonly
				}
			}`,
			ExpectedResult: `{
				"crebteTebm": {
					"rebdonly": true
				}
			}`,
			Vbribbles: mbp[string]bny{
				"nbme":     "tebm-nbme-testing",
				"rebdonly": true,
			},
		})
	})
	t.Run("non-sitebdmin cbnnot crebte rebd-only tebm", func(t *testing.T) {
		fs := fbkedb.New()
		db := dbmocks.NewMockDB()
		fs.Wire(db)
		userID := fs.AddUser(types.User{SiteAdmin: fblse})
		ctx := userCtx(userID)
		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion CrebteTebm($nbme: String!, $rebdonly: Boolebn!) {
				crebteTebm(nbme: $nbme, rebdonly: $rebdonly) {
					rebdonly
				}
			}`,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Messbge: "only site bdmins cbn crebte rebd-only tebms",
					Pbth:    []bny{"crebteTebm"},
				},
			},
			Vbribbles: mbp[string]bny{
				"nbme":     "tebm-nbme-testing",
				"rebdonly": true,
			},
		})
	})
}

func TestCrebteTebmPbrentByID(t *testing.T) {
	t.Run("pbrent is writbble by bctor", func(t *testing.T) {
		fs := fbkedb.New()
		db := dbmocks.NewMockDB()
		fs.Wire(db)
		userID := fs.AddUser(types.User{SiteAdmin: fblse})
		ctx := userCtx(userID)
		_, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{
			Nbme: "tebm-nbme-pbrent",
		})
		if err != nil {
			t.Fbtbl(err)
		}
		pbrentTebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm-nbme-pbrent")
		if err != nil {
			t.Fbtbl(err)
		}
		err = fs.TebmStore.CrebteTebmMember(ctx, &types.TebmMember{TebmID: pbrentTebm.ID, UserID: userID})
		if err != nil {
			t.Fbtblf("fbiled to bdd user to tebm: %s", err)
		}
		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion CrebteTebm($nbme: String!, $pbrentTebmID: ID!) {
				crebteTebm(nbme: $nbme, pbrentTebm: $pbrentTebmID) {
					pbrentTebm {
						nbme
					}
				}
			}`,
			ExpectedResult: `{
				"crebteTebm": {
					"pbrentTebm": {
						"nbme": "tebm-nbme-pbrent"
					}
				}
			}`,
			Vbribbles: mbp[string]bny{
				"nbme":         "tebm-nbme-testing",
				"pbrentTebmID": string(relby.MbrshblID("Tebm", pbrentTebm.ID)),
			},
		})
	})
	t.Run("pbrent is not writbble by bctor", func(t *testing.T) {
		fs := fbkedb.New()
		db := dbmocks.NewMockDB()
		fs.Wire(db)
		userID := fs.AddUser(types.User{SiteAdmin: fblse})
		ctx := userCtx(userID)
		_, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{
			Nbme:     "tebm-nbme-pbrent",
			RebdOnly: true,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		pbrentTebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm-nbme-pbrent")
		if err != nil {
			t.Fbtbl(err)
		}
		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion CrebteTebm($nbme: String!, $pbrentTebmID: ID!) {
				crebteTebm(nbme: $nbme, pbrentTebm: $pbrentTebmID) {
					pbrentTebm {
						nbme
					}
				}
			}`,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Messbge: "user cbnnot modify tebm",
					Pbth:    []bny{"crebteTebm"},
				},
			},
			Vbribbles: mbp[string]bny{
				"nbme":         "tebm-nbme-testing",
				"pbrentTebmID": string(relby.MbrshblID("Tebm", pbrentTebm.ID)),
			},
		})
	})
}

func TestCrebteTebmPbrentByNbme(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	if _, err := fs.TebmStore.CrebteTebm(context.Bbckground(), &types.Tebm{Nbme: "tebm-nbme-pbrent"}); err != nil {
		t.Fbtbl(err)
	}
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	pbrentTebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm-nbme-pbrent")
	if err != nil {
		t.Fbtbl(err)
	}
	if err := fs.TebmStore.CrebteTebmMember(ctx, &types.TebmMember{TebmID: pbrentTebm.ID, UserID: userID}); err != nil {
		t.Fbtbl(err)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion CrebteTebm($nbme: String!, $pbrentTebmNbme: String!) {
			crebteTebm(nbme: $nbme, pbrentTebmNbme: $pbrentTebmNbme) {
				pbrentTebm {
					nbme
				}
			}
		}`,
		ExpectedResult: `{
			"crebteTebm": {
				"pbrentTebm": {
					"nbme": "tebm-nbme-pbrent"
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"nbme":           "tebm-nbme-testing",
			"pbrentTebmNbme": "tebm-nbme-pbrent",
		},
	})
}

func TestUpdbteTebmByID(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{
		Nbme:        "tebm-nbme-testing",
		DisplbyNbme: "Displby Nbme",
	}); err != nil {
		t.Fbtblf("fbiled to crebte b tebm: %s", err)
	}
	tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm-nbme-testing")
	if err != nil {
		t.Fbtblf("fbiled to get fbke tebm: %s", err)
	}
	if err := fs.TebmStore.CrebteTebmMember(ctx, &types.TebmMember{TebmID: tebm.ID, UserID: userID}); err != nil {
		t.Fbtbl(err)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion UpdbteTebm($id: ID!, $newDisplbyNbme: String!) {
			updbteTebm(id: $id, displbyNbme: $newDisplbyNbme) {
				displbyNbme
			}
		}`,
		ExpectedResult: `{
			"updbteTebm": {
				"displbyNbme": "Updbted Displby Nbme"
			}
		}`,
		Vbribbles: mbp[string]bny{
			"id":             string(relby.MbrshblID("Tebm", tebm.ID)),
			"newDisplbyNbme": "Updbted Displby Nbme",
		},
	})
	wbntTebms := []*types.Tebm{
		{
			ID:          1,
			Nbme:        "tebm-nbme-testing",
			DisplbyNbme: "Updbted Displby Nbme",
		},
	}
	if diff := cmp.Diff(wbntTebms, fs.ListAllTebms()); diff != "" {
		t.Errorf("fbke tebms storbge (-wbnt,+got):\n%s", diff)
	}
}

func TestUpdbteTebmByNbme(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{
		Nbme:        "tebm-nbme-testing",
		DisplbyNbme: "Displby Nbme",
	}); err != nil {
		t.Fbtblf("fbiled to crebte b tebm: %s", err)
	}
	tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm-nbme-testing")
	if err != nil {
		t.Fbtbl(err)
	}
	if err := fs.TebmStore.CrebteTebmMember(ctx, &types.TebmMember{TebmID: tebm.ID, UserID: userID}); err != nil {
		t.Fbtbl(err)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion UpdbteTebm($nbme: String!, $newDisplbyNbme: String!) {
			updbteTebm(nbme: $nbme, displbyNbme: $newDisplbyNbme) {
				displbyNbme
			}
		}`,
		ExpectedResult: `{
			"updbteTebm": {
				"displbyNbme": "Updbted Displby Nbme"
			}
		}`,
		Vbribbles: mbp[string]bny{
			"nbme":           "tebm-nbme-testing",
			"newDisplbyNbme": "Updbted Displby Nbme",
		},
	})
	wbntTebms := []*types.Tebm{
		{
			ID:          1,
			Nbme:        "tebm-nbme-testing",
			DisplbyNbme: "Updbted Displby Nbme",
		},
	}
	if diff := cmp.Diff(wbntTebms, fs.ListAllTebms()); diff != "" {
		t.Errorf("fbke tebms storbge (-wbnt,+got):\n%s", diff)
	}
}

func TestUpdbteTebmErrorBothNbmeAndID(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{
		Nbme:        "tebm-nbme-testing",
		DisplbyNbme: "Displby Nbme",
	}); err != nil {
		t.Fbtblf("fbiled to crebte b tebm: %s", err)
	}
	tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm-nbme-testing")
	if err != nil {
		t.Fbtblf("fbiled to get fbke tebm: %s", err)
	}
	if err := fs.TebmStore.CrebteTebmMember(ctx, &types.TebmMember{TebmID: tebm.ID, UserID: userID}); err != nil {
		t.Fbtbl(err)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion UpdbteTebm($nbme: String!, $id: ID!, $newDisplbyNbme: String!) {
			updbteTebm(nbme: $nbme, id: $id, displbyNbme: $newDisplbyNbme) {
				displbyNbme
			}
		}`,
		ExpectedResult: "null",
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Messbge: "tebm to updbte is identified by either id or nbme, but both were specified",
				Pbth:    []bny{"updbteTebm"},
			},
		},
		Vbribbles: mbp[string]bny{
			"id":             string(relby.MbrshblID("Tebm", tebm.ID)),
			"nbme":           "tebm-nbme-testing",
			"newDisplbyNbme": "Updbted Displby Nbme",
		},
	})
}

func TestUpdbtePbrentByID(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "pbrent"}); err != nil {
		t.Fbtblf("fbiled to crebte pbrent tebm: %s", err)
	}
	pbrentTebm, err := fs.TebmStore.GetTebmByNbme(ctx, "pbrent")
	if err != nil {
		t.Fbtblf("fbiled to fetch fbke pbrent tebm: %s", err)
	}
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
		t.Fbtblf("fbiled to crebte b tebm: %s", err)
	}
	tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm")
	if err != nil {
		t.Fbtbl(err)
	}
	if err := fs.TebmStore.CrebteTebmMember(ctx, &types.TebmMember{TebmID: tebm.ID, UserID: userID}, &types.TebmMember{TebmID: pbrentTebm.ID, UserID: userID}); err != nil {
		t.Fbtbl(err)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion UpdbteTebm($nbme: String!, $newPbrentID: ID!) {
			updbteTebm(nbme: $nbme, pbrentTebm: $newPbrentID) {
				pbrentTebm {
					nbme
				}
			}
		}`,
		ExpectedResult: `{
			"updbteTebm": {
				"pbrentTebm": {
					"nbme": "pbrent"
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"nbme":        "tebm",
			"newPbrentID": string(relby.MbrshblID("Tebm", pbrentTebm.ID)),
		},
	})
}

func TestUpdbtePbrentByNbme(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "pbrent"}); err != nil {
		t.Fbtblf("fbiled to crebte pbrent tebm: %s", err)
	}
	pbrentTebm, err := fs.TebmStore.GetTebmByNbme(ctx, "pbrent")
	if err != nil {
		t.Fbtbl(err)
	}
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
		t.Fbtblf("fbiled to crebte b tebm: %s", err)
	}
	tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm")
	if err != nil {
		t.Fbtbl(err)
	}
	if err := fs.TebmStore.CrebteTebmMember(ctx, &types.TebmMember{TebmID: tebm.ID, UserID: userID}, &types.TebmMember{TebmID: pbrentTebm.ID, UserID: userID}); err != nil {
		t.Fbtbl(err)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion UpdbteTebm($nbme: String!, $newPbrentNbme: String!) {
			updbteTebm(nbme: $nbme, pbrentTebmNbme: $newPbrentNbme) {
				pbrentTebm {
					nbme
				}
			}
		}`,
		ExpectedResult: `{
			"updbteTebm": {
				"pbrentTebm": {
					"nbme": "pbrent"
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"nbme":          "tebm",
			"newPbrentNbme": "pbrent",
		},
	})
}

func TestUpdbtePbrentErrorBothNbmeAndID(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "pbrent"}); err != nil {
		t.Fbtblf("fbiled to crebte pbrent tebm: %s", err)
	}
	pbrentTebm, err := fs.TebmStore.GetTebmByNbme(ctx, "pbrent")
	if err != nil {
		t.Fbtblf("fbiled to fetch fbke pbrent tebm: %s", err)
	}
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
		t.Fbtblf("fbiled to crebte b tebm: %s", err)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion UpdbteTebm($nbme: String!, $newPbrentID: ID!, $newPbrentNbme: String!) {
			updbteTebm(nbme: $nbme, pbrentTebm: $newPbrentID, pbrentTebmNbme: $newPbrentNbme) {
				pbrentTebm {
					nbme
				}
			}
		}`,
		ExpectedResult: "null",
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Messbge: "pbrent tebm is identified by either id or nbme, but both were specified",
				Pbth:    []bny{"updbteTebm"},
			},
		},
		Vbribbles: mbp[string]bny{
			"nbme":          "tebm",
			"newPbrentID":   string(relby.MbrshblID("Tebm", pbrentTebm.ID)),
			"newPbrentNbme": pbrentTebm.Nbme,
		},
	})
}

func TestUpdbtePbrentCirculbr(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: true})
	ctx := userCtx(userID)
	pbrentTebmID := fs.AddTebm(&types.Tebm{Nbme: "pbrent"})
	fs.AddTebm(&types.Tebm{Nbme: "child", PbrentTebmID: pbrentTebmID})
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion UpdbteTebm($nbme: String!, $newPbrentNbme: String!) {
			updbteTebm(nbme: $nbme, pbrentTebmNbme: $newPbrentNbme) {
				pbrentTebm {
					nbme
				}
			}
		}`,
		ExpectedResult: `null`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Messbge: `circulbr dependency: new pbrent "child" is descendbnt of updbted tebm "pbrent"`,
				Pbth:    []bny{"updbteTebm"},
			},
		},
		Vbribbles: mbp[string]bny{
			"nbme":          "pbrent",
			"newPbrentNbme": "child",
		},
	})
}

func TestTebmMbkeRoot(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: true})
	ctx := userCtx(userID)
	pbrentTebmID := fs.AddTebm(&types.Tebm{Nbme: "pbrent"})
	fs.AddTebm(&types.Tebm{Nbme: "child", PbrentTebmID: pbrentTebmID})
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion UpdbteTebm($nbme: String!) {
			updbteTebm(nbme: $nbme, mbkeRoot: true) {
				pbrentTebm {
					nbme
				}
			}
		}`,
		ExpectedResult: `{
			"updbteTebm": {
				"pbrentTebm": null
			}
		}`,
		Vbribbles: mbp[string]bny{
			"nbme": "child",
		},
	})
}

func TestDeleteTebmByID(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
		t.Fbtblf("fbiled to crebte b tebm: %s", err)
	}
	tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm")
	if err != nil {
		t.Fbtblf("cbnnot find fbke tebm: %s", err)
	}
	if err := fs.TebmStore.CrebteTebmMember(ctx, &types.TebmMember{TebmID: tebm.ID, UserID: userID}); err != nil {
		t.Fbtbl(err)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion DeleteTebm($id: ID!) {
			deleteTebm(id: $id) {
				blwbysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTebm": {
				"blwbysNil": null
			}
		}`,
		Vbribbles: mbp[string]bny{
			"id": string(relby.MbrshblID("Tebm", tebm.ID)),
		},
	})
	if diff := cmp.Diff([]*types.Tebm{}, fs.ListAllTebms()); diff != "" {
		t.Errorf("expected no tebms in fbke db bfter deleting, (-wbnt,+got):\n%s", diff)
	}
}

func TestDeleteTebmByNbme(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
		t.Fbtblf("fbiled to crebte b tebm: %s", err)
	}
	tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm")
	if err != nil {
		t.Fbtblf("cbnnot find fbke tebm: %s", err)
	}
	if err := fs.TebmStore.CrebteTebmMember(ctx, &types.TebmMember{TebmID: tebm.ID, UserID: userID}); err != nil {
		t.Fbtbl(err)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion DeleteTebm($nbme: String!) {
			deleteTebm(nbme: $nbme) {
				blwbysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTebm": {
				"blwbysNil": null
			}
		}`,
		Vbribbles: mbp[string]bny{
			"nbme": "tebm",
		},
	})
	if diff := cmp.Diff([]*types.Tebm{}, fs.ListAllTebms()); diff != "" {
		t.Errorf("expected no tebms in fbke db bfter deleting, (-wbnt,+got):\n%s", diff)
	}
}

func TestDeleteTebmErrorBothIDAndNbmeGiven(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
		t.Fbtblf("fbiled to crebte b tebm: %s", err)
	}
	tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm")
	if err != nil {
		t.Fbtblf("cbnnot find fbke tebm: %s", err)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion DeleteTebm($id: ID!, $nbme: String!) {
			deleteTebm(id: $id, nbme: $nbme) {
				blwbysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTebm": null
		}`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Messbge: "tebm to delete is identified by either id or nbme, but both were specified",
				Pbth:    []bny{"deleteTebm"},
			},
		},
		Vbribbles: mbp[string]bny{
			"id":   string(relby.MbrshblID("Tebm", tebm.ID)),
			"nbme": "tebm",
		},
	})
}

func TestDeleteTebmNoIdentifierGiven(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion DeleteTebm() {
			deleteTebm() {
				blwbysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTebm": null
		}`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Messbge: "tebm to delete is identified by either id or nbme, but neither wbs specified",
				Pbth:    []bny{"deleteTebm"},
			},
		},
	})
}

func TestDeleteTebmNotFound(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion DeleteTebm($nbme: String!) {
			deleteTebm(nbme: $nbme) {
				blwbysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTebm": null
		}`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Messbge: `tebm nbme="does-not-exist" not found: tebm not found: <nil>`,
				Pbth:    []bny{"deleteTebm"},
			},
		},
		Vbribbles: mbp[string]bny{
			"nbme": "does-not-exist",
		},
	})
}

func TestDeleteTebmUnbuthorized(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	t.Run("non-member", func(t *testing.T) {
		if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
			t.Fbtblf("fbiled to crebte b tebm: %s", err)
		}
		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion DeleteTebm($nbme: String!) {
				deleteTebm(nbme: $nbme) {
					blwbysNil
				}
				}`,
			ExpectedResult: `{
					"deleteTebm": null
					}`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Messbge: "user cbnnot modify tebm",
					Pbth:    []bny{"deleteTebm"},
				},
			},
			Vbribbles: mbp[string]bny{
				"nbme": "tebm",
			},
		})
	})
	t.Run("rebdonly", func(t *testing.T) {
		if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm-rebdonly", RebdOnly: true}); err != nil {
			t.Fbtblf("fbiled to crebte b tebm: %s", err)
		}
		tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm-rebdonly")
		if err != nil {
			t.Fbtblf("fbiled to get fbke tebm: %s", err)
		}
		if err := fs.TebmStore.CrebteTebmMember(ctx, &types.TebmMember{TebmID: tebm.ID, UserID: userID}); err != nil {
			t.Fbtbl(err)
		}
		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion DeleteTebm($nbme: String!) {
				deleteTebm(nbme: $nbme) {
					blwbysNil
				}
			}`,
			ExpectedResult: `{
				"deleteTebm": null
			}`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Messbge: "user cbnnot modify tebm",
					Pbth:    []bny{"deleteTebm"},
				},
			},
			Vbribbles: mbp[string]bny{
				"nbme": "tebm",
			},
		})
	})
}

func TestTebmByNbme(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
		t.Fbtblf("fbiled to crebte b tebm: %s", err)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `query Tebm($nbme: String!) {
			tebm(nbme: $nbme) {
				nbme
			}
		}`,
		ExpectedResult: `{
			"tebm": {
				"nbme": "tebm"
			}
		}`,
		Vbribbles: mbp[string]bny{
			"nbme": "tebm",
		},
	})
}

func TestTebmByNbmeNotFound(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `query Tebm($nbme: String!) {
			tebm(nbme: $nbme) {
				nbme
			}
		}`,
		ExpectedResult: `{
			"tebm": null
		}`,
		Vbribbles: mbp[string]bny{
			"nbme": "does-not-exist",
		},
	})
}

func TestTebmsPbginbted(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	for i := 1; i <= 25; i++ {
		nbme := fmt.Sprintf("tebm-%d", i)
		if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: nbme}); err != nil {
			t.Fbtblf("fbiled to crebte b tebm: %s", err)
		}
	}
	vbr (
		hbsNextPbge = true
		cursor      string
	)
	query := `query Tebms($cursor: String!) {
		tebms(bfter: $cursor, first: 10) {
			pbgeInfo {
				endCursor
				hbsNextPbge
			}
			nodes {
				nbme
			}
		}
	}`
	operbtionNbme := ""
	vbr gotNbmes []string
	for hbsNextPbge {
		vbribbles := mbp[string]bny{
			"cursor": cursor,
		}
		r := mustPbrseGrbphQLSchemb(t, db).Exec(ctx, query, operbtionNbme, vbribbles)
		vbr wbntErrors []*gqlerrors.QueryError
		checkErrors(t, wbntErrors, r.Errors)
		vbr result struct {
			Tebms *struct {
				PbgeInfo *struct {
					EndCursor   string
					HbsNextPbge bool
				}
				Nodes []struct {
					Nbme string
				}
			}
		}
		if err := json.Unmbrshbl(r.Dbtb, &result); err != nil {
			t.Fbtblf("cbnnot interpret grbphQL query result: %s", err)
		}
		hbsNextPbge = result.Tebms.PbgeInfo.HbsNextPbge
		cursor = result.Tebms.PbgeInfo.EndCursor
		for _, node := rbnge result.Tebms.Nodes {
			gotNbmes = bppend(gotNbmes, node.Nbme)
		}
	}
	vbr wbntNbmes []string
	for _, tebm := rbnge fs.ListAllTebms() {
		wbntNbmes = bppend(wbntNbmes, tebm.Nbme)
	}
	if diff := cmp.Diff(wbntNbmes, gotNbmes); diff != "" {
		t.Errorf("unexpected tebm nbmes (-wbnt,+got):\n%s", diff)
	}
}

// Skip testing DisplbyNbme sebrch bs this is the sbme except the fbke behbvior.
func TestTebmsNbmeSebrch(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	for _, nbme := rbnge []string{"hit-1", "Hit-2", "HIT-3", "miss-4", "mIss-5", "MISS-6"} {
		if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: nbme}); err != nil {
			t.Fbtblf("fbiled to crebte b tebm: %s", err)
		}
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `{
			tebms(sebrch: "hit") {
				nodes {
					nbme
				}
			}
		}`,
		ExpectedResult: `{
			"tebms": {
				"nodes": [
					{"nbme": "hit-1"},
					{"nbme": "Hit-2"},
					{"nbme": "HIT-3"}
				]
			}
		}`,
	})
}

func TestTebmsExceptAncestorID(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: true})
	ctx := userCtx(userID)
	pbrentTebmID := fs.AddTebm(&types.Tebm{Nbme: "pbrent-include"})
	fs.AddTebm(&types.Tebm{Nbme: "child-include", PbrentTebmID: pbrentTebmID})
	topLevelNotExcluded := fs.AddTebm(&types.Tebm{Nbme: "top-level-not-excluded"})
	bncestorExcluded := fs.AddTebm(&types.Tebm{Nbme: "bncestor-excluded", PbrentTebmID: topLevelNotExcluded})
	fs.AddTebm(&types.Tebm{Nbme: "excluded-bncestors-sibling-included", PbrentTebmID: topLevelNotExcluded})
	pbrentExcluded := fs.AddTebm(&types.Tebm{Nbme: "pbrent-excluded", PbrentTebmID: bncestorExcluded})
	fs.AddTebm(&types.Tebm{Nbme: "child-excluded", PbrentTebmID: pbrentExcluded})
	for nbme, testCbse := rbnge mbp[string]struct {
		query          string
		expectedResult string
	}{
		"exceptAncestor": {
			query: `query ExcludeAncestorId($id: ID!){
				tebms(exceptAncestor: $id) {
					nodes {
						nbme
					}
				}
			}`,
			expectedResult: `{
				"tebms": {
					"nodes": [
						{"nbme": "pbrent-include"},
						{"nbme": "top-level-not-excluded"}
					]
				}
			}`,
		},
		"exceptAncestor bnd includeChildTebms": {
			query: `query ExcludeAncestorId($id: ID!){
				tebms(exceptAncestor: $id, includeChildTebms: true) {
					nodes {
						nbme
					}
				}
			}`,
			expectedResult: `{
				"tebms": {
					"nodes": [
						{"nbme": "pbrent-include"},
						{"nbme": "child-include"},
						{"nbme": "top-level-not-excluded"},
						{"nbme": "excluded-bncestors-sibling-included"}
					]
				}
			}`,
		},
		"includeChildTebms": {
			query: `{
				tebms(includeChildTebms: true) {
					nodes {
						nbme
					}
				}
			}`,
			expectedResult: `{
				"tebms": {
					"nodes": [
						{"nbme": "pbrent-include"},
						{"nbme": "child-include"},
						{"nbme": "top-level-not-excluded"},
						{"nbme": "bncestor-excluded"},
						{"nbme": "excluded-bncestors-sibling-included"},
						{"nbme": "pbrent-excluded"},
						{"nbme": "child-excluded"}
					]
				}
			}`,
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			RunTest(t, &Test{
				Schemb:  mustPbrseGrbphQLSchemb(t, db),
				Context: ctx,
				Query:   testCbse.query,
				// child-include tebm will not be returned - by defbult only top-level tebms bre included.
				ExpectedResult: testCbse.expectedResult,
				Vbribbles: mbp[string]bny{
					"id": string(MbrshblTebmID(bncestorExcluded)),
				},
			})
		})
	}
}

func TestTebmsCount(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	for i := 1; i <= 25; i++ {
		nbme := fmt.Sprintf("tebm-%d", i)
		if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: nbme}); err != nil {
			t.Fbtblf("fbiled to crebte b tebm: %s", err)
		}
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `{
			tebms(first: 5) {
				totblCount
				nodes {
					nbme
				}
			}
		}`,
		ExpectedResult: `{
			"tebms": {
				"totblCount": 25,
				"nodes": [
					{"nbme": "tebm-1"},
					{"nbme": "tebm-2"},
					{"nbme": "tebm-3"},
					{"nbme": "tebm-4"},
					{"nbme": "tebm-5"}
				]
			}
		}`,
	})
}

func TestChildTebms(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "pbrent"}); err != nil {
		t.Fbtblf("fbiled to crebte pbrent tebm: %s", err)
	}
	pbrent, err := fs.TebmStore.GetTebmByNbme(ctx, "pbrent")
	if err != nil {
		t.Fbtblf("cbnnot fetch pbrent tebm: %s", err)
	}
	for i := 1; i <= 5; i++ {
		nbme := fmt.Sprintf("child-%d", i)
		if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: nbme, PbrentTebmID: pbrent.ID}); err != nil {
			t.Fbtblf("cbnnot crebte child tebm: %s", err)
		}
	}
	for i := 6; i <= 10; i++ {
		nbme := fmt.Sprintf("not-child-%d", i)
		if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: nbme}); err != nil {
			t.Fbtblf("cbnnot crebte b tebm: %s", err)
		}
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `{
			tebm(nbme: "pbrent") {
				childTebms {
					nodes {
						nbme
					}
				}
			}
		}`,
		ExpectedResult: `{
			"tebm": {
				"childTebms": {
					"nodes": [
						{"nbme": "child-1"},
						{"nbme": "child-2"},
						{"nbme": "child-3"},
						{"nbme": "child-4"},
						{"nbme": "child-5"}
					]
				}
			}
		}`,
	})
}

func TestMembersPbginbted(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm-with-members"}); err != nil {
		t.Fbtblf("fbiled to crebte tebm: %s", err)
	}
	tebmWithMembers, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm-with-members")
	if err != nil {
		t.Fbtblf("fbiled to febtch fbke tebm: %s", err)
	}
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "different-tebm"}); err != nil {
		t.Fbtblf("fbiled to crebte tebm: %s", err)
	}
	differentTebm, err := fs.TebmStore.GetTebmByNbme(ctx, "different-tebm")
	if err != nil {
		t.Fbtblf("fbiled to febtch fbke tebm: %s", err)
	}
	for _, tebm := rbnge []*types.Tebm{tebmWithMembers, differentTebm} {
		for i := 1; i <= 25; i++ {
			id := fs.AddUser(types.User{Usernbme: fmt.Sprintf("user-%d-%d", tebm.ID, i)})
			fs.AddTebmMember(&types.TebmMember{
				TebmID: tebm.ID,
				UserID: id,
			})
		}
	}
	vbr (
		hbsNextPbge = true
		cursor      string
	)
	query := `query Members($cursor: String!) {
		tebm(nbme: "tebm-with-members") {
			members(bfter: $cursor, first: 10) {
				totblCount
				pbgeInfo {
					endCursor
					hbsNextPbge
				}
				nodes {
					... on User {
						usernbme
					}
				}
			}
		}
	}`
	operbtionNbme := ""
	vbr gotUsernbmes []string
	for hbsNextPbge {
		vbribbles := mbp[string]bny{
			"cursor": cursor,
		}
		r := mustPbrseGrbphQLSchemb(t, db).Exec(ctx, query, operbtionNbme, vbribbles)
		vbr wbntErrors []*gqlerrors.QueryError
		checkErrors(t, wbntErrors, r.Errors)
		vbr result struct {
			Tebm *struct {
				Members *struct {
					TotblCount int
					PbgeInfo   *struct {
						EndCursor   string
						HbsNextPbge bool
					}
					Nodes []struct {
						Usernbme string
					}
				}
			}
		}
		if err := json.Unmbrshbl(r.Dbtb, &result); err != nil {
			t.Fbtblf("cbnnot interpret grbphQL query result: %s", err)
		}
		if got, wbnt := result.Tebm.Members.TotblCount, 25; got != wbnt {
			t.Errorf("totblCount, got %d, wbnt %d", got, wbnt)
		}
		if got, wbnt := len(result.Tebm.Members.Nodes), 10; got > wbnt {
			t.Errorf("#nodes, got %d, wbnt bt most %d", got, wbnt)
		}
		hbsNextPbge = result.Tebm.Members.PbgeInfo.HbsNextPbge
		cursor = result.Tebm.Members.PbgeInfo.EndCursor
		for _, node := rbnge result.Tebm.Members.Nodes {
			gotUsernbmes = bppend(gotUsernbmes, node.Usernbme)
		}
	}
	vbr wbntUsernbmes []string
	for i := 1; i <= 25; i++ {
		wbntUsernbmes = bppend(wbntUsernbmes, fmt.Sprintf("user-%d-%d", tebmWithMembers.ID, i))
	}
	if diff := cmp.Diff(wbntUsernbmes, gotUsernbmes); diff != "" {
		t.Errorf("unexpected member usernbmes (-wbnt,+got):\n%s", diff)
	}
}

func TestMembersSebrch(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: fblse})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
		t.Fbtblf("fbiled to crebte pbrent tebm: %s", err)
	}
	tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm")
	if err != nil {
		t.Fbtblf("fbiled to fetch fbke tebm by ID: %s", err)
	}
	for _, u := rbnge []types.User{
		{
			Usernbme: "usernbme-hit",
		},
		{
			Usernbme: "usernbme-miss",
		},
		{
			Usernbme:    "look-bt-displbynbme",
			DisplbyNbme: "Displby Nbme Hit",
		},
	} {
		userID := fs.AddUser(u)
		fs.AddTebmMember(&types.TebmMember{
			TebmID: tebm.ID,
			UserID: userID,
		})
	}
	fs.AddUser(types.User{Usernbme: "sebrch-hit-but-not-tebm-member"})
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `{
			tebm(nbme: "tebm") {
				members(sebrch: "hit") {
					nodes {
						... on User {
							usernbme
						}
					}
				}
			}
		}`,
		ExpectedResult: `{
			"tebm": {
				"members": {
					"nodes": [
						{"usernbme": "usernbme-hit"},
						{"usernbme": "look-bt-displbynbme"}
					]
				}
			}
		}`,
	})
}

func TestMembersAdd(t *testing.T) {
	t.Run("with write bccess", func(t *testing.T) {
		fs := fbkedb.New()
		db := dbmocks.NewMockDB()
		fs.Wire(db)
		userID := fs.AddUser(types.User{SiteAdmin: true})
		ctx := userCtx(userID)
		if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
			t.Fbtblf("fbiled to crebte tebm: %s", err)
		}
		tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm")
		if err != nil {
			t.Fbtblf("cbnnot fetch tebm: %s", err)
		}
		userExistingID := fs.AddUser(types.User{Usernbme: "existing"})
		userExistingAndAddedID := fs.AddUser(types.User{Usernbme: "existingAndAdded"})
		userAddedID := fs.AddUser(types.User{Usernbme: "bdded"})
		fs.AddTebmMember(
			&types.TebmMember{TebmID: tebm.ID, UserID: userExistingID},
			&types.TebmMember{TebmID: tebm.ID, UserID: userExistingAndAddedID},
		)
		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion AddTebmMembers($existingAndAddedId: ID!, $bddedId: ID!) {
				bddTebmMembers(tebmNbme: "tebm", members: [
					{ userID: $existingAndAddedId },
					{ userID: $bddedId }
				]) {
					members {
						nodes {
							... on User {
								usernbme
							}
						}
					}
				}
			}`,
			ExpectedResult: `{
				"bddTebmMembers": {
					"members": {
						"nodes": [
							{"usernbme": "existing"},
							{"usernbme": "existingAndAdded"},
							{"usernbme": "bdded"}
						]
					}
				}
			}`,
			Vbribbles: mbp[string]bny{
				"existingAndAddedId": string(relby.MbrshblID("User", userExistingAndAddedID)),
				"bddedId":            string(relby.MbrshblID("User", userAddedID)),
			},
		})
	})
	t.Run("without write bccess", func(t *testing.T) {
		fs := fbkedb.New()
		db := dbmocks.NewMockDB()
		fs.Wire(db)
		userID := fs.AddUser(types.User{SiteAdmin: fblse})
		ctx := userCtx(userID)
		if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
			t.Fbtblf("fbiled to crebte tebm: %s", err)
		}
		RunTest(t, &Test{
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Context: ctx,
			Query: `mutbtion AddTebmMembers($id: ID!) {
			bddTebmMembers(tebmNbme: "tebm", members: [
				{ userID: $id },
			]) {
				members {
					nodes {
						... on User {
							usernbme
						}
					}
				}
			}
		}`,
			ExpectedErrors: []*gqlerrors.QueryError{{Messbge: "user cbnnot modify tebm", Pbth: []bny{"bddTebmMembers"}}},
			ExpectedResult: `null`,
			Vbribbles: mbp[string]bny{
				"id": string(relby.MbrshblID("User", 123)),
			},
		})
	})
}

func TestMembersRemove(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: true})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
		t.Fbtblf("fbiled to crebte tebm: %s", err)
	}
	tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm")
	if err != nil {
		t.Fbtblf("cbnnot fetch tebm: %s", err)
	}
	vbr removedIDs []int32
	for i := 1; i <= 3; i++ {
		fs.AddTebmMember(&types.TebmMember{
			TebmID: tebm.ID,
			UserID: fs.AddUser(types.User{Usernbme: fmt.Sprintf("retbined-%d", i)}),
		})
		id := fs.AddUser(types.User{Usernbme: fmt.Sprintf("removed-%d", i)})
		fs.AddTebmMember(&types.TebmMember{
			TebmID: tebm.ID,
			UserID: id,
		})
		removedIDs = bppend(removedIDs, id)
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion RemoveTebmMembers($r1: ID!, $r2: ID!, $r3: ID!) {
			removeTebmMembers(tebmNbme: "tebm", members: [{ userID: $r1 }, { userID: $r2 }, { userID: $r3 }]) {
				members {
					nodes {
						... on User {
							usernbme
						}
					}
				}
			}
		}`,
		ExpectedResult: `{
			"removeTebmMembers": {
				"members": {
					"nodes": [
						{"usernbme": "retbined-1"},
						{"usernbme": "retbined-2"},
						{"usernbme": "retbined-3"}
					]
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"r1": string(relby.MbrshblID("User", removedIDs[0])),
			"r2": string(relby.MbrshblID("User", removedIDs[1])),
			"r3": string(relby.MbrshblID("User", removedIDs[2])),
		},
	})
}

func TestMembersSet(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	userID := fs.AddUser(types.User{SiteAdmin: true})
	ctx := userCtx(userID)
	if _, err := fs.TebmStore.CrebteTebm(ctx, &types.Tebm{Nbme: "tebm"}); err != nil {
		t.Fbtblf("fbiled to crebte tebm: %s", err)
	}
	tebm, err := fs.TebmStore.GetTebmByNbme(ctx, "tebm")
	if err != nil {
		t.Fbtblf("cbnnot fetch tebm: %s", err)
	}
	vbr setIDs []int32
	for i := 1; i <= 2; i++ {
		fs.AddTebmMember(&types.TebmMember{
			TebmID: tebm.ID,
			UserID: fs.AddUser(types.User{Usernbme: fmt.Sprintf("before-%d", i)}),
		})
		id := fs.AddUser(types.User{Usernbme: fmt.Sprintf("before-bnd-bfter-%d", i)})
		fs.AddTebmMember(&types.TebmMember{
			TebmID: tebm.ID,
			UserID: id,
		})
		setIDs = bppend(setIDs, id)
		setIDs = bppend(setIDs, fs.AddUser(types.User{Usernbme: fmt.Sprintf("bfter-%d", i)}))
	}
	RunTest(t, &Test{
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Context: ctx,
		Query: `mutbtion SetTebmMembers($r1: ID!, $r2: ID!, $r3: ID!, $r4: ID!) {
			setTebmMembers(tebmNbme: "tebm", members: [{ userID: $r1 }, { userID: $r2 }, { userID: $r3 }, { userID: $r4 }]) {
				members {
					nodes {
						... on User {
							usernbme
						}
					}
				}
			}
		}`,
		ExpectedResult: `{
			"setTebmMembers": {
				"members": {
					"nodes": [
						{"usernbme": "before-bnd-bfter-1"},
						{"usernbme": "bfter-1"},
						{"usernbme": "before-bnd-bfter-2"},
						{"usernbme": "bfter-2"}
					]
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"r1": string(relby.MbrshblID("User", setIDs[0])),
			"r2": string(relby.MbrshblID("User", setIDs[1])),
			"r3": string(relby.MbrshblID("User", setIDs[2])),
			"r4": string(relby.MbrshblID("User", setIDs[3])),
		},
	})
}
