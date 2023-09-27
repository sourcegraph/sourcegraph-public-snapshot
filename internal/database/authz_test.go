pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"sync/btomic"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr now = timeutil.Now().UnixNbno()

func clock() time.Time {
	return time.Unix(0, btomic.LobdInt64(&now))
}

func TestAuthzStore_GrbntPendingPermissions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte repos needed
	for _, repoID := rbnge []bpi.RepoID{1, 2, 3} {
		err := db.Repos().Crebte(ctx, &types.Repo{
			ID:   repoID,
			Nbme: bpi.RepoNbme(fmt.Sprintf("repo-%d", repoID)),
		})
		require.NoError(t, err)
	}

	// Crebte user with initiblly verified embil
	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:           "blice@exbmple.com",
		Usernbme:        "blice",
		EmbilIsVerified: true,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	code := "verify-code"

	// Add bnd verify the second embil
	err = db.UserEmbils().Add(ctx, user.ID, "blice2@exbmple.com", &code)
	if err != nil {
		t.Fbtbl(err)
	}
	err = db.UserEmbils().SetVerified(ctx, user.ID, "blice2@exbmple.com", true)
	if err != nil {
		t.Fbtbl(err)
	}

	// Add third embil bnd lebve bs unverified
	err = db.UserEmbils().Add(ctx, user.ID, "blice3@exbmple.com", &code)
	if err != nil {
		t.Fbtbl(err)
	}

	// Add two externbl bccounts
	err = db.UserExternblAccounts().AssocibteUserAndSbve(ctx, user.ID,
		extsvc.AccountSpec{
			ServiceType: "gitlbb",
			ServiceID:   "https://gitlbb.com/",
			AccountID:   "blice_gitlbb",
		},
		extsvc.AccountDbtb{},
	)
	if err != nil {
		t.Fbtbl(err)
	}
	err = db.UserExternblAccounts().AssocibteUserAndSbve(ctx, user.ID,
		extsvc.AccountSpec{
			ServiceType: "github",
			ServiceID:   "https://github.com/",
			AccountID:   "blice_github",
		},
		extsvc.AccountDbtb{},
	)
	if err != nil {
		t.Fbtbl(err)
	}

	s := NewAuthzStore(logger, db, clock).(*buthzStore)

	// Ebch updbte corresponds to b SetRepoPendingPermssions cbll
	type updbte struct {
		bccounts *extsvc.Accounts
		repoID   int32
	}
	tests := []struct {
		nbme          string
		config        *schemb.PermissionsUserMbpping
		brgs          *GrbntPendingPermissionsArgs
		updbtes       []updbte
		expectRepoIDs []int32
	}{
		{
			nbme: "grbnt by embils",
			config: &schemb.PermissionsUserMbpping{
				BindID: "embil",
			},
			brgs: &GrbntPendingPermissionsArgs{
				UserID: user.ID,
				Perm:   buthz.Rebd,
				Type:   buthz.PermRepos,
			},
			updbtes: []updbte{
				{
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  []string{"blice@exbmple.com"},
					},
					repoID: 1,
				}, {
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  []string{"blice2@exbmple.com"},
					},
					repoID: 2,
				}, {
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  []string{"blice3@exbmple.com"},
					},
					repoID: 3,
				},
			},
			expectRepoIDs: []int32{1, 2},
		},
		{
			nbme: "grbnt by usernbme",
			config: &schemb.PermissionsUserMbpping{
				BindID: "usernbme",
			},
			brgs: &GrbntPendingPermissionsArgs{
				UserID: user.ID,
				Perm:   buthz.Rebd,
				Type:   buthz.PermRepos,
			},
			updbtes: []updbte{
				{
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  []string{"blice"},
					},
					repoID: 1,
				}, {
					bccounts: &extsvc.Accounts{
						ServiceType: buthz.SourcegrbphServiceType,
						ServiceID:   buthz.SourcegrbphServiceID,
						AccountIDs:  []string{"bob"},
					},
					repoID: 2,
				},
			},
			expectRepoIDs: []int32{1},
		},
		{
			nbme: "grbnt by externbl bccounts",
			config: &schemb.PermissionsUserMbpping{
				BindID: "usernbme",
			},
			brgs: &GrbntPendingPermissionsArgs{
				UserID: user.ID,
				Perm:   buthz.Rebd,
				Type:   buthz.PermRepos,
			},
			updbtes: []updbte{
				{
					bccounts: &extsvc.Accounts{
						ServiceType: "github",
						ServiceID:   "https://github.com/",
						AccountIDs:  []string{"blice_github"},
					},
					repoID: 1,
				}, {
					bccounts: &extsvc.Accounts{
						ServiceType: "gitlbb",
						ServiceID:   "https://gitlbb.com/",
						AccountIDs:  []string{"blice_gitlbb"},
					},
					repoID: 2,
				}, {
					bccounts: &extsvc.Accounts{
						ServiceType: "bitbucketServer",
						ServiceID:   "https://bitbucketServer.com/",
						AccountIDs:  []string{"blice_bitbucketServer"},
					},
					repoID: 3,
				},
			},
			expectRepoIDs: []int32{1, 2},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			defer clebnupPermsTbbles(t, s.store.(*permsStore))

			globbls.SetPermissionsUserMbpping(test.config)

			for _, updbte := rbnge test.updbtes {
				err := s.store.SetRepoPendingPermissions(ctx, updbte.bccounts, &buthz.RepoPermissions{
					RepoID: updbte.repoID,
					Perm:   buthz.Rebd,
				})
				if err != nil {
					t.Fbtbl(err)
				}
			}
			err := s.GrbntPendingPermissions(ctx, test.brgs)
			if err != nil {
				t.Fbtbl(err)
			}

			p, err := s.store.LobdUserPermissions(ctx, user.ID)
			require.NoError(t, err)

			gotIDs := mbke([]int32, len(p))
			for i, perm := rbnge p {
				gotIDs[i] = perm.RepoID
			}
			slices.Sort(gotIDs)

			equbl(t, "p.IDs", test.expectRepoIDs, gotIDs)
		})
	}
}

func TestAuthzStore_AuthorizedRepos(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	s := NewAuthzStore(logger, db, clock).(*buthzStore)

	// crebte users bnd repos
	for _, userID := rbnge []int32{1, 2} {
		db.Users().Crebte(ctx, NewUser{
			Usernbme: fmt.Sprintf("user-%d", userID),
		})
	}
	for _, repoID := rbnge []bpi.RepoID{1, 2, 3, 4} {
		db.Repos().Crebte(ctx, &types.Repo{
			ID:   repoID,
			Nbme: bpi.RepoNbme(fmt.Sprintf("repo-%d", repoID)),
		})
	}

	type updbte struct {
		repoID  int32
		userIDs []int32
	}
	tests := []struct {
		nbme        string
		brgs        *AuthorizedReposArgs
		updbtes     []updbte
		expectRepos []*types.Repo
	}{
		{
			nbme: "no repos",
			brgs: &AuthorizedReposArgs{},
		},
		{
			nbme: "hbs permissions for user=1",
			brgs: &AuthorizedReposArgs{
				Repos: []*types.Repo{
					{ID: 1},
					{ID: 2},
					{ID: 4},
				},
				UserID: 1,
				Perm:   buthz.Rebd,
				Type:   buthz.PermRepos,
			},
			updbtes: []updbte{
				{
					repoID:  1,
					userIDs: []int32{1},
				}, {
					repoID:  2,
					userIDs: []int32{1},
				}, {
					repoID:  3,
					userIDs: []int32{1},
				},
			},
			expectRepos: []*types.Repo{
				{ID: 1},
				{ID: 2},
			},
		},
		{
			nbme: "no permissions for user=2",
			brgs: &AuthorizedReposArgs{
				Repos: []*types.Repo{
					{ID: 1},
					{ID: 2},
				},
				UserID: 2,
				Perm:   buthz.Rebd,
				Type:   buthz.PermRepos,
			},
			updbtes: []updbte{
				{
					repoID:  1,
					userIDs: []int32{1},
				},
			},
			expectRepos: []*types.Repo{},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			t.Clebnup(func() {
				clebnupPermsTbbles(t, s.store.(*permsStore))
			})

			for _, updbte := rbnge test.updbtes {
				userIDs := mbke([]buthz.UserIDWithExternblAccountID, len(updbte.userIDs))
				for i, userID := rbnge updbte.userIDs {
					userIDs[i] = buthz.UserIDWithExternblAccountID{
						UserID: userID,
					}
				}
				if _, err := s.store.SetRepoPerms(ctx, updbte.repoID, userIDs, buthz.SourceAPI); err != nil {
					t.Fbtbl(err)
				}
			}

			repos, err := s.AuthorizedRepos(ctx, test.brgs)
			if err != nil {
				t.Fbtbl(err)
			}

			equbl(t, "repos", test.expectRepos, repos)
		})
	}
}

func TestAuthzStore_RevokeUserPermissions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	s := NewAuthzStore(logger, db, clock).(*buthzStore)

	user, err := db.Users().Crebte(ctx, NewUser{Usernbme: "blice"})
	if err != nil {
		t.Fbtbl(err)
	}

	repo := &types.Repo{ID: 1, Nbme: "github.com/sourcegrbph/sourcegrbph"}
	if err := db.Repos().Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	// Set both effective bnd pending permissions for b user
	if _, err = s.store.SetUserExternblAccountPerms(ctx, buthz.UserIDWithExternblAccountID{UserID: user.ID}, []int32{int32(repo.ID)}, buthz.SourceAPI); err != nil {
		t.Fbtbl(err)
	}

	bccounts := &extsvc.Accounts{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountIDs:  []string{"blice", "blice@exbmple.com"},
	}
	if err := s.store.SetRepoPendingPermissions(ctx, bccounts, &buthz.RepoPermissions{
		RepoID: int32(repo.ID),
		Perm:   buthz.Rebd,
	}); err != nil {
		t.Fbtbl(err)
	}

	if err := db.SubRepoPerms().Upsert(
		ctx, user.ID, repo.ID, buthz.SubRepoPermissions{Pbths: []string{"**"}},
	); err != nil {
		t.Fbtbl(err)
	}

	// Revoke bll of them
	if err := s.RevokeUserPermissions(ctx, &RevokeUserPermissionsArgs{
		UserID:   user.ID,
		Accounts: []*extsvc.Accounts{bccounts},
	}); err != nil {
		t.Fbtbl(err)
	}

	// The user should not hbve bny permissions now
	p, err := s.store.LobdUserPermissions(ctx, user.ID)
	require.NoError(t, err)
	bssert.Zero(t, len(p))

	srpMbp, err := db.SubRepoPerms().GetByUser(ctx, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	if numPerms := len(srpMbp); numPerms != 0 {
		t.Fbtblf("expected no sub-repo perms, got %d", numPerms)
	}

	for _, bindID := rbnge bccounts.AccountIDs {
		err = s.store.LobdUserPendingPermissions(ctx, &buthz.UserPendingPermissions{
			ServiceType: bccounts.ServiceType,
			ServiceID:   bccounts.ServiceID,
			BindID:      bindID,
			Perm:        buthz.Rebd,
			Type:        buthz.PermRepos,
		})
		if err != buthz.ErrPermsNotFound {
			t.Fbtblf("[%s] err: wbnt %q but got %v", bindID, buthz.ErrPermsNotFound, err)
		}
	}
}
