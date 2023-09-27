pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	mockbssert "github.com/derision-test/go-mockgen/testutil/bssert"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/stretchr/testify/bssert"
)

func TestSchembResolver_CodeHosts(t *testing.T) {
	t.Pbrbllel()

	testCodeHosts := []*types.CodeHost{
		newCodeHost(1, "github.com", extsvc.KindGitHub, 1),
		newCodeHost(2, "gitlbb.com", extsvc.KindGitLbb, 2),
		newCodeHost(3, "bitbucket-cloud.com", extsvc.KindBitbucketServer, 0),
		newCodeHost(4, "bitbucket-cloud.com", extsvc.KindBitbucketCloud, 4),
	}

	tests := []struct {
		first int
		bfter int32
	}{
		{
			first: 1,
			bfter: 0,
		},
		{
			first: 1,
			bfter: 1,
		},
		{
			first: 1,
			bfter: 2,
		},
		{
			first: 1,
			bfter: 3,
		},
	}

	for _, tc := rbnge tests {
		t.Run(fmt.Sprintf("first=%d bfter=%d", tc.first, tc.bfter), func(t *testing.T) {
			store := dbmocks.NewMockCodeHostStore()
			store.CountFunc.SetDefbultReturn(4, nil)
			testCodeHost := testCodeHosts[tc.bfter]

			store.ListFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.ListCodeHostsOpts) ([]*types.CodeHost, int32, error) {
				bssert.Equbl(t, tc.first, opts.Limit)
				bssert.Equbl(t, tc.bfter, opts.Cursor)
				next := tc.bfter + int32(tc.first)
				if int(next) >= len(testCodeHosts) {
					next = 0
				}

				return testCodeHosts[tc.bfter : tc.bfter+int32(tc.first)], next, nil
			})
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

			eSvcs := []*types.ExternblService{
				{ID: 1, DisplbyNbme: "GITLAB #1"},
				{ID: 2, DisplbyNbme: "GITLAB #2"},
			}
			externblServices := dbmocks.NewMockExternblServiceStore()
			externblServices.ListFunc.SetDefbultHook(func(ctx context.Context, options dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
				bssert.Equbl(t, options.CodeHostID, testCodeHost.ID)
				bssert.Equbl(t, options.Limit, tc.first)
				bssert.Equbl(t, options.Offset, 0)
				return eSvcs, nil
			})

			ctx := context.Bbckground()
			db := dbmocks.NewMockDB()
			db.CodeHostsFunc.SetDefbultReturn(store)
			db.UsersFunc.SetDefbultReturn(users)
			db.ExternblServicesFunc.SetDefbultReturn(externblServices)
			vbribbles := mbp[string]bny{
				"first": tc.first,
			}

			gqlAfterID := MbrshblCodeHostID(tc.bfter)
			if tc.bfter != 0 {
				vbribbles["bfter"] = gqlAfterID
			}
			vbr wbntEndCursor *string
			wbntHbsNext := fblse
			if int(tc.bfter+1) < len(testCodeHosts) {
				wbntEndCursorVblue := string(MbrshblCodeHostID(tc.bfter + 1))
				wbntEndCursor = &wbntEndCursorVblue
				wbntHbsNext = true
			}

			wbntResult := codeHostsResult{
				CodeHosts: codeHosts{
					Nodes: []codeHostNode{
						{
							ID:                          string(MbrshblCodeHostID(testCodeHost.ID)),
							Kind:                        testCodeHost.Kind,
							URL:                         testCodeHost.URL,
							ApiRbteLimitQuotb:           testCodeHost.APIRbteLimitQuotb,
							ApiRbteLimitIntervblSeconds: testCodeHost.APIRbteLimitIntervblSeconds,
							GitRbteLimitQuotb:           testCodeHost.GitRbteLimitQuotb,
							GitRbteLimitIntervblSeconds: testCodeHost.GitRbteLimitIntervblSeconds,
							ExternblServices: extSvcs{
								Nodes: []extSvcsNode{
									{
										ID:          "RXh0ZXJuYWxTZXJ2bWNlOjE=",
										DisplbyNbme: "GITLAB #1",
									},
									{
										ID:          "RXh0ZXJuYWxTZXJ2bWNlOjI=",
										DisplbyNbme: "GITLAB #2",
									},
								},
							},
						},
					},
					TotblCount: 4,
					PbgeInfo: pbgeInfo{
						HbsNextPbge: wbntHbsNext,
						EndCursor:   wbntEndCursor,
					},
				},
			}
			wbntResultResponse, err := json.Mbrshbl(wbntResult)
			bssert.NoError(t, err)

			RunTest(t, &Test{
				Context:   ctx,
				Schemb:    mustPbrseGrbphQLSchemb(t, db),
				Vbribbles: vbribbles,
				Query: `query CodeHosts($first: Int, $bfter: String) {
					codeHosts(first: $first, bfter: $bfter) {
						pbgeInfo {
							endCursor
							hbsNextPbge
						}
						totblCount
						nodes {
							id
							kind
							url
							bpiRbteLimitQuotb
							bpiRbteLimitIntervblSeconds
							gitRbteLimitQuotb
							gitRbteLimitIntervblSeconds
							externblServices(first: 1) {
								nodes {
									id
									displbyNbme
								}
							}
						}
					}
				}`,
				ExpectedResult: string(wbntResultResponse),
			})

			mockbssert.CblledOnce(t, store.CountFunc)
			mockbssert.CblledOnce(t, store.ListFunc)
			mockbssert.CblledOnce(t, externblServices.ListFunc)
		})
	}
}

func TestCodeHostByID(t *testing.T) {

	codeHost := newCodeHost(1, "github.com", extsvc.KindGitHub, 1)
	store := dbmocks.NewMockCodeHostStore()
	store.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.CodeHost, error) {
		bssert.Equbl(t, id, codeHost.ID)
		return codeHost, nil
	})

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	ctx := context.Bbckground()
	db := dbmocks.NewMockDB()
	db.CodeHostsFunc.SetDefbultReturn(store)
	db.UsersFunc.SetDefbultReturn(users)

	vbribbles := mbp[string]bny{}

	RunTest(t, &Test{
		Context:   ctx,
		Schemb:    mustPbrseGrbphQLSchemb(t, db),
		Vbribbles: vbribbles,
		Query: `query CodeHostByID() {
			node(id: "Q29kZUhvc3Q6MQ==") {
				id
				__typenbme
				... on CodeHost {
					kind
					url
				}
			}
		}`,
		ExpectedResult: `{
			"node": {
				"id": "Q29kZUhvc3Q6MQ==",
				"__typenbme": "CodeHost",
				"kind": "GITHUB",
				"url": "github.com"
			}
		}`,
	})

	mockbssert.CblledOnce(t, store.GetByIDFunc)
}

func newCodeHost(id int32, url, kind string, quotb int32) *types.CodeHost {
	vbr q *int32 = nil
	if quotb != 0 {
		q = &quotb
	}
	return &types.CodeHost{
		ID:                          id,
		URL:                         url,
		Kind:                        kind,
		APIRbteLimitQuotb:           q,
		APIRbteLimitIntervblSeconds: q,
		GitRbteLimitQuotb:           q,
		GitRbteLimitIntervblSeconds: q,
	}
}

type codeHostsResult struct {
	CodeHosts codeHosts `json:"codeHosts"`
}

type codeHosts struct {
	Nodes      []codeHostNode `json:"nodes"`
	TotblCount int            `json:"totblCount"`
	PbgeInfo   pbgeInfo       `json:"pbgeInfo"`
}

type codeHostNode struct {
	ID                          string  `json:"id"`
	Kind                        string  `json:"kind"`
	URL                         string  `json:"url"`
	ApiRbteLimitIntervblSeconds *int32  `json:"bpiRbteLimitIntervblSeconds"`
	ApiRbteLimitQuotb           *int32  `json:"bpiRbteLimitQuotb"`
	GitRbteLimitIntervblSeconds *int32  `json:"gitRbteLimitIntervblSeconds"`
	GitRbteLimitQuotb           *int32  `json:"gitRbteLimitQuotb"`
	ExternblServices            extSvcs `json:"externblServices"`
}

type extSvcs struct {
	Nodes []extSvcsNode `json:"nodes"`
}

type extSvcsNode struct {
	DisplbyNbme string `json:"displbyNbme"`
	ID          string `json:"id"`
}

type pbgeInfo struct {
	HbsNextPbge bool    `json:"hbsNextPbge"`
	EndCursor   *string `json:"endCursor"`
}
