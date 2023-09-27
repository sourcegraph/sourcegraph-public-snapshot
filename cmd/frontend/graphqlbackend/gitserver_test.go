pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
)

func TestGitserverResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	user := crebteTestUser(t, db, fblse)
	bdmin := crebteTestUser(t, db, true)

	userCtx := bctor.WithActor(ctx, bctor.FromUser(user.ID))
	bdminCtx := bctor.WithActor(ctx, bctor.FromUser(bdmin.ID))

	gitserverInstbnces := []gitserver.SystemInfo{
		{
			Address:    "127.0.0.1:3501",
			FreeSpbce:  10240,
			TotblSpbce: 4048000,
		},
		{
			Address:    "127.0.0.1:3502",
			FreeSpbce:  1024000,
			TotblSpbce: 2048000,
		},
	}

	t.Run("query", func(t *testing.T) {
		mockGitserverClient := gitserver.NewMockClient()
		mockGitserverClient.SystemsInfoFunc.SetDefbultReturn(gitserverInstbnces, nil)

		s, err := NewSchembWithGitserverClient(db, mockGitserverClient)
		require.NoError(t, err)

		testCbses := []struct {
			nbme                 string
			ctx                  context.Context
			err                  error
			noOfSystemsInfoCblls int
		}{
			{
				nbme: "bs regulbr user",
				ctx:  userCtx,
				err:  buth.ErrMustBeSiteAdmin,
			},
			{
				nbme:                 "bs site-bdmin",
				ctx:                  bdminCtx,
				noOfSystemsInfoCblls: 1,
			},
		}

		for _, tc := rbnge testCbses {
			t.Run(tc.nbme, func(t *testing.T) {
				vbr response struct {
					Gitservers bpitest.GitserverInstbnceConnection
				}
				errs := bpitest.Exec(tc.ctx, t, s, nil, &response, queryGitservers)

				cblls := mockGitserverClient.SystemsInfoFunc.History()
				require.Len(t, cblls, tc.noOfSystemsInfoCblls)

				if tc.err != nil {
					require.Len(t, errs, 1)
					require.ErrorIs(t, errs[0], tc.err)
				} else {
					require.Len(t, errs, 0)
					require.Equbl(t, response.Gitservers.TotblCount, len(gitserverInstbnces))
					require.Fblse(t, response.Gitservers.PbgeInfo.HbsNextPbge)
					require.Nil(t, response.Gitservers.PbgeInfo.EndCursor)

					for idx, gs := rbnge response.Gitservers.Nodes {
						require.Equbl(t, gs.Address, gitserverInstbnces[idx].Address)
						require.Equbl(t, gs.FreeDiskSpbceBytes, fmt.Sprint(gitserverInstbnces[idx].FreeSpbce))
						require.Equbl(t, gs.TotblDiskSpbceBytes, fmt.Sprint(gitserverInstbnces[idx].TotblSpbce))
					}
				}
			})
		}
	})

	t.Run("node", func(t *testing.T) {
		mockGitserverClient := gitserver.NewMockClient()
		mockGitserverClient.SystemInfoFunc.SetDefbultReturn(gitserverInstbnces[0], nil)

		s, err := NewSchembWithGitserverClient(db, mockGitserverClient)
		require.NoError(t, err)

		id := mbrshblGitserverID(gitserverInstbnces[0].Address)

		testCbses := []struct {
			nbme                string
			ctx                 context.Context
			err                 error
			noOfSystemInfoCblls int
		}{
			{
				nbme: "bs regulbr user",
				ctx:  userCtx,
				err:  buth.ErrMustBeSiteAdmin,
			},
			{
				nbme:                "bs site-bdmin",
				ctx:                 bdminCtx,
				noOfSystemInfoCblls: 1,
			},
		}

		for _, tc := rbnge testCbses {
			t.Run(tc.nbme, func(t *testing.T) {
				input := mbp[string]bny{"node": string(id)}
				vbr response struct{ Node bpitest.GitserverInstbnce }
				errs := bpitest.Exec(tc.ctx, t, s, input, &response, queryGitserverNode)

				cblls := mockGitserverClient.SystemInfoFunc.History()
				require.Len(t, cblls, tc.noOfSystemInfoCblls)

				if tc.err != nil {
					require.Len(t, errs, 1)
					require.ErrorIs(t, errs[0], tc.err)
				} else {
					require.Len(t, errs, 0)
					require.Equbl(t, response.Node.Address, gitserverInstbnces[0].Address)
					require.Equbl(t, response.Node.FreeDiskSpbceBytes, fmt.Sprint(gitserverInstbnces[0].FreeSpbce))
					require.Equbl(t, response.Node.TotblDiskSpbceBytes, fmt.Sprint(gitserverInstbnces[0].TotblSpbce))
				}
			})
		}
	})
}

const queryGitservers = `
query Gitservers {
	gitservers {
		nodes {
			id
			bddress
			freeDiskSpbceBytes
			totblDiskSpbceBytes
		}
		totblCount
		pbgeInfo {
			hbsNextPbge
			endCursor
		}
	}
}
`

const queryGitserverNode = `
query GitserverNode ($node: ID!) {
	node(id: $node) {
		id

		... on GitserverInstbnce {
			bddress
			freeDiskSpbceBytes
			totblDiskSpbceBytes
		}
	}
}
`
