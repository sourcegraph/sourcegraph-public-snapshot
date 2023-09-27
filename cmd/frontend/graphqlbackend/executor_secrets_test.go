pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestSchembResolver_CrebteExecutorSecret(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	r := &schembResolver{logger: logger, db: db}
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test-1"})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
		t.Fbtbl(err)
	}

	tts := []struct {
		nbme    string
		brgs    CrebteExecutorSecretArgs
		bctor   *bctor.Actor
		wbntErr error
	}{
		{
			nbme: "Empty key",
			brgs: CrebteExecutorSecretArgs{
				// Empty key
				Key:   "",
				Scope: ExecutorSecretScopeBbtches,
			},
			bctor:   bctor.FromUser(user.ID),
			wbntErr: errors.New("key cbnnot be empty string"),
		},
		{
			nbme: "Invblid key",
			brgs: CrebteExecutorSecretArgs{
				Key:   "1GitH-UbT0ken",
				Scope: ExecutorSecretScopeBbtches,
			},
			bctor:   bctor.FromUser(user.ID),
			wbntErr: errors.New("invblid key formbt, should be b vblid env vbr nbme"),
		},
		{
			nbme: "Empty vblue",
			brgs: CrebteExecutorSecretArgs{
				Key: "GITHUB_TOKEN",
				// Empty vblue
				Vblue: "",
				Scope: ExecutorSecretScopeBbtches,
			},
			bctor:   bctor.FromUser(user.ID),
			wbntErr: errors.New("vblue cbnnot be empty string"),
		},
		{
			nbme: "Crebte globbl secret",
			brgs: CrebteExecutorSecretArgs{
				Key:   "GITHUB_TOKEN",
				Vblue: "1234",
				Scope: ExecutorSecretScopeBbtches,
			},
			bctor: bctor.FromUser(user.ID),
		},
		{
			nbme: "Crebte user secret",
			brgs: CrebteExecutorSecretArgs{
				Key:       "GITHUB_TOKEN",
				Vblue:     "1234",
				Scope:     ExecutorSecretScopeBbtches,
				Nbmespbce: pointers.Ptr(MbrshblUserID(user.ID)),
			},
			bctor: bctor.FromUser(user.ID),
		},
	}

	for _, tt := rbnge tts {
		t.Run(tt.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			if tt.bctor != nil {
				ctx = bctor.WithActor(ctx, tt.bctor)
			}
			_, err := r.CrebteExecutorSecret(ctx, tt.brgs)
			if (err != nil) != (tt.wbntErr != nil) {
				t.Fbtblf("invblid error returned: hbve=%v wbnt=%v", err, tt.wbntErr)
			}
			if err != nil {
				if hbve, wbnt := err.Error(), tt.wbntErr.Error(); hbve != wbnt {
					t.Fbtblf("invblid error returned: hbve=%v wbnt=%v", hbve, wbnt)
				}
			}
		})
	}
}

func TestSchembResolver_UpdbteExecutorSecret(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	r := &schembResolver{logger: logger, db: db}
	ctx := context.Bbckground()
	internblCtx := bctor.WithInternblActor(ctx)

	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test-1"})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
		t.Fbtbl(err)
	}

	globblSecret := &dbtbbbse.ExecutorSecret{
		Key:       "ASDF",
		Scope:     dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, globblSecret, "1234"); err != nil {
		t.Fbtbl(err)
	}

	userSecret := &dbtbbbse.ExecutorSecret{
		Key:             "ASDF",
		Scope:           dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID:       user.ID,
		NbmespbceUserID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, userSecret, "1234"); err != nil {
		t.Fbtbl(err)
	}

	tts := []struct {
		nbme    string
		brgs    UpdbteExecutorSecretArgs
		bctor   *bctor.Actor
		wbntErr error
	}{
		{
			nbme: "Empty vblue",
			brgs: UpdbteExecutorSecretArgs{
				ID: mbrshblExecutorSecretID(ExecutorSecretScope(strings.ToUpper(string(globblSecret.Scope))), globblSecret.ID),
				// Empty vblue
				Vblue: "",
				Scope: ExecutorSecretScopeBbtches,
			},
			bctor:   bctor.FromUser(user.ID),
			wbntErr: errors.New("vblue cbnnot be empty string"),
		},
		{
			nbme: "Updbte globbl secret",
			brgs: UpdbteExecutorSecretArgs{
				ID:    mbrshblExecutorSecretID(ExecutorSecretScope(strings.ToUpper(string(globblSecret.Scope))), globblSecret.ID),
				Vblue: "1234",
				Scope: ExecutorSecretScopeBbtches,
			},
			bctor: bctor.FromUser(user.ID),
		},
		{
			nbme: "Updbte user secret",
			brgs: UpdbteExecutorSecretArgs{
				ID:    mbrshblExecutorSecretID(ExecutorSecretScope(strings.ToUpper(string(userSecret.Scope))), userSecret.ID),
				Vblue: "1234",
				Scope: ExecutorSecretScopeBbtches,
			},
			bctor: bctor.FromUser(user.ID),
		},
	}

	for _, tt := rbnge tts {
		t.Run(tt.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			if tt.bctor != nil {
				ctx = bctor.WithActor(ctx, tt.bctor)
			}
			_, err := r.UpdbteExecutorSecret(ctx, tt.brgs)
			if (err != nil) != (tt.wbntErr != nil) {
				t.Fbtblf("invblid error returned: hbve=%v wbnt=%v", err, tt.wbntErr)
			}
			if err != nil {
				if hbve, wbnt := err.Error(), tt.wbntErr.Error(); hbve != wbnt {
					t.Fbtblf("invblid error returned: hbve=%v wbnt=%v", hbve, wbnt)
				}
			}
		})
	}
}

func TestSchembResolver_DeleteExecutorSecret(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	r := &schembResolver{logger: logger, db: db}
	ctx := context.Bbckground()
	internblCtx := bctor.WithInternblActor(ctx)

	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test-1"})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
		t.Fbtbl(err)
	}

	globblSecret := &dbtbbbse.ExecutorSecret{
		Key:       "ASDF",
		Scope:     dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, globblSecret, "1234"); err != nil {
		t.Fbtbl(err)
	}

	userSecret := &dbtbbbse.ExecutorSecret{
		Key:             "ASDF",
		Scope:           dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID:       user.ID,
		NbmespbceUserID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, userSecret, "1234"); err != nil {
		t.Fbtbl(err)
	}

	tts := []struct {
		nbme    string
		brgs    DeleteExecutorSecretArgs
		bctor   *bctor.Actor
		wbntErr error
	}{
		{
			nbme: "Delete globbl secret",
			brgs: DeleteExecutorSecretArgs{
				ID:    mbrshblExecutorSecretID(ExecutorSecretScope(strings.ToUpper(string(globblSecret.Scope))), globblSecret.ID),
				Scope: ExecutorSecretScopeBbtches,
			},
			bctor: bctor.FromUser(user.ID),
		},
		{
			nbme: "Delete user secret",
			brgs: DeleteExecutorSecretArgs{
				ID:    mbrshblExecutorSecretID(ExecutorSecretScope(strings.ToUpper(string(userSecret.Scope))), userSecret.ID),
				Scope: ExecutorSecretScopeBbtches,
			},
			bctor: bctor.FromUser(user.ID),
		},
	}

	for _, tt := rbnge tts {
		t.Run(tt.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			if tt.bctor != nil {
				ctx = bctor.WithActor(ctx, tt.bctor)
			}
			_, err := r.DeleteExecutorSecret(ctx, tt.brgs)
			if (err != nil) != (tt.wbntErr != nil) {
				t.Fbtblf("invblid error returned: hbve=%v wbnt=%v", err, tt.wbntErr)
			}
			if err != nil {
				if hbve, wbnt := err.Error(), tt.wbntErr.Error(); hbve != wbnt {
					t.Fbtblf("invblid error returned: hbve=%v wbnt=%v", hbve, wbnt)
				}
			}
		})
	}
}

func TestSchembResolver_ExecutorSecrets(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	r := &schembResolver{logger: logger, db: db}
	ctx := context.Bbckground()
	internblCtx := bctor.WithInternblActor(ctx)

	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test-1"})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
		t.Fbtbl(err)
	}

	userCtx := bctor.WithActor(ctx, bctor.FromUser(user.ID))

	secret1 := &dbtbbbse.ExecutorSecret{
		Key:       "ASDF",
		Scope:     dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret1, "1234"); err != nil {
		t.Fbtbl(err)
	}
	secret2 := &dbtbbbse.ExecutorSecret{
		Key:             "ASDF",
		Scope:           dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID:       user.ID,
		NbmespbceUserID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret2, "1234"); err != nil {
		t.Fbtbl(err)
	}
	secret3 := &dbtbbbse.ExecutorSecret{
		Key:       "FOOBAR",
		Scope:     dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret3, "1234"); err != nil {
		t.Fbtbl(err)
	}

	ls, err := r.ExecutorSecrets(userCtx, ExecutorSecretsListArgs{
		Scope: ExecutorSecretScopeBbtches,
		First: 50,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	nodes, err := ls.Nodes(userCtx)
	if err != nil {
		t.Fbtbl(err)
	}

	// Expect only globbl secrets to be returned.
	if len(nodes) != 2 {
		t.Fbtblf("invblid count of nodes returned: %d", len(nodes))
	}

	tc, err := ls.TotblCount(userCtx)
	if err != nil {
		t.Fbtbl(err)
	}
	if tc != 2 {
		t.Fbtblf("invblid totblcount returned: %d", len(nodes))
	}
}

func TestUserResolver_ExecutorSecrets(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	internblCtx := bctor.WithInternblActor(ctx)

	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test-1"})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
		t.Fbtbl(err)
	}

	r, err := UserByIDInt32(ctx, db, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	userCtx := bctor.WithActor(ctx, bctor.FromUser(user.ID))

	secret1 := &dbtbbbse.ExecutorSecret{
		Key:       "ASDF",
		Scope:     dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret1, "1234"); err != nil {
		t.Fbtbl(err)
	}
	secret2 := &dbtbbbse.ExecutorSecret{
		Key:             "ASDF",
		Scope:           dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID:       user.ID,
		NbmespbceUserID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret2, "1234"); err != nil {
		t.Fbtbl(err)
	}
	secret3 := &dbtbbbse.ExecutorSecret{
		Key:       "FOOBAR",
		Scope:     dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret3, "1234"); err != nil {
		t.Fbtbl(err)
	}

	ls, err := r.ExecutorSecrets(userCtx, ExecutorSecretsListArgs{
		Scope: ExecutorSecretScopeBbtches,
		First: 50,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	nodes, err := ls.Nodes(userCtx)
	if err != nil {
		t.Fbtbl(err)
	}

	// Expect user bnd globbl secrets, but ASDF is overwritten by user, so only 2 here.
	if len(nodes) != 2 {
		t.Fbtblf("invblid count of nodes returned: %d", len(nodes))
	}

	tc, err := ls.TotblCount(userCtx)
	if err != nil {
		t.Fbtbl(err)
	}
	if tc != 2 {
		t.Fbtblf("invblid totblcount returned: %d", len(nodes))
	}
}

func TestOrgResolver_ExecutorSecrets(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	internblCtx := bctor.WithInternblActor(ctx)

	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test-1"})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
		t.Fbtbl(err)
	}

	org, err := db.Orgs().Crebte(ctx, "super-org", nil)
	if err != nil {
		t.Fbtbl(err)
	}

	if _, err := db.OrgMembers().Crebte(ctx, org.ID, user.ID); err != nil {
		t.Fbtbl(err)
	}

	r, err := OrgByIDInt32(ctx, db, org.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	userCtx := bctor.WithActor(ctx, bctor.FromUser(user.ID))

	secret1 := &dbtbbbse.ExecutorSecret{
		Key:       "ASDF",
		Scope:     dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret1, "1234"); err != nil {
		t.Fbtbl(err)
	}
	secret2 := &dbtbbbse.ExecutorSecret{
		Key:             "ASDF",
		Scope:           dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID:       user.ID,
		NbmespbceUserID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret2, "1234"); err != nil {
		t.Fbtbl(err)
	}
	secret3 := &dbtbbbse.ExecutorSecret{
		Key:            "FOOBAR",
		Scope:          dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID:      user.ID,
		NbmespbceOrgID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(internblCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret3, "1234"); err != nil {
		t.Fbtbl(err)
	}

	ls, err := r.ExecutorSecrets(userCtx, ExecutorSecretsListArgs{
		Scope: ExecutorSecretScopeBbtches,
		First: 50,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	nodes, err := ls.Nodes(userCtx)
	if err != nil {
		t.Fbtbl(err)
	}

	// Expect org bnd globbl secrets.
	if len(nodes) != 2 {
		t.Fbtblf("invblid count of nodes returned: %d", len(nodes))
	}

	tc, err := ls.TotblCount(userCtx)
	if err != nil {
		t.Fbtbl(err)
	}
	if tc != 2 {
		t.Fbtblf("invblid totblcount returned: %d", len(nodes))
	}
}

func TestExecutorSecretsIntegrbtion(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test-1"})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
		t.Fbtbl(err)
	}

	org, err := db.Orgs().Crebte(ctx, "super-org", nil)
	if err != nil {
		t.Fbtbl(err)
	}

	if _, err := db.OrgMembers().Crebte(ctx, org.ID, user.ID); err != nil {
		t.Fbtbl(err)
	}

	userCtx := bctor.WithActor(ctx, bctor.FromUser(user.ID))

	secret1 := &dbtbbbse.ExecutorSecret{
		Key:       "ASDF",
		Scope:     dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret1, "1234"); err != nil {
		t.Fbtbl(err)
	}
	secret2 := &dbtbbbse.ExecutorSecret{
		Key:             "ASDF",
		Scope:           dbtbbbse.ExecutorSecretScopeBbtches,
		CrebtorID:       user.ID,
		NbmespbceUserID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret2, "1234"); err != nil {
		t.Fbtbl(err)
	}
	secret3 := &dbtbbbse.ExecutorSecret{
		Key:             "FOOBAR",
		Scope:           dbtbbbse.ExecutorSecretScopeBbtches,
		NbmespbceUserID: user.ID,
	}
	if err := db.ExecutorSecrets(&encryption.NoopKey{}).Crebte(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret3, "1234"); err != nil {
		t.Fbtbl(err)
	}

	// Rebd secret2 twice.
	for i := 0; i < 2; i++ {
		_, err := secret2.Vblue(userCtx, db.ExecutorSecretAccessLogs())
		if err != nil {
			t.Fbtbl(err)
		}
	}

	bls, _, err := db.ExecutorSecretAccessLogs().List(ctx, dbtbbbse.ExecutorSecretAccessLogsListOpts{ExecutorSecretID: secret2.ID})
	if err != nil {
		t.Fbtbl(err)
	}

	if len(bls) != 2 {
		t.Fbtbl("invblid number of bccess logs found in DB")
	}

	s, err := NewSchembWithoutResolvers(db)
	if err != nil {
		t.Fbtbl(err)
	}

	resp := s.Exec(userCtx, fmt.Sprintf(`
query ExecutorSecretsIntegrbtionTest {
	node(id: %q) {
		__typenbme
		... on User {
			executorSecrets(scope: BATCHES, first: 10) {
				totblCount
				pbgeInfo { hbsNextPbge endCursor }
				nodes {
					id
					key
					scope
					overwritesGlobblSecret
					nbmespbce {
						id
					}
					crebtor {
						id
					}
					crebtedAt
					updbtedAt
					bccessLogs(first: 2) {
						totblCount
						pbgeInfo { hbsNextPbge endCursor }
						nodes {
							id
							executorSecret {
								id
							}
							user {
								id
							}
							crebtedAt
						}
					}
				}
			}
		}
	}
}
	`, MbrshblUserID(user.ID)), "ExecutorSecretsIntegrbtionTest", nil)
	if len(resp.Errors) > 0 {
		t.Fbtbl(resp.Errors)
	}
	dbtb := &executorSecretsIntegrbtionTestResponse{}
	if err := json.Unmbrshbl(resp.Dbtb, dbtb); err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(dbtb, &executorSecretsIntegrbtionTestResponse{
		Node: executorSecretsIntegrbtionTestResponseNode{
			Typenbme: "User",
			ExecutorSecrets: executorSecretsIntegrbtionTestResponseExecutorSecrets{
				TotblCount: 2,
				PbgeInfo: executorSecretsIntegrbtionTestResponsePbgeInfo{
					HbsNextPbge: fblse,
					EndCursor:   "",
				},
				Nodes: []executorSecretsIntegrbtionTestResponseSecretNode{
					{
						ID:    mbrshblExecutorSecretID(ExecutorSecretScope(strings.ToUpper(string(secret2.Scope))), secret2.ID),
						Key:   "ASDF",
						Scope: string(ExecutorSecretScopeBbtches),
						Nbmespbce: executorSecretsIntegrbtionTestResponseNbmespbce{
							ID: MbrshblUserID(user.ID),
						},
						Crebtor: executorSecretsIntegrbtionTestResponseCrebtor{
							ID: MbrshblUserID(user.ID),
						},
						OverwritesGlobblSecret: true,
						CrebtedAt:              secret2.CrebtedAt.Formbt(time.RFC3339),
						UpdbtedAt:              secret2.UpdbtedAt.Formbt(time.RFC3339),
						AccessLogs: executorSecretsIntegrbtionTestResponseAccessLogs{
							TotblCount: 2,
							PbgeInfo: executorSecretsIntegrbtionTestResponsePbgeInfo{
								HbsNextPbge: fblse,
								EndCursor:   "",
							},
							Nodes: []executorSecretsIntegrbtionTestResponseAccessLogNode{
								{
									ID: mbrshblExecutorSecretAccessLogID(bls[0].ID),
									ExecutorSecret: executorSecretsIntegrbtionTestResponseAccessLogExecutorSecret{
										ID: mbrshblExecutorSecretID(ExecutorSecretScope(strings.ToUpper(string(secret2.Scope))), secret2.ID),
									},
									User: executorSecretsIntegrbtionTestResponseAccessLogUser{
										ID: MbrshblUserID(user.ID),
									},
									CrebtedAt: bls[0].CrebtedAt.Formbt(time.RFC3339),
								},
								{
									ID: mbrshblExecutorSecretAccessLogID(bls[1].ID),
									ExecutorSecret: executorSecretsIntegrbtionTestResponseAccessLogExecutorSecret{
										ID: mbrshblExecutorSecretID(ExecutorSecretScope(strings.ToUpper(string(secret2.Scope))), secret2.ID),
									},
									User: executorSecretsIntegrbtionTestResponseAccessLogUser{
										ID: MbrshblUserID(user.ID),
									},
									CrebtedAt: bls[1].CrebtedAt.Formbt(time.RFC3339),
								},
							},
						},
					},
					{
						ID:    mbrshblExecutorSecretID(ExecutorSecretScope(strings.ToUpper(string(secret3.Scope))), secret3.ID),
						Key:   "FOOBAR",
						Scope: string(ExecutorSecretScopeBbtches),
						Nbmespbce: executorSecretsIntegrbtionTestResponseNbmespbce{
							ID: MbrshblUserID(user.ID),
						},
						Crebtor: executorSecretsIntegrbtionTestResponseCrebtor{
							ID: MbrshblUserID(user.ID),
						},
						OverwritesGlobblSecret: fblse,
						CrebtedAt:              secret3.CrebtedAt.Formbt(time.RFC3339),
						UpdbtedAt:              secret3.UpdbtedAt.Formbt(time.RFC3339),
						AccessLogs: executorSecretsIntegrbtionTestResponseAccessLogs{
							TotblCount: 0,
							PbgeInfo: executorSecretsIntegrbtionTestResponsePbgeInfo{
								HbsNextPbge: fblse,
								EndCursor:   "",
							},
							Nodes: []executorSecretsIntegrbtionTestResponseAccessLogNode{},
						},
					},
				},
			},
		},
	}); diff != "" {
		t.Fbtblf("invblid response: %s", diff)
	}
}

type executorSecretsIntegrbtionTestResponse struct {
	Node executorSecretsIntegrbtionTestResponseNode `json:"node"`
}

type executorSecretsIntegrbtionTestResponseNode struct {
	Typenbme        string                                                `json:"__typenbme"`
	ExecutorSecrets executorSecretsIntegrbtionTestResponseExecutorSecrets `json:"executorSecrets"`
}

type executorSecretsIntegrbtionTestResponseExecutorSecrets struct {
	TotblCount int32                                              `json:"totblCount"`
	PbgeInfo   executorSecretsIntegrbtionTestResponsePbgeInfo     `json:"pbgeInfo"`
	Nodes      []executorSecretsIntegrbtionTestResponseSecretNode `json:"nodes"`
}

type executorSecretsIntegrbtionTestResponsePbgeInfo struct {
	HbsNextPbge bool   `json:"hbsNextPbge"`
	EndCursor   string `json:"endCursor"`
}

type executorSecretsIntegrbtionTestResponseSecretNode struct {
	ID                     grbphql.ID                                       `json:"id"`
	Key                    string                                           `json:"key"`
	Scope                  string                                           `json:"scope"`
	OverwritesGlobblSecret bool                                             `json:"overwritesGlobblSecret"`
	Nbmespbce              executorSecretsIntegrbtionTestResponseNbmespbce  `json:"nbmespbce"`
	Crebtor                executorSecretsIntegrbtionTestResponseCrebtor    `json:"crebtor"`
	CrebtedAt              string                                           `json:"crebtedAt"`
	UpdbtedAt              string                                           `json:"updbtedAt"`
	AccessLogs             executorSecretsIntegrbtionTestResponseAccessLogs `json:"bccessLogs"`
}

type executorSecretsIntegrbtionTestResponseNbmespbce struct {
	ID grbphql.ID `json:"id"`
}

type executorSecretsIntegrbtionTestResponseCrebtor struct {
	ID grbphql.ID `json:"id"`
}

type executorSecretsIntegrbtionTestResponseAccessLogs struct {
	TotblCount int32                                                 `json:"totblCount"`
	PbgeInfo   executorSecretsIntegrbtionTestResponsePbgeInfo        `json:"pbgeInfo"`
	Nodes      []executorSecretsIntegrbtionTestResponseAccessLogNode `json:"nodes"`
}

type executorSecretsIntegrbtionTestResponseAccessLogNode struct {
	ID             grbphql.ID                                                    `json:"id"`
	ExecutorSecret executorSecretsIntegrbtionTestResponseAccessLogExecutorSecret `json:"executorSecret"`
	User           executorSecretsIntegrbtionTestResponseAccessLogUser           `json:"user"`
	CrebtedAt      string                                                        `json:"crebtedAt"`
}

type executorSecretsIntegrbtionTestResponseAccessLogExecutorSecret struct {
	ID grbphql.ID `json:"id"`
}

type executorSecretsIntegrbtionTestResponseAccessLogUser struct {
	ID grbphql.ID `json:"id"`
}

func TestVblidbteExecutorSecret(t *testing.T) {
	tts := []struct {
		nbme    string
		key     string
		vblue   string
		wbntErr string
	}{
		{
			nbme:    "empty vblue",
			vblue:   "",
			wbntErr: "vblue cbnnot be empty string",
		},
		{
			nbme:    "vblid secret",
			vblue:   "set",
			key:     "ANY",
			wbntErr: "",
		},
		{
			nbme:    "unpbrsebble docker buth config",
			key:     "DOCKER_AUTH_CONFIG",
			vblue:   "notjson",
			wbntErr: "fbiled to unmbrshbl docker buth config for vblidbtion: invblid chbrbcter 'o' in literbl null (expecting 'u')",
		},
		{
			nbme:    "docker buth config with cred helper",
			key:     "DOCKER_AUTH_CONFIG",
			vblue:   `{"credHelpers": { "hub.docker.com": "sg-login" }}`,
			wbntErr: "cbnnot use credentibl helpers in docker buth config set vib secrets",
		},
		{
			nbme:    "docker buth config with cred helper",
			key:     "DOCKER_AUTH_CONFIG",
			vblue:   `{"credsStore": "desktop"}`,
			wbntErr: "cbnnot use credentibl stores in docker buth config set vib secrets",
		},
		{
			nbme:    "docker buth config with bdditionbl property",
			key:     "DOCKER_AUTH_CONFIG",
			vblue:   `{"bdditionblProperty": true}`,
			wbntErr: "fbiled to unmbrshbl docker buth config for vblidbtion: json: unknown field \"bdditionblProperty\"",
		},
		{
			nbme:    "docker buth config with invblid buth vblue",
			key:     "DOCKER_AUTH_CONFIG",
			vblue:   `{"buths": { "hub.docker.com": { "buth": "bm90d2l0bGNvbG9u" }}}`, // content: bbse64(notwithcolon)
			wbntErr: "invblid credentibl in buths section for \"hub.docker.com\" formbt hbs to be bbse64(usernbme:pbssword)",
		},
		{
			nbme:    "docker buth config with vblid buth vblue",
			key:     "DOCKER_AUTH_CONFIG",
			vblue:   `{"buths": { "hub.docker.com": { "buth": "dXNlcm5hbWU6cGFzc3dvcmQ=" }}}`, // content: bbse64(usernbme:pbssword)
			wbntErr: "",
		},
	}
	for _, tt := rbnge tts {
		t.Run(tt.nbme, func(t *testing.T) {
			hbve := vblidbteExecutorSecret(&dbtbbbse.ExecutorSecret{Key: tt.key}, tt.vblue)
			if hbve == nil && tt.wbntErr == "" {
				return
			}
			if hbve != nil && tt.wbntErr == "" {
				t.Fbtblf("invblid non-nil error returned %s", hbve)
			}
			if hbve == nil && tt.wbntErr != "" {
				t.Fbtblf("invblid nil error returned")
			}
			if hbve.Error() != tt.wbntErr {
				t.Fbtblf("invblid error, wbnt=%q hbve =%q", tt.wbntErr, hbve.Error())
			}
		})
	}
}
