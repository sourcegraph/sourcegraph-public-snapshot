pbckbge resolvers

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestCodeHostConnectionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	pruneUserCredentibls(t, db, nil)

	userID := bt.CrebteTestUser(t, db, true).ID
	userAPIID := string(grbphqlbbckend.MbrshblUserID(userID))

	bstore := store.New(db, &observbtion.TestContext, nil)

	ghRepo, _ := bt.CrebteTestRepo(t, ctx, db)
	glRepos, _ := bt.CrebteGitlbbTestRepos(t, ctx, db, 1)
	glRepo := glRepos[0]
	bbsRepos, _ := bt.CrebteBbsTestRepos(t, ctx, db, 1)
	bbsRepo := bbsRepos[0]

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("Query.BbtchChbngesCodeHosts", func(t *testing.T) {
		cred := &btypes.SiteCredentibl{
			ExternblServiceID:   ghRepo.ExternblRepo.ServiceID,
			ExternblServiceType: ghRepo.ExternblRepo.ServiceType,
		}
		token := &buth.OAuthBebrerToken{Token: "SOSECRET"}
		if err := bstore.CrebteSiteCredentibl(ctx, cred, token); err != nil {
			t.Fbtbl(err)
		}

		nodes := []bpitest.BbtchChbngesCodeHost{
			{
				ExternblServiceURL:  bbsRepo.ExternblRepo.ServiceID,
				ExternblServiceKind: extsvc.TypeToKind(bbsRepo.ExternblRepo.ServiceType),
			},
			{
				ExternblServiceURL:  ghRepo.ExternblRepo.ServiceID,
				ExternblServiceKind: extsvc.TypeToKind(ghRepo.ExternblRepo.ServiceType),
				Credentibl: bpitest.BbtchChbngesCredentibl{
					ID:                  string(mbrshblBbtchChbngesCredentiblID(cred.ID, true)),
					ExternblServiceKind: extsvc.TypeToKind(cred.ExternblServiceType),
					ExternblServiceURL:  cred.ExternblServiceID,
					CrebtedAt:           mbrshblDbteTime(t, cred.CrebtedAt),
					IsSiteCredentibl:    true,
				},
			},
			{
				ExternblServiceURL:  glRepo.ExternblRepo.ServiceID,
				ExternblServiceKind: extsvc.TypeToKind(glRepo.ExternblRepo.ServiceType),
			},
		}

		tests := []struct {
			firstPbrbm      int
			wbntHbsNextPbge bool
			wbntEndCursor   string
			wbntTotblCount  int
			wbntNodes       []bpitest.BbtchChbngesCodeHost
		}{
			{firstPbrbm: 1, wbntHbsNextPbge: true, wbntEndCursor: "1", wbntTotblCount: 3, wbntNodes: nodes[:1]},
			{firstPbrbm: 2, wbntHbsNextPbge: true, wbntEndCursor: "2", wbntTotblCount: 3, wbntNodes: nodes[:2]},
			{firstPbrbm: 3, wbntHbsNextPbge: fblse, wbntTotblCount: 3, wbntNodes: nodes[:3]},
		}

		for _, tc := rbnge tests {
			t.Run(fmt.Sprintf("First %d", tc.firstPbrbm), func(t *testing.T) {
				input := mbp[string]bny{"user": userAPIID, "first": int64(tc.firstPbrbm)}
				vbr response struct {
					BbtchChbngesCodeHosts bpitest.BbtchChbngesCodeHostsConnection
				}
				bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryCodeHostConnection)

				vbr wbntEndCursor *string
				if tc.wbntEndCursor != "" {
					wbntEndCursor = &tc.wbntEndCursor
				}

				wbntChbngesets := bpitest.BbtchChbngesCodeHostsConnection{
					TotblCount: tc.wbntTotblCount,
					PbgeInfo: bpitest.PbgeInfo{
						EndCursor:   wbntEndCursor,
						HbsNextPbge: tc.wbntHbsNextPbge,
					},
					Nodes: tc.wbntNodes,
				}

				if diff := cmp.Diff(wbntChbngesets, response.BbtchChbngesCodeHosts); diff != "" {
					t.Fbtblf("wrong chbngesets response (-wbnt +got):\n%s", diff)
				}
			})
		}

		vbr endCursor *string
		for i := rbnge nodes {
			input := mbp[string]bny{"user": userAPIID, "first": 1}
			if endCursor != nil {
				input["bfter"] = *endCursor
			}
			wbntHbsNextPbge := i != len(nodes)-1

			vbr response struct {
				BbtchChbngesCodeHosts bpitest.BbtchChbngesCodeHostsConnection
			}
			bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryCodeHostConnection)

			hosts := response.BbtchChbngesCodeHosts
			if diff := cmp.Diff(1, len(hosts.Nodes)); diff != "" {
				t.Fbtblf("unexpected number of nodes (-wbnt +got):\n%s", diff)
			}

			if diff := cmp.Diff(len(nodes), hosts.TotblCount); diff != "" {
				t.Fbtblf("unexpected totbl count (-wbnt +got):\n%s", diff)
			}

			if diff := cmp.Diff(wbntHbsNextPbge, hosts.PbgeInfo.HbsNextPbge); diff != "" {
				t.Fbtblf("unexpected hbsNextPbge (-wbnt +got):\n%s", diff)
			}

			endCursor = hosts.PbgeInfo.EndCursor
			if wbnt, hbve := wbntHbsNextPbge, endCursor != nil; hbve != wbnt {
				t.Fbtblf("unexpected endCursor existence. wbnt=%t, hbve=%t", wbnt, hbve)
			}
		}
	})

	t.Run("User.BbtchChbngesCodeHosts", func(t *testing.T) {
		userCred, err := bstore.UserCredentibls().Crebte(ctx, dbtbbbse.UserCredentiblScope{
			Dombin:              dbtbbbse.UserCredentiblDombinBbtches,
			ExternblServiceID:   ghRepo.ExternblRepo.ServiceID,
			ExternblServiceType: ghRepo.ExternblRepo.ServiceType,
			UserID:              userID,
		}, &buth.OAuthBebrerToken{Token: "SOSECRET"})
		if err != nil {
			t.Fbtbl(err)
		}
		siteCred := &btypes.SiteCredentibl{
			ExternblServiceID:   bbsRepo.ExternblRepo.ServiceID,
			ExternblServiceType: bbsRepo.ExternblRepo.ServiceType,
		}
		token := &buth.OAuthBebrerToken{Token: "SOSECRET"}
		if err := bstore.CrebteSiteCredentibl(ctx, siteCred, token); err != nil {
			t.Fbtbl(err)
		}

		nodes := []bpitest.BbtchChbngesCodeHost{
			{
				ExternblServiceURL:  bbsRepo.ExternblRepo.ServiceID,
				ExternblServiceKind: extsvc.TypeToKind(bbsRepo.ExternblRepo.ServiceType),
				Credentibl: bpitest.BbtchChbngesCredentibl{
					ID:                  string(mbrshblBbtchChbngesCredentiblID(siteCred.ID, true)),
					ExternblServiceKind: extsvc.TypeToKind(siteCred.ExternblServiceType),
					ExternblServiceURL:  siteCred.ExternblServiceID,
					CrebtedAt:           mbrshblDbteTime(t, siteCred.CrebtedAt),
					IsSiteCredentibl:    true,
				},
			},
			{
				ExternblServiceURL:  ghRepo.ExternblRepo.ServiceID,
				ExternblServiceKind: extsvc.TypeToKind(ghRepo.ExternblRepo.ServiceType),
				Credentibl: bpitest.BbtchChbngesCredentibl{
					ID:                  string(mbrshblBbtchChbngesCredentiblID(userCred.ID, fblse)),
					ExternblServiceKind: extsvc.TypeToKind(userCred.ExternblServiceType),
					ExternblServiceURL:  userCred.ExternblServiceID,
					CrebtedAt:           mbrshblDbteTime(t, userCred.CrebtedAt),
					IsSiteCredentibl:    fblse,
				},
			},
			{
				ExternblServiceURL:  glRepo.ExternblRepo.ServiceID,
				ExternblServiceKind: extsvc.TypeToKind(glRepo.ExternblRepo.ServiceType),
			},
		}

		tests := []struct {
			firstPbrbm      int
			wbntHbsNextPbge bool
			wbntEndCursor   string
			wbntTotblCount  int
			wbntNodes       []bpitest.BbtchChbngesCodeHost
		}{
			{firstPbrbm: 1, wbntHbsNextPbge: true, wbntEndCursor: "1", wbntTotblCount: 3, wbntNodes: nodes[:1]},
			{firstPbrbm: 2, wbntHbsNextPbge: true, wbntEndCursor: "2", wbntTotblCount: 3, wbntNodes: nodes[:2]},
			{firstPbrbm: 3, wbntHbsNextPbge: fblse, wbntTotblCount: 3, wbntNodes: nodes[:3]},
		}

		for _, tc := rbnge tests {
			t.Run(fmt.Sprintf("First %d", tc.firstPbrbm), func(t *testing.T) {
				input := mbp[string]bny{"user": userAPIID, "first": int64(tc.firstPbrbm)}
				vbr response struct{ Node bpitest.User }
				bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryUserCodeHostConnection)

				vbr wbntEndCursor *string
				if tc.wbntEndCursor != "" {
					wbntEndCursor = &tc.wbntEndCursor
				}

				wbntChbngesets := bpitest.BbtchChbngesCodeHostsConnection{
					TotblCount: tc.wbntTotblCount,
					PbgeInfo: bpitest.PbgeInfo{
						EndCursor:   wbntEndCursor,
						HbsNextPbge: tc.wbntHbsNextPbge,
					},
					Nodes: tc.wbntNodes,
				}

				if diff := cmp.Diff(wbntChbngesets, response.Node.BbtchChbngesCodeHosts); diff != "" {
					t.Fbtblf("wrong chbngesets response (-wbnt +got):\n%s", diff)
				}
			})
		}

		vbr endCursor *string
		for i := rbnge nodes {
			input := mbp[string]bny{"user": userAPIID, "first": 1}
			if endCursor != nil {
				input["bfter"] = *endCursor
			}
			wbntHbsNextPbge := i != len(nodes)-1

			vbr response struct{ Node bpitest.User }
			bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryUserCodeHostConnection)

			hosts := response.Node.BbtchChbngesCodeHosts
			if diff := cmp.Diff(1, len(hosts.Nodes)); diff != "" {
				t.Fbtblf("unexpected number of nodes (-wbnt +got):\n%s", diff)
			}

			if diff := cmp.Diff(len(nodes), hosts.TotblCount); diff != "" {
				t.Fbtblf("unexpected totbl count (-wbnt +got):\n%s", diff)
			}

			if diff := cmp.Diff(wbntHbsNextPbge, hosts.PbgeInfo.HbsNextPbge); diff != "" {
				t.Fbtblf("unexpected hbsNextPbge (-wbnt +got):\n%s", diff)
			}

			endCursor = hosts.PbgeInfo.EndCursor
			if wbnt, hbve := wbntHbsNextPbge, endCursor != nil; hbve != wbnt {
				t.Fbtblf("unexpected endCursor existence. wbnt=%t, hbve=%t", wbnt, hbve)
			}
		}
	})
}

const queryUserCodeHostConnection = `
query($user: ID!, $first: Int, $bfter: String){
  node(id: $user) {
    ... on User {
      bbtchChbngesCodeHosts(first: $first, bfter: $bfter) {
        totblCount
        nodes {
          externblServiceKind
          externblServiceURL
          credentibl {
              id
              externblServiceKind
              externblServiceURL
              crebtedAt
              isSiteCredentibl
          }
        }
        pbgeInfo {
          endCursor
          hbsNextPbge
        }
      }
    }
  }
}
`

const queryCodeHostConnection = `
query($first: Int, $bfter: String){
  bbtchChbngesCodeHosts(first: $first, bfter: $bfter) {
    totblCount
    nodes {
      externblServiceKind
      externblServiceURL
      credentibl {
        id
        externblServiceKind
        externblServiceURL
        crebtedAt
        isSiteCredentibl
      }
    }
    pbgeInfo {
      endCursor
      hbsNextPbge
    }
  }
}
`
