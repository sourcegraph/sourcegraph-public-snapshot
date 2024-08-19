package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func TestGitserverResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))

	user := createTestUser(t, db, false)
	admin := createTestUser(t, db, true)

	userCtx := actor.WithActor(ctx, actor.FromUser(user.ID))
	adminCtx := actor.WithActor(ctx, actor.FromUser(admin.ID))

	gitserverInstances := []protocol.SystemInfo{
		{
			Address:    "127.0.0.1:3501",
			FreeSpace:  10240,
			TotalSpace: 4048000,
		},
		{
			Address:    "127.0.0.1:3502",
			FreeSpace:  1024000,
			TotalSpace: 2048000,
		},
	}

	t.Run("query", func(t *testing.T) {
		mockGitserverClient := gitserver.NewMockClient()
		mockGitserverClient.SystemsInfoFunc.SetDefaultReturn(gitserverInstances, nil)

		s, err := NewSchemaWithGitserverClient(db, nil, mockGitserverClient)
		require.NoError(t, err)

		testCases := []struct {
			name                 string
			ctx                  context.Context
			err                  error
			noOfSystemsInfoCalls int
		}{
			{
				name: "as regular user",
				ctx:  userCtx,
				err:  auth.ErrMustBeSiteAdmin,
			},
			{
				name:                 "as site-admin",
				ctx:                  adminCtx,
				noOfSystemsInfoCalls: 1,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var response struct {
					Gitservers apitest.GitserverInstanceConnection
				}
				errs := apitest.Exec(tc.ctx, t, s, nil, &response, queryGitservers)

				calls := mockGitserverClient.SystemsInfoFunc.History()
				require.Len(t, calls, tc.noOfSystemsInfoCalls)

				if tc.err != nil {
					require.Len(t, errs, 1)
					require.ErrorIs(t, errs[0], tc.err)
				} else {
					require.Len(t, errs, 0)
					require.Equal(t, response.Gitservers.TotalCount, len(gitserverInstances))
					require.False(t, response.Gitservers.PageInfo.HasNextPage)
					require.Nil(t, response.Gitservers.PageInfo.EndCursor)

					for idx, gs := range response.Gitservers.Nodes {
						require.Equal(t, gs.Address, gitserverInstances[idx].Address)
						require.Equal(t, gs.FreeDiskSpaceBytes, fmt.Sprint(gitserverInstances[idx].FreeSpace))
						require.Equal(t, gs.TotalDiskSpaceBytes, fmt.Sprint(gitserverInstances[idx].TotalSpace))
					}
				}
			})
		}
	})

	t.Run("node", func(t *testing.T) {
		mockGitserverClient := gitserver.NewMockClient()
		mockGitserverClient.SystemInfoFunc.SetDefaultReturn(gitserverInstances[0], nil)

		s, err := NewSchemaWithGitserverClient(db, nil, mockGitserverClient)
		require.NoError(t, err)

		id := marshalGitserverID(gitserverInstances[0].Address)

		testCases := []struct {
			name                string
			ctx                 context.Context
			err                 error
			noOfSystemInfoCalls int
		}{
			{
				name: "as regular user",
				ctx:  userCtx,
				err:  auth.ErrMustBeSiteAdmin,
			},
			{
				name:                "as site-admin",
				ctx:                 adminCtx,
				noOfSystemInfoCalls: 1,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				input := map[string]any{"node": string(id)}
				var response struct{ Node apitest.GitserverInstance }
				errs := apitest.Exec(tc.ctx, t, s, input, &response, queryGitserverNode)

				calls := mockGitserverClient.SystemInfoFunc.History()
				require.Len(t, calls, tc.noOfSystemInfoCalls)

				if tc.err != nil {
					require.Len(t, errs, 1)
					require.ErrorIs(t, errs[0], tc.err)
				} else {
					require.Len(t, errs, 0)
					require.Equal(t, response.Node.Address, gitserverInstances[0].Address)
					require.Equal(t, response.Node.FreeDiskSpaceBytes, fmt.Sprint(gitserverInstances[0].FreeSpace))
					require.Equal(t, response.Node.TotalDiskSpaceBytes, fmt.Sprint(gitserverInstances[0].TotalSpace))
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
			address
			freeDiskSpaceBytes
			totalDiskSpaceBytes
		}
		totalCount
		pageInfo {
			hasNextPage
			endCursor
		}
	}
}
`

const queryGitserverNode = `
query GitserverNode ($node: ID!) {
	node(id: $node) {
		id

		... on GitserverInstance {
			address
			freeDiskSpaceBytes
			totalDiskSpaceBytes
		}
	}
}
`
