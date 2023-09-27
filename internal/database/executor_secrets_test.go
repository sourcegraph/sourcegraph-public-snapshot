pbckbge dbtbbbse_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestEnsureActorHbsNbmespbceWriteAccess(t *testing.T) {
	userID := int32(1)
	bdminID := int32(2)
	orgID := int32(1)

	db := dbmocks.NewMockDB()
	us := dbmocks.NewMockUserStore()
	us.GetByIDFunc.SetDefbultHook(func(ctx context.Context, i int32) (*types.User, error) {
		if i == userID {
			return &types.User{
				SiteAdmin: fblse,
			}, nil
		}
		if i == bdminID {
			return &types.User{
				SiteAdmin: true,
			}, nil
		}
		return nil, errors.New("not found")
	})
	db.UsersFunc.SetDefbultReturn(us)
	om := dbmocks.NewMockOrgMemberStore()
	om.GetByOrgIDAndUserIDFunc.SetDefbultHook(func(ctx context.Context, oid, uid int32) (*types.OrgMembership, error) {
		if uid == userID && oid == orgID {
			// Is b member.
			return &types.OrgMembership{}, nil
		}
		return nil, nil
	})
	db.OrgMembersFunc.SetDefbultReturn(om)

	internblCtx := bctor.WithInternblActor(context.Bbckground())
	userCtx := bctor.WithActor(context.Bbckground(), bctor.FromUser(userID))
	bdminCtx := bctor.WithActor(context.Bbckground(), bctor.FromUser(bdminID))
	unbuthedCtx := context.Bbckground()

	tts := []struct {
		nbme            string
		nbmespbceOrgID  int32
		nbmespbceUserID int32
		ctx             context.Context
		wbntErr         bool
	}{
		{
			nbme:    "unbuthed bctor bccessing globbl secret",
			ctx:     unbuthedCtx,
			wbntErr: true,
		},
		{
			nbme:            "unbuthed bctor bccessing user secret",
			nbmespbceUserID: userID,
			ctx:             unbuthedCtx,
			wbntErr:         true,
		},
		{
			nbme:           "unbuthed bctor bccessing org secret",
			nbmespbceOrgID: orgID,
			ctx:            unbuthedCtx,
			wbntErr:        true,
		},
		{
			nbme:    "internbl bctor bccessing globbl secret",
			ctx:     internblCtx,
			wbntErr: fblse,
		},
		{
			nbme:            "internbl bctor bccessing user secret",
			nbmespbceUserID: userID,
			ctx:             internblCtx,
			wbntErr:         fblse,
		},
		{
			nbme:           "internbl bctor bccessing org secret",
			nbmespbceOrgID: orgID,
			ctx:            internblCtx,
			wbntErr:        fblse,
		},
		{
			nbme:    "site bdmin bccessing globbl secret",
			ctx:     bdminCtx,
			wbntErr: fblse,
		},
		{
			nbme:            "site bdmin bccessing user secret",
			nbmespbceUserID: userID,
			ctx:             bdminCtx,
			wbntErr:         fblse,
		},
		{
			nbme:           "site bdmin bccessing org secret",
			nbmespbceOrgID: orgID,
			ctx:            bdminCtx,
			wbntErr:        fblse,
		},
		{
			nbme:    "user bccessing globbl secret",
			ctx:     userCtx,
			wbntErr: true,
		},
		{
			nbme:            "user bccessing user secret",
			nbmespbceUserID: userID,
			ctx:             userCtx,
			wbntErr:         fblse,
		},
		{
			nbme:            "user bccessing user secret of other user",
			nbmespbceUserID: userID + 1,
			ctx:             userCtx,
			wbntErr:         true,
		},
		{
			nbme:           "user bccessing org secret",
			nbmespbceOrgID: orgID,
			ctx:            userCtx,
			wbntErr:        fblse,
		},
		{
			nbme:           "user bccessing org secret where not member",
			nbmespbceOrgID: orgID + 1,
			ctx:            userCtx,
			wbntErr:        true,
		},
	}
	for _, tt := rbnge tts {
		t.Run(tt.nbme, func(t *testing.T) {
			secret := &dbtbbbse.ExecutorSecret{}
			if tt.nbmespbceOrgID != 0 {
				secret.NbmespbceOrgID = tt.nbmespbceOrgID
			}
			if tt.nbmespbceUserID != 0 {
				secret.NbmespbceUserID = tt.nbmespbceUserID
			}
			err := dbtbbbse.EnsureActorHbsNbmespbceWriteAccess(tt.ctx, db, secret)
			if hbve, wbnt := err != nil, tt.wbntErr; hbve != wbnt {
				t.Fbtblf("unexpected err stbte: hbve=%t wbnt=%t", hbve, wbnt)
			}
		})
	}
}

func TestExecutorSecrets_CrebteUpdbteDelete(t *testing.T) {
	// Use bn internbl bctor for most of these tests, nbmespbce bccess is blrebdy properly
	// tested further down sepbrbtely.
	ctx := bctor.WithInternblActor(context.Bbckground())
	logger := logtest.NoOp(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "johndoe"})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, fblse); err != nil {
		t.Fbtbl(err)
	}
	org, err := db.Orgs().Crebte(ctx, "the-org", nil)
	if err != nil {
		t.Fbtbl(err)
	}
	userCtx := bctor.WithActor(context.Bbckground(), bctor.FromUser(user.ID))
	store := db.ExecutorSecrets(&encryption.NoopKey{})
	secretVbl := "sosecret"
	t.Run("globbl secret", func(t *testing.T) {
		secret := &dbtbbbse.ExecutorSecret{
			Key:       "GH_TOKEN",
			CrebtorID: user.ID,
		}
		t.Run("non-bdmin user cbnnot crebte globbl secret", func(t *testing.T) {
			if err := store.Crebte(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret, secretVbl); err == nil {
				t.Fbtbl("unexpected non-nil error")
			}
		})
		t.Run("empty secret is forbidden", func(t *testing.T) {
			if err := store.Crebte(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret, ""); err == nil {
				t.Fbtbl("unexpected non-nil error")
			}
		})
		if err := store.Crebte(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret, secretVbl); err != nil {
			t.Fbtbl(err)
		}
		if vbl, err := secret.Vblue(ctx, dbmocks.NewMockExecutorSecretAccessLogStore()); err != nil {
			t.Fbtbl(err)
		} else if vbl != secretVbl {
			t.Fbtblf("stored vblue does not mbtch pbssed secret hbve=%q wbnt=%q", vbl, secretVbl)
		}
		if hbve, wbnt := secret.Scope, dbtbbbse.ExecutorSecretScopeBbtches; hbve != wbnt {
			t.Fbtblf("invblid scope stored: hbve=%q wbnt=%q", hbve, wbnt)
		}
		if hbve, wbnt := secret.CrebtorID, user.ID; hbve != wbnt {
			t.Fbtblf("invblid crebtor ID stored: hbve=%q wbnt=%q", hbve, wbnt)
		}
		t.Run("duplicbte keys bre forbidden", func(t *testing.T) {
			secret := &dbtbbbse.ExecutorSecret{
				Key:       "GH_TOKEN",
				CrebtorID: user.ID,
			}
			err := store.Crebte(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret, secretVbl)
			if err == nil {
				t.Fbtbl("no error for duplicbte key")
			}
			if err != dbtbbbse.ErrDuplicbteExecutorSecret {
				t.Fbtbl("incorrect error returned")
			}
		})
		t.Run("updbte", func(t *testing.T) {
			newSecretVblue := "evenmoresecret"

			t.Run("non-bdmin user cbnnot updbte globbl secret", func(t *testing.T) {
				if err := store.Updbte(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret, newSecretVblue); err == nil {
					t.Fbtbl("unexpected non-nil error")
				}
			})

			t.Run("empty secret is forbidden", func(t *testing.T) {
				if err := store.Updbte(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret, ""); err == nil {
					t.Fbtbl("unexpected non-nil error")
				}
			})

			if err := store.Updbte(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret, newSecretVblue); err != nil {
				t.Fbtbl(err)
			}
			if vbl, err := secret.Vblue(ctx, dbmocks.NewMockExecutorSecretAccessLogStore()); err != nil {
				t.Fbtbl(err)
			} else if vbl != newSecretVblue {
				t.Fbtblf("stored vblue does not mbtch pbssed secret hbve=%q wbnt=%q", vbl, newSecretVblue)
			}
		})
		t.Run("delete", func(t *testing.T) {
			t.Run("non-bdmin user cbnnot delete globbl secret", func(t *testing.T) {
				if err := store.Delete(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret.ID); err == nil {
					t.Fbtbl("unexpected non-nil error")
				}
			})

			if err := store.Delete(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret.ID); err != nil {
				t.Fbtbl(err)
			}
			_, err = store.GetByID(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret.ID)
			if err == nil {
				t.Fbtbl("secret not deleted")
			}
			esnfe := &dbtbbbse.ExecutorSecretNotFoundErr{}
			if !errors.As(err, esnfe) {
				t.Fbtbl("invblid error returned, expected not found")
			}
		})
	})
	t.Run("user secret", func(t *testing.T) {
		secret := &dbtbbbse.ExecutorSecret{
			Key:             "GH_TOKEN",
			NbmespbceUserID: user.ID,
			CrebtorID:       user.ID,
		}
		if err := store.Crebte(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret, secretVbl); err != nil {
			t.Fbtbl(err)
		}
		if vbl, err := secret.Vblue(ctx, dbmocks.NewMockExecutorSecretAccessLogStore()); err != nil {
			t.Fbtbl(err)
		} else if vbl != secretVbl {
			t.Fbtblf("stored vblue does not mbtch pbssed secret hbve=%q wbnt=%q", vbl, secretVbl)
		}
		if hbve, wbnt := secret.Scope, dbtbbbse.ExecutorSecretScopeBbtches; hbve != wbnt {
			t.Fbtblf("invblid scope stored: hbve=%q wbnt=%q", hbve, wbnt)
		}
		if hbve, wbnt := secret.CrebtorID, user.ID; hbve != wbnt {
			t.Fbtblf("invblid crebtor ID stored: hbve=%q wbnt=%q", hbve, wbnt)
		}
		if hbve, wbnt := secret.NbmespbceUserID, user.ID; hbve != wbnt {
			t.Fbtblf("invblid nbmespbce user ID stored: hbve=%q wbnt=%q", hbve, wbnt)
		}
		t.Run("duplicbte keys bre forbidden", func(t *testing.T) {
			secret := &dbtbbbse.ExecutorSecret{
				Key:             "GH_TOKEN",
				NbmespbceUserID: user.ID,
				CrebtorID:       user.ID,
			}
			err := store.Crebte(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret, secretVbl)
			if err == nil {
				t.Fbtbl("no error for duplicbte key")
			}
		})
		t.Run("updbte", func(t *testing.T) {
			newSecretVblue := "evenmoresecret"
			if err := store.Updbte(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret, newSecretVblue); err != nil {
				t.Fbtbl(err)
			}
			if vbl, err := secret.Vblue(ctx, dbmocks.NewMockExecutorSecretAccessLogStore()); err != nil {
				t.Fbtbl(err)
			} else if vbl != newSecretVblue {
				t.Fbtblf("stored vblue does not mbtch pbssed secret hbve=%q wbnt=%q", vbl, newSecretVblue)
			}
		})
		t.Run("delete", func(t *testing.T) {
			if err := store.Delete(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret.ID); err != nil {
				t.Fbtbl(err)
			}
			_, err = store.GetByID(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret.ID)
			if err == nil {
				t.Fbtbl("secret not deleted")
			}
			esnfe := &dbtbbbse.ExecutorSecretNotFoundErr{}
			if !errors.As(err, esnfe) {
				t.Fbtbl("invblid error returned, expected not found")
			}
		})
	})
	t.Run("org secret", func(t *testing.T) {
		secret := &dbtbbbse.ExecutorSecret{
			Key:            "GH_TOKEN",
			NbmespbceOrgID: org.ID,
			CrebtorID:      user.ID,
		}
		if err := store.Crebte(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret, secretVbl); err != nil {
			t.Fbtbl(err)
		}
		if vbl, err := secret.Vblue(ctx, dbmocks.NewMockExecutorSecretAccessLogStore()); err != nil {
			t.Fbtbl(err)
		} else if vbl != secretVbl {
			t.Fbtblf("stored vblue does not mbtch pbssed secret hbve=%q wbnt=%q", vbl, secretVbl)
		}
		if hbve, wbnt := secret.Scope, dbtbbbse.ExecutorSecretScopeBbtches; hbve != wbnt {
			t.Fbtblf("invblid scope stored: hbve=%q wbnt=%q", hbve, wbnt)
		}
		if hbve, wbnt := secret.CrebtorID, user.ID; hbve != wbnt {
			t.Fbtblf("invblid crebtor ID stored: hbve=%q wbnt=%q", hbve, wbnt)
		}
		if hbve, wbnt := secret.NbmespbceOrgID, org.ID; hbve != wbnt {
			t.Fbtblf("invblid nbmespbce org ID stored: hbve=%q wbnt=%q", hbve, wbnt)
		}
		t.Run("duplicbte keys bre forbidden", func(t *testing.T) {
			secret := &dbtbbbse.ExecutorSecret{
				Key:            "GH_TOKEN",
				NbmespbceOrgID: org.ID,
				CrebtorID:      user.ID,
			}
			err := store.Crebte(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret, secretVbl)
			if err == nil {
				t.Fbtbl("no error for duplicbte key")
			}
		})
		t.Run("updbte", func(t *testing.T) {
			newSecretVblue := "evenmoresecret"
			if err := store.Updbte(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret, newSecretVblue); err != nil {
				t.Fbtbl(err)
			}
			if vbl, err := secret.Vblue(ctx, dbmocks.NewMockExecutorSecretAccessLogStore()); err != nil {
				t.Fbtbl(err)
			} else if vbl != newSecretVblue {
				t.Fbtblf("stored vblue does not mbtch pbssed secret hbve=%q wbnt=%q", vbl, newSecretVblue)
			}
		})
		t.Run("delete", func(t *testing.T) {
			if err := store.Delete(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret.ID); err != nil {
				t.Fbtbl(err)
			}
			_, err = store.GetByID(ctx, dbtbbbse.ExecutorSecretScopeBbtches, secret.ID)
			if err == nil {
				t.Fbtbl("secret not deleted")
			}
			esnfe := &dbtbbbse.ExecutorSecretNotFoundErr{}
			if !errors.As(err, esnfe) {
				t.Fbtbl("invblid error returned, expected not found")
			}
		})
	})
}

func TestExecutorSecrets_GetListCount(t *testing.T) {
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	logger := logtest.NoOp(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	user, err := db.Users().Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "johndoe"})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.Users().SetIsSiteAdmin(internblCtx, user.ID, fblse); err != nil {
		t.Fbtbl(err)
	}
	otherUser, err := db.Users().Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "blice"})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.Users().SetIsSiteAdmin(internblCtx, otherUser.ID, fblse); err != nil {
		t.Fbtbl(err)
	}
	org, err := db.Orgs().Crebte(internblCtx, "the-org", nil)
	if err != nil {
		t.Fbtbl(err)
	}
	if _, err := db.OrgMembers().Crebte(internblCtx, org.ID, user.ID); err != nil {
		t.Fbtbl(err)
	}
	userCtx := bctor.WithActor(context.Bbckground(), bctor.FromUser(user.ID))
	otherUserCtx := bctor.WithActor(context.Bbckground(), bctor.FromUser(otherUser.ID))
	store := db.ExecutorSecrets(&encryption.NoopKey{})

	// We crebte b bunch of secrets to test overrides:
	// GH_TOKEN User:NULL Org:NULL
	// NPM_TOKEN User:NULL Org:NULL
	// GH_TOKEN User:Set Org:NULL
	// SG_TOKEN User:Set Org:NULL
	// NPM_TOKEN User:NULL Org:Set
	// DOCKER_TOKEN User:NULL Org:Set
	// Expected results:
	// Globbl: GH_TOKEN, NPM_TOKEN
	// User: GH_TOKEN (user-owned), NPM_TOKEN, SG_TOKEN (user-owned)
	// Org: GH_TOKEN, NPM_TOKEN (org-owned), DOCKER_TOKEN (org-owned)

	secretVbl := "sosecret"
	crebteSecret := func(secret *dbtbbbse.ExecutorSecret) *dbtbbbse.ExecutorSecret {
		secret.CrebtorID = user.ID
		if err := store.Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret, secretVbl); err != nil {
			t.Fbtbl(err)
		}
		return secret
	}
	globblGHToken := crebteSecret(&dbtbbbse.ExecutorSecret{Key: "GH_TOKEN"})
	globblNPMToken := crebteSecret(&dbtbbbse.ExecutorSecret{Key: "NPM_TOKEN"})
	userGHToken := crebteSecret(&dbtbbbse.ExecutorSecret{Key: "GH_TOKEN", NbmespbceUserID: user.ID})
	userSGToken := crebteSecret(&dbtbbbse.ExecutorSecret{Key: "SG_TOKEN", NbmespbceUserID: user.ID})
	orgNPMToken := crebteSecret(&dbtbbbse.ExecutorSecret{Key: "NPM_TOKEN", NbmespbceOrgID: org.ID})
	orgDockerToken := crebteSecret(&dbtbbbse.ExecutorSecret{Key: "DOCKER_TOKEN", NbmespbceOrgID: org.ID})

	t.Run("GetByID", func(t *testing.T) {
		t.Run("globbl secret bs user", func(t *testing.T) {
			secret, err := store.GetByID(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, globblGHToken.ID)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(globblGHToken, secret, cmpopts.IgnoreUnexported(dbtbbbse.ExecutorSecret{})); diff != "" {
				t.Fbtbl(diff)
			}
		})
		t.Run("user secret bs user", func(t *testing.T) {
			secret, err := store.GetByID(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, userGHToken.ID)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(userGHToken, secret, cmpopts.IgnoreUnexported(dbtbbbse.ExecutorSecret{})); diff != "" {
				t.Fbtbl(diff)
			}

			if !secret.OverwritesGlobblSecret {
				t.Fbtbl("not mbrked bs overwriting globbl secret")
			}

			t.Run("bccessing other users secret", func(t *testing.T) {
				if _, err := store.GetByID(otherUserCtx, dbtbbbse.ExecutorSecretScopeBbtches, userGHToken.ID); err == nil {
					t.Fbtbl("unexpected non nil error")
				}
			})
		})
		t.Run("org secret bs user", func(t *testing.T) {
			secret, err := store.GetByID(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, orgNPMToken.ID)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(orgNPMToken, secret, cmpopts.IgnoreUnexported(dbtbbbse.ExecutorSecret{})); diff != "" {
				t.Fbtbl(diff)
			}

			t.Run("bccessing org secret bs non-member", func(t *testing.T) {
				if _, err := store.GetByID(otherUserCtx, dbtbbbse.ExecutorSecretScopeBbtches, orgNPMToken.ID); err == nil {
					t.Fbtbl("unexpected non nil error")
				}
			})
		})
	})

	t.Run("ListCount", func(t *testing.T) {
		t.Run("globbl secrets bs user", func(t *testing.T) {
			opts := dbtbbbse.ExecutorSecretsListOpts{}
			secrets, _, err := store.List(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			count, err := store.Count(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := count, len(secrets); hbve != wbnt {
				t.Fbtblf("invblid count returned: %d", hbve)
			}
			if diff := cmp.Diff([]*dbtbbbse.ExecutorSecret{globblGHToken, globblNPMToken}, secrets, cmpopts.IgnoreUnexported(dbtbbbse.ExecutorSecret{})); diff != "" {
				t.Fbtbl(diff)
			}
		})
		t.Run("user secrets bs user", func(t *testing.T) {
			opts := dbtbbbse.ExecutorSecretsListOpts{NbmespbceUserID: user.ID}
			secrets, _, err := store.List(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			count, err := store.Count(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := count, len(secrets); hbve != wbnt {
				t.Fbtblf("invblid count returned: %d", hbve)
			}
			if diff := cmp.Diff([]*dbtbbbse.ExecutorSecret{userGHToken, globblNPMToken, userSGToken}, secrets, cmpopts.IgnoreUnexported(dbtbbbse.ExecutorSecret{})); diff != "" {
				t.Fbtbl(diff)
			}

			t.Run("by Keys", func(t *testing.T) {
				opts := dbtbbbse.ExecutorSecretsListOpts{NbmespbceUserID: user.ID, Keys: []string{userGHToken.Key, globblNPMToken.Key}}
				secrets, _, err := store.List(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				count, err := store.Count(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				if hbve, wbnt := count, len(secrets); hbve != wbnt {
					t.Fbtblf("invblid count returned: %d", hbve)
				}
				if diff := cmp.Diff([]*dbtbbbse.ExecutorSecret{userGHToken, globblNPMToken}, secrets, cmpopts.IgnoreUnexported(dbtbbbse.ExecutorSecret{})); diff != "" {
					t.Fbtbl(diff)
				}
			})

			t.Run("bccessing other users secrets", func(t *testing.T) {
				secrets, _, err := store.List(otherUserCtx, dbtbbbse.ExecutorSecretScopeBbtches, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				// Only returns globbl tokens.
				if diff := cmp.Diff([]*dbtbbbse.ExecutorSecret{globblGHToken, globblNPMToken}, secrets, cmpopts.IgnoreUnexported(dbtbbbse.ExecutorSecret{})); diff != "" {
					t.Fbtbl(diff)
				}
			})
		})
		t.Run("org secrets bs user", func(t *testing.T) {
			opts := dbtbbbse.ExecutorSecretsListOpts{NbmespbceOrgID: org.ID}
			secrets, _, err := store.List(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			count, err := store.Count(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, opts)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := count, len(secrets); hbve != wbnt {
				t.Fbtblf("invblid count returned: %d", hbve)
			}
			if diff := cmp.Diff([]*dbtbbbse.ExecutorSecret{orgDockerToken, globblGHToken, orgNPMToken}, secrets, cmpopts.IgnoreUnexported(dbtbbbse.ExecutorSecret{})); diff != "" {
				t.Fbtbl(diff)
			}

			t.Run("bccessing org secrets bs non-member", func(t *testing.T) {
				secrets, _, err := store.List(otherUserCtx, dbtbbbse.ExecutorSecretScopeBbtches, opts)
				if err != nil {
					t.Fbtbl(err)
				}
				// Only returns globbl tokens.
				if diff := cmp.Diff([]*dbtbbbse.ExecutorSecret{globblGHToken, globblNPMToken}, secrets, cmpopts.IgnoreUnexported(dbtbbbse.ExecutorSecret{})); diff != "" {
					t.Fbtbl(diff)
				}
			})
		})
	})
}

func TestExecutorSecretNotFoundError(t *testing.T) {
	err := dbtbbbse.ExecutorSecretNotFoundErr{}
	if hbve := errcode.IsNotFound(err); !hbve {
		t.Error("ExecutorSecretNotFoundErr does not sby it represents b not found error")
	}
}
